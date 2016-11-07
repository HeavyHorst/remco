/*
 * This file is part of remco.
 * Based on code from confd.
 * https://github.com/kelseyhightower/confd/blob/30663b9822fe8e800d1f2ea78447fba0ebce8f6c/resource/template/resource.go
 * Users who have contributed to this file
 * © 2013 Kelsey Hightower
 * © 2014 Armon Dadgar
 * © 2014 Ernesto Jimenez
 * © 2014 Nathan Fritz
 * © 2014 John Engelman
 * © 2014 Joanna Solmon
 * © 2014 Chris Armstrong
 * © 2014 Chris McNabb
 * © 2015 Phil Kates
 * © 2015 Matthew Fisher
 *
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/memkv"
	"github.com/HeavyHorst/pongo2"
	berr "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template/fileutil"
	"github.com/Sirupsen/logrus"
)

// A BackendConfig - Every backend implements this interface. If Connect is called a new connection to the underlaying kv-store will be established.
// Connect should also set the name and the StoreClient of the Backend. The other values of Backend will be loaded from the configuration file.
type BackendConfig interface {
	Connect() (Backend, error)
}

// Backend is the representation of a template backend like etcd or consul
type Backend struct {
	easyKV.ReadWatcher
	Name     string
	Onetime  bool
	Watch    bool
	Prefix   string
	Interval int
	Keys     []string
	store    *memkv.Store
}

type Backends []Backend

// Close is calling the close method on all backends
func (b Backends) Close() {
	for _, v := range b {
		log.WithFields(logrus.Fields{
			"backend": v.Name,
		}).Debug("Closing client connection")
		v.ReadWatcher.Close()
	}
}

// Resource is the representation of a parsed template resource.
type Resource struct {
	backends []Backend
	funcMap  map[string]interface{}
	store    memkv.Store
	sources  []*ProcessConfig
}

// ErrEmptySrc is returned if an emty src template is passed to NewResource
var ErrEmptySrc = errors.New("empty src template")

// NewResource creates a Resource.
func NewResource(backends Backends, sources []*ProcessConfig) (*Resource, error) {
	if len(backends) == 0 {
		return nil, errors.New("A valid StoreClient is required.")
	}

	for _, v := range sources {
		if v.Src == "" {
			return nil, ErrEmptySrc
		}
	}

	tr := &Resource{
		backends: backends,
		store:    memkv.New(),
		funcMap:  newFuncMap(),
		sources:  sources,
	}

	// initialize the inidividual backend memkv Stores
	for i := range tr.backends {
		store := memkv.New()
		tr.backends[i].store = &store

		if tr.backends[i].Interval <= 0 && tr.backends[i].Onetime == false && tr.backends[i].Watch == false {
			log.Warning("Interval needs to be > 0: setting interval to 60")
			tr.backends[i].Interval = 60
		}
	}

	addFuncs(tr.funcMap, tr.store.FuncMap)

	return tr, nil
}

// setVars sets the Vars for a template resource.
func (t *Resource) setVars(storeClient Backend) error {
	var err error

	log.WithFields(logrus.Fields{
		"backend":    storeClient.Name,
		"key_prefix": storeClient.Prefix,
	}).Debug("Retrieving keys")

	result, err := storeClient.GetValues(appendPrefix(storeClient.Prefix, storeClient.Keys))
	if err != nil {
		return err
	}

	storeClient.store.Purge()

	for key, value := range result {
		storeClient.store.Set(path.Join("/", strings.TrimPrefix(key, storeClient.Prefix)), value)
	}

	//merge all stores
	t.store.Purge()
	for _, v := range t.backends {
		for _, kv := range v.store.GetAllKVs() {
			if t.store.Exists(kv.Key) {
				log.Warning("Key collision - " + kv.Key)
			}
			t.store.Set(kv.Key, kv.Value)
		}
	}

	return nil
}

// createStageFile stages the src configuration file by processing the src
// template and setting the desired owner, group, and mode. It also sets the
// StageFile for the template resource.
// It returns an error if any.
func (t *Resource) createStageFileAndSync() error {
	for _, s := range t.sources {
		if !fileutil.IsFileExist(s.Src) {
			return errors.New("Missing template: " + s.Src)
		}

		log.WithFields(logrus.Fields{
			"template": s.Src,
		}).Debug("Compiling source template")

		set := pongo2.NewSet("local", &pongo2.LocalFilesystemLoader{})
		tmpl, err := set.FromFile(s.Src)
		if err != nil {
			return fmt.Errorf("Unable to process template %s, %s", s.Src, err)
		}

		// create TempFile in Dest directory to avoid cross-filesystem issues
		temp, err := ioutil.TempFile(filepath.Dir(s.Dst), "."+filepath.Base(s.Dst))
		if err != nil {
			return err
		}

		if err = tmpl.ExecuteWriter(t.funcMap, temp); err != nil {
			temp.Close()
			os.Remove(temp.Name())
			return err
		}

		temp.Close()

		fileMode, err := s.getFileMode()
		if err != nil {
			return err
		}
		// Set the owner, group, and mode on the stage file now to make it easier to
		// compare against the destination configuration file later.
		os.Chmod(temp.Name(), fileMode)
		os.Chown(temp.Name(), s.UID, s.GID)
		s.stageFile = temp

		if err = t.sync(s); err != nil {
			return err
		}
	}
	return nil
}

// sync compares the staged and dest config files and attempts to sync them
// if they differ. sync will run a config check command if set before
// overwriting the target config file. Finally, sync will run a reload command
// if set to have the application or service pick up the changes.
// It returns an error if any.
func (t *Resource) sync(s *ProcessConfig) error {
	staged := s.stageFile.Name()
	defer os.Remove(staged)

	log.WithFields(logrus.Fields{
		"staged": path.Base(staged),
		"dest":   s.Dst,
	}).Debug("Comparing staged and dest config files")

	ok, err := fileutil.SameFile(staged, s.Dst)
	if err != nil {
		log.Error(err.Error())
	}

	if !ok {
		log.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Info("Target config out of sync")

		if err := s.check(staged); err != nil {
			return errors.New("Config check failed: " + err.Error())
		}

		log.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Debug("Overwriting target config")

		fileMode, err := s.getFileMode()
		if err != nil {
			return err
		}
		if err := fileutil.ReplaceFile(staged, s.Dst, fileMode); err != nil {
			return err
		}

		// make sure owner and group match the temp file, in case the file was created with WriteFile
		os.Chown(s.Dst, s.UID, s.GID)

		if err := s.reload(); err != nil {
			return err
		}

		log.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Info("Target config has been updated")

	} else {
		log.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Debug("Target config in sync")

	}
	return nil
}

// Process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *Resource) process(storeClients []Backend) error {
	for _, storeClient := range storeClients {
		if err := t.setVars(storeClient); err != nil {
			return berr.BackendError{
				Message: err.Error(),
				Backend: storeClient.Name,
			}
		}
	}
	if err := t.createStageFileAndSync(); err != nil {
		return err
	}
	return nil
}

func (s Backend) watch(ctx context.Context, processChan chan Backend) {
	if s.Onetime {
		return
	}

	var lastIndex uint64
	keysPrefix := appendPrefix(s.Prefix, s.Keys)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			index, err := s.WatchPrefix(s.Prefix, ctx, easyKV.WithKeys(keysPrefix), easyKV.WithWaitIndex(lastIndex))
			if err != nil {
				if err != easyKV.ErrWatchCanceled {
					log.WithFields(logrus.Fields{
						"backend": s.Name,
					}).Error(err)
					time.Sleep(2 * time.Second)
				}
				continue
			}
			processChan <- s
			lastIndex = index
		}
	}
}

func (s Backend) interval(ctx context.Context, processChan chan Backend) {
	if s.Onetime {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(s.Interval) * time.Second):
			processChan <- s
		}
	}
}

// Monitor will start to monitor all given Backends for changes.
// It will process all given tamplates on changes.
func (t *Resource) Monitor(ctx context.Context) {
	wg := &sync.WaitGroup{}

	processChan := make(chan Backend)
	defer close(processChan)

	if err := t.process(t.backends); err != nil {
		switch err.(type) {
		case berr.BackendError:
			err := err.(berr.BackendError)
			log.WithFields(logrus.Fields{
				"backend": err.Backend,
			}).Error(err)
		default:
			log.Error(err)
		}
	}

	for _, sc := range t.backends {
		if sc.Watch {
			wg.Add(1)
			go func(s Backend) {
				defer wg.Done()
				s.watch(ctx, processChan)
			}(sc)
		}

		if sc.Interval > 0 {
			wg.Add(1)
			go func(s Backend) {
				defer wg.Done()
				s.interval(ctx, processChan)
			}(sc)
		}
	}

	done := make(chan struct{})
	go func() {
		// If there is no goroutine left - quit
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case storeClient := <-processChan:
			if err := t.process([]Backend{storeClient}); err != nil {
				if _, ok := err.(berr.BackendError); ok {
					log.WithFields(logrus.Fields{
						"backend": storeClient.Name,
					}).Error(err)
				} else {
					log.Error(err)
				}
			}
		case <-ctx.Done():
			go func() {
				for range processChan {
				}
			}()
			wg.Wait()
			return
		case <-done:
			return
		}
	}
}

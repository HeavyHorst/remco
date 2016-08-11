/*
 * This file is part of remco.
 * Based on code from confd. https://github.com/kelseyhightower/confd
 * © 2013 Kelsey Hightower
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/memkv"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template/fileutil"
	"github.com/Sirupsen/logrus"
	"github.com/flosch/pongo2"
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
func NewResource(backends []Backend, sources []*ProcessConfig) (*Resource, error) {
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

		if tr.backends[i].Interval <= 0 && tr.backends[i].Onetime == false {
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
		"store":      storeClient.Name,
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

		tmpl, err := pongo2.FromFile(s.Src)
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

		// Set the owner, group, and mode on the stage file now to make it easier to
		// compare against the destination configuration file later.
		os.Chmod(temp.Name(), s.fileMode)
		os.Chown(temp.Name(), s.UID, s.GID)
		s.stageFile = temp

		temp.Close()

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

		if s.CheckCmd != "" {
			if err := s.check(staged); err != nil {
				return errors.New("Config check failed: " + err.Error())
			}
		}
		log.WithFields(logrus.Fields{
			"config": s.Dst,
		}).Debug("Overwriting target config")
		err := os.Rename(staged, s.Dst)
		if err != nil {
			if strings.Contains(err.Error(), "device or resource busy") {
				log.Debug("Rename failed - target is likely a mount. Trying to write instead")
				// try to open the file and write to it
				var contents []byte
				var rerr error
				contents, rerr = ioutil.ReadFile(staged)
				if rerr != nil {
					return rerr
				}
				err := ioutil.WriteFile(s.Dst, contents, s.fileMode)
				// make sure owner and group match the temp file, in case the file was created with WriteFile
				os.Chown(s.Dst, s.UID, s.GID)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}
		if s.ReloadCmd != "" {
			if err := s.reload(); err != nil {
				return err
			}
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

// setFileMode sets the FileMode.
func (t *Resource) setFileMode() error {
	for _, s := range t.sources {
		if err := s.setFileMode(); err != nil {
			return nil
		}
	}
	return nil
}

// Process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *Resource) process(storeClient Backend) error {
	if err := t.setFileMode(); err != nil {
		return err
	}
	if err := t.setVars(storeClient); err != nil {
		return err
	}
	if err := t.createStageFileAndSync(); err != nil {
		return err
	}
	return nil
}

func (t *Resource) processAll() error {
	//load all KV-pairs the first time
	var err error
	if err = t.setFileMode(); err != nil {
		return err
	}
	for _, storeClient := range t.backends {
		if err := t.setVars(storeClient); err != nil {
			return err
		}
	}
	if err = t.createStageFileAndSync(); err != nil {
		return err
	}
	return nil
}

func (s Backend) watch(stopChan chan bool, processChan chan Backend) {
	var lastIndex uint64
	keysPrefix := appendPrefix(s.Prefix, s.Keys)

	for {
		select {
		case <-stopChan:
			return
		default:
			index, err := s.WatchPrefix(s.Prefix, stopChan, easyKV.WithKeys(keysPrefix), easyKV.WithWaitIndex(lastIndex))
			if err != nil {
				log.Error(err)
				// Prevent backend errors from consuming all resources.
				time.Sleep(time.Second * 2)
				continue
			}
			processChan <- s
			lastIndex = index
		}
	}
}

func (s Backend) interval(stopChan chan bool, processChan chan Backend) {
	if s.Onetime {
		return
	}
	for {
		select {
		case <-stopChan:
			return
		case <-time.After(time.Duration(s.Interval) * time.Second):
			processChan <- s
		}
	}
}

// Monitor will start to monitor all given Backends for changes.
// It will process all given tamplates on changes.
func (t *Resource) Monitor(stopChan chan bool) {
	wg := &sync.WaitGroup{}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)
	stopIntervalWatch := make(chan bool)
	processChan := make(chan Backend)

	defer close(processChan)

	if err := t.processAll(); err != nil {
		log.Error(err)
	}

	for _, sc := range t.backends {
		wg.Add(1)
		if sc.Watch {
			go func(s Backend) {
				defer wg.Done()
				//t.watch(s, stopChan, processChan)
				s.watch(stopIntervalWatch, processChan)
			}(sc)
		} else {
			go func(s Backend) {
				defer wg.Done()
				s.interval(stopIntervalWatch, processChan)
			}(sc)
		}
	}

	go func() {
		// If there is no goroutine left - quit
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case storeClient := <-processChan:
			if err := t.process(storeClient); err != nil {
				log.Error(err)
			}
		case <-stopChan:
			close(stopIntervalWatch)
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

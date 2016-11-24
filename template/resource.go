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
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/HeavyHorst/memkv"
	berr "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/executor"
	"github.com/HeavyHorst/remco/log"
	"github.com/Sirupsen/logrus"
)

// Resource is the representation of a parsed template resource.
type Resource struct {
	backends []Backend
	funcMap  map[string]interface{}
	store    *memkv.Store
	sources  []*Renderer
	logger   *logrus.Entry

	exec       executor.Executor
	SignalChan chan os.Signal

	Failed bool
}

// ErrEmptySrc is returned if an emty src template is passed to NewResource
var ErrEmptySrc = errors.New("empty src template")

// NewResource creates a Resource.
func NewResource(backends []Backend, sources []*Renderer, name string, exec executor.Executor) (*Resource, error) {
	if len(backends) == 0 {
		return nil, errors.New("A valid StoreClient is required.")
	}

	logger := log.WithFields(logrus.Fields{"resource": name})

	for _, v := range sources {
		if v.Src == "" {
			return nil, ErrEmptySrc
		}
		v.logger = logger
	}

	tr := &Resource{
		backends:   backends,
		store:      memkv.New(),
		funcMap:    newFuncMap(),
		sources:    sources,
		logger:     logger,
		SignalChan: make(chan os.Signal, 1),
		exec:       exec,
	}

	// initialize the inidividual backend memkv Stores
	for i := range tr.backends {
		store := memkv.New()
		tr.backends[i].store = store

		if tr.backends[i].Interval <= 0 && tr.backends[i].Onetime == false && tr.backends[i].Watch == false {
			logger.Warning("Interval needs to be > 0: setting interval to 60")
			tr.backends[i].Interval = 60
		}
	}

	addFuncs(tr.funcMap, tr.store.FuncMap)

	return tr, nil
}

// Close is calling the close method on all backends
func (t *Resource) Close() {
	for _, v := range t.backends {
		t.logger.WithFields(logrus.Fields{
			"backend": v.Name,
		}).Debug("Closing client connection")
		v.Close()
	}
}

// setVars sets the Vars for a template resource.
func (t *Resource) setVars(storeClient Backend) error {
	var err error

	t.logger.WithFields(logrus.Fields{
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
				t.logger.Warning("Key collision - " + kv.Key)
			}
			t.store.Set(kv.Key, kv.Value)
		}
	}

	return nil
}

func (t *Resource) createStageFileAndSync() (bool, error) {
	var changed bool
	for _, s := range t.sources {
		err := s.createStageFile(t.funcMap)
		if err != nil {
			return changed, err
		}
		c, err := s.syncFiles()
		changed = changed || c
		if err != nil {
			return changed, err
		}
	}
	return changed, nil
}

// Process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *Resource) process(storeClients []Backend) (bool, error) {
	var changed bool
	var err error
	for _, storeClient := range storeClients {
		if err = t.setVars(storeClient); err != nil {
			return changed, berr.BackendError{
				Message: err.Error(),
				Backend: storeClient.Name,
			}
		}
	}
	if changed, err = t.createStageFileAndSync(); err != nil {
		return changed, err
	}
	return changed, nil
}

// Monitor will start to monitor all given Backends for changes.
// It will process all given tamplates on changes.
func (t *Resource) Monitor(ctx context.Context) {
	t.Failed = false
	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	processChan := make(chan Backend)
	defer close(processChan)
	errChan := make(chan berr.BackendError, 10)

	// try to process the template resource with all given backends
	// we wait a random amount of time (between 0 - 30 seconds)
	// to prevent ddossing our backends and try again (with all backends - no stale data)
	// if some error occurs

	retryChan := make(chan struct{}, 1)
	retryChan <- struct{}{}

retryloop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-retryChan:
			if _, err := t.process(t.backends); err != nil {
				switch err.(type) {
				case berr.BackendError:
					err := err.(berr.BackendError)
					t.logger.WithFields(logrus.Fields{
						"backend": err.Backend,
					}).Error(err)
				default:
					t.logger.Error(err)
				}
				go func() {
					rn := rand.Int63n(30)
					t.logger.Error(fmt.Sprintf("Not all templates could be rendered, trying again after %d seconds", rn))
					time.Sleep(time.Duration(rn) * time.Second)
					select {
					case <-ctx.Done():
						return
					default:
						retryChan <- struct{}{}
					}
				}()
				continue retryloop
			}
			break retryloop
		}
	}

	err := t.exec.SpawnChild()
	if err != nil {
		t.logger.Error(err)
		t.Failed = true
		cancel()
	} else {
		defer t.exec.StopChild()
	}

	done := make(chan struct{})
	wg.Add(1)
	go func() {
		// run the cancel func if the childProcess quit
		defer wg.Done()
		stopped := t.exec.Wait(ctx)
		if stopped {
			t.Failed = true
			cancel()
		}
	}()

	// start the watch and interval processors so that we get notfied on changes
	for _, sc := range t.backends {
		if sc.Watch {
			wg.Add(1)
			go func(s Backend) {
				defer wg.Done()
				s.watch(ctx, processChan, errChan)
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

	go func() {
		// If there is no goroutine left - quit
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case storeClient := <-processChan:
			changed, err := t.process([]Backend{storeClient})
			if err != nil {
				if _, ok := err.(berr.BackendError); ok {
					t.logger.WithFields(logrus.Fields{
						"backend": storeClient.Name,
					}).Error(err)
				} else {
					t.logger.Error(err)
				}
			} else if changed {
				if err := t.exec.Reload(); err != nil {
					t.logger.Error(err)
				}
			}
		case s := <-t.SignalChan:
			t.exec.SignalChild(s)
		case err := <-errChan:
			t.logger.WithFields(logrus.Fields{
				"backend": err.Backend,
			}).Error(err.Message)
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

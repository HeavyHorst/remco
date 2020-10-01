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
	"fmt"
	"github.com/armon/go-metrics"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/HeavyHorst/memkv"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Resource is the representation of a parsed template resource.
type Resource struct {
	backends []Backend
	funcMap  map[string]interface{}
	store    *memkv.Store
	sources  []*Renderer
	logger   *logrus.Entry

	exec      Executor
	startCmd  string
	reloadCmd string
	// SignalChan is a channel to send os.Signal's to all child processes.
	SignalChan chan os.Signal

	// Failed is true if we run Monitor() in exec mode and the child process exits unexpectedly.
	// If the monitor context is canceled as usual Failed is false.
	// Failed is used to restart the Resource on failure.
	Failed bool
}

// ResourceConfig is a configuration struct to create a new resource.
type ResourceConfig struct {
	Exec      ExecConfig
	StartCmd  string
	ReloadCmd string

	// Template is the configuration for all template options.
	// You can configure as much template-destination pairs as you like.
	Template []*Renderer

	// Name gives the Resource a name.
	// This name is added to the logs to distinguish between different resources.
	Name string

	// Connectors is a list of BackendConnectors.
	// The Resource will establish a connection to all of these.
	Connectors []BackendConnector
}

// ErrEmptySrc is returned if an emty src template is passed to NewResource
var ErrEmptySrc = fmt.Errorf("empty src template")

// NewResourceFromResourceConfig creates a new resource from the given ResourceConfig.
func NewResourceFromResourceConfig(ctx context.Context, reapLock *sync.RWMutex, r ResourceConfig) (*Resource, error) {
	backendList, err := connectAllBackends(ctx, r.Connectors)
	if err != nil {
		return nil, errors.Wrap(err, "connectAllBackends failed")
	}

	for _, p := range r.Template {
		p.ReapLock = reapLock
	}

	logger := log.WithFields(logrus.Fields{"resource": r.Name})
	exec := NewExecutor(r.Exec.Command, r.Exec.ReloadSignal, r.Exec.KillSignal, r.Exec.KillTimeout, r.Exec.Splay, logger)
	res, err := NewResource(backendList, r.Template, r.Name, exec, r.StartCmd, r.ReloadCmd)
	if err != nil {
		for _, v := range backendList {
			v.Close()
		}
	}
	return res, err
}

// NewResource creates a Resource.
func NewResource(backends []Backend, sources []*Renderer, name string, exec Executor, startCmd, reloadCmd string) (*Resource, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("a valid StoreClient is required")
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
		startCmd:   startCmd,
		reloadCmd:  reloadCmd,
	}

	// initialize the inidividual backend memkv Stores
	for i := range tr.backends {
		store := memkv.New()
		tr.backends[i].store = store

		if tr.backends[i].Interval <= 0 && !tr.backends[i].Onetime && !tr.backends[i].Watch {
			logger.Warning("interval needs to be > 0: setting interval to 60")
			tr.backends[i].Interval = 60
		}
	}

	addFuncs(tr.funcMap, tr.store.FuncMap)

	return tr, nil
}

// Close closes the connection to all underlying backends.
func (t *Resource) Close() {
	for _, v := range t.backends {
		t.logger.WithFields(logrus.Fields{
			"backend": v.Name,
		}).Debug("closing client connection")
		v.Close()
	}
}

// setVars reads all KV-Pairs for the backend
// and writes these pairs to the individual (per backend) memkv store.
// After that, the instance wide memkv store gets purged and is recreated with all individual
// memkv KV-Pairs.
// Key collisions are logged.
// It returns an error if any.
func (t *Resource) setVars(storeClient Backend) error {
	var err error

	t.logger.WithFields(logrus.Fields{
		"backend":    storeClient.Name,
		"key_prefix": storeClient.Prefix,
	}).Debug("retrieving keys")

	result, err := storeClient.GetValues(appendPrefix(storeClient.Prefix, storeClient.Keys))
	if err != nil {
		return errors.Wrap(err, "getValues failed")
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
				t.logger.Warning("key collision - " + kv.Key)
			}
			t.store.Set(kv.Key, kv.Value)
		}
	}

	return nil
}

func (t *Resource) createStageFileAndSync(runCommands bool) (bool, error) {
	var changed bool
	for _, s := range t.sources {
		err := s.createStageFile(t.funcMap)
		if err != nil {
			metrics.IncrCounter([]string{"files", "stage_errors_total"}, 1)
			return changed, errors.Wrap(err, "create stage file failed")
		}
		metrics.IncrCounter([]string{"files", "staged_total"}, 1)
		c, err := s.syncFiles(runCommands)
		changed = changed || c
		if err != nil {
			metrics.IncrCounter([]string{"files", "sync_errors_total"}, 1)
			return changed, errors.Wrap(err, "sync files failed")
		}
		metrics.IncrCounter([]string{"files", "synced_total"}, 1)
	}
	return changed, nil
}

// Process is a convenience function that wraps calls to the three main tasks
// required to keep local configuration files in sync. First we gather vars
// from the store, then we stage a candidate configuration file, and finally sync
// things up.
// It returns an error if any.
func (t *Resource) process(storeClients []Backend, runCommands bool) (bool, error) {
	var changed bool
	var err error
	for _, storeClient := range storeClients {
		labels := []metrics.Label{{Name: "name", Value: storeClient.Name}}
		if err = t.setVars(storeClient); err != nil {
			metrics.IncrCounterWithLabels([]string{"backends", "sync_errors_total"}, 1, labels)
			return changed, berr.BackendError{
				Message: errors.Wrap(err, "setVars failed").Error(),
				Backend: storeClient.Name,
			}
		}
		metrics.IncrCounterWithLabels([]string{"backends", "synced_total"}, 1, labels)
	}
	if changed, err = t.createStageFileAndSync(runCommands); err != nil {
		return changed, errors.Wrap(err, "createStageFileAndSync failed")
	}
	return changed, nil
}

// Monitor will start to monitor all given Backends for changes.
// It accepts a ctx.Context for cancelation.
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
	retryChan := make(chan struct{}, 1)
	retryChan <- struct{}{}
retryloop:
	for {
		select {
		case <-ctx.Done():
			return
		case <-retryChan:
			if _, err := t.process(t.backends, t.startCmd == ""); err != nil {
				switch err := err.(type) {
				case berr.BackendError:
					t.logger.WithFields(logrus.Fields{
						"backend": err.Backend,
					}).Error(err)
				default:
					t.logger.Error(err)
				}
				go func() {
					rn := rand.Int63n(30)
					t.logger.Error(fmt.Sprintf("not all templates could be rendered, trying again after %d seconds", rn))
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

	if t.startCmd != "" {
		output, err := execCommand(t.startCmd, t.logger, nil)
		if err != nil {
			t.logger.Error(fmt.Sprintf("failed to execute the start cmd - %q", string(output)))
			t.Failed = true
			cancel()
		} else {
			t.logger.Debug(fmt.Sprintf("%q", string(output)))
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
		// Wait for the child process to quit.
		// If the process terminates unexpectedly (the context was NOT canceled), we set t.Failed to true
		// and cancel the resource context. Remco will try to restart the resource if t.Failed is true.
		defer wg.Done()
		failed := t.exec.Wait(ctx)
		if failed {
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
			changed, err := t.process([]Backend{storeClient}, true)
			if err != nil {
				switch err.(type) {
				case berr.BackendError:
					t.logger.WithField("backend", storeClient.Name).Error(err)
				default:
					t.logger.Error(err)
				}
			} else if changed {
				if err := t.exec.Reload(); err != nil {
					t.logger.Error(err)
				}

				if t.reloadCmd != "" {
					output, err := execCommand(t.reloadCmd, t.logger, nil)
					if err != nil {
						t.logger.Error(fmt.Sprintf("failed to execute the resource reload cmd - %q", string(output)))
					}
				}
			}
		case s := <-t.SignalChan:
			err := t.exec.SignalChild(s)
			if err != nil {
				t.logger.Error(err)
			}
		case err := <-errChan:
			t.logger.WithField("backend", err.Backend).Error(err.Message)
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

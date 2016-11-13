/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/pborman/uuid"
)

type status struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	State    string `json:"state"`
	Exec     exec
	Template []*template.Processor
}

// the configWatcher watches the config file for changes
type configWatcher struct {
	stoppedW         chan struct{}
	stopWatch        chan struct{}
	stopWatchConf    chan struct{}
	stoppedWatchConf chan struct{}
	reloadChan       chan struct{}
	wg               sync.WaitGroup
	mu               sync.Mutex
	canceled         bool
	configPath       string

	signalChans      map[string]chan os.Signal
	signalChansMutex sync.RWMutex

	resourceMap      map[string]status
	resourceMapMutex sync.RWMutex
}

func (w *configWatcher) addSignalChan(id string, sigchan chan os.Signal) {
	w.signalChansMutex.Lock()
	defer w.signalChansMutex.Unlock()
	w.signalChans[id] = sigchan
}

func (w *configWatcher) removeSignalChan(id string) {
	w.signalChansMutex.Lock()
	defer w.signalChansMutex.Unlock()
	delete(w.signalChans, id)
}

func (w *configWatcher) sendSignal(s os.Signal) {
	w.signalChansMutex.RLock()
	defer w.signalChansMutex.RUnlock()
	for _, v := range w.signalChans {
		v <- s
	}
}

func (w *configWatcher) setResource(id, state string, r resource) {
	w.resourceMapMutex.Lock()
	defer w.resourceMapMutex.Unlock()
	w.resourceMap[id] = status{
		ID:       id,
		Name:     r.Name,
		State:    state,
		Exec:     r.Exec,
		Template: r.Template,
	}
}

func (w *configWatcher) resetResourceMap() {
	w.resourceMapMutex.Lock()
	defer w.resourceMapMutex.Unlock()
	w.resourceMap = make(map[string]status)
}

func (w *configWatcher) status() ([]byte, error) {
	w.resourceMapMutex.RLock()
	defer w.resourceMapMutex.RUnlock()
	vs := make([]status, len(w.resourceMap))

	idx := 0
	for _, v := range w.resourceMap {
		vs[idx] = v
		idx++
	}
	dat, err := json.MarshalIndent(vs, "", "    ")
	return dat, err
}

func (w *configWatcher) runConfig(c configuration) {
	defer func() {
		w.stoppedW <- struct{}{}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	wait := sync.WaitGroup{}
	for _, v := range c.Resource {
		wait.Add(1)
		go func(r resource) {
			defer wait.Done()
			res, err := r.init(ctx)
			if err != nil {
				log.Error(err)
			} else {
				defer res.Close()
				id := uuid.New()
				w.setResource(id, "running", r)
				w.addSignalChan(id, res.SignalChan)
				defer w.removeSignalChan(id)
				defer w.setResource(id, "stopped", r)
				res.Monitor(ctx)
			}
		}(v)
	}

	go func() {
		// If there is no goroutine left - quit
		wait.Wait()
		close(done)
	}()

	for {
		select {
		case <-w.stopWatch:
			cancel()
			wait.Wait()
			return
		case <-done:
			return
		}
	}
}

func (w *configWatcher) getCanceled() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.canceled
}

// startWatchConfig starts to watch the configuration file and all files under include_dir which ends with .toml
// if there is an event (write, remove, create) we write to w.reloadChan to trigger an relaod and return.
func (w *configWatcher) startWatchConfig(config configuration) {
	w.wg.Add(1)
	defer w.wg.Done()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer watcher.Close()

	// add the configfile to the watcher
	err = watcher.Add(w.configPath)
	if err != nil {
		log.Error(err)
	}

	// add the include_dir to the watcher
	if config.IncludeDir != "" {
		err = watcher.Add(config.IncludeDir)
		if err != nil {
			log.Error(err)
		}
	}

	// watch the config for changes
	for {
		select {
		case <-w.stopWatchConf:
			return
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Create == fsnotify.Create {
				// only watch .toml files
				if strings.HasSuffix(event.Name, ".toml") || event.Name == w.configPath {
					log.WithFields(logrus.Fields{
						"file": event.Name,
					}).Info("config change detected")
					time.Sleep(500 * time.Millisecond)

					// don't try to reload if w is already canceled
					if w.getCanceled() {
						return
					}

					w.reloadChan <- struct{}{}
					return
				}
			}
		}
	}
}

func newConfigWatcher(configPath string, config configuration, done chan struct{}) *configWatcher {
	w := &configWatcher{
		stoppedW:      make(chan struct{}),
		stopWatch:     make(chan struct{}),
		stopWatchConf: make(chan struct{}),
		reloadChan:    make(chan struct{}),
		configPath:    configPath,
		signalChans:   make(map[string]chan os.Signal),
		resourceMap:   make(map[string]status),
	}

	go w.runConfig(config)
	go w.startWatchConfig(config)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.reloadChan:
				log.WithFields(logrus.Fields{
					"file": w.configPath,
				}).Info("loading new config")
				newConf, err := newConfiguration(w.configPath)
				if err != nil {
					log.Error(err)
					continue
				}

				// don't try to relaod anything if w is already canceled
				if w.getCanceled() {
					continue
				}

				// stop the old config and wait until it has stopped
				log.WithFields(logrus.Fields{
					"file": w.configPath,
				}).Info("stopping the old config")
				w.stopWatch <- struct{}{}
				<-w.stoppedW
				go w.runConfig(newConf)
				w.resetResourceMap()

				// restart the fsnotify config watcher (the include_dir folder may have changed)
				// startWatchConfig returns when calling reload (we don't need to stop it)
				go w.startWatchConfig(newConf)
			case <-w.stoppedW:
				// close the reloadChan
				// every attempt to write to reloadChan would block forever otherwise
				close(w.reloadChan)

				// close the done channel
				// this signals the main function that the configWatcher has completed all work
				// for example all backends are configured with onetime=true
				close(done)
				return
			}
		}
	}()

	return w
}

func (w *configWatcher) stop() {
	w.mu.Lock()
	w.canceled = true
	w.mu.Unlock()
	close(w.stopWatch)
	close(w.stopWatchConf)

	// wait for the main routine and startWatchConfig to exit
	w.wg.Wait()
}

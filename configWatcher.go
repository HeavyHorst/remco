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
	"strings"
	"sync"
	"time"

	"github.com/HeavyHorst/remco/log"
	"github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

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
			r.run(ctx)
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

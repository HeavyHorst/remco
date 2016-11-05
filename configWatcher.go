/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HeavyHorst/remco/log"
	"github.com/fsnotify/fsnotify"
)

// the configWatcher watches the config file for changes
type configWatcher struct {
	stoppedW  chan struct{}
	stopWatch chan struct{}

	stopWatchConf    chan struct{}
	stoppedWatchConf chan struct{}

	reloadChan chan struct{}

	wg sync.WaitGroup

	configPath string
}

// call c.run in its own goroutine
// and write to the w.stoppedW chan if its done
func (w *configWatcher) runConfig(c configuration) {
	defer func() {
		w.stoppedW <- struct{}{}
	}()
	c.run(w.stopWatch)
}

// reload stops the old config and starts a new one
// this function blocks forever if runConfig was never called before
func (w *configWatcher) reload() {
	defer func() {
		// we may try to send on the closed channel w.stopWatch
		// we need to recover from this panic
		if r := recover(); r != nil {
			if fmt.Sprintf("%v", r) != "send on closed channel" {
				panic(r)
			}
		}
	}()

	newConf, err := newConfiguration(w.configPath)
	if err != nil {
		log.Error(err)
		return
	}
	// stop the old config and wait until it has stopped
	w.stopWatch <- struct{}{}
	<-w.stoppedW
	go w.runConfig(newConf)

	// restart the fsnotify config watcher (the include_dir folder may have changed)
	// startWatchConfig returns when calling reload
	go w.startWatchConfig(newConf)
}

func (w *configWatcher) startWatchConfig(config configuration) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer func() {
		watcher.Close()
		w.stoppedWatchConf <- struct{}{}
	}()

	// add theconfigfile to the watcher
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
				if strings.HasSuffix(event.Name, ".toml") {
					time.Sleep(500 * time.Millisecond)
					w.reloadChan <- struct{}{}
					return
				}
			}
		}
	}
}

func newConfigWatcher(configPath string, config configuration, done chan struct{}) *configWatcher {
	w := &configWatcher{
		stoppedW:         make(chan struct{}),
		stopWatch:        make(chan struct{}),
		stopWatchConf:    make(chan struct{}),
		stoppedWatchConf: make(chan struct{}),
		reloadChan:       make(chan struct{}),
		configPath:       configPath,
	}

	go w.runConfig(config)
	go w.startWatchConfig(config)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.reloadChan:
				w.reload()
			case <-w.stoppedW:
				close(done)

				// there is no running runConfig function which can answer to the stop method
				// we need to send to the w.stoppedW channel so that we don't block
				w.stoppedW <- struct{}{}
				return
			}
		}
	}()

	return w
}

func (w *configWatcher) stop() {
	close(w.stopWatch)
	<-w.stoppedW
	close(w.stopWatchConf)
	<-w.stoppedWatchConf
	w.wg.Wait()
}

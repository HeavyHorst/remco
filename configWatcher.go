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
	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/remco/log"
)

// the configWatcher watches the config file for changes
type configWatcher struct {
	stoppedCW     chan struct{}
	stoppedW      chan struct{}
	stopWatchConf chan bool
	stopWatch     chan bool
	filePath      string
}

func (w *configWatcher) startWatch(c tomlConf) {
	go func() {
		defer func() {
			w.stoppedW <- struct{}{}
		}()
		c.run(w.stopWatch)
	}()
}

// reload stops the old watcher and starts a new one
// this function blocks forever if startWatch was never called before
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

	newConf, err := NewConf(w.filePath)
	if err != nil {
		log.Error(err)
		return
	}
	// stop the old watcher and wait until it has stopped
	w.stopWatch <- true
	<-w.stoppedW
	// start a new watcher
	w.startWatch(newConf)
}

func newConfigWatcher(filepath string, watcher easyKV.ReadWatcher, config tomlConf, done chan struct{}) *configWatcher {
	w := &configWatcher{
		stoppedW:      make(chan struct{}),
		stoppedCW:     make(chan struct{}),
		stopWatchConf: make(chan bool),
		stopWatch:     make(chan bool),
		filePath:      filepath,
	}

	w.startWatch(config)

	reload := make(chan struct{})
	go func() {
		// watch the config for changes
		for {
			select {
			case <-w.stopWatchConf:
				w.stoppedCW <- struct{}{}
				return
			default:
				_, err := watcher.WatchPrefix("", w.stopWatchConf, easyKV.WithKeys([]string{""}))
				if err != nil {
					if err != easyKV.ErrWatchCanceled {
						log.Error(err)
						time.Sleep(2 * time.Second)
					}
					continue
				}
				time.Sleep(1 * time.Second)
				//w.reload()
				reload <- struct{}{}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-reload:
				w.reload()
			case <-w.stoppedW:
				close(done)
				// there is no runnign startWatch function which can answer to the stop method
				// we need to send to the w.stoppedW channel so that we not block
				w.stoppedW <- struct{}{}
			}
		}
	}()

	return w
}

func (w *configWatcher) stop() {
	close(w.stopWatchConf)
	close(w.stopWatch)
	<-w.stoppedW
	<-w.stoppedCW
}

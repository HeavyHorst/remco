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
	"fmt"
	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/remco/log"
)

// the configWatcher watches the config file for changes
type configWatcher struct {
	stoppedW  chan struct{}
	stopWatch chan struct{}
	filePath  string
	cancel    context.CancelFunc
}

// call c.run in its own goroutine
// and write to the w.stoppedW chan if its done
func (w *configWatcher) runConfig(c configuration) {
	go func() {
		defer func() {
			w.stoppedW <- struct{}{}
		}()
		c.run(w.stopWatch)
	}()
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

	newConf, err := newConfiguration(w.filePath)
	if err != nil {
		log.Error(err)
		return
	}
	// stop the old watcher and wait until it has stopped
	w.stopWatch <- struct{}{}
	<-w.stoppedW
	// start a new watcher
	w.runConfig(newConf)
}

func newConfigWatcher(filepath string, watcher easyKV.ReadWatcher, config configuration, done chan struct{}) *configWatcher {
	w := &configWatcher{
		stoppedW:  make(chan struct{}),
		stopWatch: make(chan struct{}),
		filePath:  filepath,
	}

	w.runConfig(config)

	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	reload := make(chan struct{})
	go func() {
		// watch the config for changes
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, err := watcher.WatchPrefix("", ctx, easyKV.WithKeys([]string{""}))
				if err != nil {
					if err != easyKV.ErrWatchCanceled {
						log.Error(err)
						time.Sleep(2 * time.Second)
					}
					continue
				}
				time.Sleep(1 * time.Second)
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
				// there is no runnign runConfig function which can answer to the stop method
				// we need to send to the w.stoppedW channel so that we don't block
				w.stoppedW <- struct{}{}
			}
		}
	}()

	return w
}

func (w *configWatcher) stop() {
	w.cancel()
	close(w.stopWatch)
	<-w.stoppedW
}

/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"sync"
	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/remco/log"
)

// the configWatcher watches the config file for changes
type configWatcher struct {
	waitAndExit   *sync.WaitGroup
	stopped       chan struct{}
	stopWatchConf chan bool
	stopWatch     chan bool
	filePath      string
}

func newConfigWatcher(filepath string, watcher easyKV.ReadWatcher, oldConf tomlConf) *configWatcher {
	c := oldConf
	w := &configWatcher{
		waitAndExit:   &sync.WaitGroup{},
		stopped:       make(chan struct{}),
		stopWatchConf: make(chan bool),
		stopWatch:     make(chan bool),
		filePath:      filepath,
	}

	startWatch := func(c tomlConf) {
		w.waitAndExit.Add(1)
		go func() {
			defer func() {
				w.waitAndExit.Done()
				w.stopped <- struct{}{}
			}()
			c.watch(w.stopWatch)
		}()
	}

	// reload the config
	reload := func() {
		defer func() {
			// we may try to send on the closed channel w.stopWatch
			// we need to recover from this panic
			if r := recover(); r != nil {
				err := r.(error)
				if err.Error() != "send on closed channel" {
					panic(r)
				}
			}
		}()

		newConf, err := NewConf(w.filePath)
		if err != nil {
			log.Error(err)
			return
		}
		// waitAndExit is temporally increased by 1, so that the program don't terminate after stopWatch
		w.waitAndExit.Add(1)
		// stop the old watcher and wait until it has stopped
		w.stopWatch <- true
		<-w.stopped
		// start a new watcher
		startWatch(newConf)
		w.waitAndExit.Done()

		c = newConf
	}

	startWatch(c)

	go func() {
		// watch the config for changes
		for {
			select {
			case <-w.stopWatchConf:
				w.stopped <- struct{}{}
				return
			default:
				_, err := watcher.WatchPrefix("", w.stopWatchConf, easyKV.WithKeys([]string{""}))
				if err != nil {
					log.Error(err)
					time.Sleep(2 * time.Second)
					continue
				}
				time.Sleep(1 * time.Second)
				reload()
			}
		}
	}()

	return w
}

func (w *configWatcher) stop() {
	close(w.stopWatchConf)
	close(w.stopWatch)
	<-w.stopped
	<-w.stopped
}

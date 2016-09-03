/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/log"
)

var (
	fileConfig          = &file.Config{}
	printVersionAndExit bool
)

func init() {
	const (
		defaultConfig = "/etc/remco/config"
	)
	flag.StringVar(&fileConfig.Filepath, "config", defaultConfig, "path to the configuration file")
	flag.BoolVar(&printVersionAndExit, "version", false, "print version and exit")
}

func run() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	s, err := fileConfig.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	// we need a working config here - exit on error
	c, err := NewConf(fileConfig.Filepath)
	if err != nil {
		log.Fatal(err.Error())
	}

	waitAndExit := &sync.WaitGroup{}
	newCfgChan := make(chan tomlConf)
	stopped := make(chan struct{})
	stopWatch := make(chan bool)
	stopWatchConf := make(chan bool)
	startWatch := func(c tomlConf) {
		waitAndExit.Add(1)
		go func() {
			defer func() {
				waitAndExit.Done()
				stopped <- struct{}{}
			}()
			c.watch(stopWatch)
		}()
	}

	// reload the config
	reload := func() {
		newConf, err := NewConf(fileConfig.Filepath)
		if err != nil {
			log.Error(err)
			return
		}
		if newConf.hash != c.hash {
			newCfgChan <- newConf
		}
		c = newConf
	}

	startWatch(c)

	go func() {
		// watch the config for changes
		var lastIndex uint64
		isWatcher := true
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !isWatcher {
					reload()
				}
			case <-stopWatchConf:
				stopped <- struct{}{}
				return
			default:
				index, err := s.ReadWatcher.WatchPrefix("", stopWatchConf, easyKV.WithWaitIndex(lastIndex), easyKV.WithKeys([]string{""}))
				if err != nil {
					if err == easyKV.ErrWatchNotSupported {
						isWatcher = false
					} else {
						log.Error(err)
						time.Sleep(2 * time.Second)
					}
					continue
				}
				lastIndex = index
				time.Sleep(1 * time.Second)
				reload()
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		// If there is no goroutine left - quit
		// this is needed for the onetime mode
		waitAndExit.Wait()
		close(done)
	}()

	exit := func() {
		// we need to drain the newCfgChan
		go func() {
			for range newCfgChan {
			}
		}()
		close(stopWatch)
		close(stopWatchConf)
		<-stopped
		<-stopped
	}

	for {
		select {
		case cfg := <-newCfgChan:
			// waitAndExit is temporally increased by 1, so that the program don't terminate after stopWatch
			waitAndExit.Add(1)
			// stop the old watcher and wait until it has stopped
			stopWatch <- true
			<-stopped
			// start a new watcher
			startWatch(cfg)
			waitAndExit.Done()
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			exit()
			return
		case <-done:
			exit()
			return
		}
	}
}

func main() {
	flag.Parse()

	if printVersionAndExit {
		printVersion()
		return
	}

	run()
}

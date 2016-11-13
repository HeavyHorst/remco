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
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/HeavyHorst/consul-template/signals"
	"github.com/HeavyHorst/remco/log"
)

var (
	configPath          string
	printVersionAndExit bool
)

func init() {
	const defaultConfig = "/etc/remco/config"
	flag.StringVar(&configPath, "config", defaultConfig, "path to the configuration file")
	flag.BoolVar(&printVersionAndExit, "version", false, "print version and exit")
}

func run() {
	// catch all signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)

	// we need a working config here - exit on error
	c, err := newConfiguration(configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	done := make(chan struct{})
	// watch the configuration file and all files under include_dir for changes
	cfgWatcher := newConfigWatcher(configPath, c, done)
	defer cfgWatcher.stop()

	go func() {

		http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			s, _ := cfgWatcher.status()
			w.Write(s)
		})
		if c.Http != "" {
			http.ListenAndServe(c.Http, nil)
		}
	}()

	for {
		select {
		case s := <-signalChan:
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
				return
			case signals.SignalLookup["SIGCHLD"]:
			default:
				cfgWatcher.sendSignal(s)
			}
		case <-done:
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

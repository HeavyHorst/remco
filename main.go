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
	"syscall"

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
	// handle exit signals
	// signalExitChan := make(chan os.Signal, 1)
	// signal.Notify(signalExitChan, syscall.SIGINT, syscall.SIGTERM)

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

	for {
		select {
		case s := <-signalChan:
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
				return
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

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

	"github.com/HeavyHorst/consul-template/signals"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/runner"
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

	done := make(chan struct{})
	// watch the configuration file and all files under include_dir for changes
	run, err := runner.New(configPath, done)
	if err != nil {
		log.Fatal(err)
	}
	defer run.Stop()
	run.StartStatusHandler()

	for {
		select {
		case s := <-signalChan:
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
				return
			case syscall.SIGHUP:
				run.Reload()
			case signals.SignalLookup["SIGCHLD"]:
			default:
				run.SendSignal(s)
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

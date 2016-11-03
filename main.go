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

	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/log"
)

var (
	fileConfig          = &file.Config{}
	printVersionAndExit bool
)

func init() {
	const defaultConfig = "/etc/remco/config"
	flag.StringVar(&fileConfig.Filepath, "config", defaultConfig, "path to the configuration file")
	flag.BoolVar(&printVersionAndExit, "version", false, "print version and exit")
}

func run() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// we need a working config here - exit on error
	c, err := newConfiguration(fileConfig.Filepath)
	if err != nil {
		log.Fatal(err.Error())
	}

	s, err := fileConfig.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	done := make(chan struct{})
	cfgWatcher := newConfigWatcher(fileConfig.Filepath, s.ReadWatcher, c, done)
	defer cfgWatcher.stop()

	for {
		select {
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			return
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

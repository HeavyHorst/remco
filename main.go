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
	const (
		defaultConfig = "/etc/remco/config"
	)
	flag.StringVar(&fileConfig.Filepath, "config", defaultConfig, "path to the configuration file")
	flag.BoolVar(&printVersionAndExit, "version", false, "print version and exit")
}

func run() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// we need a working config here - exit on error
	c, err := NewConf(fileConfig.Filepath)
	if err != nil {
		log.Fatal(err.Error())
	}
	c.loadGlobals()

	s, err := fileConfig.Connect()
	if err != nil {
		log.Fatal(err.Error())
	}

	cfgWatcher := newConfigWatcher(fileConfig.Filepath, s.ReadWatcher, c)
	defer cfgWatcher.stop()

	done := make(chan struct{})
	go func() {
		// If there is no goroutine left - quit
		// this is needed for the onetime mode
		cfgWatcher.waitAndExit.Wait()
		close(done)
	}()

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

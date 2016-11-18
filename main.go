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

	"github.com/HeavyHorst/consul-template/signals"
	"github.com/HeavyHorst/remco/config"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/runner"
	"github.com/Sirupsen/logrus"
	reap "github.com/hashicorp/go-reap"
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
	reapLock := &sync.RWMutex{}

	cfg, err := config.NewConfiguration(configPath)
	if err != nil {
		log.Fatal(err)
	}

	run := runner.New(cfg, reapLock, done)
	defer run.Stop()

	// reap zombies if pid is 1
	pidReapChan := make(reap.PidCh, 1)
	errorReapChan := make(reap.ErrorCh, 1)
	if os.Getpid() == 1 {
		if !reap.IsSupported() {
			log.Warning("the pid is 1 but zombie reaping is not supported on this platform")
		} else {
			go reap.ReapChildren(pidReapChan, errorReapChan, done, reapLock)
		}
	}

	for {
		select {
		case s := <-signalChan:
			switch s {
			case syscall.SIGHUP:
				log.WithFields(logrus.Fields{
					"file": configPath,
				}).Info("loading new config")
				newConf, err := config.NewConfiguration(configPath)
				if err != nil {
					log.Error(err)
					continue
				}
				run.Reload(newConf)
			case signals.SignalLookup["SIGCHLD"]:
			case os.Interrupt, syscall.SIGTERM:
				log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
				return
			default:
				run.SendSignal(s)
			}
		case pid := <-pidReapChan:
			log.Debug(fmt.Sprintf("Reaped child process %d", pid))
		case err := <-errorReapChan:
			log.Error(fmt.Sprintf("Error reaping child process %v", err))
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

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
	"reflect"
	"sync"
	"syscall"

	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/hashicorp/consul-template/signals"
	"github.com/hashicorp/go-reap"
)

var (
	configPath          string
	printVersionAndExit bool
	onetime             bool
)

func init() {
	const defaultConfig = "/etc/remco/config"
	flag.StringVar(&configPath, "config", defaultConfig, "path to the configuration file")
	flag.BoolVar(&printVersionAndExit, "version", false, "print version and exit")
	flag.BoolVar(&onetime, "onetime", false, "run templating process once and exit")
}

func run() int32 {
	// catch all signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan)

	done := make(chan struct{})
	reapLock := &sync.RWMutex{}

	cfg, err := NewConfiguration(configPath)
	if err != nil {
		log.Fatal("failed to read config", err)
	}

	if onetime {
		for _, res := range cfg.Resource {
			for _, b := range res.Backends.GetBackends() {
				if !reflect.ValueOf(b).IsZero() {
					backend := b.GetBackend()
					backend.Onetime = true
				}
			}
		}
	}

	run := NewSupervisor(cfg, reapLock, done)
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
				log.WithFields(
					"file", configPath,
				).Info("loading new config")
				newConf, err := NewConfiguration(configPath)
				if err != nil {
					log.Error("failed to read config", err)
					continue
				}
				run.Reload(newConf)
			case signals.SignalLookup["SIGCHLD"]:
			case os.Interrupt, syscall.SIGTERM:
				log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
				return 0
			default:
				run.SendSignal(s)
			}
		case pid := <-pidReapChan:
			log.Debug(fmt.Sprintf("Reaped child process %d", pid))
		case err := <-errorReapChan:
			log.Error(fmt.Sprintf("Error reaping child process %v", err))
		case <-done:
			return run.getNumResourceErrors()
		}
	}
}

func main() {
	flag.Parse()

	if printVersionAndExit {
		printVersion()
		return
	}
	rc := run()
	// be on the safe side for portability and lots of backend resources
	if rc > 125 {
		rc = 125
	}
	os.Exit(int(rc))
}

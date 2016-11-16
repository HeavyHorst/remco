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
	"github.com/kardianos/service"
)

var (
	configPath          string
	serviceFlag         string
	printVersionAndExit bool
)

type program struct {
	runner   *runner.Runner
	stopChan chan struct{}
}

func init() {
	const defaultConfig = "/etc/remco/config"
	flag.StringVar(&configPath, "config", defaultConfig, "path to the configuration file")
	flag.BoolVar(&printVersionAndExit, "version", false, "print version and exit")
	flag.StringVar(&serviceFlag, "service", "", "operate on the service")
}

func (p *program) Start(s service.Service) error {
	p.stopChan = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	close(p.stopChan)
	if p.runner != nil {
		p.runner.Stop()
	}

	return nil
}

func (p *program) run() {
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
	p.runner = run

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
				// the service package stops remco on these signals
				// don't forward these signals to the subprocesses
				// so that the configured kill and reload signals work as expected
			default:
				run.SendSignal(s)
			}
		case pid := <-pidReapChan:
			log.Debug(fmt.Sprintf("Reaped child process %d", pid))
		case err := <-errorReapChan:
			log.Error(fmt.Sprintf("Error reaping child process %v", err))
		case <-p.stopChan:
			return
		case <-done:
			run.Stop()
			os.Exit(0)
		}
	}
}

func main() {
	flag.Parse()

	if printVersionAndExit {
		printVersion()
	}

	svcConfig := &service.Config{
		Name:        "remco",
		DisplayName: "remco",
		Description: "Remco remote configuration manager",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if serviceFlag != "" {
		if configPath != "" {
			svcConfig.Arguments = []string{"-config", configPath}
		}
		err := service.Control(s, serviceFlag)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = s.Run()
		if err != nil {
			log.Error(err)
		}
	}
}

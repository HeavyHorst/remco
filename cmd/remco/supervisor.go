/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
	"github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

type reloadSignal struct {
	c        Configuration
	reloaded chan<- struct{}
}

// Supervisor runs
type Supervisor struct {
	stopChan   chan struct{}
	reloadChan chan reloadSignal
	wg         sync.WaitGroup

	signalChans      map[string]chan os.Signal
	signalChansMutex sync.RWMutex

	pidFile string

	reapLock *sync.RWMutex
}

// NewSupervisor creates a new Supervisor
func NewSupervisor(cfg Configuration, reapLock *sync.RWMutex, done chan struct{}) *Supervisor {
	w := &Supervisor{
		stopChan:    make(chan struct{}),
		reloadChan:  make(chan reloadSignal),
		signalChans: make(map[string]chan os.Signal),
		reapLock:    reapLock,
	}

	w.pidFile = cfg.PidFile
	pid := os.Getpid()
	err := w.writePid(pid)
	if err != nil {
		log.WithFields(logrus.Fields{"pid_file": w.pidFile}).Error(err)
	}

	stopChan := make(chan struct{})
	stoppedChan := make(chan struct{})

	go w.runResource(cfg.Resource, stopChan, stoppedChan)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		// close the done channel
		// this signals the main function that the Runner has completed all work
		// for example all backends are configured with onetime=true
		defer close(done)
		for {
			select {
			case rs := <-w.reloadChan:
				// write a new pidfile if the pid filepath has changed
				if rs.c.PidFile != w.pidFile {
					err := w.deletePid()
					if err != nil {
						log.WithFields(logrus.Fields{"pid_file": w.pidFile}).Error(err)
					}
					w.pidFile = rs.c.PidFile
					err = w.writePid(pid)
					if err != nil {
						log.WithFields(logrus.Fields{"pid_file": w.pidFile}).Error(err)
					}
				}
				stopChan <- struct{}{}
				<-stoppedChan
				go w.runResource(rs.c.Resource, stopChan, stoppedChan)
				rs.reloaded <- struct{}{}
			case <-stoppedChan:
				return
			case <-w.stopChan:
				stopChan <- struct{}{}
				<-stoppedChan
				return
			}
		}
	}()

	return w
}

func (ru *Supervisor) writePid(pid int) error {
	if ru.pidFile == "" {
		return nil
	}

	log.Info(fmt.Sprintf("creating pid file at %q", ru.pidFile))

	err := ioutil.WriteFile(ru.pidFile, []byte(fmt.Sprintf("%d", pid)), 0666)
	if err != nil {
		return errors.Wrap(err, "couldn't write pid file")
	}
	return nil
}

func (ru *Supervisor) deletePid() error {
	if ru.pidFile == "" {
		return nil
	}

	log.Debug(fmt.Sprintf("removing pid file at %q", ru.pidFile))

	stat, err := os.Stat(ru.pidFile)
	if err != nil {
		return errors.Wrap(err, "couldn't get file stats")
	}

	if stat.IsDir() {
		return fmt.Errorf("the pid file path seems to be a directory")
	}

	return os.Remove(ru.pidFile)
}

func (ru *Supervisor) addSignalChan(id string, sigchan chan os.Signal) {
	ru.signalChansMutex.Lock()
	defer ru.signalChansMutex.Unlock()
	ru.signalChans[id] = sigchan
}

func (ru *Supervisor) removeSignalChan(id string) {
	ru.signalChansMutex.Lock()
	defer ru.signalChansMutex.Unlock()
	delete(ru.signalChans, id)
}

// SendSignal forwards the given Signal to all child processes
func (ru *Supervisor) SendSignal(s os.Signal) {
	ru.signalChansMutex.RLock()
	defer ru.signalChansMutex.RUnlock()
	// try to send the signal to all child processes
	// we don't block here if the signal can't be send
	for _, v := range ru.signalChans {
		select {
		case v <- s:
		default:
		}
	}
}

func (ru *Supervisor) runResource(r []Resource, stop, stopped chan struct{}) {
	defer func() {
		if stopped != nil {
			stopped <- struct{}{}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	wait := sync.WaitGroup{}
	for _, v := range r {
		wait.Add(1)
		go func(r Resource) {
			defer wait.Done()

			backendConfigs := r.Backend.GetBackends()
			for _, v := range r.Backend.Plugin {
				backendConfigs = append(backendConfigs, &v)
			}

			rsc := template.ResourceConfig{
				Exec:       r.Exec,
				Template:   r.Template,
				Name:       r.Name,
				Connectors: backendConfigs,
			}
			res, err := template.NewResourceFromResourceConfig(ctx, ru.reapLock, rsc)
			if err != nil {
				log.Error(err)
				return
			}
			defer res.Close()

			id := uuid.New()
			ru.addSignalChan(id, res.SignalChan)
			defer ru.removeSignalChan(id)

			restartChan := make(chan struct{}, 1)
			restartChan <- struct{}{}

			for {
				select {
				case <-ctx.Done():
					return
				case <-restartChan:
					res.Monitor(ctx)
					if res.Failed {
						go func() {
							// try to restart the resource after a random amount of time
							rn := rand.Int63n(30)
							log.WithFields(logrus.Fields{
								"resource": r.Name,
							}).Error(fmt.Sprintf("resource execution failed, restarting after %d seconds", rn))
							time.Sleep(time.Duration(rn) * time.Second)
							select {
							case <-ctx.Done():
								return
							default:
								restartChan <- struct{}{}
							}
						}()
					} else {
						return
					}
				}
			}
		}(v)
	}

	go func() {
		// If there is no goroutine left - quit
		// this is necessary for the onetime mode
		wait.Wait()
		close(done)
	}()

	for {
		select {
		case <-stop:
			cancel()
			wait.Wait()
			return
		case <-done:
			return
		}
	}
}

// Reload rereads the configuration, stops the old Runner and starts a new one.
func (ru *Supervisor) Reload(cfg Configuration) {
	reloaded := make(chan struct{})
	ru.reloadChan <- reloadSignal{
		c:        cfg,
		reloaded: reloaded,
	}
	<-reloaded
}

// Stop stops the Runner gracefully.
func (ru *Supervisor) Stop() {
	close(ru.stopChan)
	// wait for the main routine to exit
	ru.wg.Wait()

	// remove the pidfile
	err := ru.deletePid()
	if err != nil {
		log.WithFields(logrus.Fields{"pid_file": ru.pidFile}).Error(err)
	}
}

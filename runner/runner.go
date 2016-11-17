/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package runner

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/HeavyHorst/remco/config"
	"github.com/HeavyHorst/remco/log"
	"github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// Runner runs
type Runner struct {
	stoppedW         chan struct{}
	stopWatch        chan struct{}
	stopWatchConf    chan struct{}
	stoppedWatchConf chan struct{}
	reloadChan       chan config.Configuration
	wg               sync.WaitGroup
	mu               sync.Mutex
	canceled         bool

	signalChans      map[string]chan os.Signal
	signalChansMutex sync.RWMutex

	pidFile string

	reapLock *sync.RWMutex
}

// New creates a new Runner
func New(cfg config.Configuration, reapLock *sync.RWMutex, done chan struct{}) *Runner {
	w := &Runner{
		stoppedW:      make(chan struct{}),
		stopWatch:     make(chan struct{}),
		stopWatchConf: make(chan struct{}),
		reloadChan:    make(chan config.Configuration),
		signalChans:   make(map[string]chan os.Signal),
		reapLock:      reapLock,
	}

	w.pidFile = cfg.PidFile
	pid := os.Getpid()
	err := w.writePid(pid)
	if err != nil {
		log.WithFields(logrus.Fields{"pid_file": w.pidFile}).Error(err)
	}

	go w.runConfig(cfg)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case newConf := <-w.reloadChan:
				// don't try to relaod anything if w is already canceled
				if w.getCanceled() {
					continue
				}
				// write a new pidfile if the pid filepath has changed
				if newConf.PidFile != w.pidFile {
					err := w.deletePid()
					if err != nil {
						log.WithFields(logrus.Fields{"pid_file": w.pidFile}).Error(err)
					}
					w.pidFile = newConf.PidFile
					err = w.writePid(pid)
					if err != nil {
						log.WithFields(logrus.Fields{"pid_file": w.pidFile}).Error(err)
					}
				}
				w.stopWatch <- struct{}{}
				<-w.stoppedW
				go w.runConfig(newConf)
			case <-w.stoppedW:
				// close the reloadChan
				// every attempt to write to reloadChan would block forever otherwise
				close(w.reloadChan)

				// close the done channel
				// this signals the main function that the Runner has completed all work
				// for example all backends are configured with onetime=true
				close(done)
				return
			}
		}
	}()

	return w
}

func (ru *Runner) writePid(pid int) error {
	if ru.pidFile == "" {
		return nil
	}

	log.Info(fmt.Sprintf("creating pid file at %q", ru.pidFile))

	err := ioutil.WriteFile(ru.pidFile, []byte(fmt.Sprintf("%d", pid)), 0666)
	if err != nil {
		return fmt.Errorf("could not create pid file: %s", err)
	}
	return nil
}

func (ru *Runner) deletePid() error {
	if ru.pidFile == "" {
		return nil
	}

	log.Debug(fmt.Sprintf("removing pid file at %q", ru.pidFile))

	stat, err := os.Stat(ru.pidFile)
	if err != nil {
		return fmt.Errorf("could not remove pid file: %s", err)
	}

	if stat.IsDir() {
		return fmt.Errorf("the pid file path seems to be a directory")
	}

	return os.Remove(ru.pidFile)
}

func (ru *Runner) addSignalChan(id string, sigchan chan os.Signal) {
	ru.signalChansMutex.Lock()
	defer ru.signalChansMutex.Unlock()
	ru.signalChans[id] = sigchan
}

func (ru *Runner) removeSignalChan(id string) {
	ru.signalChansMutex.Lock()
	defer ru.signalChansMutex.Unlock()
	delete(ru.signalChans, id)
}

// SendSignal forwards the given Signal to all child processes
func (ru *Runner) SendSignal(s os.Signal) {
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

func (ru *Runner) runConfig(c config.Configuration) {
	defer func() {
		ru.stoppedW <- struct{}{}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	wait := sync.WaitGroup{}
	for _, v := range c.Resource {
		wait.Add(1)
		go func(r config.Resource) {
			defer wait.Done()
			res, err := r.Init(ctx, ru.reapLock)
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
		wait.Wait()
		close(done)
	}()

	for {
		select {
		case <-ru.stopWatch:
			cancel()
			wait.Wait()
			return
		case <-done:
			return
		}
	}
}

func (ru *Runner) getCanceled() bool {
	ru.mu.Lock()
	defer ru.mu.Unlock()
	return ru.canceled
}

// Reload rereads the configuration, stops the old Runner and starts a new one.
func (ru *Runner) Reload(cfg config.Configuration) {
	ru.reloadChan <- cfg
}

// Stop stops the Runner gracefully.
func (ru *Runner) Stop() {
	ru.mu.Lock()
	if ru.canceled {
		ru.mu.Unlock()
		return
	}
	ru.canceled = true
	ru.mu.Unlock()
	close(ru.stopWatch)
	close(ru.stopWatchConf)

	// wait for the main routine and startWatchConfig to exit
	ru.wg.Wait()

	// remove the pidfile
	err := ru.deletePid()
	if err != nil {
		log.WithFields(logrus.Fields{"pid_file": ru.pidFile}).Error(err)
	}
}

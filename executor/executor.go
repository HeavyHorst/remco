/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package executor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/HeavyHorst/consul-template/child"
	"github.com/HeavyHorst/consul-template/signals"
	"github.com/Sirupsen/logrus"
	shellwords "github.com/mattn/go-shellwords"
)

// Executor provides some methods to control a subprocess
type Executor struct {
	child            *child.Child
	childLock        *sync.RWMutex
	execCommand      string
	reloadSignal     os.Signal
	killSignal       os.Signal
	killTimeout      time.Duration
	splay            time.Duration
	logger           *logrus.Entry
	restartOnFailure bool
}

// New creates a new Executor
func New(execCommand, reloadSignal, killSignal string, killTimeout, splay int, logger *logrus.Entry, restartOnFailure bool) Executor {
	var rs, ks os.Signal
	var err error
	if reloadSignal != "" {
		rs, err = signals.Parse(reloadSignal)
		if err != nil {
			logger.Error(err)
		}
	}

	// default killSignal is SIGTERM
	if killSignal == "" {
		killSignal = "SIGTERM"
	}
	ks, err = signals.Parse(killSignal)
	if err != nil {
		logger.Error(err)
		ks = syscall.SIGTERM
	}

	// set killTimeout to 10 if its not defined
	if killTimeout == 0 {
		killTimeout = 10
	}

	return Executor{
		childLock:        &sync.RWMutex{},
		execCommand:      execCommand,
		reloadSignal:     rs,
		killSignal:       ks,
		killTimeout:      time.Duration(killTimeout) * time.Second,
		splay:            time.Duration(splay) * time.Second,
		logger:           logger,
		restartOnFailure: restartOnFailure,
	}
}

// SignalChild forwards the os.Signal to the child process
func (e *Executor) SignalChild(s os.Signal) error {
	e.childLock.RLock()
	defer e.childLock.RUnlock()

	if e.child != nil {
		if err := e.child.Signal(s); err != nil {
			return err
		}
	}
	return nil
}

// StopChild stops the child process
// it blocks until the child quits.
// the child will be killed if it takes longer than
// killTimeout to stop it.
func (e *Executor) StopChild() {
	e.childLock.RLock()
	defer e.childLock.RUnlock()

	if e.child != nil {
		e.child.Stop()
	}
}

// SpawnChild parses e.execCommand and starts the child process accordingly
func (e *Executor) SpawnChild() error {
	if e.execCommand != "" {
		e.childLock.Lock()
		defer e.childLock.Unlock()

		p := shellwords.NewParser()
		p.ParseBacktick = true
		args, err := p.Parse(e.execCommand)
		if err != nil {
			return err
		}

		child, err := child.New(&child.NewInput{
			Stdin:        os.Stdin,
			Stdout:       os.Stdout,
			Stderr:       os.Stderr,
			Command:      args[0],
			Args:         args[1:],
			ReloadSignal: e.reloadSignal,
			KillSignal:   e.killSignal,
			KillTimeout:  e.killTimeout,
			Splay:        e.splay,
			Logger:       e.logger,
		})

		if err != nil {
			return fmt.Errorf("error creating child: %s", err)
		}
		e.child = child
		if err := e.child.Start(); err != nil {
			return fmt.Errorf("error starting child: %s", err)
		}
		return nil
	}
	return nil
}

// CancelOnExit waits for the children to stop.
// If the children stop it calls the cancel function.
// CancelOnExit ignores reloads.
func (e *Executor) CancelOnExit(ctx context.Context, cancel context.CancelFunc) {
	var exitChan <-chan int

	e.childLock.RLock()
	notNil := (e.child != nil)
	if !notNil {
		e.childLock.RUnlock()
		return
	}
	exitChan = e.child.ExitCh()
	e.childLock.RUnlock()

	for {
		if notNil {
			select {
			case <-ctx.Done():
				return
			case code := <-exitChan:
				// wait a little bit to give the process time to start
				// in case of a reload
				time.Sleep(1 * time.Second)
				e.childLock.RLock()
				nexitChan := e.child.ExitCh()
				// the exitChan has changed which means the process was reloaded
				// don't exit in this case
				if nexitChan != exitChan {
					exitChan = nexitChan
					e.childLock.RUnlock()
					continue
				}
				// the process exited - stop
				e.childLock.RUnlock()
				if code != 0 {
					e.logger.Error(fmt.Sprintf("(child) process exited unexpectedly with code: %d", code))
					if e.restartOnFailure {
						e.restartChild()
						e.childLock.RLock()
						exitChan = e.child.ExitCh()
						e.childLock.RUnlock()
						continue
					}
				}
				cancel()
			}
		}
	}
}

func (e *Executor) restartChild() {
	e.childLock.Lock()
	defer e.childLock.Unlock()

	if err := e.child.Start(); err != nil {
		e.logger.Error(fmt.Sprintf("error starting child: %s", err))
	}
}

// Reload reloads the child process.
// If a reloadSignal is provided it will send this signal to the child.
// Otherwise the child process will be killed and a new one started afterwards.
func (e *Executor) Reload() error {
	e.childLock.RLock()
	defer e.childLock.RUnlock()
	if e.child != nil {
		if err := e.child.Reload(); err != nil {
			return err
		}
	}
	return nil
}

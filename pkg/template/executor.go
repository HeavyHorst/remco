/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"os"
	"syscall"
	"time"

	"github.com/hashicorp/consul-template/child"
	"github.com/hashicorp/consul-template/signals"
	"github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
)

// ExecConfig represents the configuration values for the exec mode.
type ExecConfig struct {
	// Command is the command to execute.	// Backtick parsing is supported:
	//   ./foo `echo $SHELL`
	Command string `json:"command"`

	// ReloadSignal is the Signal that is sended to the subpocess if we want to reload it.
	// If no signal is specified the child process will be killed (gracefully) and started again.
	ReloadSignal string `toml:"reload_signal" json:"reload_signal"`

	// KillSignal defines the signal sent to the child process when remco is gracefully shutting down.
	//
	// The application needs to exit before the kill_timeout, it will be terminated otherwise (like kill -9).
	// The default value is “SIGTERM”.
	KillSignal string `toml:"kill_signal" json:"kill_signal"`

	// KillTimeout - the maximum amount of time in seconds to wait for the child process to gracefully terminate.
	KillTimeout int `toml:"kill_timeout" json:"kill_timeout"`

	// A random splay to wait before killing the command.
	// May be useful in large clusters to prevent all child processes to reload at the same time when configuration changes occur.
	Splay int `json:"splay"`
}

type childSignal struct {
	signal os.Signal
	err    chan<- error
}

type exitC struct {
	exitChan <-chan int
	valid    bool
}

// An Executor controls a subprocess.
// It can control the whole process lifecycle and can
// reload and stop the process gracefully or send other signals to
// the child process using channels.
type Executor struct {
	execCommand  string
	reloadSignal os.Signal
	killSignal   os.Signal
	killTimeout  time.Duration
	splay        time.Duration
	logger       hclog.Logger

	stopChan   chan chan<- error
	reloadChan chan chan<- error
	signalChan chan childSignal
	exitChan   chan chan exitC
}

// NewExecutor creates a new Executor.
func NewExecutor(execCommand, reloadSignal, killSignal string, killTimeout, splay int, logger hclog.Logger) Executor {
	var rs, ks os.Signal
	var err error

	if logger == nil {
		logger = hclog.Default()
	}

	if reloadSignal != "" {
		rs, err = signals.Parse(reloadSignal)
		if err != nil {
			logger.Error("failed to parse reload signal", err)
		}
	}

	// default killSignal is SIGTERM
	if killSignal == "" {
		killSignal = "SIGTERM"
	}
	ks, err = signals.Parse(killSignal)
	if err != nil {
		logger.Error("failed to parse kill signal", err)
		ks = syscall.SIGTERM
	}

	// set killTimeout to 10 if its not defined
	if killTimeout == 0 {
		killTimeout = 10
	}

	return Executor{
		execCommand:  execCommand,
		reloadSignal: rs,
		killSignal:   ks,
		killTimeout:  time.Duration(killTimeout) * time.Second,
		splay:        time.Duration(splay) * time.Second,
		logger:       logger,
		stopChan:     make(chan chan<- error),
		reloadChan:   make(chan chan<- error),
		signalChan:   make(chan childSignal),
		exitChan:     make(chan chan exitC),
	}
}

// SpawnChild parses e.execCommand and starts the child process accordingly.
// Backtick parsing is supported:
//   ./foo `echo $SHELL`
//
// only call this once !
func (e *Executor) SpawnChild() error {
	var c *child.Child
	if e.execCommand != "" {
		p := shellwords.NewParser()
		p.ParseBacktick = true
		args, err := p.Parse(e.execCommand)
		if err != nil {
			return err
		}

		c, err = child.New(&child.NewInput{
			Stdin:        os.Stdin,
			Stdout:       os.Stdout,
			Stderr:       os.Stderr,
			Command:      args[0],
			Args:         args[1:],
			ReloadSignal: e.reloadSignal,
			KillSignal:   e.killSignal,
			KillTimeout:  e.killTimeout,
			Splay:        e.splay,
			Logger:       e.logger.StandardLogger(&hclog.StandardLoggerOptions{InferLevels: true}),
		})

		if err != nil {
			return fmt.Errorf("error creating child: %s", err)
		}
		//e.child = child
		if err := c.Start(); err != nil {
			return fmt.Errorf("error starting child: %s", err)
		}
	}

	go func() {
		for {
			select {
			case errchan := <-e.stopChan:
				if c != nil {
					c.Stop()
				}
				errchan <- nil
				return
			case errchan := <-e.reloadChan:
				var err error
				if c != nil {
					err = c.Reload()
				}
				errchan <- err
			case s := <-e.signalChan:
				var err error
				if c != nil {
					err = c.Signal(s.signal)
				}
				s.err <- err
			case exit := <-e.exitChan:
				if c != nil {
					ex := exitC{
						valid:    true,
						exitChan: c.ExitCh(),
					}
					exit <- ex
				} else {
					ex := exitC{
						valid: false,
					}
					exit <- ex
				}
			}
		}
	}()

	return nil
}

// SignalChild forwards the os.Signal to the child process.
func (e *Executor) SignalChild(s os.Signal) error {
	err := make(chan error)

	signal := childSignal{
		signal: s,
		err:    err,
	}

	e.signalChan <- signal
	return <-err

}

// StopChild stops the child process.
//
// It blocks until the child quits or the killTimeout is reached.
// The child will be killed if it takes longer than killTimeout to stop it.
func (e *Executor) StopChild() {
	errchan := make(chan error)
	e.stopChan <- errchan
	<-errchan
}

// Reload reloads the child process.
// If a reloadSignal is provided it will send this signal to the child.
// The child process will be killed and restarted otherwise.
func (e *Executor) Reload() error {
	errchan := make(chan error)
	e.reloadChan <- errchan
	if err := <-errchan; err != nil {
		return errors.Wrap(err, "reload failed")
	}
	return nil
}

func (e *Executor) getExitChan() (<-chan int, bool) {
	ecc := make(chan exitC)
	e.exitChan <- ecc
	exit := <-ecc
	return exit.exitChan, exit.valid
}

// Wait waits for the child to stop.
// Returns true if the command stops unexpectedly and false if the context is canceled.
//
// Wait ignores reloads.
func (e *Executor) Wait(ctx context.Context) bool {
	exitChan, valid := e.getExitChan()
	if !valid {
		return false
	}

	for {
		select {
		case <-ctx.Done():
			return false
		case <-exitChan:
			// wait a little bit to give the process time to start
			// in case of a reload
			time.Sleep(1 * time.Second)
			nexitChan, _ := e.getExitChan()
			// the exitChan has changed which means the process was reloaded
			// don't exit in this case
			if nexitChan != exitChan {
				exitChan = nexitChan
				continue
			}
			// the process exited - stop
			return true
		}
	}
}

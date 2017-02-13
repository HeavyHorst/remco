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
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
)

func TestNew(t *testing.T) {
	command := "echo"
	reloadSignal := "SIGHUP"
	killSignal := "SIGKILL"
	killTimeout := 1
	splay := 0
	logger := &logrus.Entry{}

	exec := NewExecutor(command, reloadSignal, killSignal, killTimeout, splay, logger)

	if exec.killSignal != os.Kill {
		t.Errorf("killSignal should be: %v", os.Kill)
	}

	if exec.reloadSignal != syscall.SIGHUP {
		t.Errorf("reloadSignal should be: %v", syscall.SIGHUP)
	}

	if exec.execCommand != command {
		t.Errorf("execCommand should be: %s", command)
	}

	if exec.killTimeout != time.Duration(killTimeout)*time.Second {
		t.Errorf("killTimeout should be: %v", time.Duration(killTimeout)*time.Second)
	}

	if exec.splay != time.Duration(splay)*time.Second {
		t.Errorf("splay should be: %v", time.Duration(splay)*time.Second)
	}

	if exec.logger != logger {
		t.Errorf("logger should be: %v", logger)
	}
}

func TestNewDefaults(t *testing.T) {
	exec := NewExecutor("", "", "", 0, 0, &logrus.Entry{})

	if exec.killSignal != syscall.SIGTERM {
		t.Errorf("default killSignal should be: %v", syscall.SIGTERM)
	}

	if exec.reloadSignal != nil {
		t.Error("default reloadSignal should be nil")
	}

	if exec.killTimeout != 10*time.Second {
		t.Errorf("default killTimeout should be: %v", 10*time.Second)
	}
}

func TestNewInvalidSignals(t *testing.T) {
	logger := logrus.New()
	logger.Out = ioutil.Discard
	exec := NewExecutor("", "SIGBLA", "SIGBLA", 0, 0, logrus.NewEntry(logger))

	if exec.reloadSignal != nil {
		t.Error("reloadSignal should be nil")
	}

	if exec.killSignal != syscall.SIGTERM {
		t.Errorf("killSignal should be: %v", syscall.SIGTERM)
	}
}

func spawnChild(command string) (Executor, error) {
	reloadSignal := "SIGINT"
	killSignal := "SIGTERM"
	killTimeout := 2
	logger := logrus.New()
	logger.Out = ioutil.Discard

	exec := NewExecutor(command, reloadSignal, killSignal, killTimeout, 0, logrus.NewEntry(logger))
	err := exec.SpawnChild()

	return exec, err
}

func spawnTrapChild() (Executor, error) {
	command := `bash -c "trap ' ' SIGINT SIGTERM; while true; do sleep 10; done"`
	return spawnChild(command)
}

func spawnTimeOutChild() (Executor, error) {
	command := "bash -c 'sleep 5'"
	return spawnChild(command)
}

func TestStopChildTimeOut(t *testing.T) {
	exec, err := spawnTrapChild()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)

	ticker := time.Tick(4 * time.Second)
	stopped := make(chan struct{})

	go func() {
		exec.StopChild()
		stopped <- struct{}{}
	}()

	select {
	case <-ticker:
		t.Error("killTimeout is not working")
	case <-stopped:
		return
	}
}

func TestWait(t *testing.T) {
	exec, err := spawnTimeOutChild()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)

	nc := make(chan bool)
	go func() {
		s := exec.Wait(context.Background())
		nc <- s
	}()

	if <-nc != true {
		t.Error("the context was not canceled, should be true")
	}
}

func TestWaitCancel(t *testing.T) {
	exec, err := spawnTimeOutChild()
	if err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)

	nc := make(chan bool)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		s := exec.Wait(ctx)
		nc <- s
	}()

	if <-nc != false {
		t.Error("the context was canceled, should be false")
	}

	exec.StopChild()
}

func TestReload(t *testing.T) {
	command := "bash -c 'sleep 5'"
	logger := logrus.New()
	logger.Out = ioutil.Discard

	exec := NewExecutor(command, "", "", 0, 0, logrus.NewEntry(logger))
	err := exec.SpawnChild()
	if err != nil {
		t.Error(err)
	}

	// exitChan := exec.child.ExitCh()
	exitChan, valid := exec.getExitChan()
	if !valid {
		t.Error("we should have a valid exitChan")
	}

	nc := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		s := exec.Wait(ctx)
		nc <- s
	}()

	ticker := time.Tick(2 * time.Second)

	err = exec.Reload()
	if err != nil {
		t.Error(err)
	}

	select {
	case <-nc:
		t.Error("exec.Wait returned, that should never happen on a reload")
	case <-ticker:
	}

	// should be different after the reload
	//nexitChan := exec.child.ExitCh()
	nexitChan, valid := exec.getExitChan()
	if !valid {
		t.Error("we should have a valid exitChan")
	}
	if exitChan == nexitChan {
		t.Error("reload failed")
	}

	exec.StopChild()
}

func TestSignalChild(t *testing.T) {
	command := "bash -c 'sleep 5'"
	logger := logrus.New()
	logger.Out = ioutil.Discard

	exec := NewExecutor(command, "", "", 0, 0, logrus.NewEntry(logger))
	err := exec.SpawnChild()
	if err != nil {
		t.Error(err)
	}

	nc := make(chan bool)
	go func() {
		s := exec.Wait(context.Background())
		nc <- s
	}()

	err = exec.SignalChild(os.Interrupt)
	if err != nil {
		t.Error(err)
	}

	// the program should exit when it receives the os.Interrupt
	n := <-nc
	if n != true {
		t.Error("the context wasn't canceled, exec.Wait should have returned true")
	}

	exec.StopChild()
}

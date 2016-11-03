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
	"io/ioutil"
	"os"
	"sync"

	"github.com/HeavyHorst/remco/backends"
	backendErrors "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/naoina/toml"
)

type resource struct {
	Template []*template.ProcessConfig
	Backend  backends.Config
}

// configuration is the representation of an config file
type configuration struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
	Resource  []resource
}

// newConfiguration reads the file at `path`, expand the environment variables
// and unmarshals it to a new configuration struct.
// it returns an error if any.
func newConfiguration(path string) (configuration, error) {
	var c configuration
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	buf = []byte(os.ExpandEnv(string(buf)))
	if err := toml.Unmarshal(buf, &c); err != nil {
		return c, err
	}

	c.loadGlobals()

	return c, nil
}

// loadGlobals configures remco with the global configuration options
// for example it sets the log level and log formatting
func (c *configuration) loadGlobals() {
	if c.LogLevel != "" {
		err := log.SetLevel(c.LogLevel)
		if err != nil {
			log.Error(err)
		}
	}
	if c.LogFormat != "" {
		log.SetFormatter(c.LogFormat)
	}
}

// run connects to all given backends and starts the template processing as defined in the config file
func (c *configuration) run(stop chan bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})

	wait := sync.WaitGroup{}
	for _, v := range c.Resource {
		var backendList template.Backends

		// try to connect to all supported backends
		// add them to the backendList on success
		// log an error otherwise
		for _, config := range v.Backend.GetBackends() {
			b, err := config.Connect()
			if err == nil {
				backendList = append(backendList, b)
			} else if err != backendErrors.ErrNilConfig {
				log.WithFields(logrus.Fields{
					"backend": b.Name,
				}).Error(err)
			}
		}

		// create a new template resource with given backends and template processing configuration
		t, err := template.NewResource(backendList, v.Template)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		wait.Add(1)
		go func() {
			defer wait.Done()
			// make sure that all backend clients are closed cleanly
			defer backendList.Close()
			t.Monitor(ctx)
		}()
	}

	go func() {
		// If there is no goroutine left - quit
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

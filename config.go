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

// tomlConf is the representation of an config file
type tomlConf struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
	Resource  []struct {
		Template []*template.ProcessConfig
		Backend  backends.Config
	}
}

func NewConf(path string) (tomlConf, error) {
	var c tomlConf
	err := c.fromFile(path)
	if err != nil {
		return c, err
	}
	return c, nil
}

// load a config from file
func (c *tomlConf) fromFile(cfg string) error {
	buf, err := ioutil.ReadFile(cfg)
	if err != nil {
		return err
	}
	buf = []byte(os.ExpandEnv(string(buf)))
	if err := toml.Unmarshal(buf, c); err != nil {
		return err
	}
	return nil
}

func (c *tomlConf) loadGlobals() {
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

func (c *tomlConf) run(stop chan bool) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	c.loadGlobals()

	wait := &sync.WaitGroup{}
	for _, v := range c.Resource {
		var backendList []template.Backend
		for name, config := range v.Backend.GetBackends() {
			b, err := config.Connect()
			if err == nil {
				backendList = append(backendList, b)
			} else if err != backendErrors.ErrNilConfig {
				log.WithFields(logrus.Fields{
					"backend": name,
				}).Error(err)
			}
		}

		t, err := template.NewResource(backendList, v.Template)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		wait.Add(1)
		go func() {
			defer func() {
				// close all client connections
				for _, v := range backendList {
					log.WithFields(logrus.Fields{
						"backend": v.Name,
					}).Debug("Closing client connection")
					v.ReadWatcher.Close()
				}
				wait.Done()
			}()
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

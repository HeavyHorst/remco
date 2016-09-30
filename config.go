/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

// tomlConf is the representation of an config file
import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/configmap"
	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/HeavyHorst/remco/backends/env"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/backends/redis"
	"github.com/HeavyHorst/remco/backends/vault"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/naoina/toml"
)

type tomlConf struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
	Resource  []struct {
		Template []*template.ProcessConfig
		Backend  struct {
			Etcd      *etcd.Config
			File      *file.Config
			Env       *env.Config
			Consul    *consul.Config
			Vault     *vault.Config
			Redis     *redis.Config
			Configmap *configmap.Config
		}
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
	stopChan := make(chan bool)
	done := make(chan struct{})

	c.loadGlobals()

	wait := &sync.WaitGroup{}
	for _, v := range c.Resource {
		var backendList []template.Backend
		backendConfigMap := map[string]template.BackendConfig{
			"etcd":      v.Backend.Etcd,
			"file":      v.Backend.File,
			"env":       v.Backend.Env,
			"consul":    v.Backend.Consul,
			"vault":     v.Backend.Vault,
			"redis":     v.Backend.Redis,
			"configmap": v.Backend.Configmap,
		}

		for name, config := range backendConfigMap {
			b, err := config.Connect()
			if err == nil {
				backendList = append(backendList, b)
			} else if err != backends.ErrNilConfig {
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
			defer wait.Done()
			t.Monitor(stopChan)
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
			close(stopChan)
			wait.Wait()
			return
		case <-done:
			return
		}
	}
}

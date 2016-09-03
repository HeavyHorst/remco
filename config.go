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
	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/HeavyHorst/remco/backends/env"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/backends/redis"
	"github.com/HeavyHorst/remco/backends/vault"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/hashstructure"
	"github.com/naoina/toml"
)

type tomlConf struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
	Resource  []struct {
		Template []*template.ProcessConfig
		Backend  struct {
			Etcdconfig   *etcd.Config
			Fileconfig   *file.Config
			Envconfig    *env.Config
			Consulconfig *consul.Config
			Vaultconfig  *vault.Config
			Redisconfig  *redis.Config
		}
	}
	hash uint64
}

func NewConf(path string) (tomlConf, error) {
	var c tomlConf
	err := c.fromFile(path)
	if err != nil {
		return c, err
	}
	return c, nil
}

// getHash is a helper function to calculate the hash of the current configuration
func (c *tomlConf) getHash() uint64 {
	hash, err := hashstructure.Hash(c, nil)
	if err != nil {
		return 0
	}
	return hash
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
	c.hash = c.getHash()
	return nil
}

func (c *tomlConf) watch(stop chan bool) {
	stopChan := make(chan bool)
	done := make(chan struct{})

	if c.LogLevel != "" {
		err := log.SetLevel(c.LogLevel)
		if err != nil {
			log.Error(err)
		}
	}

	if c.LogFormat != "" {
		log.SetFormatter(c.LogFormat)
	}

	wait := &sync.WaitGroup{}
	for _, v := range c.Resource {
		var backendList []template.Backend
		backendConfigMap := map[string]template.BackendConfig{
			"etcd":   v.Backend.Etcdconfig,
			"file":   v.Backend.Fileconfig,
			"env":    v.Backend.Envconfig,
			"consul": v.Backend.Consulconfig,
			"vault":  v.Backend.Vaultconfig,
			"redis":  v.Backend.Redisconfig,
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

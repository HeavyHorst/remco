/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/backends/vault"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/hashstructure"
	"github.com/naoina/toml"
	"github.com/spf13/cobra"
)

// CfgCmd represents the config command
var CfgCmd = &cobra.Command{
	Use:   "config",
	Short: "parses the provided config file and process any number of templates",
}

type tomlConf struct {
	LogLevel  string `toml:"log_level"`
	LogFormat string `toml:"log_format"`
	Resource  []struct {
		Template []*template.ProcessConfig
		Backend  struct {
			Etcdconfig   *etcd.Config
			Fileconfig   *file.Config
			Consulconfig *consul.Config
			Vaultconfig  *vault.Config
		}
	}
	hash uint64
}

// the default config load function - works with every ReadWatcher
func defaultReload(client easyKV.ReadWatcher, config string) func() (tomlConf, error) {
	//load the new config
	return func() (tomlConf, error) {
		var c tomlConf
		values, err := client.GetValues([]string{config})
		if err != nil {
			return c, err
		}
		if len(values) == 0 {
			return c, errors.New("configuration is empty")
		}
		if err := toml.Unmarshal([]byte(os.ExpandEnv(values[config])), &c); err != nil {
			return c, err
		}
		c.hash = c.getHash()
		return c, nil
	}
}

func defaultConfigRunCMD(config template.BackendConfig) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		s, err := config.Connect()
		if err != nil {
			log.Fatal(err.Error())
		}
		config, _ := cmd.Flags().GetString("config")
		loadConf := defaultReload(s.ReadWatcher, config)
		// we need a working config here - exit on error
		c, err := loadConf()
		if err != nil {
			log.Fatal(err.Error())
		}
		c.configWatch(s.ReadWatcher, config, loadConf)
	}
}

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
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
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
			"consul": v.Backend.Consulconfig,
			"vault":  v.Backend.Vaultconfig,
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
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(stopChan)
			wait.Wait()
			return
		case <-stop:
			close(stopChan)
			wait.Wait()
			return
		case <-done:
			return
		}
	}
}

func (c *tomlConf) configWatch(cli easyKV.ReadWatcher, prefix string, reloadFunc func() (tomlConf, error)) {
	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	wg.Add(1)
	stopWatch := make(chan bool)
	go func() {
		defer wg.Done()
		c.watch(stopWatch)
	}()

	watchConfChan := make(chan struct{})
	go func() {
		var lastIndex uint64
		stop := make(chan bool)
		for {
			index, err := cli.WatchPrefix(prefix, stop, easyKV.WithWaitIndex(lastIndex), easyKV.WithKeys([]string{""}))
			if err != nil {
				log.Error(err)
				// Prevent backend errors from consuming all resources.
				time.Sleep(time.Second * 2)
				continue
			}
			lastIndex = index
			watchConfChan <- struct{}{}
		}
	}()

	go func() {
		// If there is no goroutine left - quit
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-watchConfChan:
			time.Sleep(1 * time.Second)

			newConf, err := reloadFunc()
			if err != nil {
				log.Error(err.Error())
				continue
			}

			if newConf.hash != c.hash {
				log.Info("Configuration has changed - reload remco")
				wg.Add(1)
				// stop the old Resource
				log.Debug("Stopping the old instance")
				stopWatch <- true
				// and start the new Resource
				log.Debug("Starting the new instance")
				go func() {
					defer wg.Done()
					newConf.watch(stopWatch)
				}()
				c.hash = newConf.hash
			} else {
				log.Debug("config is unchanged")
			}
		case <-done:
			return
		}
	}
}

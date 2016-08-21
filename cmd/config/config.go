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
	"github.com/HeavyHorst/remco/backends/redis"
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
			Redisconfig  *redis.Config
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

func defaultConfigRunCMD(config template.BackendConfig, reloadFunc ...func() (tomlConf, error)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		var loadConf func() (tomlConf, error)

		s, err := config.Connect()
		if err != nil {
			log.Fatal(err.Error())
		}
		config, _ := cmd.Flags().GetString("config")

		if len(reloadFunc) > 0 {
			loadConf = reloadFunc[0]
		} else {
			loadConf = defaultReload(s.ReadWatcher, config)
		}

		// we need a working config here - exit on error
		c, err := loadConf()
		if err != nil {
			log.Fatal(err.Error())
		}

		wg := &sync.WaitGroup{}
		stopWatch := make(chan bool)

		startWatch := func(c tomlConf) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				c.watch(stopWatch)
			}()
		}

		startWatch(c)

		go func() {
			for {
				newConf := c.watchConfig(s.ReadWatcher, config, stopWatch, loadConf)
				stopWatch <- true
				wg.Wait()
				startWatch(newConf)
				c = newConf
			}
		}()

		for {
			select {
			case s := <-signalChan:
				log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
				close(stopWatch)
				wg.Wait()
				return
			}
		}
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

func (c *tomlConf) watchConfig(cli easyKV.ReadWatcher, prefix string, stop chan bool, reloadFunc func() (tomlConf, error)) tomlConf {
	var lastIndex uint64

	for {
		select {
		case <-stop:
			return tomlConf{}
		default:
			index, err := cli.WatchPrefix(prefix, stop, easyKV.WithWaitIndex(lastIndex), easyKV.WithKeys([]string{""}))
			if err != nil {
				if err == easyKV.ErrWatchNotSupported {
					// Just wait some time - then reread the config
					time.Sleep(time.Second * 30)
				} else {
					log.Error(err)
					// Prevent backend errors from consuming all resources.
					time.Sleep(time.Second * 2)
					continue
				}
			}

			lastIndex = index
			time.Sleep(1 * time.Second)

			newConf, err := reloadFunc()
			if err != nil {
				log.Error(err.Error())
				continue
			}

			if newConf.hash != c.hash {
				return newConf
			}
		}
	}
}

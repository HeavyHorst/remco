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

// tomlConf is the representation of an config file
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
func defaultLoad(client easyKV.ReadWatcher, config string) func() (tomlConf, error) {
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
		cfgPath, _ := cmd.Flags().GetString("config")
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		var loadConf func() (tomlConf, error)

		s, err := config.Connect()
		if err != nil {
			log.Fatal(err.Error())
		}

		if len(reloadFunc) > 0 {
			loadConf = reloadFunc[0]
		} else {
			loadConf = defaultLoad(s.ReadWatcher, cfgPath)
		}

		// we need a working config here - exit on error
		c, err := loadConf()
		if err != nil {
			log.Fatal(err.Error())
		}

		waitAndExit := &sync.WaitGroup{}
		newCfgChan := make(chan tomlConf)
		stopped := make(chan struct{})
		stopWatch := make(chan bool)
		stopWatchConf := make(chan bool)
		startWatch := func(c tomlConf) {
			waitAndExit.Add(1)
			go func() {
				defer func() {
					waitAndExit.Done()
					stopped <- struct{}{}
				}()
				c.watch(stopWatch)
			}()
		}

		// reload the config
		reload := func() {
			newConf, err := loadConf()
			if err != nil {
				log.Error(err)
				return
			}
			if newConf.hash != c.hash {
				newCfgChan <- newConf
			}
			c = newConf
		}

		startWatch(c)

		go func() {
			// watch the config for changes
			var lastIndex uint64
			isWatcher := true
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if !isWatcher {
						reload()
					}
				case <-stopWatchConf:
					stopped <- struct{}{}
					return
				default:
					index, err := s.ReadWatcher.WatchPrefix(cfgPath, stopWatchConf, easyKV.WithWaitIndex(lastIndex), easyKV.WithKeys([]string{""}))
					if err != nil {
						if err == easyKV.ErrWatchNotSupported {
							isWatcher = false
						} else {
							log.Error(err)
							time.Sleep(2 * time.Second)
						}
						continue
					}
					lastIndex = index
					time.Sleep(1 * time.Second)
					reload()
				}
			}
		}()

		done := make(chan struct{})
		go func() {
			// If there is no goroutine left - quit
			// this is needed for the onetime mode
			waitAndExit.Wait()
			close(done)
		}()

		exit := func() {
			// we need to drain the newCfgChan
			go func() {
				for range newCfgChan {
				}
			}()
			close(stopWatch)
			close(stopWatchConf)
			<-stopped
			<-stopped
		}

		for {
			select {
			case cfg := <-newCfgChan:
				// waitAndExit is temporally increased by 1, so that the program don't terminate after stopWatch
				waitAndExit.Add(1)
				// stop the old watcher and wait until it has stopped
				stopWatch <- true
				<-stopped
				// start a new watcher
				startWatch(cfg)
				waitAndExit.Done()
			case s := <-signalChan:
				log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
				exit()
				return
			case <-done:
				exit()
				return
			}
		}
	}
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

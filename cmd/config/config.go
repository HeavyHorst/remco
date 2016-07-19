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

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/backends/vault"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
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
}

func defaultReload(client backends.StoreClient, config string) func() (tomlConf, error) {
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

		if err := toml.Unmarshal([]byte(values[config]), &c); err != nil {
			return c, err
		}

		return c, nil
	}
}

func (c *tomlConf) fromFile(cfg string) error {
	buf, err := ioutil.ReadFile(cfg)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(buf, c); err != nil {
		return err
	}
	return nil
}

func (c *tomlConf) watch(stop chan bool) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	stopChan := make(chan bool)
	done := make(chan bool)

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
		var backends []template.Backend

		if v.Backend.Etcdconfig != nil {
			_, err := v.Backend.Etcdconfig.Connect()
			if err == nil {
				backends = append(backends, v.Backend.Etcdconfig.Backend)
			} else {
				log.WithFields(logrus.Fields{
					"backend": "etcd",
				}).Error(err)
			}
		}
		if v.Backend.Fileconfig != nil {
			_, err := v.Backend.Fileconfig.Connect()
			if err == nil {
				backends = append(backends, v.Backend.Fileconfig.Backend)
			} else {
				log.WithFields(logrus.Fields{
					"backend": "file",
				}).Error(err)
			}
		}
		if v.Backend.Consulconfig != nil {
			_, err := v.Backend.Consulconfig.Connect()
			if err == nil {
				backends = append(backends, v.Backend.Consulconfig.Backend)
			} else {
				log.WithFields(logrus.Fields{
					"backend": "consul",
				}).Error(err)
			}
		}
		if v.Backend.Vaultconfig != nil {
			_, err := v.Backend.Vaultconfig.Connect()
			if err == nil {
				backends = append(backends, v.Backend.Vaultconfig.Backend)
			} else {
				log.WithFields(logrus.Fields{
					"backend": "vault",
				}).Error(err)
			}
		}

		t, err := template.NewResource(backends, v.Template)
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

func (c *tomlConf) configWatch(cli backends.StoreClient, prefix string, reloadFunc func() (tomlConf, error)) {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	stopWatch := make(chan bool)
	go func() {
		defer wg.Done()
		c.watch(stopWatch)
	}()

	go func() {
		var lastIndex uint64
		stop := make(chan bool)
		for {
			index, err := cli.WatchPrefix(prefix, []string{""}, lastIndex, stop)
			if err != nil {
				log.Error(err)
				// Prevent backend errors from consuming all resources.
				time.Sleep(time.Second * 2)
				continue
			}
			lastIndex = index
			log.Info("Configuration has changed - reload remco")
			time.Sleep(1 * time.Second)

			newConf, err := reloadFunc()
			if err != nil {
				log.Error(err.Error())
				continue
			}
			//c = &newConf

			wg.Add(1)
			// stop the old Resource
			stopWatch <- true
			log.Debug("Stopping the old instance")
			// and start the new Resource
			log.Debug("Starting the new instance")
			go func() {
				defer wg.Done()
				newConf.watch(stopWatch)
			}()
		}
	}()
	wg.Wait()
}

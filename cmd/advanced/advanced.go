package advanced

import (
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
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/naoina/toml"
	"github.com/spf13/cobra"
)

type tomlConf struct {
	LogLevel string `toml:"log-level"`
	Resource []struct {
		Template []*template.SrcDst
		Backend  struct {
			Etcdconfig   *etcd.Config
			Fileconfig   *file.Config
			Consulconfig *consul.Config
		}
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

	wait := &sync.WaitGroup{}
	for _, v := range c.Resource {
		var storeClients []template.StoreConfig

		if v.Backend.Etcdconfig != nil {
			_, err := v.Backend.Etcdconfig.Connect()
			if err == nil {
				storeClients = append(storeClients, v.Backend.Etcdconfig.StoreConfig)
			}
		}
		if v.Backend.Fileconfig != nil {
			_, err := v.Backend.Fileconfig.Connect()
			if err == nil {
				storeClients = append(storeClients, v.Backend.Fileconfig.StoreConfig)
			}
		}
		if v.Backend.Consulconfig != nil {
			_, err := v.Backend.Consulconfig.Connect()
			if err == nil {
				storeClients = append(storeClients, v.Backend.Consulconfig.StoreConfig)
			}
		}

		t, err := template.NewResource(storeClients, v.Template)
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

func (c *tomlConf) configWatch(cli backends.StoreClient, reloadFunc func() (tomlConf, error)) {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	stopWatch := make(chan bool)
	go func() {
		defer wg.Done()
		c.watch(stopWatch)
	}()

	go func() {
		stop := make(chan bool)
		for {
			cli.WatchPrefix("", []string{}, 0, stop)
			log.Info("Config changed on disk - reload remco")
			time.Sleep(1 * time.Second)

			newConf, err := reloadFunc()
			if err != nil {
				log.Error(err.Error())
				time.Sleep(2 * time.Second)
				continue
			}
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

// Cmd represents the advanced command
var Cmd = &cobra.Command{
	Use:   "advanced",
	Short: "advanced mode - parses the provided config file and process any number of templates",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := cmd.Flags().GetString("config")
		f, _ := file.NewFileClient(cfg)

		loadConf := func() (tomlConf, error) {
			//load the new config
			var c tomlConf
			err := c.fromFile(cfg)
			if err != nil {
				return c, err
			}
			return c, nil
		}

		// we need a working config here - exit on error
		c, err := loadConf()
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		c.configWatch(f, loadConf)
	},
}

func init() {
	Cmd.Flags().String("config", "", "Absolute path to the config file")
}

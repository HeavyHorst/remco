package advanced

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/template"
	"github.com/cloudflare/cfssl/log"
	"github.com/naoina/toml"
	"github.com/spf13/cobra"
)

type tomlConf struct {
	Config []struct {
		Template []*template.SrcDst
		Backend  struct {
			Etcdconfig   *etcd.Config
			Fileconfig   *file.Config
			Consulconfig *consul.Config
		}
	}
}

func (t *tomlConf) fromFile(cfg string) error {
	f, err := os.Open(cfg)
	if err != nil {
		return err
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(buf, t); err != nil {
		return err
	}
	return nil
}

// Cmd represents the advanced command
var Cmd = &cobra.Command{
	Use:   "advanced",
	Short: "advanced mode - parses the provided config file and process any number of templates",
	Run: func(cmd *cobra.Command, args []string) {
		var c tomlConf

		cfg, _ := cmd.Flags().GetString("config")
		err := c.fromFile(cfg)
		if err != nil {
			log.Critical(err)
			os.Exit(1)
		}

		wait := &sync.WaitGroup{}
		for _, v := range c.Config {
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
				log.Error(err)
				continue
			}

			wait.Add(1)
			go func() {
				defer wait.Done()
				t.Monitor()
			}()
		}
		wait.Wait()
	},
}

func init() {
	Cmd.Flags().String("config", "", "Absolute path to the config file")
}

// Copyright © 2016 The Remco Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		Cmd struct {
			Check  string
			Reload string
		}
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

			t, err := template.NewResource(storeClients, v.Template, v.Cmd.Reload, v.Cmd.Check)
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

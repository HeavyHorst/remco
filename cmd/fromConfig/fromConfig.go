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

package fromConfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/cloudflare/cfssl/log"
	"github.com/naoina/toml"
	"github.com/spf13/cobra"
)

type tomlConf struct {
	Config []struct {
		FileMode string
		Cmd      struct {
			Check  string
			Reload string
		}
		Template struct {
			Src string
			Dst string
		}
		Backend struct {
			Name         string
			Prefix       string
			Interval     int
			Keys         []string
			Etcdconfig   etcd.Config
			Fileconfig   file.Config
			Consulconfig consul.Config
		}
	}
}

// Cmd represents the fromConfig command
var Cmd = &cobra.Command{
	Use:   "fromConfig",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := cmd.Flags().GetString("config")

		f, err := os.Open(cfg)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		defer f.Close()
		buf, err := ioutil.ReadAll(f)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		var c tomlConf
		if err := toml.Unmarshal(buf, &c); err != nil {
			log.Error(err)
			os.Exit(1)
		}

		//wait := &sync.WaitGroup{}
		for _, v := range c.Config {
			data, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(data))
		}

	},
}

func init() {
	Cmd.Flags().String("config", "", "Absolute path to the config file")
}

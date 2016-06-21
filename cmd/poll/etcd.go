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

package poll

import (
	"os"
	"strings"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/template"
	"github.com/HeavyHorst/temp/backends/etcdv3"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

type etcdConfig struct {
	nodes     []string
	cert      string
	key       string
	caCert    string
	basicAuth bool
	username  string
	password  string
	version   int
}

var config = etcdConfig{}

var pollEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "use etcd2/3 as the backend source",

	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Backend nodes set to " + strings.Join(config.nodes, ", "))

		var err error
		var client backends.StoreClient
		if config.version == 3 {
			client, err = etcdv3.NewEtcdClient(config.nodes, config.cert, config.key, config.caCert, config.basicAuth, config.username, config.password)
		} else {
			client, err = etcd.NewEtcdClient(config.nodes, config.cert, config.key, config.caCert, config.basicAuth, config.username, config.password)
		}

		t, err := template.NewTemplateResource(client, "/", cmd.Flags())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		interval, _ := cmd.Flags().GetInt("interval")
		t.Interval(interval)
	},
}

func init() {
	pollEtcdCmd.PersistentFlags().StringSliceVar(&config.nodes, "nodes", []string{"http://127.0.0.1:2379"}, "The etcd nodes")
	pollEtcdCmd.PersistentFlags().StringVar(&config.cert, "cert", "", "The client cert file")
	pollEtcdCmd.PersistentFlags().StringVar(&config.key, "key", "", "The client key file")
	pollEtcdCmd.PersistentFlags().StringVar(&config.caCert, "caCert", "", "The client CA key file")
	pollEtcdCmd.PersistentFlags().BoolVar(&config.basicAuth, "basicAuth", false, "Enable etcd basic auth with username and password")
	pollEtcdCmd.PersistentFlags().StringVar(&config.username, "username", "", "username")
	pollEtcdCmd.PersistentFlags().StringVar(&config.password, "password", "", "password")
	pollEtcdCmd.PersistentFlags().IntVar(&config.version, "apiversion", 2, "The etcd version (2/3)")

	Cmd.AddCommand(pollEtcdCmd)
}

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

package watch

import (
	"os"
	"strings"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/etcdv3"
	"github.com/HeavyHorst/remco/template"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func (e *etcdConfig) newTemplateRes(flags *pflag.FlagSet) (*template.TemplateResource, error) {
	var err error
	var client backends.StoreClient

	log.Info("Backend nodes set to " + strings.Join(config.nodes, ", "))
	if e.version == 3 {
		client, err = etcdv3.NewEtcdClient(config.nodes, config.cert, config.key, config.caCert, config.basicAuth, config.username, config.password)
	} else {
		client, err = etcd.NewEtcdClient(config.nodes, config.cert, config.key, config.caCert, config.basicAuth, config.username, config.password)
	}
	if err != nil {
		return nil, err
	}
	return template.NewTemplateResource(client, flags)
}

var config = etcdConfig{}

var watchEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "use etcd2/3 as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		t, err := config.newTemplateRes(cmd.Flags())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		t.Monitor()
	},
}

var pollEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "use etcd2/3 as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		t, err := config.newTemplateRes(cmd.Flags())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		interval, _ := cmd.Flags().GetInt("interval")
		t.Interval(interval)
	},
}

func init() {
	cmds := []*cobra.Command{watchEtcdCmd, pollEtcdCmd}
	for _, v := range cmds {
		v.Flags().StringSliceVar(&config.nodes, "nodes", []string{"http://127.0.0.1:2379"}, "The etcd nodes")
		v.Flags().StringVar(&config.cert, "cert", "", "The client cert file")
		v.Flags().StringVar(&config.key, "key", "", "The client key file")
		v.Flags().StringVar(&config.caCert, "caCert", "", "The client CA key file")
		v.Flags().BoolVar(&config.basicAuth, "basicAuth", false, "Enable etcd basic auth with username and password")
		v.Flags().StringVar(&config.username, "username", "", "username")
		v.Flags().StringVar(&config.password, "password", "", "password")
		v.Flags().IntVar(&config.version, "apiversion", 2, "The etcd version (2/3)")
	}

	PollCmd.AddCommand(pollEtcdCmd)
	WatchCmd.AddCommand(watchEtcdCmd)
}

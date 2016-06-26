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
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/spf13/cobra"
)

var config = &etcd.Config{}

var watchEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "use etcd2/3 as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		watch(config, cmd)
	},
}

var pollEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "use etcd2/3 as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		poll(config, cmd)
	},
}

func init() {
	cmds := []*cobra.Command{watchEtcdCmd, pollEtcdCmd}
	for _, v := range cmds {
		v.Flags().StringSliceVar(&config.Nodes, "nodes", []string{"http://127.0.0.1:2379"}, "The etcd nodes")
		v.Flags().StringVar(&config.Cert, "cert", "", "The client cert file")
		v.Flags().StringVar(&config.Key, "key", "", "The client key file")
		v.Flags().StringVar(&config.CaCert, "caCert", "", "The client CA key file")
		v.Flags().BoolVar(&config.BasicAuth, "basicAuth", false, "Enable etcd basic auth with username and password")
		v.Flags().StringVar(&config.Username, "username", "", "username")
		v.Flags().StringVar(&config.Password, "password", "", "password")
		v.Flags().IntVar(&config.Version, "apiversion", 2, "The etcd version (2/3)")
	}

	PollCmd.AddCommand(pollEtcdCmd)
	WatchCmd.AddCommand(watchEtcdCmd)
}

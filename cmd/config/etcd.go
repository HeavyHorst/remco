/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/spf13/cobra"
)

var etcdConfig = &etcd.Config{}

// EtcdCmd represents the etcd command
var EtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "load a config file from etcd",
	Run:   defaultConfigRunCMD(etcdConfig),
}

func init() {
	EtcdCmd.Flags().StringSliceVar(&etcdConfig.Nodes, "nodes", []string{"http://127.0.0.1:2379"}, "The etcd nodes")
	EtcdCmd.Flags().StringVar(&etcdConfig.ClientCert, "client-cert", "", "The client cert file")
	EtcdCmd.Flags().StringVar(&etcdConfig.ClientKey, "client-key", "", "The client key file")
	EtcdCmd.Flags().StringVar(&etcdConfig.ClientCaKeys, "client-ca-keys", "", "The client CA key file")
	EtcdCmd.Flags().BoolVar(&etcdConfig.BasicAuth, "basicAuth", false, "Enable etcd basic auth with username and password")
	EtcdCmd.Flags().StringVar(&etcdConfig.Username, "username", "", "The username")
	EtcdCmd.Flags().StringVar(&etcdConfig.Password, "password", "", "The password")
	EtcdCmd.Flags().IntVar(&etcdConfig.Version, "api-version", 2, "The etcd version (2/3)")
	EtcdCmd.Flags().StringP("config", "c", "", "The path in etcd where the config is stored")

	CfgCmd.AddCommand(EtcdCmd)
}

/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"github.com/HeavyHorst/remco/log"

	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/spf13/cobra"
)

var consulConfig = &consul.Config{}

// ConsulCmd represents the consul command
var ConsulCmd = &cobra.Command{
	Use:   "consul",
	Short: "load a config file from consul",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := consulConfig.Connect()
		if err != nil {
			log.Fatal(err.Error())
		}

		config, _ := cmd.Flags().GetString("config")

		loadConf := defaultReload(s.StoreClient, config)
		// we need a working config here - exit on error
		c, err := loadConf()
		if err != nil {
			log.Fatal(err.Error())
		}
		c.configWatch(s.StoreClient, config, loadConf)
	},
}

func init() {
	ConsulCmd.Flags().StringSliceVar(&consulConfig.Nodes, "nodes", []string{"127.0.0.1:8500"}, "The consul nodes")
	ConsulCmd.Flags().StringVar(&consulConfig.ClientCert, "client-cert", "", "The client cert file")
	ConsulCmd.Flags().StringVar(&consulConfig.ClientKey, "client-key", "", "The client key file")
	ConsulCmd.Flags().StringVar(&consulConfig.ClientCaKeys, "client-ca-keys", "", "The client CA key file")
	ConsulCmd.Flags().StringVar(&consulConfig.Scheme, "scheme", "http", "The backend URI scheme")
	ConsulCmd.Flags().StringP("config", "c", "", "The path in consul where the config is stored")

	CfgCmd.AddCommand(ConsulCmd)
}

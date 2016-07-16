package config

import (
	"errors"
	"log"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv2"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv3"
	"github.com/naoina/toml"
	"github.com/spf13/cobra"
)

// EtcdCmd represents the etcd command
var EtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "load a config file from etcd",
	Run: func(cmd *cobra.Command, args []string) {
		nodes, _ := cmd.Flags().GetStringSlice("nodes")
		cert, _ := cmd.Flags().GetString("cert")
		key, _ := cmd.Flags().GetString("key")
		caCert, _ := cmd.Flags().GetString("caCert")
		basicAuth, _ := cmd.Flags().GetBool("basicAuth")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		api, _ := cmd.Flags().GetInt("apiversion")
		config, _ := cmd.Flags().GetString("config")

		var e backends.StoreClient
		var err error
		if api == 2 {
			e, err = etcdv2.NewEtcdClient(nodes, cert, key, caCert, basicAuth, username, password)
		} else if api == 3 {
			e, err = etcdv3.NewEtcdClient(nodes, cert, key, caCert, basicAuth, username, password)
		}

		if err != nil {
			log.Fatal(err.Error())
		}

		loadConf := func() (tomlConf, error) {
			//load the new config
			var c tomlConf
			values, err := e.GetValues([]string{config})
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

		// we need a working config here - exit on error
		c, err := loadConf()
		if err != nil {
			log.Fatal(err.Error())
		}
		c.configWatch(e, "", loadConf)
	},
}

func init() {
	EtcdCmd.Flags().StringSlice("nodes", []string{"http://127.0.0.1:2379"}, "The etcd nodes")
	EtcdCmd.Flags().String("cert", "", "The client cert file")
	EtcdCmd.Flags().String("key", "", "The client key file")
	EtcdCmd.Flags().String("caCert", "", "The client CA key file")
	EtcdCmd.Flags().Bool("basicAuth", false, "Enable etcd basic auth with username and password")
	EtcdCmd.Flags().String("username", "", "username")
	EtcdCmd.Flags().String("password", "", "password")
	EtcdCmd.Flags().Int("apiversion", 2, "The etcd version (2/3)")
	EtcdCmd.Flags().StringP("config", "c", "", "The path in etcd where the config is stored")

	CfgCmd.AddCommand(EtcdCmd)
}

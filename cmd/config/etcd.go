package config

import (
	"github.com/HeavyHorst/remco/log"

	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/spf13/cobra"
)

var etcdConfig = &etcd.Config{}

// EtcdCmd represents the etcd command
var EtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "load a config file from etcd",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := etcdConfig.Connect()
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
		c.configWatch(s.StoreClient, "", loadConf)
	},
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

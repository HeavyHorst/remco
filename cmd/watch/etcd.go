package watch

import (
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

var config = &etcd.Config{}

var watchEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "use etcd2/3 as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		err := watch(config, cmd)
		if err != nil {
			log.Error(err)
		}
	},
}

var pollEtcdCmd = &cobra.Command{
	Use:   "etcd",
	Short: "use etcd2/3 as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		err := poll(config, cmd)
		if err != nil {
			log.Error(err)
		}
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

package watch

import (
	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

var consulConfig = &consul.Config{}

var watchConsulCmd = &cobra.Command{
	Use:   "consul",
	Short: "use consul k/v as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		err := watch(consulConfig, cmd)
		if err != nil {
			log.Error(err)
		}
	},
}

var pollConsulCmd = &cobra.Command{
	Use:   "consul",
	Short: "use consul k/v as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		err := poll(consulConfig, cmd)
		if err != nil {
			log.Error(err)
		}
	},
}

func init() {
	cmds := []*cobra.Command{watchConsulCmd, pollConsulCmd}
	for _, v := range cmds {
		v.Flags().StringSliceVar(&consulConfig.Nodes, "nodes", []string{"127.0.0.1:8500"}, "The consul nodes")
		v.Flags().StringVar(&consulConfig.Cert, "cert", "", "The client cert file")
		v.Flags().StringVar(&consulConfig.Key, "key", "", "The client key file")
		v.Flags().StringVar(&consulConfig.CaCert, "caCert", "", "The client CA key file")
		v.Flags().StringVar(&consulConfig.Scheme, "scheme", "http", "The backend URI scheme")
	}

	PollCmd.AddCommand(pollConsulCmd)
	WatchCmd.AddCommand(watchConsulCmd)
}

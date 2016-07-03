package watch

import (
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

var fc = &file.Config{}

// Cmd represents the file command
var watchFileCmd = &cobra.Command{
	Use:   "file",
	Short: "use a simple json/yaml file as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		err := watch(fc, cmd)
		if err != nil {
			log.Error(err)
		}
	},
}

// Cmd represents the file command
var pollFileCmd = &cobra.Command{
	Use:   "file",
	Short: "use a simple json/yaml file as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		err := poll(fc, cmd)
		if err != nil {
			log.Error(err)
		}
	},
}

func init() {
	watchFileCmd.Flags().StringVar(&fc.Filepath, "filepath", "", "The filepath of the yaml/json file")
	pollFileCmd.Flags().StringVar(&fc.Filepath, "filepath", "", "The filepath of the yaml/json file")

	WatchCmd.AddCommand(watchFileCmd)
	PollCmd.AddCommand(pollFileCmd)
}

package advanced

import (
	"os"

	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/log"
	"github.com/spf13/cobra"
)

// Cmd represents the advanced command
var Cmd = &cobra.Command{
	Use:   "advanced",
	Short: "advanced mode - parses the provided config file and process any number of templates",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := cmd.Flags().GetString("config")
		f, _ := file.NewFileClient(cfg)

		loadConf := func() (tomlConf, error) {
			//load the new config
			var c tomlConf
			err := c.fromFile(cfg)
			if err != nil {
				return c, err
			}
			return c, nil
		}

		// we need a working config here - exit on error
		c, err := loadConf()
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
		c.configWatch(f, "", loadConf)
	},
}

func init() {
	Cmd.Flags().String("config", "", "Absolute path to the config file")
}

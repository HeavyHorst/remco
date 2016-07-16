package config

import (
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/log"
	"github.com/spf13/cobra"
)

// Cmd represents the advanced command
var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "load a config file from a file",
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
			log.Fatal(err.Error())
		}
		c.configWatch(f, "", loadConf)
	},
}

func init() {
	FileCmd.Flags().StringP("config", "c", "", "Relative path to the config file")
	CfgCmd.AddCommand(FileCmd)
}

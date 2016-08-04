/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"github.com/HeavyHorst/easyKV/file"
	"github.com/HeavyHorst/remco/log"
	"github.com/spf13/cobra"
)

// FileCmd represents the file command
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

/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/spf13/cobra"
)

var fileConfig = &file.Config{}

func fileConfigRunCMD() func(cmd *cobra.Command, args []string) {
	loadConf := func() (tomlConf, error) {
		var c tomlConf
		err := c.fromFile(fileConfig.Filepath)
		if err != nil {
			return c, err
		}
		return c, nil
	}
	return defaultConfigRunCMD(fileConfig, loadConf)
}

// FileCmd represents the file command
var FileCmd = &cobra.Command{
	Use:   "file",
	Short: "load a config file from a file",
	Run:   fileConfigRunCMD(),
}

func init() {
	FileCmd.Flags().StringVarP(&fileConfig.Filepath, "config", "c", "", "Relative path to the config file")
	CfgCmd.AddCommand(FileCmd)
}

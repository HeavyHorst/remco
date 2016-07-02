// Copyright © 2016 The Remco Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"

	"github.com/HeavyHorst/remco/cmd/advanced"
	"github.com/HeavyHorst/remco/cmd/watch"
)

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "remco",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringP("src", "s", "/etc/remco/default.template", "The absolute path of a configuration template")
	RootCmd.PersistentFlags().StringP("dst", "d", "", "The target file")
	RootCmd.PersistentFlags().StringSliceP("keys", "k", []string{"/"}, "An array of keys")
	RootCmd.PersistentFlags().StringP("fileMode", "m", "0644", "The permission mode of the target file")
	RootCmd.PersistentFlags().StringP("prefix", "p", "/", "The string to prefix to keys")
	RootCmd.PersistentFlags().StringP("reload_cmd", "r", "", "The command to reload the config")
	RootCmd.PersistentFlags().StringP("check_cmd", "c", "", "The command to check the config")
	RootCmd.PersistentFlags().String("log-level", "INFO", "The log Level (DEBUG, INFO, ERROR, ...)")

	RootCmd.AddCommand(watch.WatchCmd, watch.PollCmd, advanced.Cmd)

	cobra.OnInitialize(func() {
		l, _ := RootCmd.Flags().GetString("log-level")
		switch l {
		case "info":
			log.Level = log.LevelInfo
		case "warn":
			log.Level = log.LevelWarning
		case "debug":
			log.Level = log.LevelDebug
		case "error":
			log.Level = log.LevelError
		case "critical":
			log.Level = log.LevelCritical
		}
	})
}

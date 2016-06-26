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

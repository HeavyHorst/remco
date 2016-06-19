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
	"os"

	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/template"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

type fileConfig struct {
	filepath string
}

var fc = fileConfig{}

// Cmd represents the file command
var watchFileCmd = &cobra.Command{
	Use:   "file",
	Short: "A brief description of your command",

	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Filepath set to " + fc.filepath)
		client, err := file.NewFileClient(fc.filepath)
		if err != nil {
			log.Error(err)
		}

		t, err := template.NewTemplateResource(client, "/", cmd.Flags())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		t.Monitor()
	},
}

func init() {
	watchFileCmd.PersistentFlags().StringVar(&fc.filepath, "filepath", "", "The filepath of the yaml/json file")

	Cmd.AddCommand(watchFileCmd)
}

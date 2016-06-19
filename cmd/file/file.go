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

package file

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
)

type fileConfig struct {
	filepath string
	client   backends.StoreClient
}

// Cmd represents the file command
var Cmd = &cobra.Command{
	Use:   "file",
	Short: "A brief description of your command",

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Info("Filepath set to " + config.filepath)
		client, err := file.NewFileClient(config.filepath)
		if err != nil {
			log.Error(err)
		}
		config.client = client
	},
}

var config = fileConfig{}

func init() {
	Cmd.PersistentFlags().StringVar(&config.filepath, "filepath", "", "The filepath of the yaml/json file")
}

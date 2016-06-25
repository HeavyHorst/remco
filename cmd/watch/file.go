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
	"github.com/spf13/pflag"
)

type fileConfig struct {
	filepath string
}

func (f *fileConfig) newTemplateRes(flags *pflag.FlagSet) (*template.TemplateResource, error) {
	log.Info("Filepath set to " + f.filepath)
	client, err := file.NewFileClient(f.filepath)
	if err != nil {
		return nil, err
	}

	return template.NewTemplateResource(client, flags)
}

var fc = fileConfig{}

// Cmd represents the file command
var watchFileCmd = &cobra.Command{
	Use:   "file",
	Short: "use a simple json/yaml file as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		t, err := fc.newTemplateRes(cmd.Flags())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		t.Monitor()
	},
}

// Cmd represents the file command
var pollFileCmd = &cobra.Command{
	Use:   "file",
	Short: "use a simple json/yaml file as the backend source",
	Run: func(cmd *cobra.Command, args []string) {
		t, err := fc.newTemplateRes(cmd.Flags())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		interval, _ := cmd.Flags().GetInt("interval")
		t.Interval(interval)
	},
}

func init() {
	watchFileCmd.Flags().StringVar(&fc.filepath, "filepath", "", "The filepath of the yaml/json file")
	pollFileCmd.Flags().StringVar(&fc.filepath, "filepath", "", "The filepath of the yaml/json file")

	WatchCmd.AddCommand(watchFileCmd)
	PollCmd.AddCommand(pollFileCmd)
}

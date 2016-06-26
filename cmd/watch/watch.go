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
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/template"
	"github.com/spf13/cobra"
)

// WatchCmd represents the watch command
var WatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch a backend for changes and render the template accordingly",
}

// PollCmd represents the watch command
var PollCmd = &cobra.Command{
	Use:   "poll",
	Short: "poll a backend for changes and render the template accordingly",
}

func init() {
	PollCmd.PersistentFlags().IntP("interval", "i", 60, "The backend polling interval in seconds")
	PollCmd.PersistentFlags().Bool("onetime", false, "run once and exit")
}

func watch(bc backends.BackendConfig, cmd *cobra.Command) error {
	client, err := bc.NewClient()
	if err != nil {
		return err
	}

	t, err := template.NewTemplateResource(client, cmd.Flags())
	if err != nil {
		return err
	}
	t.Monitor()

	return nil
}

func poll(bc backends.BackendConfig, cmd *cobra.Command) error {
	client, err := bc.NewClient()
	if err != nil {
		return err
	}

	t, err := template.NewTemplateResource(client, cmd.Flags())
	if err != nil {
		return err
	}
	interval, _ := cmd.Flags().GetInt("interval")
	t.Interval(interval)

	return nil
}

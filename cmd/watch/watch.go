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
	store, err := bc.Connect()
	if err != nil {
		return err
	}

	t, err := template.NewResourceFromFlags(store, cmd.Flags(), true)
	if err != nil {
		return err
	}
	t.Monitor()

	return nil
}

func poll(bc backends.BackendConfig, cmd *cobra.Command) error {
	store, err := bc.Connect()
	if err != nil {
		return err
	}

	t, err := template.NewResourceFromFlags(store, cmd.Flags(), false)
	if err != nil {
		return err
	}
	t.Monitor()

	return nil
}

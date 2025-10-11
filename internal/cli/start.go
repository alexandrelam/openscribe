package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the OpenScribe service",
	Long: `Start OpenScribe and begin listening for hotkey activation.
Once started, press the configured hotkey (default: Right Option) twice to start/stop recording.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start command - Not yet implemented")
		fmt.Println("This will start the OpenScribe service and listen for hotkey activation")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Add flags for the start command
	startCmd.Flags().StringP("microphone", "m", "", "Override microphone selection")
	startCmd.Flags().String("model", "", "Override model selection")
	startCmd.Flags().StringP("language", "l", "", "Override language setting")
	startCmd.Flags().Bool("no-paste", false, "Disable auto-paste")
	startCmd.Flags().BoolP("verbose", "v", false, "Enable verbose debug output")
}

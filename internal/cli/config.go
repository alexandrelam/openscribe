package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `View and modify OpenScribe configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no flags are provided, show help
		if !cmd.Flags().Changed("show") &&
			!cmd.Flags().Changed("list-microphones") &&
			!cmd.Flags().Changed("set-microphone") &&
			!cmd.Flags().Changed("set-model") &&
			!cmd.Flags().Changed("set-language") &&
			!cmd.Flags().Changed("set-hotkey") {
			cmd.Help()
			return
		}

		fmt.Println("Config command - Not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add flags for the config command
	configCmd.Flags().Bool("show", false, "Display current configuration")
	configCmd.Flags().Bool("list-microphones", false, "List available microphones")
	configCmd.Flags().String("set-microphone", "", "Set default microphone")
	configCmd.Flags().String("set-model", "", "Set default model")
	configCmd.Flags().String("set-language", "", "Set default language")
	configCmd.Flags().String("set-hotkey", "", "Configure activation hotkey")
}

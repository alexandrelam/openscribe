package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initial setup for OpenScribe",
	Long: `Download and configure whisper.cpp and the default Whisper model.
This command will:
  - Check for or download whisper.cpp
  - Compile whisper.cpp if needed
  - Download the default small model
  - Create necessary configuration directories`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Setup command - Not yet implemented")
		fmt.Println("This will download whisper.cpp and the default model")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

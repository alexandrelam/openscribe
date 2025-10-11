// Package cli provides the command-line interface for OpenScribe.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openscribe",
	Short: "OpenScribe - Real-time speech transcription with hotkey activation",
	Long: `OpenScribe is a CLI application for macOS that enables real-time speech transcription.
It records audio via a double-press of a configurable button, transcribes the speech using Whisper,
and automatically pastes the transcribed text at the current cursor position.`,
	// Run is not specified as we want subcommands to be required
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
}

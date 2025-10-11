package cli

import (
	"fmt"

	"github.com/alexandrelam/openscribe/internal/logging"
	"github.com/spf13/cobra"
)

var logtestCmd = &cobra.Command{
	Use:    "logtest",
	Short:  "Test logging functionality (for development)",
	Hidden: true, // Hide from main help
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating test transcription logs...")

		testEntries := []struct {
			duration float64
			model    string
			language string
			text     string
		}{
			{3.5, "whisper-small", "en", "Hello, this is a test transcription."},
			{5.2, "whisper-small", "en", "This is the second test entry with more text to demonstrate how logs are displayed."},
			{2.1, "whisper-base", "fr", "Bonjour, ceci est un test en français."},
			{4.7, "whisper-small", "en", "OpenScribe is working great for speech transcription!"},
			{6.3, "whisper-medium", "es", "Esta es una prueba en español para demostrar múltiples idiomas."},
		}

		for i, entry := range testEntries {
			err := logging.LogTranscription(entry.duration, entry.model, entry.language, entry.text)
			if err != nil {
				fmt.Printf("Error logging entry %d: %v\n", i+1, err)
				continue
			}
			fmt.Printf("✓ Logged entry %d\n", i+1)
		}

		fmt.Println("\nTest logs created successfully!")
		fmt.Println("Run 'openscribe logs show' to view them.")
	},
}

func init() {
	rootCmd.AddCommand(logtestCmd)
}

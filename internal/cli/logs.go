package cli

import (
	"fmt"
	"os"

	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/logging"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Log management",
	Long:  `View and manage transcription logs.`,
}

var logsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display recent transcription logs",
	Long:  `Show recent transcription logs from the log file.`,
	Run: func(cmd *cobra.Command, _ []string) {
		tail, _ := cmd.Flags().GetInt("tail")

		// Get transcription entries
		entries, err := logging.GetTranscriptions(tail)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading logs: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Println("No transcription logs found.")
			fmt.Println()
			logPath, _ := config.GetTranscriptionLogPath()
			fmt.Printf("Log file location: %s\n", logPath)
			return
		}

		// Display entries
		fmt.Printf("Showing %d transcription(s):\n\n", len(entries))
		for i, entry := range entries {
			fmt.Printf("─────────────────────────────────────────────────────────────\n")
			fmt.Printf("[%d] %s\n", i+1, entry.Timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("Duration: %.2f seconds | Model: %s | Language: %s\n",
				entry.Duration, entry.Model, entry.Language)
			fmt.Printf("\nTranscription:\n%s\n", entry.Text)
		}
		fmt.Printf("─────────────────────────────────────────────────────────────\n")

		// Show total count
		total, _ := logging.CountTranscriptions()
		if total > len(entries) {
			fmt.Printf("\nShowing %d of %d total transcriptions.\n", len(entries), total)
			fmt.Printf("Use --tail/-n flag to show more: openscribe logs show -n %d\n", total)
		} else {
			fmt.Printf("\nTotal transcriptions: %d\n", total)
		}

		// Show log file location
		logPath, _ := config.GetTranscriptionLogPath()
		fmt.Printf("Log file: %s\n", logPath)
	},
}

var logsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear transcription logs",
	Long:  `Delete all transcription logs.`,
	Run: func(_ *cobra.Command, _ []string) {
		// Get count before clearing
		count, err := logging.CountTranscriptions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading logs: %v\n", err)
			os.Exit(1)
		}

		if count == 0 {
			fmt.Println("No transcription logs to clear.")
			return
		}

		// Clear logs
		if err := logging.ClearTranscriptions(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing logs: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Cleared %d transcription log(s).\n", count)

		// Show log file location
		logPath, _ := config.GetTranscriptionLogPath()
		fmt.Printf("Log file removed: %s\n", logPath)
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.AddCommand(logsShowCmd)
	logsCmd.AddCommand(logsClearCmd)

	// Add flags for logs show command
	logsShowCmd.Flags().IntP("tail", "n", 10, "Show last N transcriptions")
}

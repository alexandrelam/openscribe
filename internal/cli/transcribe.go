package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/models"
	"github.com/alexandrelam/openscribe/internal/transcription"
	"github.com/spf13/cobra"
)

var transcribeCmd = &cobra.Command{
	Use:   "transcribe [audio-file]",
	Short: "Transcribe an audio file (for testing)",
	Long: `Transcribe an audio file using Whisper.

This command is useful for testing transcription without recording.
Provide the path to a WAV audio file (16kHz, mono recommended).`,
	Args: cobra.ExactArgs(1),
	RunE: runTranscribe,
}

var (
	transcribeModel    string
	transcribeLanguage string
	transcribeVerbose  bool
)

func init() {
	transcribeCmd.Flags().StringVarP(&transcribeModel, "model", "m", "small", "Whisper model to use (tiny, base, small, medium, large)")
	transcribeCmd.Flags().StringVarP(&transcribeLanguage, "language", "l", "", "Language code (e.g., en, fr, es). Empty = auto-detect")
	transcribeCmd.Flags().BoolVarP(&transcribeVerbose, "verbose", "v", false, "Enable verbose output from whisper")

	rootCmd.AddCommand(transcribeCmd)
}

func runTranscribe(cmd *cobra.Command, args []string) error {
	audioPath := args[0]

	// Check if file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return fmt.Errorf("audio file not found: %s", audioPath)
	}

	// Parse model
	modelSize, err := models.ParseModelSize(transcribeModel)
	if err != nil {
		return err
	}

	// Check if model is downloaded
	isDownloaded, err := models.IsModelDownloaded(modelSize)
	if err != nil {
		return fmt.Errorf("failed to check if model is downloaded: %w", err)
	}
	if !isDownloaded {
		return fmt.Errorf("model %s is not downloaded. Run 'openscribe models download %s' first", modelSize, modelSize)
	}

	// Get model path for display
	modelPath, _ := models.GetModelPath(modelSize)

	// Display configuration
	fmt.Printf("Transcribing audio file: %s\n", audioPath)
	fmt.Printf("Using model: %s (%s)\n", modelSize, modelPath)
	if transcribeLanguage != "" {
		fmt.Printf("Language: %s\n", transcribeLanguage)
	} else {
		fmt.Printf("Language: auto-detect\n")
	}
	fmt.Println()

	// Create transcriber
	transcriber, err := transcription.NewTranscriber()
	if err != nil {
		return err
	}

	// Prepare options
	opts := transcription.Options{
		Model:    modelSize,
		Language: transcribeLanguage,
		Verbose:  transcribeVerbose,
	}

	// Transcribe
	fmt.Println("Transcribing... (this may take a few seconds)")
	startTime := time.Now()

	result, err := transcriber.TranscribeFile(audioPath, opts)
	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	}

	duration := time.Since(startTime)

	// Display results
	fmt.Println()
	fmt.Println("=== Transcription Result ===")
	fmt.Printf("Text: %s\n", result.Text)
	if result.Language != "" {
		fmt.Printf("Language: %s\n", result.Language)
	}
	fmt.Printf("Processing time: %.2f seconds\n", duration.Seconds())

	// Optionally log the transcription
	cfg, err := config.Load()
	if err == nil && cfg.Verbose {
		fmt.Println()
		if logsDir, err := config.GetLogsDir(); err == nil {
			fmt.Printf("You can view this in logs at: %s\n", logsDir)
		}
	}

	return nil
}

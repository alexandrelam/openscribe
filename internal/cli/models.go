package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/alexandrelam/openscribe/internal/models"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Model management",
	Long:  `List and download Whisper models.`,
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and downloaded models",
	Long:  `Display all available Whisper models and indicate which are downloaded.`,
	Run: func(cmd *cobra.Command, args []string) {
		listModels()
	},
}

var modelsDownloadCmd = &cobra.Command{
	Use:   "download [model]",
	Short: "Download a specific Whisper model",
	Long:  `Download a specific Whisper model (tiny, base, small, medium, large).`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify a model to download (tiny, base, small, medium, large)")
			fmt.Println("\nExample: openscribe models download small")
			return
		}
		downloadModel(args[0])
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsDownloadCmd)
}

func listModels() {
	fmt.Println("Available Whisper Models:")
	fmt.Println()

	// Get list of downloaded models
	downloaded, err := models.ListDownloadedModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking downloaded models: %v\n", err)
		os.Exit(1)
	}

	downloadedMap := make(map[models.ModelSize]bool)
	for _, model := range downloaded {
		downloadedMap[model] = true
	}

	// Display all models in order
	modelOrder := []models.ModelSize{models.Tiny, models.Base, models.Small, models.Medium, models.Large}

	for _, modelName := range modelOrder {
		info := models.AvailableModels[modelName]
		status := " "
		if downloadedMap[modelName] {
			status = "✓"
		}

		fmt.Printf("  [%s] %-8s %s\n", status, info.Name, info.Description)
	}

	fmt.Println()
	fmt.Println("Legend: [✓] Downloaded  [ ] Not downloaded")
	fmt.Println()

	if len(downloaded) == 0 {
		fmt.Println("No models downloaded yet. Run 'openscribe setup' or 'openscribe models download <model>'")
	}
}

func downloadModel(modelName string) {
	// Parse model name
	model, err := models.ParseModelSize(modelName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check if already downloaded
	isDownloaded, err := models.IsModelDownloaded(model)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking model: %v\n", err)
		os.Exit(1)
	}

	if isDownloaded {
		fmt.Printf("Model '%s' is already downloaded.\n", modelName)
		return
	}

	modelInfo := models.AvailableModels[model]
	fmt.Printf("Downloading %s model (%d MB)...\n", modelInfo.Name, modelInfo.SizeMB)
	fmt.Println()

	// Track download progress
	startTime := time.Now()

	progressCallback := func(downloaded, total int64, percent float64) {
		elapsed := time.Since(startTime).Seconds()
		bytesPerSecond := float64(downloaded) / elapsed

		// Calculate progress bar
		barWidth := 40
		filled := int(percent / 100.0 * float64(barWidth))
		bar := ""
		for i := 0; i < barWidth; i++ {
			if i < filled {
				bar += "="
			} else if i == filled {
				bar += ">"
			} else {
				bar += " "
			}
		}

		// Format output
		downloadedStr := models.FormatBytes(downloaded)
		totalStr := models.FormatBytes(total)
		speedStr := models.FormatSpeed(bytesPerSecond)
		eta := models.EstimateTimeRemaining(downloaded, total, bytesPerSecond)

		fmt.Printf("\r[%s] %.1f%% - %s / %s - %s - ETA: %s",
			bar, percent, downloadedStr, totalStr, speedStr, eta)
	}

	// Download the model
	if err := models.DownloadModel(model, progressCallback); err != nil {
		fmt.Fprintf(os.Stderr, "\n\nError downloading model: %v\n", err)
		os.Exit(1)
	}

	fmt.Println() // New line after progress bar
	fmt.Println()
	fmt.Printf("✓ Model '%s' downloaded successfully!\n", modelName)

	// Show the path
	modelPath, _ := models.GetModelPath(model)
	fmt.Printf("  Location: %s\n", modelPath)
}

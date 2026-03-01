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
	Long:  `List and download Whisper and Moonshine models.`,
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and downloaded models",
	Long:  `Display all available models and indicate which are downloaded.`,
	Run: func(cmd *cobra.Command, _ []string) {
		backend, _ := cmd.Flags().GetString("backend")
		if backend == "moonshine" {
			listMoonshineModels()
		} else {
			listModels()
		}
	},
}

var modelsDownloadCmd = &cobra.Command{
	Use:   "download [model]",
	Short: "Download a specific model",
	Long:  `Download a specific model. Use --backend to select whisper (default) or moonshine.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			backend, _ := cmd.Flags().GetString("backend")
			if backend == "moonshine" {
				fmt.Println("Please specify a model to download (tiny, base)")
				fmt.Println("\nExample: openscribe models download --backend moonshine base")
			} else {
				fmt.Println("Please specify a model to download (tiny, base, small, medium, large)")
				fmt.Println("\nExample: openscribe models download small")
			}
			return
		}
		backend, _ := cmd.Flags().GetString("backend")
		if backend == "moonshine" {
			downloadMoonshineModel(args[0])
		} else {
			downloadModel(args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsDownloadCmd)

	// Add --backend flag to subcommands
	modelsListCmd.Flags().String("backend", "whisper", "Backend to list models for (whisper or moonshine)")
	modelsDownloadCmd.Flags().String("backend", "whisper", "Backend to download models for (whisper or moonshine)")
}

func listModels() {
	fmt.Println("Available Whisper Models:")
	fmt.Println()

	downloaded, err := models.ListDownloadedModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking downloaded models: %v\n", err)
		os.Exit(1)
	}

	downloadedMap := make(map[models.ModelSize]bool)
	for _, model := range downloaded {
		downloadedMap[model] = true
	}

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

func listMoonshineModels() {
	fmt.Println("Available Moonshine Models:")
	fmt.Println()

	downloaded, err := models.ListDownloadedMoonshineModels()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking downloaded models: %v\n", err)
		os.Exit(1)
	}

	downloadedMap := make(map[models.MoonshineModelSize]bool)
	for _, model := range downloaded {
		downloadedMap[model] = true
	}

	modelOrder := []models.MoonshineModelSize{models.MoonshineTiny, models.MoonshineBase}

	for _, modelName := range modelOrder {
		info := models.AvailableMoonshineModels[modelName]
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
		fmt.Println("No Moonshine models downloaded. Run 'openscribe models download --backend moonshine <model>'")
	}
}

func downloadModel(modelName string) {
	model, err := models.ParseModelSize(modelName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

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

	startTime := time.Now()

	progressCallback := func(downloaded, total int64, percent float64) {
		elapsed := time.Since(startTime).Seconds()
		bytesPerSecond := float64(downloaded) / elapsed

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

		downloadedStr := models.FormatBytes(downloaded)
		totalStr := models.FormatBytes(total)
		speedStr := models.FormatSpeed(bytesPerSecond)
		eta := models.EstimateTimeRemaining(downloaded, total, bytesPerSecond)

		fmt.Printf("\r[%s] %.1f%% - %s / %s - %s - ETA: %s",
			bar, percent, downloadedStr, totalStr, speedStr, eta)
	}

	if err := models.DownloadModel(model, progressCallback); err != nil {
		fmt.Fprintf(os.Stderr, "\n\nError downloading model: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println()
	fmt.Printf("✓ Model '%s' downloaded successfully!\n", modelName)

	modelPath, _ := models.GetModelPath(model)
	fmt.Printf("  Location: %s\n", modelPath)
}

func downloadMoonshineModel(modelName string) {
	model, err := models.ParseMoonshineModelSize(modelName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	isDownloaded, err := models.IsMoonshineModelDownloaded(model)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking model: %v\n", err)
		os.Exit(1)
	}

	if isDownloaded {
		fmt.Printf("Moonshine model '%s' is already downloaded.\n", modelName)
		return
	}

	info := models.AvailableMoonshineModels[model]
	fmt.Printf("Downloading Moonshine %s model (%d files)...\n", info.Name, len(info.RequiredFiles))
	fmt.Println()

	startTime := time.Now()

	progressCallback := func(downloaded, total int64, percent float64) {
		elapsed := time.Since(startTime).Seconds()
		if elapsed == 0 {
			elapsed = 0.001
		}
		bytesPerSecond := float64(downloaded) / elapsed

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

		speedStr := models.FormatSpeed(bytesPerSecond)
		fmt.Printf("\r[%s] %.1f%% - %s", bar, percent, speedStr)
	}

	if err := models.DownloadMoonshineModel(model, progressCallback); err != nil {
		fmt.Fprintf(os.Stderr, "\n\nError downloading moonshine model: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println()
	fmt.Printf("✓ Moonshine model '%s' downloaded successfully!\n", modelName)

	modelDir, _ := models.GetMoonshineModelDir(model)
	fmt.Printf("  Location: %s\n", modelDir)
}

package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/models"
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
		runSetup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup() {
	fmt.Println("OpenScribe Setup")
	fmt.Println("================")
	fmt.Println()

	// Step 1: Ensure directories exist
	fmt.Println("[1/4] Creating directories...")
	if err := config.EnsureDirectories(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Directories created")
	fmt.Println()

	// Step 2: Check dependencies
	fmt.Println("[2/4] Checking system dependencies...")
	if err := models.CheckDependencies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println()
		fmt.Println("Please install the missing dependencies:")
		fmt.Println("  - On macOS: Install Xcode Command Line Tools")
		fmt.Println("    $ xcode-select --install")
		os.Exit(1)
	}
	fmt.Println("✓ All dependencies found (git, make, C++ compiler)")
	fmt.Println()

	// Step 3: Setup whisper.cpp
	fmt.Println("[3/4] Setting up whisper.cpp...")

	// Check if already installed
	installed, err := models.IsWhisperCppInstalled()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking whisper.cpp: %v\n", err)
		os.Exit(1)
	}

	if installed {
		fmt.Println("✓ whisper.cpp already installed")
		whisperPath, _ := models.GetWhisperCppBinaryPath()
		fmt.Printf("  Location: %s\n", whisperPath)
	} else {
		// Check if we need to download or just compile
		whisperDir, _ := models.GetWhisperCppDir()
		if _, err := os.Stat(whisperDir); os.IsNotExist(err) {
			fmt.Println("  Downloading whisper.cpp from GitHub...")
			if err := models.DownloadWhisperCpp(); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading whisper.cpp: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("  ✓ Downloaded")
		}

		fmt.Println("  Compiling whisper.cpp (this may take a few minutes)...")
		if err := models.CompileWhisperCpp(); err != nil {
			fmt.Fprintf(os.Stderr, "Error compiling whisper.cpp: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ whisper.cpp compiled successfully")
	}
	fmt.Println()

	// Step 4: Download default model
	fmt.Println("[4/4] Downloading default model (small)...")

	defaultModel := models.Small
	isDownloaded, err := models.IsModelDownloaded(defaultModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking model: %v\n", err)
		os.Exit(1)
	}

	if isDownloaded {
		fmt.Println("✓ Model 'small' already downloaded")
		modelPath, _ := models.GetModelPath(defaultModel)
		fmt.Printf("  Location: %s\n", modelPath)
	} else {
		modelInfo := models.AvailableModels[defaultModel]
		fmt.Printf("  Downloading %s model (%d MB)...\n", modelInfo.Name, modelInfo.SizeMB)
		fmt.Println()

		// Progress tracking
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

			fmt.Printf("\r  [%s] %.1f%% - %s / %s - %s - ETA: %s",
				bar, percent, downloadedStr, totalStr, speedStr, eta)
		}

		if err := models.DownloadModel(defaultModel, progressCallback); err != nil {
			fmt.Fprintf(os.Stderr, "\n\nError downloading model: %v\n", err)
			os.Exit(1)
		}

		fmt.Println() // New line after progress bar
		fmt.Println()
		fmt.Println("✓ Model downloaded successfully")
	}

	// Final summary
	fmt.Println()
	fmt.Println("================")
	fmt.Println("Setup Complete!")
	fmt.Println("================")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Configure your microphone (optional):")
	fmt.Println("     $ openscribe config --list-microphones")
	fmt.Println()
	fmt.Println("  2. Start OpenScribe:")
	fmt.Println("     $ openscribe start")
	fmt.Println()
	fmt.Println("  3. View available models:")
	fmt.Println("     $ openscribe models list")
	fmt.Println()
}

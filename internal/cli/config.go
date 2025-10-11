package cli

import (
	"fmt"
	"os"

	"github.com/alexandrelam/openscribe/internal/audio"
	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/hotkey"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `View and modify OpenScribe configuration settings.`,
	Run: func(cmd *cobra.Command, _ []string) {
		// If no flags are provided, show help
		if !cmd.Flags().Changed("show") &&
			!cmd.Flags().Changed("list-microphones") &&
			!cmd.Flags().Changed("list-hotkeys") &&
			!cmd.Flags().Changed("set-microphone") &&
			!cmd.Flags().Changed("set-model") &&
			!cmd.Flags().Changed("set-language") &&
			!cmd.Flags().Changed("set-hotkey") {
			_ = cmd.Help()
			return
		}

		// Handle --show flag
		if cmd.Flags().Changed("show") {
			handleShowConfig()
			return
		}

		// Handle --list-microphones flag
		if cmd.Flags().Changed("list-microphones") {
			handleListMicrophones()
			return
		}

		// Handle --list-hotkeys flag
		if cmd.Flags().Changed("list-hotkeys") {
			handleListHotkeys()
			return
		}

		// Handle set commands
		if cmd.Flags().Changed("set-microphone") {
			value, _ := cmd.Flags().GetString("set-microphone")
			handleSetConfig("microphone", value)
			return
		}

		if cmd.Flags().Changed("set-model") {
			value, _ := cmd.Flags().GetString("set-model")
			handleSetConfig("model", value)
			return
		}

		if cmd.Flags().Changed("set-language") {
			value, _ := cmd.Flags().GetString("set-language")
			handleSetConfig("language", value)
			return
		}

		if cmd.Flags().Changed("set-hotkey") {
			value, _ := cmd.Flags().GetString("set-hotkey")
			handleSetConfig("hotkey", value)
			return
		}
	},
}

func handleShowConfig() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(cfg.String())
}

func handleListMicrophones() {
	fmt.Println("Detecting available microphones...")

	devices, err := audio.ListMicrophones()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing microphones: %v\n", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		fmt.Println("No microphones found.")
		return
	}

	fmt.Println("\nAvailable microphones:")
	for i, device := range devices {
		defaultMarker := ""
		if device.IsDefault {
			defaultMarker = " (default)"
		}
		fmt.Printf("  %d. %s%s\n", i+1, device.Name, defaultMarker)
	}

	fmt.Println("\nTo set a microphone, use either the number or name:")
	fmt.Println("  openscribe config --set-microphone 1")
	fmt.Println("  openscribe config --set-microphone \"<microphone name>\"")
}

func handleListHotkeys() {
	fmt.Println("Available hotkeys:")

	keys := hotkey.GetAvailableKeys()
	for i, key := range keys {
		fmt.Printf("  %d. %s\n", i+1, key)
	}

	fmt.Println("\nTo set a hotkey, use:")
	fmt.Println("  openscribe config --set-hotkey \"Right Option\"")
}

func handleSetConfig(key, value string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Update the appropriate field
	switch key {
	case "microphone":
		// Validate that the microphone exists
		if value != "" {
			device, err := audio.FindMicrophoneByNameOrIndex(value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Println("\nRun 'openscribe config --list-microphones' to see available devices.")
				os.Exit(1)
			}
			// Store the actual device name (not the index)
			cfg.Microphone = device.Name
			fmt.Printf("Microphone set to: %s\n", device.Name)
		} else {
			cfg.Microphone = value
			fmt.Println("Microphone set to: (system default)")
		}
	case "model":
		cfg.Model = value
		fmt.Printf("Model set to: %s\n", value)
	case "language":
		cfg.Language = value
		if value == "" {
			fmt.Println("Language set to: auto-detect")
		} else {
			fmt.Printf("Language set to: %s\n", value)
		}
	case "hotkey":
		// Validate hotkey
		if err := hotkey.ValidateKeyName(value); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Println("\nRun 'openscribe config --list-hotkeys' to see available hotkeys.")
			os.Exit(1)
		}
		cfg.Hotkey = value
		fmt.Printf("Hotkey set to: %s\n", value)
	}

	// Validate before saving
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Save the updated config
	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully!")
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add flags for the config command
	configCmd.Flags().Bool("show", false, "Display current configuration")
	configCmd.Flags().Bool("list-microphones", false, "List available microphones")
	configCmd.Flags().Bool("list-hotkeys", false, "List available hotkeys")
	configCmd.Flags().String("set-microphone", "", "Set default microphone")
	configCmd.Flags().String("set-model", "", "Set default model")
	configCmd.Flags().String("set-language", "", "Set default language")
	configCmd.Flags().String("set-hotkey", "", "Configure activation hotkey")
}

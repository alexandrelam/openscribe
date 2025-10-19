package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

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
			!cmd.Flags().Changed("open") &&
			!cmd.Flags().Changed("list-microphones") &&
			!cmd.Flags().Changed("list-hotkeys") &&
			!cmd.Flags().Changed("list-sounds") &&
			!cmd.Flags().Changed("test-sounds") &&
			!cmd.Flags().Changed("set-microphone") &&
			!cmd.Flags().Changed("set-model") &&
			!cmd.Flags().Changed("set-language") &&
			!cmd.Flags().Changed("set-hotkey") &&
			!cmd.Flags().Changed("enable-audio-feedback") &&
			!cmd.Flags().Changed("disable-audio-feedback") &&
			!cmd.Flags().Changed("show-preferences") &&
			!cmd.Flags().Changed("add-preference") &&
			!cmd.Flags().Changed("remove-preference") &&
			!cmd.Flags().Changed("clear-preferences") {
			_ = cmd.Help()
			return
		}

		// Handle --open flag
		if cmd.Flags().Changed("open") {
			handleOpenConfig()
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

		// Handle --list-sounds flag
		if cmd.Flags().Changed("list-sounds") {
			handleListSounds()
			return
		}

		// Handle --test-sounds flag
		if cmd.Flags().Changed("test-sounds") {
			handleTestSounds()
			return
		}

		// Handle audio feedback enable/disable
		if cmd.Flags().Changed("enable-audio-feedback") {
			handleSetAudioFeedback(true)
			return
		}

		if cmd.Flags().Changed("disable-audio-feedback") {
			handleSetAudioFeedback(false)
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

		// Handle preference management commands
		if cmd.Flags().Changed("show-preferences") {
			handleShowPreferences()
			return
		}

		if cmd.Flags().Changed("add-preference") {
			value, _ := cmd.Flags().GetString("add-preference")
			handleAddPreference(value)
			return
		}

		if cmd.Flags().Changed("remove-preference") {
			value, _ := cmd.Flags().GetString("remove-preference")
			handleRemovePreference(value)
			return
		}

		if cmd.Flags().Changed("clear-preferences") {
			handleClearPreferences()
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

func handleOpenConfig() {
	// Ensure config exists (this will create it with defaults if it doesn't exist)
	_, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Get the config file path
	configPath, err := config.GetConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting config path: %v\n", err)
		os.Exit(1)
	}

	// Open the config file with the default editor using macOS 'open' command
	fmt.Printf("Opening config file: %s\n", configPath)
	cmd := exec.Command("open", configPath)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening config file: %v\n", err)
		fmt.Fprintf(os.Stderr, "You can manually open the file at: %s\n", configPath)
		os.Exit(1)
	}

	fmt.Println("Config file opened in default editor.")
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

	fmt.Println("\nTo set preferences:")
	fmt.Println("  openscribe config --add-preference \"<microphone name>\"")
	fmt.Println("  openscribe config --show-preferences")
	fmt.Println("\nTo set a single microphone (legacy):")
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

func handleListSounds() {
	sounds := audio.ListSystemSounds()
	fmt.Println("Available macOS system sounds:")
	for i, sound := range sounds {
		fmt.Printf("  %d. %s\n", i+1, sound)
	}
	fmt.Println("\nOpenScribe uses the following sounds by default:")
	fmt.Println("  - Start recording: Tink (short ascending beep)")
	fmt.Println("  - Stop recording: Pop (short neutral beep)")
	fmt.Println("  - Transcription complete: Glass (pleasant ding)")
	fmt.Println("\nTo test the sounds, run:")
	fmt.Println("  openscribe config --test-sounds")
}

func handleTestSounds() {
	fmt.Println("Testing audio feedback sounds...")

	feedback, err := audio.NewFeedback()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing audio feedback: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := feedback.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close audio feedback: %v\n", err)
		}
	}()

	fmt.Println("Playing start sound (Tink)...")
	if err := feedback.PlayStartSound(); err != nil {
		fmt.Fprintf(os.Stderr, "Error playing start sound: %v\n", err)
	}

	// Brief pause between sounds
	fmt.Println("Waiting 1 second...")
	_ = os.Stdout.Sync()
	// Use a simple busy-wait for demo purposes
	for i := 0; i < 100000000; i++ {
		_ = i // prevent empty block warning
	}

	fmt.Println("Playing stop sound (Pop)...")
	if err := feedback.PlayStopSound(); err != nil {
		fmt.Fprintf(os.Stderr, "Error playing stop sound: %v\n", err)
	}

	fmt.Println("Waiting 1 second...")
	_ = os.Stdout.Sync()
	for i := 0; i < 100000000; i++ {
		_ = i // prevent empty block warning
	}

	fmt.Println("Playing complete sound (Glass)...")
	if err := feedback.PlayCompleteSound(); err != nil {
		fmt.Fprintf(os.Stderr, "Error playing complete sound: %v\n", err)
	}

	fmt.Println("\nAudio feedback test complete!")
}

func handleSetAudioFeedback(enabled bool) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	cfg.AudioFeedback = enabled

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	if enabled {
		fmt.Println("Audio feedback enabled!")
	} else {
		fmt.Println("Audio feedback disabled.")
	}
	fmt.Println("Configuration saved successfully!")
}

func handleShowPreferences() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.PreferredMicrophones) == 0 {
		fmt.Println("No preferred microphones configured.")
		fmt.Println("\nUsing: System default microphone")
		fmt.Println("\nTo add preferences:")
		fmt.Println("  openscribe config --add-preference \"<microphone name>\"")
		return
	}

	fmt.Println("Preferred Microphones (in priority order):")
	for i, mic := range cfg.PreferredMicrophones {
		fmt.Printf("  %d. %s\n", i+1, mic)
	}
	fmt.Println("\nFallback: System default microphone")
}

func handleAddPreference(name string) {
	// Validate input
	if name == "" {
		fmt.Fprintf(os.Stderr, "Error: Microphone name cannot be empty\n")
		fmt.Println("\nUsage:")
		fmt.Println("  openscribe config --add-preference \"<microphone name>\"")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Check for duplicates (case-insensitive)
	for i, existingMic := range cfg.PreferredMicrophones {
		if existingMic == name {
			fmt.Fprintf(os.Stderr, "Error: \"%s\" is already in your preferences (priority %d)\n", name, i+1)
			os.Exit(1)
		}
	}

	// Check if device is currently connected (warning, not error)
	devices, err := audio.ListMicrophones()
	if err == nil {
		found := false
		for _, device := range devices {
			if device.Name == name {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("⚠ Warning: \"%s\" is not currently connected\n", name)
		}
	}

	// Add to preferences
	cfg.PreferredMicrophones = append(cfg.PreferredMicrophones, name)

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

	fmt.Printf("✓ Added \"%s\" to preferred microphones (priority %d)\n", name, len(cfg.PreferredMicrophones))
	fmt.Println("Configuration saved successfully!")
}

func handleRemovePreference(nameOrIndex string) {
	if nameOrIndex == "" {
		fmt.Fprintf(os.Stderr, "Error: Must provide microphone name or index\n")
		fmt.Println("\nUsage:")
		fmt.Println("  openscribe config --remove-preference \"<microphone name>\"")
		fmt.Println("  openscribe config --remove-preference <index>")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.PreferredMicrophones) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No preferred microphones configured\n")
		os.Exit(1)
	}

	// Try to find and remove the preference
	removedName := ""
	newPreferences := []string{}

	// Try parsing as index first
	if idx, parseErr := strconv.Atoi(nameOrIndex); parseErr == nil {
		// It's a number - validate index range
		if idx < 1 || idx > len(cfg.PreferredMicrophones) {
			fmt.Fprintf(os.Stderr, "Error: Invalid index: %d (valid range: 1-%d)\n", idx, len(cfg.PreferredMicrophones))
			os.Exit(1)
		}
		// Remove by index (convert from 1-based to 0-based)
		removedName = cfg.PreferredMicrophones[idx-1]
		for i, mic := range cfg.PreferredMicrophones {
			if i != idx-1 {
				newPreferences = append(newPreferences, mic)
			}
		}
	} else {
		// It's a name - find and remove (case-sensitive match)
		found := false
		for _, mic := range cfg.PreferredMicrophones {
			if mic == nameOrIndex {
				removedName = mic
				found = true
			} else {
				newPreferences = append(newPreferences, mic)
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Error: \"%s\" not found in preferences\n", nameOrIndex)
			os.Exit(1)
		}
	}

	cfg.PreferredMicrophones = newPreferences

	// Save the updated config
	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Removed \"%s\" from preferred microphones\n", removedName)
	fmt.Println("Configuration saved successfully!")
}

func handleClearPreferences() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	cfg.PreferredMicrophones = []string{}

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Cleared all preferred microphones")
	fmt.Println("  Will now use system default microphone")
	fmt.Println("Configuration saved successfully!")
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add flags for the config command
	configCmd.Flags().Bool("show", false, "Display current configuration")
	configCmd.Flags().Bool("open", false, "Open configuration file in default editor")
	configCmd.Flags().Bool("list-microphones", false, "List available microphones")
	configCmd.Flags().Bool("list-hotkeys", false, "List available hotkeys")
	configCmd.Flags().Bool("list-sounds", false, "List available system sounds")
	configCmd.Flags().Bool("test-sounds", false, "Test audio feedback sounds")
	configCmd.Flags().Bool("enable-audio-feedback", false, "Enable audio feedback")
	configCmd.Flags().Bool("disable-audio-feedback", false, "Disable audio feedback")
	configCmd.Flags().String("set-microphone", "", "Set default microphone")
	configCmd.Flags().String("set-model", "", "Set default model")
	configCmd.Flags().String("set-language", "", "Set default language")
	configCmd.Flags().String("set-hotkey", "", "Configure activation hotkey")

	// Add flags for preference management
	configCmd.Flags().Bool("show-preferences", false, "Show current preferred microphones list")
	configCmd.Flags().String("add-preference", "", "Add a microphone to the preferences list")
	configCmd.Flags().String("remove-preference", "", "Remove a microphone from preferences (by name or index)")
	configCmd.Flags().Bool("clear-preferences", false, "Clear all preferred microphones")
}

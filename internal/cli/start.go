package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/hotkey"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the OpenScribe service",
	Long: `Start OpenScribe and begin listening for hotkey activation.
Once started, press the configured hotkey (default: Right Option) twice to start/stop recording.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStart(cmd)
	},
}

func runStart(cmd *cobra.Command) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Apply command-line overrides
	if cmd.Flags().Changed("microphone") {
		cfg.Microphone, _ = cmd.Flags().GetString("microphone")
	}
	if cmd.Flags().Changed("model") {
		cfg.Model, _ = cmd.Flags().GetString("model")
	}
	if cmd.Flags().Changed("language") {
		cfg.Language, _ = cmd.Flags().GetString("language")
	}
	if cmd.Flags().Changed("no-paste") {
		noPaste, _ := cmd.Flags().GetBool("no-paste")
		cfg.AutoPaste = !noPaste
	}
	if cmd.Flags().Changed("verbose") {
		cfg.Verbose, _ = cmd.Flags().GetBool("verbose")
	}

	// Display current configuration
	microphone := cfg.Microphone
	if microphone == "" {
		microphone = "(system default)"
	}
	language := cfg.Language
	if language == "" {
		language = "auto-detect"
	}

	fmt.Println("OpenScribe Starting...")
	fmt.Printf("  Microphone:      %s\n", microphone)
	fmt.Printf("  Model:           %s\n", cfg.Model)
	fmt.Printf("  Language:        %s\n", language)
	fmt.Printf("  Hotkey:          %s (double-press)\n", cfg.Hotkey)
	fmt.Printf("  Auto-paste:      %t\n", cfg.AutoPaste)
	fmt.Printf("  Audio Feedback:  %t\n", cfg.AudioFeedback)
	fmt.Println()

	// State management
	var (
		mu         sync.Mutex
		isRecording bool
	)

	// Create hotkey callback
	hotkeyCallback := func() {
		mu.Lock()
		defer mu.Unlock()

		if !isRecording {
			// Start recording
			isRecording = true
			fmt.Println("🔴 Recording started... (double-press hotkey again to stop)")
			// TODO: Start actual audio recording in Phase 10
		} else {
			// Stop recording
			isRecording = false
			fmt.Println("⏹  Recording stopped. Transcribing...")
			// TODO: Stop recording and transcribe in Phase 10
			fmt.Println("✅ Transcription complete!")
		}
	}

	// Create and start hotkey listener
	listener, err := hotkey.NewListener(cfg.Hotkey, hotkeyCallback)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating hotkey listener: %v\n", err)
		os.Exit(1)
	}

	if err := listener.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting hotkey listener: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nNote: Hotkey detection requires accessibility permissions.\n")
		fmt.Fprintf(os.Stderr, "Please grant accessibility permissions in System Preferences > Security & Privacy > Privacy > Accessibility\n")
		os.Exit(1)
	}
	defer listener.Stop()

	fmt.Println("Ready! Press hotkey to start recording...")
	fmt.Println("Press Ctrl+C to exit.")
	fmt.Println()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan

	fmt.Println("\n\nShutting down...")
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Add flags for the start command
	startCmd.Flags().StringP("microphone", "m", "", "Override microphone selection")
	startCmd.Flags().String("model", "", "Override model selection")
	startCmd.Flags().StringP("language", "l", "", "Override language setting")
	startCmd.Flags().Bool("no-paste", false, "Disable auto-paste")
	startCmd.Flags().BoolP("verbose", "v", false, "Enable verbose debug output")
}

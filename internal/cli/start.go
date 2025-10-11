package cli

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexandrelam/openscribe/internal/audio"
	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/hotkey"
	"github.com/alexandrelam/openscribe/internal/keyboard"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the OpenScribe service",
	Long: `Start OpenScribe and begin listening for hotkey activation.
Once started, press the configured hotkey (default: Right Option) twice to start/stop recording.`,
	Run: func(cmd *cobra.Command, _ []string) {
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

	// Initialize audio feedback if enabled
	var feedback audio.Feedback
	if cfg.AudioFeedback {
		var err error
		feedback, err = audio.NewFeedback()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize audio feedback: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing without audio feedback...\n\n")
		} else {
			defer func() {
				if err := feedback.Close(); err != nil && cfg.Verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to close audio feedback: %v\n", err)
				}
			}()
		}
	}

	// Initialize keyboard simulation if auto-paste is enabled
	var kb keyboard.Keyboard
	if cfg.AutoPaste {
		var err error
		kb, err = keyboard.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize keyboard simulation: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := kb.Close(); err != nil && cfg.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: Failed to close keyboard: %v\n", err)
			}
		}()

		// Check accessibility permissions
		if err := kb.CheckPermissions(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Accessibility permissions not granted.\n\n")
			fmt.Fprintf(os.Stderr, "Auto-paste requires accessibility permissions to simulate keyboard input.\n")
			fmt.Fprintf(os.Stderr, "Please grant permissions in:\n")
			fmt.Fprintf(os.Stderr, "  System Preferences > Security & Privacy > Privacy > Accessibility\n\n")
			fmt.Fprintf(os.Stderr, "Add 'Terminal' (or your terminal app) to the list of allowed applications.\n\n")
			fmt.Fprintf(os.Stderr, "Alternatively, run with --no-paste to disable auto-paste:\n")
			fmt.Fprintf(os.Stderr, "  openscribe start --no-paste\n\n")

			// Prompt user to grant permissions
			fmt.Fprintf(os.Stderr, "Would you like to open System Preferences now? This will prompt for permissions.\n")
			fmt.Fprintf(os.Stderr, "After granting permissions, please restart OpenScribe.\n\n")
			keyboard.RequestPermissions()
			os.Exit(1)
		}
	}

	// State management
	var (
		mu          sync.Mutex
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

			// Play start sound
			if feedback != nil {
				if err := feedback.PlayStartSound(); err != nil && cfg.Verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to play start sound: %v\n", err)
				}
			}

			// TODO: Start actual audio recording in Phase 10
		} else {
			// Stop recording
			isRecording = false
			fmt.Println("⏹  Recording stopped. Transcribing...")

			// Play stop sound
			if feedback != nil {
				if err := feedback.PlayStopSound(); err != nil && cfg.Verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to play stop sound: %v\n", err)
				}
			}

			// TODO: Stop recording and transcribe in Phase 10
			// For now, we'll use a sample transcription for testing
			sampleTranscription := "Hello, this is a test of the OpenScribe application."

			// Play complete sound when transcription is done
			if feedback != nil {
				if err := feedback.PlayCompleteSound(); err != nil && cfg.Verbose {
					fmt.Fprintf(os.Stderr, "Warning: Failed to play complete sound: %v\n", err)
				}
			}

			fmt.Printf("Transcription: \"%s\"\n", sampleTranscription)

			// Auto-paste if enabled
			if cfg.AutoPaste && kb != nil {
				if err := kb.TypeText(sampleTranscription); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to paste text: %v\n", err)
				} else {
					fmt.Println("✅ Text pasted to cursor position!")
				}
			} else {
				fmt.Println("✅ Transcription complete!")
			}

			// TODO: Log transcription in Phase 10
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

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alexandrelam/openscribe/internal/audio"
	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/spf13/cobra"
)

var audioTestCmd = &cobra.Command{
	Use:    "audio-test",
	Short:  "Test audio recording (records 5 seconds)",
	Long:   `Test command to verify microphone and audio recording functionality. Records 5 seconds of audio and saves it to the cache directory.`,
	Hidden: true, // Hidden command for testing purposes
	Run: func(cmd *cobra.Command, args []string) {
		duration, _ := cmd.Flags().GetInt("duration")
		runAudioTest(duration)
	},
}

func runAudioTest(durationSeconds int) {
	fmt.Println("Audio Recording Test")
	fmt.Println("====================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Determine which microphone to use
	micName := cfg.Microphone
	if micName == "" {
		fmt.Println("No microphone configured, using system default...")
		device, err := audio.GetDefaultMicrophone()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting default microphone: %v\n", err)
			os.Exit(1)
		}
		micName = device.Name
	}

	fmt.Printf("Microphone: %s\n", micName)
	fmt.Printf("Duration: %d seconds\n", durationSeconds)
	fmt.Printf("Sample Rate: 16000 Hz (Whisper-compatible)\n")
	fmt.Printf("Channels: 1 (mono)\n\n")

	// Create recorder
	recorder := audio.NewRecorder(micName)

	// Start recording
	fmt.Printf("Starting recording...\n")
	if err := recorder.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting recording: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Recording for %d seconds...\n", durationSeconds)

	// Show progress
	for i := 0; i < durationSeconds; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("  %d/%d seconds\n", i+1, durationSeconds)
	}

	// Stop recording
	fmt.Printf("\nStopping recording...\n")
	audioData, err := recorder.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping recording: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Captured %d bytes of audio data\n", len(audioData))

	// Save to cache directory
	cacheDir, err := config.GetCacheDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting cache directory: %v\n", err)
		os.Exit(1)
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("test-recording-%s.wav", time.Now().Format("20060102-150405"))
	filepath := filepath.Join(cacheDir, filename)

	fmt.Printf("Saving to: %s\n", filepath)
	err = audio.SaveWAV(filepath, audioData, recorder.GetSampleRate(), recorder.GetChannels())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving WAV file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nRecording test completed successfully!\n")
	fmt.Printf("Audio file saved to: %s\n", filepath)
	fmt.Printf("\nYou can play this file to verify the recording:\n")
	fmt.Printf("  afplay %s\n", filepath)
}

func init() {
	rootCmd.AddCommand(audioTestCmd)

	// Add flags
	audioTestCmd.Flags().IntP("duration", "d", 5, "Recording duration in seconds")
}

// Package audio provides audio recording, device management, and feedback functionality for OpenScribe.
//
// This package handles:
//   - Audio input device enumeration and selection
//   - Audio recording with configurable parameters (sample rate, channels, format)
//   - WAV file generation from recorded audio data
//   - Audio feedback (system sounds) for user notifications
//
// The audio recording is implemented using the malgo library which provides
// cross-platform audio I/O capabilities. On macOS, audio feedback uses NSSound
// to play system sounds.
//
// Example usage:
//
//	// List available microphones
//	devices, err := audio.ListDevices()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create recorder with default settings
//	recorder, err := audio.NewRecorder(devices[0].ID)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer recorder.Close()
//
//	// Record audio
//	if err := recorder.Start(); err != nil {
//	    log.Fatal(err)
//	}
//	time.Sleep(5 * time.Second)
//	data, err := recorder.Stop()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Save to WAV file
//	if err := audio.SaveWAV("output.wav", data, 16000, 1, 16); err != nil {
//	    log.Fatal(err)
//	}
package audio

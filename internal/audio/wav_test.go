package audio

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadWAV(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.wav")

	// Create test audio data (simple sine-like pattern)
	sampleRate := uint32(16000)
	channels := uint32(1)
	duration := 0.5 // 0.5 seconds
	numSamples := int(float64(sampleRate) * duration)

	// Generate some test audio data (alternating high/low values)
	audioData := make([]byte, numSamples*2) // 2 bytes per sample (16-bit)
	for i := 0; i < numSamples; i++ {
		// Create a simple alternating pattern
		value := int16(i % 2 * 10000)
		audioData[i*2] = byte(value & 0xFF)
		audioData[i*2+1] = byte((value >> 8) & 0xFF)
	}

	// Save the WAV file
	err := SaveWAV(testFile, audioData, sampleRate, channels)
	if err != nil {
		t.Fatalf("Failed to save WAV file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("WAV file was not created")
	}

	// Load the WAV file
	loadedData, loadedRate, loadedChannels, err := LoadWAV(testFile)
	if err != nil {
		t.Fatalf("Failed to load WAV file: %v", err)
	}

	// Verify loaded data matches original
	if loadedRate != sampleRate {
		t.Errorf("Expected sample rate %d, got %d", sampleRate, loadedRate)
	}
	if loadedChannels != channels {
		t.Errorf("Expected %d channels, got %d", channels, loadedChannels)
	}
	if !bytes.Equal(loadedData, audioData) {
		t.Error("Loaded audio data does not match original")
	}
}

func TestLoadWAV_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.wav")

	// Create a file with invalid content
	err := os.WriteFile(invalidFile, []byte("not a wav file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	// Try to load it
	_, _, _, err = LoadWAV(invalidFile)
	if err == nil {
		t.Error("Expected error when loading invalid WAV file, got nil")
	}
}

func TestLoadWAV_NonExistentFile(t *testing.T) {
	_, _, _, err := LoadWAV("/nonexistent/file.wav")
	if err == nil {
		t.Error("Expected error when loading non-existent file, got nil")
	}
}

func TestSaveWAV_ValidFormats(t *testing.T) {
	testCases := []struct {
		name       string
		sampleRate uint32
		channels   uint32
		dataSize   int
	}{
		{"16kHz Mono", 16000, 1, 16000 * 2},     // 1 second of 16-bit audio
		{"44.1kHz Stereo", 44100, 2, 44100 * 4}, // 1 second of 16-bit stereo
		{"8kHz Mono", 8000, 1, 8000 * 2},        // 1 second of 16-bit audio
	}

	tmpDir := t.TempDir()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tc.name+".wav")
			audioData := make([]byte, tc.dataSize)

			// Fill with test data
			for i := range audioData {
				audioData[i] = byte(i % 256)
			}

			err := SaveWAV(testFile, audioData, tc.sampleRate, tc.channels)
			if err != nil {
				t.Fatalf("Failed to save WAV file: %v", err)
			}

			// Load and verify
			loadedData, loadedRate, loadedChannels, err := LoadWAV(testFile)
			if err != nil {
				t.Fatalf("Failed to load WAV file: %v", err)
			}

			if loadedRate != tc.sampleRate {
				t.Errorf("Sample rate mismatch: expected %d, got %d", tc.sampleRate, loadedRate)
			}
			if loadedChannels != tc.channels {
				t.Errorf("Channels mismatch: expected %d, got %d", tc.channels, loadedChannels)
			}
			if !bytes.Equal(loadedData, audioData) {
				t.Error("Audio data mismatch")
			}
		})
	}
}

//go:build integration
// +build integration

package transcription

import (
	"testing"

	"github.com/alexandrelam/openscribe/internal/models"
)

// Integration tests for transcription with real audio files and whisper-cli
// These tests are NOT run in CI by default to avoid slow setup times
//
// To run these tests locally:
//   1. Install whisper-cpp: brew install whisper-cpp
//   2. Download a model: openscribe models download tiny
//   3. Run tests: go test -tags=integration ./internal/transcription/...

func TestTranscribeFile_Integration(t *testing.T) {
	// Verify whisper-cli is available
	transcriber, err := NewTranscriber()
	if err != nil {
		t.Fatalf("whisper-cli not installed: %v\nInstall with: brew install whisper-cpp", err)
	}

	// Verify tiny model is downloaded
	isDownloaded, err := models.IsModelDownloaded(models.Tiny)
	if err != nil || !isDownloaded {
		t.Fatalf("tiny model not downloaded\nDownload with: openscribe models download tiny")
	}

	// Test transcription with the test audio file
	opts := Options{
		Model:    models.Tiny, // Use tiny model for faster testing
		Language: "en",
		Verbose:  false,
	}

	result, err := transcriber.TranscribeFile("testdata/test-english.wav", opts)
	if err != nil {
		t.Fatalf("TranscribeFile() failed: %v", err)
	}

	// Check that we got some transcription
	if result.Text == "" {
		t.Error("TranscribeFile() returned empty text")
	}

	// Check that language was set
	if result.Language != "en" {
		t.Errorf("TranscribeFile() language = %q, want %q", result.Language, "en")
	}

	t.Logf("Transcription result: %q", result.Text)
}

func TestTranscribeFile_AutoDetectLanguage(t *testing.T) {
	transcriber, err := NewTranscriber()
	if err != nil {
		t.Skipf("whisper-cli not installed: %v", err)
	}

	isDownloaded, err := models.IsModelDownloaded(models.Tiny)
	if err != nil || !isDownloaded {
		t.Skipf("tiny model not downloaded")
	}

	// Test with auto-detect language
	opts := Options{
		Model:    models.Tiny,
		Language: "", // auto-detect
		Verbose:  false,
	}

	result, err := transcriber.TranscribeFile("testdata/test-english.wav", opts)
	if err != nil {
		t.Fatalf("TranscribeFile() with auto-detect failed: %v", err)
	}

	if result.Text == "" {
		t.Error("TranscribeFile() returned empty text")
	}

	t.Logf("Auto-detected language: %q, Text: %q", result.Language, result.Text)
}

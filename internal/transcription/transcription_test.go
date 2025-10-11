package transcription

import (
	"testing"

	"github.com/alexandrelam/openscribe/internal/models"
)

func TestParseOutput(t *testing.T) {
	transcriber := &Transcriber{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple text",
			input:    "Hello, this is a test.",
			expected: "Hello, this is a test.",
		},
		{
			name: "Text with timestamps",
			input: `[00:00:00.000 --> 00:00:02.000]  Hello, this is a test.
[00:00:02.000 --> 00:00:04.000]  This is another line.`,
			expected: "Hello, this is a test. This is another line.",
		},
		{
			name: "Text with metadata",
			input: `Detected language: en
Processing audio...
Hello, this is a test.`,
			expected: "Hello, this is a test.",
		},
		{
			name: "Multi-line text",
			input: `First line.
Second line.
Third line.`,
			expected: "First line. Second line. Third line.",
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
		{
			name: "Only metadata",
			input: `[00:00:00.000 --> 00:00:02.000]
[Processing...]`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transcriber.parseOutput(tt.input)
			if result != tt.expected {
				t.Errorf("parseOutput() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractLanguage(t *testing.T) {
	transcriber := &Transcriber{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "English detection",
			input:    "Detected language: en",
			expected: "en",
		},
		{
			name:     "French detection",
			input:    "Detected language: fr (probability: 0.98)",
			expected: "fr",
		},
		{
			name:     "Spanish detection with extra info",
			input:    "some text\nDetected language: es\nmore text",
			expected: "es",
		},
		{
			name:     "No language detection",
			input:    "Just some text without language info",
			expected: "",
		},
		{
			name:     "Case insensitive",
			input:    "DETECTED LANGUAGE: de",
			expected: "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transcriber.extractLanguage(tt.input)
			if result != tt.expected {
				t.Errorf("extractLanguage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Model != models.Small {
		t.Errorf("DefaultOptions().Model = %v, want %v", opts.Model, models.Small)
	}

	if opts.Language != "" {
		t.Errorf("DefaultOptions().Language = %q, want empty string", opts.Language)
	}

	if opts.Verbose {
		t.Errorf("DefaultOptions().Verbose = true, want false")
	}
}

func TestNewTranscriber(t *testing.T) {
	// This test checks if whisper-cli can be found
	// It will skip gracefully if not installed (e.g., in CI)
	transcriber, err := NewTranscriber()

	// If whisper-cli is not installed, we expect an error
	if err != nil {
		if transcriber != nil {
			t.Error("NewTranscriber() returned both error and non-nil transcriber")
		}
		// This is expected if whisper-cli is not installed (e.g., in CI)
		t.Skipf("whisper-cli not installed, skipping test: %v", err)
		return
	}

	// If no error, transcriber should be valid
	if transcriber == nil {
		t.Fatal("NewTranscriber() returned nil transcriber with no error")
	}

	if transcriber.whisperPath == "" {
		t.Error("NewTranscriber() created transcriber with empty whisperPath")
	}
}

// Note: Integration tests with real audio files are in integration_test.go
// Run them with: go test -tags=integration ./internal/transcription/...

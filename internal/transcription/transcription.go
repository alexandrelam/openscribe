// Package transcription provides speech-to-text transcription backends.
package transcription

import (
	"fmt"

	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/models"
)

// Transcriber is the interface for speech-to-text backends.
type Transcriber interface {
	TranscribeFile(audioPath string, opts Options) (*Result, error)
}

// Options contains options for transcription
type Options struct {
	// Model is the Whisper model to use (tiny, base, small, medium, large)
	Model models.ModelSize

	// Language is the target language code (e.g., "en", "fr", "es")
	// Empty string means auto-detect
	Language string

	// Verbose enables detailed output
	Verbose bool
}

// Result contains the transcription result and metadata
type Result struct {
	// Text is the transcribed text
	Text string

	// Language is the detected or specified language
	Language string

	// Duration is the audio duration in seconds (if available)
	Duration float64
}

// DefaultOptions returns default transcription options
func DefaultOptions() Options {
	return Options{
		Model:    models.Small,
		Language: "", // auto-detect
		Verbose:  false,
	}
}

// New creates a Transcriber based on the configured backend.
// For "moonshine" backend, build with -tags moonshine.
func New(cfg *config.Config) (Transcriber, error) {
	switch cfg.Backend {
	case "", "whisper":
		return NewWhisperTranscriber()
	case "moonshine":
		return newMoonshineTranscriber(cfg)
	case "openai":
		return NewOpenAITranscriber(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	default:
		return nil, fmt.Errorf("unknown transcription backend: %s", cfg.Backend)
	}
}

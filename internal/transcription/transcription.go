package transcription

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/alexandrelam/openscribe/internal/models"
)

// Options contains options for transcription
type Options struct {
	// Model is the Whisper model to use (tiny, base, small, medium, large)
	Model models.ModelSize

	// Language is the target language code (e.g., "en", "fr", "es")
	// Empty string means auto-detect
	Language string

	// Verbose enables detailed output from whisper
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

// Transcriber handles speech-to-text transcription using whisper.cpp
type Transcriber struct {
	whisperPath string
}

// NewTranscriber creates a new transcriber
func NewTranscriber() (*Transcriber, error) {
	// Check if whisper-cli is available
	whisperPath, err := exec.LookPath("whisper-cli")
	if err != nil {
		return nil, fmt.Errorf("whisper-cli not found in PATH. Please install whisper-cpp via Homebrew: brew install whisper-cpp")
	}

	return &Transcriber{
		whisperPath: whisperPath,
	}, nil
}

// TranscribeFile transcribes an audio file and returns the text
func (t *Transcriber) TranscribeFile(audioPath string, opts Options) (*Result, error) {
	// Validate that the model is downloaded
	isDownloaded, err := models.IsModelDownloaded(opts.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to check if model is downloaded: %w", err)
	}
	if !isDownloaded {
		return nil, fmt.Errorf("model %s is not downloaded. Run 'openscribe models download %s' first", opts.Model, opts.Model)
	}

	// Get the model path
	modelPath, err := models.GetModelPath(opts.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get model path: %w", err)
	}

	// Build whisper-cli command
	args := []string{
		"-m", modelPath,
		"-f", audioPath,
		"--no-timestamps", // We want clean text output without timestamps
		"--output-txt",    // Output as text
		// Note: --print-colors is NOT used here, as it would enable colors
	}

	// Add language if specified
	if opts.Language != "" {
		args = append(args, "-l", opts.Language)
	}

	// Add threads for faster processing (use 4 threads by default)
	args = append(args, "-t", "4")

	// Verbose mode
	if !opts.Verbose {
		args = append(args, "--no-prints")
	}

	// Execute whisper-cli
	cmd := exec.Command(t.whisperPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("whisper-cli failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse the output
	output := stdout.String()
	text := t.parseOutput(output)

	if text == "" {
		return nil, fmt.Errorf("transcription produced empty result")
	}

	result := &Result{
		Text:     text,
		Language: opts.Language,
	}

	// If language was auto-detected, try to extract it from output
	if opts.Language == "" {
		detectedLang := t.extractLanguage(output)
		if detectedLang != "" {
			result.Language = detectedLang
		}
	}

	return result, nil
}

// parseOutput extracts the transcribed text from whisper-cli output
func (t *Transcriber) parseOutput(output string) string {
	lines := strings.Split(output, "\n")
	var textLines []string

	for _, line := range lines {
		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip common metadata lines
		lower := strings.ToLower(line)
		if strings.Contains(lower, "detected language") ||
			strings.Contains(lower, "processing") ||
			strings.HasPrefix(lower, "whisper") {
			continue
		}

		// If line starts with timestamp like [00:00:00.000 --> 00:00:02.000], extract text after it
		if strings.HasPrefix(line, "[") {
			// Find the closing bracket
			closingIndex := strings.Index(line, "]")
			if closingIndex != -1 && closingIndex < len(line)-1 {
				// Extract text after the timestamp
				text := strings.TrimSpace(line[closingIndex+1:])
				if text != "" {
					textLines = append(textLines, text)
				}
			}
			continue
		}

		textLines = append(textLines, line)
	}

	// Join all text lines and clean up
	text := strings.Join(textLines, " ")
	text = strings.TrimSpace(text)

	// Strip ANSI color codes
	text = stripAnsiCodes(text)

	return text
}

// stripAnsiCodes removes ANSI escape codes from a string
func stripAnsiCodes(s string) string {
	// Regex to match ANSI escape codes (both \x1b[ and plain [ variants)
	// Matches sequences like \x1b[38;5;160m or [38;5;160m or [0m
	ansiRegex := regexp.MustCompile(`(\x1b)?\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(s, "")
}

// extractLanguage tries to extract the detected language from whisper output
func (t *Transcriber) extractLanguage(output string) string {
	// Whisper-cli usually prints something like "Detected language: en"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "detected language") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				lang := strings.TrimSpace(parts[1])
				// Extract just the language code if present
				fields := strings.Fields(lang)
				if len(fields) > 0 {
					return fields[0]
				}
			}
		}
	}
	return ""
}

// DefaultOptions returns default transcription options
func DefaultOptions() Options {
	return Options{
		Model:    models.Small,
		Language: "", // auto-detect
		Verbose:  false,
	}
}

package transcription

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/alexandrelam/openscribe/internal/models"
)

// WhisperTranscriber handles speech-to-text transcription using whisper.cpp
type WhisperTranscriber struct {
	whisperPath string
}

// NewWhisperTranscriber creates a new whisper-based transcriber
func NewWhisperTranscriber() (*WhisperTranscriber, error) {
	whisperPath, err := exec.LookPath("whisper-cli")
	if err != nil {
		return nil, fmt.Errorf("whisper-cli not found in PATH. Please install whisper-cpp via Homebrew: brew install whisper-cpp")
	}

	return &WhisperTranscriber{
		whisperPath: whisperPath,
	}, nil
}

// TranscribeFile transcribes an audio file and returns the text
func (t *WhisperTranscriber) TranscribeFile(audioPath string, opts Options) (*Result, error) {
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
		"--no-timestamps",
		"--output-txt",
	}

	// Add language if specified
	if opts.Language != "" {
		args = append(args, "-l", opts.Language)
	}

	// Add threads for faster processing
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
	text := parseWhisperOutput(output)

	if text == "" {
		return nil, fmt.Errorf("transcription produced empty result")
	}

	result := &Result{
		Text:     text,
		Language: opts.Language,
	}

	// If language was auto-detected, try to extract it from output
	if opts.Language == "" {
		detectedLang := extractWhisperLanguage(output)
		if detectedLang != "" {
			result.Language = detectedLang
		}
	}

	return result, nil
}

// parseWhisperOutput extracts the transcribed text from whisper-cli output
func parseWhisperOutput(output string) string {
	lines := strings.Split(output, "\n")
	var textLines []string

	for _, line := range lines {
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
			closingIndex := strings.Index(line, "]")
			if closingIndex != -1 && closingIndex < len(line)-1 {
				text := strings.TrimSpace(line[closingIndex+1:])
				if text != "" {
					textLines = append(textLines, text)
				}
			}
			continue
		}

		textLines = append(textLines, line)
	}

	text := strings.Join(textLines, " ")
	text = strings.TrimSpace(text)
	text = stripAnsiCodes(text)

	return text
}

// stripAnsiCodes removes ANSI escape codes from a string
func stripAnsiCodes(s string) string {
	ansiRegex := regexp.MustCompile(`(\x1b)?\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(s, "")
}

// extractWhisperLanguage tries to extract the detected language from whisper output
func extractWhisperLanguage(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "detected language") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				lang := strings.TrimSpace(parts[1])
				fields := strings.Fields(lang)
				if len(fields) > 0 {
					return fields[0]
				}
			}
		}
	}
	return ""
}

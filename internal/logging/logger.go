package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alexandrelam/openscribe/internal/config"
)

// TranscriptionEntry represents a single transcription log entry
type TranscriptionEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Duration  float64   `json:"duration_seconds"`
	Model     string    `json:"model"`
	Language  string    `json:"language"`
	Text      string    `json:"text"`
}

// LogTranscription writes a transcription entry to the log file
func LogTranscription(duration float64, model, language, text string) error {
	// Ensure log directory exists
	if err := config.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Get log file path
	logPath, err := config.GetTranscriptionLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Create entry
	entry := TranscriptionEntry{
		Timestamp: time.Now(),
		Duration:  duration,
		Model:     model,
		Language:  language,
		Text:      text,
	}

	// Open file in append mode (create if doesn't exist)
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Marshal to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Write JSON line
	if _, err := file.Write(append(jsonData, '\n')); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

// GetTranscriptions reads transcription entries from the log file
func GetTranscriptions(tail int) ([]TranscriptionEntry, error) {
	logPath, err := config.GetTranscriptionLogPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get log path: %w", err)
	}

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return []TranscriptionEntry{}, nil
	}

	// Open log file
	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines
	var entries []TranscriptionEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry TranscriptionEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			// Skip malformed lines
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Return last N entries if tail is specified
	if tail > 0 && len(entries) > tail {
		return entries[len(entries)-tail:], nil
	}

	return entries, nil
}

// ClearTranscriptions removes all transcription log entries
func ClearTranscriptions() error {
	logPath, err := config.GetTranscriptionLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// File doesn't exist, nothing to clear
		return nil
	}

	// Remove the file
	if err := os.Remove(logPath); err != nil {
		return fmt.Errorf("failed to remove log file: %w", err)
	}

	return nil
}

// CountTranscriptions returns the total number of transcription entries
func CountTranscriptions() (int, error) {
	entries, err := GetTranscriptions(0)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

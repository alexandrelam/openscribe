// Package models provides Whisper model management and downloading functionality.
package models

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexandrelam/openscribe/internal/config"
)

// ProgressCallback is called periodically during download
type ProgressCallback func(downloaded, total int64, percent float64)

// DownloadModel downloads a Whisper model with progress reporting
func DownloadModel(modelName ModelSize, progress ProgressCallback) error {
	modelInfo, ok := AvailableModels[modelName]
	if !ok {
		return fmt.Errorf("unknown model: %s", modelName)
	}

	// Ensure models directory exists
	modelsDir, err := config.GetModelsDir()
	if err != nil {
		return fmt.Errorf("failed to get models directory: %w", err)
	}

	if mkdirErr := os.MkdirAll(modelsDir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create models directory: %w", mkdirErr)
	}

	// Download to a temporary file first
	tempFile := filepath.Join(modelsDir, modelInfo.FileName+".tmp")
	finalPath := filepath.Join(modelsDir, modelInfo.FileName)

	// Check if model already exists
	if _, statErr := os.Stat(finalPath); statErr == nil {
		return fmt.Errorf("model already exists: %s", modelName)
	}

	// Create the HTTP request
	resp, err := http.Get(modelInfo.URL)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Best effort close
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download model: HTTP %d", resp.StatusCode)
	}

	// Create the temporary file
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		_ = out.Close() // Will be closed explicitly before rename
	}()

	// Get the total size
	totalSize := resp.ContentLength

	// Create a progress reader
	reader := &progressReader{
		reader:   resp.Body,
		total:    totalSize,
		callback: progress,
	}

	// Copy with progress
	_, err = io.Copy(out, reader)
	if err != nil {
		_ = os.Remove(tempFile) // Clean up on error
		return fmt.Errorf("failed to write model file: %w", err)
	}

	// Close the file before renaming
	if err := out.Close(); err != nil {
		_ = os.Remove(tempFile) // Clean up on error
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Move temp file to final location
	if err := os.Rename(tempFile, finalPath); err != nil {
		_ = os.Remove(tempFile) // Clean up on error
		return fmt.Errorf("failed to finalize model file: %w", err)
	}

	// Validate the downloaded model
	if err := ValidateModel(modelName); err != nil {
		_ = os.Remove(finalPath) // Remove invalid file
		return fmt.Errorf("model validation failed: %w", err)
	}

	return nil
}

// progressReader wraps an io.Reader to report download progress
type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	callback   ProgressCallback
	lastUpdate time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	// Update progress every 100ms to avoid too many callbacks
	if pr.callback != nil && time.Since(pr.lastUpdate) > 100*time.Millisecond {
		percent := 0.0
		if pr.total > 0 {
			percent = float64(pr.downloaded) / float64(pr.total) * 100.0
		}
		pr.callback(pr.downloaded, pr.total, percent)
		pr.lastUpdate = time.Now()
	}

	return n, err
}

// FormatBytes converts bytes to a human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSpeed formats a speed in bytes per second
func FormatSpeed(bytesPerSecond float64) string {
	return fmt.Sprintf("%s/s", FormatBytes(int64(bytesPerSecond)))
}

// EstimateTimeRemaining estimates the time remaining for a download
func EstimateTimeRemaining(downloaded, total int64, bytesPerSecond float64) string {
	if bytesPerSecond == 0 || total == 0 {
		return "calculating..."
	}

	remaining := total - downloaded
	seconds := float64(remaining) / bytesPerSecond

	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	}
	if seconds < 3600 {
		minutes := seconds / 60
		return fmt.Sprintf("%.0fm", minutes)
	}
	hours := seconds / 3600
	return fmt.Sprintf("%.1fh", hours)
}

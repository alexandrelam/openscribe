package logging

import (
	"os"
	"testing"
	"time"

	"github.com/alexandrelam/openscribe/internal/config"
)

func TestLogTranscription(t *testing.T) {
	// Create a test log entry
	duration := 5.5
	model := "whisper-small"
	language := "en"
	text := "This is a test transcription"

	// Log the transcription (uses actual config paths)
	err := LogTranscription(duration, model, language, text)
	if err != nil {
		t.Fatalf("LogTranscription failed: %v", err)
	}

	// Read back the transcriptions
	entries, err := GetTranscriptions(0)
	if err != nil {
		t.Fatalf("GetTranscriptions failed: %v", err)
	}

	// Verify at least one entry exists
	if len(entries) == 0 {
		t.Fatal("Expected at least one transcription entry")
	}

	// Verify the last entry matches what we logged
	lastEntry := entries[len(entries)-1]
	if lastEntry.Duration != duration {
		t.Errorf("Expected duration %.2f, got %.2f", duration, lastEntry.Duration)
	}
	if lastEntry.Model != model {
		t.Errorf("Expected model %s, got %s", model, lastEntry.Model)
	}
	if lastEntry.Language != language {
		t.Errorf("Expected language %s, got %s", language, lastEntry.Language)
	}
	if lastEntry.Text != text {
		t.Errorf("Expected text %s, got %s", text, lastEntry.Text)
	}

	// Verify timestamp is recent (within last minute)
	if time.Since(lastEntry.Timestamp) > time.Minute {
		t.Errorf("Expected recent timestamp, got %v", lastEntry.Timestamp)
	}
}

func TestGetTranscriptionsWithTail(t *testing.T) {
	// Clear existing logs first
	_ = ClearTranscriptions()

	// Log multiple transcriptions
	for i := 1; i <= 5; i++ {
		err := LogTranscription(float64(i), "whisper-small", "en", "Test "+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("Failed to log transcription %d: %v", i, err)
		}
	}

	// Test tail = 3
	entries, err := GetTranscriptions(3)
	if err != nil {
		t.Fatalf("GetTranscriptions failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Test tail = 0 (all entries)
	allEntries, err := GetTranscriptions(0)
	if err != nil {
		t.Fatalf("GetTranscriptions failed: %v", err)
	}

	if len(allEntries) < 5 {
		t.Errorf("Expected at least 5 entries, got %d", len(allEntries))
	}
}

func TestGetTranscriptionsNoFile(t *testing.T) {
	// Clear logs to ensure file doesn't exist
	_ = ClearTranscriptions()

	// Try to read from non-existent file
	entries, err := GetTranscriptions(0)
	if err != nil {
		t.Fatalf("Expected no error when file doesn't exist, got: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected empty slice, got %d entries", len(entries))
	}
}

func TestClearTranscriptions(t *testing.T) {
	// Log a transcription
	err := LogTranscription(1.0, "whisper-small", "en", "Test")
	if err != nil {
		t.Fatalf("Failed to log transcription: %v", err)
	}

	// Clear logs
	err = ClearTranscriptions()
	if err != nil {
		t.Fatalf("ClearTranscriptions failed: %v", err)
	}

	// Verify log file is gone
	logPath, _ := config.GetTranscriptionLogPath()
	if _, err := os.Stat(logPath); err == nil {
		t.Error("Expected log file to be deleted")
	}

	// Verify reading returns empty
	entries, err := GetTranscriptions(0)
	if err != nil {
		t.Fatalf("GetTranscriptions failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}

	// Test clearing when no file exists (should not error)
	err = ClearTranscriptions()
	if err != nil {
		t.Fatalf("Expected no error when clearing non-existent file, got: %v", err)
	}
}

func TestCountTranscriptions(t *testing.T) {
	// Clear logs
	_ = ClearTranscriptions()

	// Count when empty
	count, err := CountTranscriptions()
	if err != nil {
		t.Fatalf("CountTranscriptions failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Add some entries
	for i := 0; i < 3; i++ {
		_ = LogTranscription(1.0, "whisper-small", "en", "Test")
	}

	// Count again
	count, err = CountTranscriptions()
	if err != nil {
		t.Fatalf("CountTranscriptions failed: %v", err)
	}

	if count < 3 {
		t.Errorf("Expected count >= 3, got %d", count)
	}
}

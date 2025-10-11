// Package logging provides transcription history logging for OpenScribe.
//
// This package handles:
//   - Logging transcriptions to file with structured format
//   - JSON-based log entries with timestamps
//   - Reading and displaying transcription history
//   - Log file management and clearing
//
// Each transcription log entry includes:
//   - Timestamp (ISO 8601 format)
//   - Transcribed text
//   - Audio duration
//   - Model used
//   - Language detected/specified
//
// Log file location:
//
//	~/Library/Logs/openscribe/transcriptions.log
//
// The log file uses JSON Lines format (newline-delimited JSON) for easy
// parsing and processing. Each line is a complete JSON object representing
// one transcription.
//
// Example usage:
//
//	// Create logger
//	logger := logging.NewLogger()
//	defer logger.Close()
//
//	// Log a transcription
//	entry := logging.Entry{
//	    Timestamp: time.Now(),
//	    Text:      "Hello, world!",
//	    Duration:  5 * time.Second,
//	    Model:     "small",
//	    Language:  "en",
//	}
//	if err := logger.Log(entry); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Read history
//	entries, err := logger.ReadRecent(10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, e := range entries {
//	    fmt.Printf("[%s] %s\n", e.Timestamp.Format(time.RFC3339), e.Text)
//	}
package logging

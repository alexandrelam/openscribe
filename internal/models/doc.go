// Package models provides Whisper model management for OpenScribe.
//
// This package handles:
//   - Whisper model discovery and listing
//   - Model downloading from HuggingFace
//   - Model validation and integrity checking
//   - whisper-cpp installation detection
//   - Progress reporting during downloads
//
// Supported Whisper models:
//   - tiny: ~75MB, fastest, least accurate
//   - base: ~145MB, fast, good for simple speech
//   - small: ~500MB, balanced (recommended)
//   - medium: ~1.5GB, slower, more accurate
//   - large: ~3GB, slowest, most accurate
//
// Models are downloaded from HuggingFace (ggerganov/whisper.cpp) and stored in:
//
//	~/Library/Application Support/openscribe/models/
//
// The package also handles whisper-cpp binary detection via Homebrew.
// whisper-cpp must be installed separately using: brew install whisper-cpp
//
// Example usage:
//
//	// Check if whisper-cpp is installed
//	if !models.IsWhisperCppInstalled() {
//	    fmt.Println("Please install whisper-cpp: brew install whisper-cpp")
//	    return
//	}
//
//	// List available models
//	available := models.ListAvailableModels()
//	for _, m := range available {
//	    fmt.Printf("%s: %s, Size: %s\n", m.Name, m.Description, m.Size)
//	}
//
//	// Download a model
//	downloader := models.NewDownloader()
//	if err := downloader.Download("small", func(progress float64) {
//	    fmt.Printf("\rProgress: %.1f%%", progress*100)
//	}); err != nil {
//	    log.Fatal(err)
//	}
//
//	// List downloaded models
//	downloaded, err := models.ListDownloadedModels()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, m := range downloaded {
//	    fmt.Printf("Model: %s\n", m)
//	}
package models

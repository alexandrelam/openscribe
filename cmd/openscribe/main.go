// Package main is the entry point for the OpenScribe CLI application.
//
// OpenScribe is a free, private, universal speech-to-text application for macOS
// that provides real-time audio transcription with hotkey activation. It uses
// OpenAI's Whisper model (via whisper.cpp) for high-quality offline transcription.
//
// Key features:
//   - Hotkey-activated recording (double-press configurable modifier key)
//   - Offline transcription using Whisper (no internet required)
//   - Auto-paste transcribed text at cursor position
//   - Multiple model sizes (tiny to large) for speed/accuracy tradeoffs
//   - Multi-language support with auto-detection
//   - Audio feedback for state transitions
//   - Transcription history logging
//
// Usage:
//
//	openscribe start                  # Start transcription service
//	openscribe setup                  # First-time setup
//	openscribe config --show          # View configuration
//	openscribe models list            # List models
//	openscribe logs show              # View history
//	openscribe version                # Show version
//
// For more information, see: https://github.com/alexandrelam/openscribe-go
package main

import "github.com/alexandrelam/openscribe/internal/cli"

func main() {
	cli.Execute()
}

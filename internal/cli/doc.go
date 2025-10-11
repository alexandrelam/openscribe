// Package cli implements the command-line interface for OpenScribe using Cobra.
//
// This package provides all CLI commands:
//   - start: Start the transcription service with hotkey activation
//   - setup: Initial setup (download models, verify whisper-cpp)
//   - config: Configuration management (microphones, models, language, hotkeys)
//   - models: Model management (list, download)
//   - logs: Transcription history viewing
//   - version: Show version information
//
// The CLI is built using the Cobra library (github.com/spf13/cobra) which
// provides automatic help generation, flag parsing, and command organization.
//
// Main workflow:
//  1. User runs `openscribe start` to begin listening for hotkeys
//  2. Double-press configured hotkey to start recording
//  3. Speak into microphone
//  4. Double-press hotkey again to stop recording
//  5. Audio is transcribed using Whisper
//  6. Text is automatically pasted at cursor position
//  7. Transcription is logged to history
//
// Example usage:
//
//	# Basic usage
//	openscribe start
//
//	# With options
//	openscribe start --model base --language en --no-paste
//
//	# Configuration
//	openscribe config --set-microphone "MacBook Pro Microphone"
//	openscribe config --set-hotkey "Right Option"
//
//	# Model management
//	openscribe models list
//	openscribe models download small
//
//	# View history
//	openscribe logs show -n 10
package cli

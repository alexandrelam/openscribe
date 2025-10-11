// Package transcription provides speech-to-text transcription using Whisper.
//
// This package handles:
//   - Integration with whisper-cpp (via command-line invocation)
//   - Model loading and validation
//   - Audio file transcription with various options
//   - Language detection and specification
//   - Error handling for transcription failures
//
// The transcription process:
//  1. Takes an audio file path (WAV format, 16kHz, mono)
//  2. Validates the specified Whisper model exists
//  3. Invokes whisper-cli with appropriate parameters
//  4. Parses the transcription output
//  5. Returns the transcribed text
//
// Supported Whisper models:
//   - tiny: Fastest, least accurate (~75MB)
//   - base: Fast, good for simple speech (~145MB)
//   - small: Balanced speed/accuracy (~500MB) - Recommended
//   - medium: Slower, more accurate (~1.5GB)
//   - large: Slowest, most accurate (~3GB)
//
// Example usage:
//
//	// Create transcriber
//	transcriber := transcription.NewTranscriber()
//
//	// Transcribe audio file
//	opts := transcription.Options{
//	    Model:    "small",
//	    Language: "en",
//	    Verbose:  false,
//	}
//	text, err := transcriber.Transcribe("audio.wav", opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(text)
package transcription

// Package transcription provides speech-to-text transcription backends.
//
// Three backends sit behind the Transcriber interface, selected by the
// "backend" config key:
//
//   - whisper (default): local transcription via whisper-cpp, invoked as a
//     subprocess. Requires a downloaded model. Fully offline.
//   - moonshine: local transcription via the Moonshine models. Requires
//     building with -tags moonshine. Fully offline.
//   - openrouter: cloud transcription via the OpenRouter API. Requires an API
//     key. This is the only backend that sends audio off-device.
//
// The transcription process:
//  1. Takes an audio file path (WAV format, 16kHz, mono)
//  2. Dispatches to the configured backend
//  3. Returns the transcribed text plus any detected language and duration
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
//	// Create a transcriber for the configured backend
//	transcriber, err := transcription.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Transcribe an audio file
//	opts := transcription.Options{
//	    Model:    models.Small,
//	    Language: "en", // empty means auto-detect
//	    Verbose:  false,
//	}
//	result, err := transcriber.TranscribeFile("audio.wav", opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Text)
package transcription

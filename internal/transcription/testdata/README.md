# Test Audio Files

This directory contains small audio files used for **local-only** integration testing of the transcription engine.

## Files

- `test-english.wav` - Short English audio clip for testing transcription (95KB)

## Running Integration Tests Locally

Integration tests are **NOT run in CI** to avoid slow setup times. They use Go build tags and must be explicitly enabled.

To run integration tests locally:

```bash
# 1. Install whisper-cpp
brew install whisper-cpp

# 2. Download the tiny model for fast testing
openscribe models download tiny

# 3. Run integration tests with the build tag
go test -tags=integration ./internal/transcription/... -v
```

## CI/CD

In CI, only **unit tests** run:
- Output parsing tests
- Language extraction tests
- Configuration tests

These don't require `whisper-cli` or audio files, so CI is fast and doesn't need any special setup.

## Why Separate Integration Tests?

Installing whisper-cpp and downloading models would add significant time to CI:
- Homebrew install: ~1-2 minutes
- Model download (tiny): ~75MB, 30+ seconds
- Transcription: 1-2 seconds per test

By using build tags (`//go:build integration`), we keep CI fast while still having comprehensive tests available for local development.

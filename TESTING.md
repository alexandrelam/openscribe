# Testing Strategy

This document outlines the testing approach for OpenScribe, optimized for both fast CI and comprehensive local testing.

## Test Types

### 1. Unit Tests (Always Run)

Fast, dependency-free tests that run in CI automatically:

```bash
go test ./...
```

**Characteristics:**
- No external dependencies (whisper-cli, models, audio files)
- Run in < 1 second total
- Test parsing logic, configuration, error handling

**Examples:**
- `internal/transcription/transcription_test.go` - Output parsing, language extraction
- `internal/config/config_test.go` - Configuration validation
- `internal/audio/devices_test.go` - Audio device enumeration logic

### 2. Integration Tests (Local Only)

Comprehensive tests with real dependencies, using Go build tags:

```bash
go test -tags=integration ./...
```

**Characteristics:**
- Require `whisper-cli` (via `brew install whisper-cpp`)
- Require downloaded models (`openscribe models download tiny`)
- Test actual transcription with real audio files
- Take 1-2 seconds per test

**Files:**
- `internal/transcription/integration_test.go` - Real audio transcription tests

**Why separate?** Installing whisper-cpp and downloading models would add 2+ minutes to CI time.

## Running Tests

### In CI (Fast)

```bash
# Runs only unit tests - no setup required
go test ./...

# Result: ~0.3 seconds total
```

### Locally (Full Coverage)

```bash
# 1. One-time setup
brew install whisper-cpp
openscribe models download tiny

# 2. Run all tests including integration
go test -tags=integration ./... -v

# Result: Unit tests + integration tests
```

## CI Configuration

Your CI doesn't need any special setup. The default `go test ./...` command will:
- ✓ Run all unit tests
- ⊘ Skip integration tests (due to build tag)
- ✓ Pass successfully

Example GitHub Actions workflow:

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test ./...
```

No whisper-cpp installation needed! ✅

## Adding New Tests

### Unit Test

```go
// internal/foo/foo_test.go
func TestSomething(t *testing.T) {
    // No build tag needed
    result := DoSomething()
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Integration Test

```go
//go:build integration
// +build integration

// internal/foo/integration_test.go
func TestSomethingWithRealDependency(t *testing.T) {
    // Will only run with: go test -tags=integration
    result := DoSomethingWithWhisper()
    // ...
}
```

## Test Coverage

To see coverage (unit tests only):

```bash
go test ./... -cover
```

To see coverage including integration tests:

```bash
go test -tags=integration ./... -cover
```

## Summary

| Test Type   | CI | Local | Speed | Dependencies |
|-------------|----|----|-------|--------------|
| Unit        | ✓  | ✓  | < 1s  | None         |
| Integration | ✗  | ✓  | ~2s   | whisper-cli, models, audio files |

This approach keeps CI fast while ensuring comprehensive testing is available for local development.

# Contributing to OpenScribe

Thank you for your interest in contributing to OpenScribe! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and welcoming environment for all contributors. Please:

- Be respectful and considerate in all interactions
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Accept criticism gracefully
- Respect differing viewpoints and experiences

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/openscribe-go.git
   cd openscribe-go
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/alexandrelam/openscribe-go.git
   ```

## Development Setup

### Prerequisites

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **macOS**: This project currently only supports macOS
- **Xcode Command Line Tools**: Install with `xcode-select --install`
- **whisper-cpp**: Install with `brew install whisper-cpp`
- **golangci-lint** (optional): For running linter
  ```bash
  brew install golangci-lint
  ```

### Setup

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Build the project**:
   ```bash
   make build
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Format code**:
   ```bash
   make fmt
   ```

5. **Run linter** (optional):
   ```bash
   make lint
   ```

## How to Contribute

### Reporting Bugs

Before submitting a bug report:
1. Check the [existing issues](https://github.com/alexandrelam/openscribe-go/issues) to avoid duplicates
2. Try the latest version to see if the bug still exists
3. Collect relevant information (macOS version, OpenScribe version, error messages)

When submitting a bug report, include:
- **Clear title** describing the issue
- **Steps to reproduce** the bug
- **Expected behavior** vs. **actual behavior**
- **Environment details**: macOS version, OpenScribe version (`openscribe version`)
- **Logs/screenshots**: Run with `--verbose` flag and include output
- **Configuration**: Output of `openscribe config --show`

### Suggesting Features

Before suggesting a feature:
1. Check [existing issues](https://github.com/alexandrelam/openscribe-go/issues) and [discussions](https://github.com/alexandrelam/openscribe-go/discussions)
2. Consider if it fits the project's scope and goals

When suggesting a feature, include:
- **Clear description** of the feature
- **Use case**: Why is this feature useful?
- **Proposed implementation** (if you have ideas)
- **Examples**: Show how it would work

### Submitting Code

1. **Create a branch** for your work:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/bug-description
   ```

2. **Make your changes**:
   - Follow the [coding standards](#coding-standards)
   - Add tests for new functionality
   - Update documentation as needed
   - Keep commits focused and atomic

3. **Test your changes**:
   ```bash
   make test
   make build
   ./bin/openscribe start  # Test manually
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Brief description of changes"
   ```

   Good commit message format:
   ```
   Short summary (50 chars or less)

   More detailed explanation if needed. Wrap at 72 characters.

   - Bullet points are okay
   - Explain what and why, not how

   Fixes #123
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request** on GitHub

## Pull Request Process

1. **Update your branch** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Ensure all tests pass**:
   ```bash
   make test
   make lint  # If you have golangci-lint installed
   ```

3. **Write a clear PR description**:
   - What does this PR do?
   - Why is this change needed?
   - How was it tested?
   - Related issues (use "Fixes #123" or "Relates to #123")

4. **Wait for review**:
   - Maintainers will review your PR
   - Be responsive to feedback and questions
   - Make requested changes promptly

5. **After approval**:
   - Maintainers will merge your PR
   - Your contribution will be included in the next release!

## Coding Standards

### Go Style

- Follow standard Go conventions: [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` to format code (run `make fmt`)
- Use `golangci-lint` for linting (run `make lint`)

### Code Organization

- Keep functions focused and small
- Use descriptive variable and function names
- Add comments for exported functions and complex logic
- Organize code into logical packages

### Example:

```go
// Good: Clear, descriptive function with documentation
// TranscribeAudio transcribes the given audio file using the specified model.
// It returns the transcribed text or an error if transcription fails.
func TranscribeAudio(audioPath string, model string) (string, error) {
    // Implementation
}

// Bad: Unclear function name, no documentation
func DoStuff(p string, m string) (string, error) {
    // Implementation
}
```

### Error Handling

- Always handle errors explicitly
- Provide context in error messages
- Use `fmt.Errorf` with `%w` to wrap errors

```go
// Good
if err != nil {
    return fmt.Errorf("failed to transcribe audio: %w", err)
}

// Bad
if err != nil {
    return err  // No context
}
```

### Package Documentation

- Add package-level documentation in `doc.go` files
- Document exported types, functions, and constants

## Testing

### Writing Tests

- Write tests for new functionality
- Use table-driven tests for multiple test cases
- Test both success and failure scenarios
- Use meaningful test names

Example:

```go
func TestTranscribe(t *testing.T) {
    tests := []struct {
        name      string
        audioPath string
        model     string
        wantErr   bool
    }{
        {
            name:      "valid audio file",
            audioPath: "testdata/audio.wav",
            model:     "small",
            wantErr:   false,
        },
        {
            name:      "missing audio file",
            audioPath: "nonexistent.wav",
            model:     "small",
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Transcribe(tt.audioPath, tt.model)
            if (err != nil) != tt.wantErr {
                t.Errorf("Transcribe() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/audio/...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...
```

## Documentation

### Code Documentation

- Add godoc comments for all exported symbols
- Use complete sentences starting with the item name
- Include examples where helpful

### User Documentation

When adding new features, update:
- **README.md**: For user-facing features and installation instructions
- **TROUBLESHOOTING.md**: For common issues and solutions
- **PLAN.md**: For development status and implementation notes

### Example Documentation

```go
// Recorder captures audio from a microphone device.
// It buffers audio data in memory until Stop is called.
type Recorder struct {
    // ...
}

// NewRecorder creates a new Recorder for the specified device.
// The deviceID should match one returned by ListDevices.
func NewRecorder(deviceID string) (*Recorder, error) {
    // ...
}
```

## Project Structure

```
openscribe-go/
├── cmd/                    # Application entry point
│   └── openscribe/        # Main executable
├── internal/              # Private application code
│   ├── audio/            # Audio recording and feedback
│   ├── config/           # Configuration management
│   ├── hotkey/           # Hotkey detection (macOS)
│   ├── keyboard/         # Keyboard simulation (macOS)
│   ├── logging/          # Transcription logging
│   ├── models/           # Model management
│   ├── transcription/    # Whisper integration
│   └── cli/              # CLI commands (Cobra)
├── docs/                  # Documentation
├── scripts/               # Build and release scripts
└── Makefile              # Build automation
```

## Getting Help

- **Questions**: Open a [GitHub Discussion](https://github.com/alexandrelam/openscribe-go/discussions)
- **Bugs**: Open a [GitHub Issue](https://github.com/alexandrelam/openscribe-go/issues)
- **Chat**: (Coming soon - Discord/Slack?)

## Recognition

Contributors will be:
- Listed in the project's contributor list
- Mentioned in release notes for significant contributions
- Appreciated for their time and effort! 🙏

## License

By contributing to OpenScribe, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to OpenScribe! Your help makes this project better for everyone. 🚀

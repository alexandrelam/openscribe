# Product Requirements Document: OpenScribe

## Project Overview

OpenScribe is a Go-based CLI application for macOS that enables real-time speech transcription with hotkey activation. The tool records audio via a double-press of a configurable button, transcribes the speech using Whisper, and automatically pastes the transcribed text at the current cursor position.

## Platform & Distribution

- **Platform**: macOS only (initial release)
- **Distribution**: Homebrew
- **Installation**: `brew install openscribe`

## Technology Stack

- **Language**: Go
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) - Modern CLI library used by Kubernetes, Hugo, and GitHub CLI
- **Speech Recognition**: [whisper.cpp](https://github.com/ggerganov/whisper.cpp) - C++ implementation of OpenAI's Whisper
- **Default Model**: Whisper Small (will be downloaded during setup)
- **Audio Recording**: Go audio libraries with macOS Core Audio support
- **Audio Feedback**: AVFoundation or NSSound for playing distinct sounds (start/stop/complete)
- **System Integration**: macOS-native APIs for keyboard simulation (CGEvent for direct text input)

## Project Structure

The Go project should follow standard Go best practices with the following high-level organization:

- **`cmd/`**: Application entry point (minimal main.go that calls CLI commands)
- **`internal/`**: Private application code, organized by domain:
  - **Audio management**: Recording, device detection, audio feedback/sounds
  - **Transcription**: Whisper.cpp integration, model management, language support
  - **System integration**: Hotkey detection, keyboard simulation, macOS permissions
  - **Configuration**: Config file management, default settings, macOS standard paths
  - **Logging**: Log file management, transcription history
  - **CLI commands**: Cobra command definitions (setup, start, config, models, logs, version)
- **`assets/`**: Static resources like bundled sound files (start/stop/complete beeps)
- **`scripts/`**: Build, installation, and release automation scripts
- **`homebrew/`**: Homebrew formula template
- **`docs/`**: Documentation and troubleshooting guides

Note: The `internal/` directory ensures code is private to the project. Specific file organization within each domain will be determined during implementation.

## Core Features

### 1. Setup & Installation
- **Homebrew Installation**: `brew install` only installs the binary, NO models downloaded
- **First Run Detection**: When `openscribe start` is run without models, display helpful message prompting user to run setup
- **Model Download**: User must explicitly run `openscribe setup` or `openscribe models download [model]` to download models
- **whisper.cpp Integration**: Download and compile whisper.cpp binaries during setup if not present
- **Configuration**: Create config file for user preferences on first run

### 2. Microphone Selection
- List all available audio input devices
- Allow user to select preferred microphone via CLI command
- Save microphone preference to config file
- Support for `--microphone` or `-m` flag to specify device

### 3. Model Selection
- Support multiple Whisper model sizes (tiny, base, small, medium, large)
- Default to small model
- Allow model selection via `--model` flag
- List available models with size/accuracy trade-offs

### 4. Language Configuration
- Allow language specification via `--language` or `-l` flag
- Support auto-detection (default behavior)
- List supported languages via `--list-languages` command

### 5. Recording Control
- **Hotkey Activation**: Double-press of a configurable button (default: Right Option)
- **Start Recording**: First double-press activates recording
- **Audio Feedback**: Play distinct sound when recording starts
- **Visual Feedback**: Display recording status in terminal
- **Stop Recording**: Second double-press stops recording
- **Audio Feedback**: Play distinct sound when recording stops (different from start sound)
- **Hotkey Configuration**: Allow users to configure the activation button
- **Exit Behavior**: Ctrl+C stops the application and any in-progress recording

### 6. Transcription & Output
- Transcribe audio using selected Whisper model
- **Audio Feedback**: Play distinct sound when transcription completes (different from start/stop sounds)
- Display full transcription in terminal
- Log all transcriptions to log file with timestamp
- **Auto-paste**: Directly paste transcribed text at cursor position (simulate keyboard input, NOT via clipboard)
- Option to disable auto-paste and only show text in terminal (`--no-paste` flag)
- Clipboard contents remain unaffected by transcription

### 7. Logging & Transcription History
- **Log Location**: `~/Library/Logs/openscribe/transcriptions.log`
- Log each transcription with timestamp, duration, model used, and language detected
- Terminal output shows real-time transcription
- Logs persist for debugging and history review
- Optional `--verbose` flag for detailed debug output

### 8. Audio Feedback System
- **Three Distinct Sounds**: Each state change plays a unique sound
  - **Start Recording Sound**: Short, friendly beep (e.g., ascending tone)
  - **Stop Recording Sound**: Different tone (e.g., descending tone)
  - **Transcription Complete Sound**: Success indicator (e.g., pleasant "ding")
- Sounds should be brief (< 0.5 seconds) and non-intrusive
- Consider bundling custom sound files or using system sounds
- Optional configuration to disable audio feedback

## User Flows

### Installation via Homebrew
```bash
$ brew tap yourusername/openscribe
$ brew install openscribe
✅ OpenScribe installed successfully!
```

### First-Time Setup Flow
```bash
# Try to start without models
$ openscribe start
⚠️  No Whisper models found!

Please run setup to download a model:
  $ openscribe setup

Or download a specific model:
  $ openscribe models download small

# Run setup
$ openscribe setup
Checking for whisper.cpp... Not found
Downloading whisper.cpp...
Compiling whisper.cpp...
Downloading whisper-small model to ~/Library/Application Support/openscribe/models/...
Setup complete!

Configure your microphone (optional):
$ openscribe config --list-microphones
```

### Recording & Transcription Flow
```bash
$ openscribe start
Using microphone: MacBook Pro Microphone
Using model: whisper-small
Language: auto-detect
Hotkey: Right Option (double-press)
Ready! Press hotkey to start recording...

[User double-presses Right Option]
🔊 *beep* (start sound)
🔴 Recording... (press hotkey again to stop)

[User double-presses Right Option]
🔊 *boop* (stop sound)
⏹  Recording stopped. Transcribing...
🔊 *ding* (transcription complete sound)
Transcription: "Hello, this is a test of the OpenScribe application."
✅ Text pasted to cursor position!

[2025-01-15 14:23:45] Logged to ~/Library/Logs/openscribe/transcriptions.log
```

### Configuration Commands
```bash
# List available microphones
$ openscribe config --list-microphones

# Set microphone
$ openscribe config --set-microphone "MacBook Pro Microphone"

# Set default model
$ openscribe config --set-model small

# Set default language
$ openscribe config --set-language en

# Configure hotkey
$ openscribe config --set-hotkey "Right Option"

# View current configuration
$ openscribe config --show
```

## Technical Requirements

### 1. Audio Recording
- Capture audio from selected microphone device
- Support common audio formats (WAV recommended for whisper.cpp)
- Configurable sample rate (16kHz default for Whisper)
- Real-time audio buffer management

### 2. Whisper Integration
- Bind to whisper.cpp via CGo or command-line invocation
- Model loading and caching
- Efficient transcription processing
- Handle different model sizes

### 3. System Integration (macOS)
- **Hotkey Detection**: Global hotkey listener using macOS APIs
- **Keyboard Simulation**: Directly simulate keyboard input using CGEvent APIs (NOT clipboard-based)
- **Audio Feedback**: Play system sounds for recording states (start/stop/complete) using AVFoundation or NSSound
- **Accessibility Permissions**: Request and verify accessibility permissions for keyboard simulation
- **Microphone Permissions**: Request microphone access permissions
- **Permission Error Handling**: Display clear error messages with System Preferences navigation instructions if permissions denied

### 4. Configuration Management
- **Config file location**: `~/Library/Application Support/openscribe/config.yaml`
- **Models directory**: `~/Library/Application Support/openscribe/models/`
- **Cache directory**: `~/Library/Caches/openscribe/`
- **Logs directory**: `~/Library/Logs/openscribe/`
- **Store**: microphone device, model preference, language, hotkey mapping, audio feedback settings
- **Default values** if config doesn't exist

### 5. Error Handling
- **No models found**: Display helpful message directing user to run `openscribe setup`
- **No microphone available**: List available devices and suggest configuration
- **Model download failures**: Network errors, disk space issues, corrupted downloads
- **Transcription errors**: Invalid audio, unsupported format, model loading failures
- **Permission issues**: Microphone access denied, accessibility permissions denied (provide System Preferences navigation instructions)

## CLI Commands Structure

```
openscribe
├── setup                    # Initial setup (download models, compile whisper.cpp)
├── start                    # Start the OpenScribe service
│   ├── --microphone, -m     # Override microphone selection
│   ├── --model              # Override model selection
│   ├── --language, -l       # Override language setting
│   ├── --no-paste           # Disable auto-paste
│   └── --verbose, -v        # Enable verbose debug output
├── config                   # Configuration management
│   ├── --list-microphones   # List available microphones
│   ├── --set-microphone     # Set default microphone
│   ├── --set-model          # Set default model
│   ├── --set-language       # Set default language
│   ├── --set-hotkey         # Configure activation hotkey
│   └── --show               # Display current configuration
├── models                   # Model management
│   ├── list                 # List available/downloaded models
│   └── download [model]     # Download specific model
├── logs                     # Log management
│   ├── show                 # Display recent transcription logs
│   ├── --tail, -n [count]   # Show last N transcriptions
│   └── clear                # Clear transcription logs
└── version                  # Show version info
```

## Non-Functional Requirements

### Performance
- Transcription latency: < 5 seconds for 30-second audio clips (small model)
- Memory usage: < 500MB during transcription
- Quick startup time: < 2 seconds

### Usability
- Clear terminal feedback during all operations
- Helpful error messages with suggested fixes
- Progress indicators for long operations (downloads, transcription)

### Security & Privacy
- All processing happens locally (no cloud services)
- Audio files can be optionally saved or immediately deleted
- Require explicit microphone permissions

## Homebrew Distribution

### Formula Structure
- **Binary**: Compiled Go binary for macOS (ARM64 and x86_64)
- **Dependencies**: List any brew dependencies (if needed)
- **Post-install**: Instructions to run `openscribe setup` for first-time setup
- **Caveats**: Display accessibility and microphone permissions requirements
- **Models NOT Bundled**: Whisper models are downloaded separately by user via `openscribe setup` (keeps installation size small)

### Installation Artifacts
```
/opt/homebrew/bin/openscribe                           # Binary
~/Library/Application Support/openscribe/              # User data directory
  ├── config.yaml                                       # User configuration
  └── models/                                           # Downloaded models
~/Library/Caches/openscribe/                           # Temporary files
~/Library/Logs/openscribe/                             # Log files
  └── transcriptions.log                                # Transcription history
```

## Future Enhancements (Out of Scope for v1)

- Cross-platform support (Linux, Windows)
- Real-time streaming transcription
- Custom vocabulary/terminology support
- Multiple language detection in single recording
- GUI application with menu bar integration
- Transcription history and management
- Custom hotkey for different languages/models
- Audio file transcription (not just live recording)
- Homebrew cask for GUI version

## Success Metrics

- Successfully transcribe speech with >90% accuracy (English)
- End-to-end flow (record → transcribe → paste) completes in <10 seconds
- Zero crashes during normal operation
- Smooth installation experience via Homebrew
- Proper integration with macOS permissions system

## Dependencies & Risks

### Dependencies
- whisper.cpp availability and compatibility
- macOS Core Audio framework for audio recording
- macOS Accessibility permissions for keyboard simulation
- macOS Microphone permissions
- Homebrew for distribution and installation
- macOS system APIs (CGEvent for keyboard simulation)
- AVFoundation or NSSound for audio feedback

### Risks
- **Whisper.cpp compilation issues**: Bundle pre-compiled binaries or download/compile during setup
- **Accessibility permissions**: Users must grant permissions; provide clear instructions and error messages
- **Model download size**: Small model is ~500MB; provide clear download progress indicator and graceful failure handling
- **macOS version compatibility**: Test on multiple macOS versions (minimum version TBD)
- **Homebrew tap maintenance**: Need to maintain tap repository and update formula for new releases
- **First-run experience**: Users might try to run `openscribe start` without models; clear error messages are critical

## Open Questions

1. Should we support recording time limits (e.g., max 60 seconds)?
2. What should happen if hotkey is pressed while transcription is in progress?
3. Should we support multiple hotkeys for different languages/models?
4. Should audio recordings be saved by default for debugging/review?
5. How should we handle background noise detection/filtering?
6. What's the minimum macOS version we should support?
7. Should we bundle whisper.cpp with the Homebrew formula or download on first run?
8. How should we handle updates to models when new versions are available?
9. Should we support running as a background service/daemon?

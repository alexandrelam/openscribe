<div align="center">

<!-- Logo placeholder - Replace with actual logo -->
<img src="https://github.com/user-attachments/assets/877ec94a-782c-4156-b4e4-49885bd010e3" alt="OpenScribe Logo" width="200"/>

# OpenScribe

**The Free, Private, Universal Speech-to-Text App for macOS**

*Transform your voice into text instantly, anywhere, without compromising your privacy*

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Go 1.21+](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![macOS](https://img.shields.io/badge/platform-macOS-lightgrey.svg)](https://www.apple.com/macos/)

---

### 🎥 See OpenScribe in Action

<!-- Video demo placeholder - Replace with actual demo video -->
<a href="https://github.com/alexandrelam/openscribe">
  <img src="https://via.placeholder.com/800x450/FF6B6B/FFFFFF?text=📹+Demo+Video+Coming+Soon" alt="OpenScribe Demo Video" width="800"/>
</a>

*Click to watch how OpenScribe works across any macOS application*

</div>

## 🚀 Why Choose OpenScribe?

**Fed up with expensive speech-to-text subscriptions?** OpenScribe is the **100% free, open-source alternative** to paid apps like Whisper Transcribe and Ink Voice.

### ✨ **Completely Free Forever**
- No subscriptions, no usage limits, no hidden costs
- Full-featured app with professional-grade accuracy
- Open source - modify and customize as you wish

### 🔒 **Your Privacy Guaranteed**
- **100% offline processing** - your voice never leaves your Mac
- No internet required, no data sent to servers
- Unlike cloud-based services, your conversations stay private

### 🌍 **Works Everywhere**
- **Universal compatibility** - works in any macOS application
- One hotkey activation across all your apps
- No need for app-specific integrations or plugins

### ⚡ **Blazing Fast & Accurate**
- Real-time transcription with industry-leading Whisper AI
- Multiple model sizes (tiny to large) for speed/accuracy tradeoffs
- Auto-language detection or manual selection
- Smart audio processing for optimal transcription quality

---

## 📋 Table of Contents

- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage](#-usage)
- [Configuration](#️-configuration)
- [Commands Reference](#-commands-reference)
- [Troubleshooting](#-troubleshooting)
- [How It Works](#-how-it-works)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🔧 Installation

### Prerequisites

- **macOS** (tested on macOS 10.15+)
- **Homebrew** package manager
- **whisper-cpp** for transcription engine

### Step 1: Install whisper-cpp

```bash
brew install whisper-cpp
```

### Step 2: Install OpenScribe

**Option A: From Source** (recommended for development)

```bash
# Clone the repository
git clone https://github.com/alexandrelam/openscribe-go.git
cd openscribe-go

# Build and install
make build
make install
```

**Option B: Download Binary** (coming soon)

Pre-built binaries will be available in the [Releases](https://github.com/alexandrelam/openscribe-go/releases) section.

### Step 3: First-Time Setup

Run the setup command to download the default Whisper model:

```bash
openscribe setup
```

This will:
- Verify whisper-cpp installation
- Download the small Whisper model (~500MB)
- Create configuration directories
- Set up default preferences

### Step 4: Grant Permissions

OpenScribe requires two macOS permissions:

**Microphone Access**
- macOS will prompt you automatically when you first start recording
- Grant permission to allow audio recording

**Accessibility Access** (for auto-paste feature)
1. Open **System Preferences** → **Security & Privacy** → **Accessibility**
2. Click the lock icon to make changes
3. Add `openscribe` to the list of allowed apps
4. Check the box next to `openscribe`

---

## 🎯 Quick Start

1. **Start OpenScribe:**
   ```bash
   openscribe start
   ```

2. **Record and transcribe:**
   - Double-press **Right Option** key to start recording (you'll hear a beep)
   - Speak your message
   - Double-press **Right Option** again to stop recording (you'll hear a different beep)
   - Wait for transcription (you'll hear a completion sound)
   - Your text will automatically appear at your cursor!

3. **Stop OpenScribe:**
   - Press `Ctrl+C` in the terminal

**That's it!** You're now ready to use speech-to-text anywhere on your Mac.

---

## 📖 Usage

### Basic Recording

```bash
# Start with default settings
openscribe start

# Start with a specific microphone
openscribe start --microphone "MacBook Pro Microphone"

# Start with a specific model
openscribe start --model base  # faster but less accurate
openscribe start --model large # slower but more accurate

# Start with a specific language
openscribe start --language en  # English
openscribe start --language es  # Spanish
openscribe start --language fr  # French

# Disable auto-paste (only show text in terminal)
openscribe start --no-paste

# Enable verbose output for debugging
openscribe start --verbose
```

### Workflow Example

```bash
$ openscribe start
Using microphone: MacBook Pro Microphone
Using model: small
Language: auto-detect
Hotkey: Right Option (double-press)
Audio feedback: enabled
Ready! Press hotkey to start recording...

[Double-press Right Option]
🔴 Recording... (press hotkey again to stop)

[Double-press Right Option]
⏹  Recording stopped. Transcribing...
✅ Transcription: "Hello, this is a test of OpenScribe."
📝 Text pasted to cursor position!

[2025-01-15 14:23:45] Logged to ~/Library/Logs/openscribe/transcriptions.log
```

---

## ⚙️ Configuration

### View Current Configuration

```bash
openscribe config --show
```

### Configure Microphone

```bash
# List available microphones
openscribe config --list-microphones

# Set default microphone
openscribe config --set-microphone "MacBook Pro Microphone"
```

### Configure Model

```bash
# Set default model
openscribe config --set-model small
```

Available models:
- `tiny` - Fastest, least accurate (~75MB)
- `base` - Fast, good for simple speech (~145MB)
- `small` - **Recommended** - Balanced speed/accuracy (~500MB)
- `medium` - Slower, more accurate (~1.5GB)
- `large` - Slowest, most accurate (~3GB)

### Configure Language

```bash
# Set default language
openscribe config --set-language en  # English
openscribe config --set-language auto  # Auto-detect
```

### Configure Hotkey

```bash
# List available hotkeys
openscribe config --list-hotkeys

# Set activation hotkey
openscribe config --set-hotkey "Right Option"
openscribe config --set-hotkey "Left Shift"
openscribe config --set-hotkey "Right Command"
```

### Audio Feedback

```bash
# Enable audio feedback (beeps for start/stop/complete)
openscribe config --enable-audio-feedback

# Disable audio feedback
openscribe config --disable-audio-feedback

# List available system sounds
openscribe config --list-sounds

# Test audio feedback sounds
openscribe config --test-sounds
```

### Configuration File

All settings are stored in:
```
~/Library/Application Support/openscribe/config.yaml
```

You can edit this file directly if needed. Example:

```yaml
microphone: "MacBook Pro Microphone"
model: "small"
language: "auto"
hotkey: "Right Option"
audio_feedback: true
start_sound: "Tink"
stop_sound: "Pop"
complete_sound: "Glass"
```

---

## 📚 Commands Reference

### Main Commands

| Command | Description |
|---------|-------------|
| `openscribe start` | Start the transcription service |
| `openscribe setup` | Download default model and verify installation |
| `openscribe config` | Manage configuration settings |
| `openscribe models` | Manage Whisper models |
| `openscribe logs` | View transcription history |
| `openscribe version` | Show version information |

### Start Command Flags

| Flag | Description |
|------|-------------|
| `-m, --microphone` | Override microphone selection |
| `--model` | Override model selection |
| `-l, --language` | Override language setting |
| `--no-paste` | Disable auto-paste feature |
| `-v, --verbose` | Enable verbose debug output |

### Config Command Flags

| Flag | Description |
|------|-------------|
| `--show` | Display current configuration |
| `--list-microphones` | List available microphones |
| `--set-microphone` | Set default microphone |
| `--set-model` | Set default model |
| `--set-language` | Set default language |
| `--set-hotkey` | Configure activation hotkey |
| `--list-hotkeys` | List available hotkeys |
| `--enable-audio-feedback` | Enable audio feedback |
| `--disable-audio-feedback` | Disable audio feedback |
| `--list-sounds` | List available system sounds |
| `--test-sounds` | Test audio feedback sounds |

### Models Commands

| Command | Description |
|---------|-------------|
| `openscribe models list` | List downloaded models |
| `openscribe models download <model>` | Download a specific model |

Available models: `tiny`, `base`, `small`, `medium`, `large`

### Logs Commands

| Command | Description |
|---------|-------------|
| `openscribe logs show` | Display recent transcriptions |
| `openscribe logs show -n 10` | Show last 10 transcriptions |
| `openscribe logs clear` | Clear transcription history |

---

## 🔧 Troubleshooting

### Common Issues

**"No Whisper models found!"**
```bash
# Run setup to download the default model
openscribe setup

# Or download a specific model
openscribe models download small
```

**"whisper-cpp is not installed"**
```bash
# Install whisper-cpp via Homebrew
brew install whisper-cpp
```

**"Microphone permission denied"**
- Go to **System Preferences** → **Security & Privacy** → **Privacy** → **Microphone**
- Check the box next to `openscribe` or your terminal application

**"Accessibility permission denied" / Auto-paste not working**
- Go to **System Preferences** → **Security & Privacy** → **Accessibility**
- Click the lock icon to make changes
- Add `openscribe` or your terminal application to the list
- Check the box to enable it

**Hotkey not detected**
- Make sure OpenScribe has Accessibility permissions (see above)
- Try a different hotkey with `openscribe config --set-hotkey "Left Shift"`
- Check for conflicts with other apps using the same hotkey

**Poor transcription quality**
- Try a larger model: `openscribe config --set-model medium`
- Specify your language: `openscribe start --language en`
- Make sure you're in a quiet environment
- Speak clearly and at a moderate pace
- Check microphone selection: `openscribe config --list-microphones`

**Transcription is slow**
- Use a smaller model: `openscribe config --set-model base`
- Close other resource-intensive applications
- Note: First transcription is slower due to model loading

For more detailed troubleshooting, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).

---

## 🔍 How It Works

OpenScribe combines several technologies to provide seamless speech-to-text:

1. **Global Hotkey Detection**: Uses macOS Carbon Event Manager APIs to detect double-press of your configured hotkey
2. **Audio Recording**: Captures audio from your selected microphone using Go audio libraries with macOS Core Audio
3. **Speech Recognition**: Transcribes audio using [whisper.cpp](https://github.com/ggerganov/whisper.cpp), a high-performance C++ implementation of OpenAI's Whisper
4. **Text Injection**: Simulates keyboard input using CGEvent APIs to paste text at your cursor (not clipboard-based!)
5. **Audio Feedback**: Plays distinct system sounds for each state transition
6. **Logging**: Saves all transcriptions with timestamps to `~/Library/Logs/openscribe/transcriptions.log`

### File Locations

```
/opt/homebrew/bin/openscribe                          # Binary (if installed via Homebrew)
~/Library/Application Support/openscribe/             # Configuration and models
  ├── config.yaml                                      # Configuration file
  └── models/                                          # Downloaded Whisper models
~/Library/Caches/openscribe/                          # Temporary audio files
~/Library/Logs/openscribe/                            # Log files
  └── transcriptions.log                               # Transcription history
```

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Report bugs**: Open an issue with details about the problem
2. **Suggest features**: Open an issue with your feature request
3. **Submit PRs**: Fork the repo, make changes, and submit a pull request
4. **Improve documentation**: Help make the docs better

### Development Setup

```bash
# Clone the repository
git clone https://github.com/alexandrelam/openscribe-go.git
cd openscribe-go

# Install dependencies
make deps

# Build
make build

# Run
make run

# Run tests
make test

# Format code
make fmt

# Run linter
make lint
```

### Project Structure

```
openscribe-go/
├── cmd/                    # Application entry point
├── internal/               # Private application code
│   ├── audio/             # Audio recording and feedback
│   ├── config/            # Configuration management
│   ├── hotkey/            # Hotkey detection (macOS)
│   ├── keyboard/          # Keyboard simulation (macOS)
│   ├── logger/            # Logging and history
│   ├── models/            # Model management
│   └── transcription/     # Whisper integration
├── assets/                 # Static resources
├── docs/                   # Documentation
└── scripts/                # Build and installation scripts
```

---

## 📄 License

OpenScribe is licensed under the MIT License. See [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) - High-performance Whisper implementation by Georgi Gerganov
- [OpenAI Whisper](https://github.com/openai/whisper) - Original Whisper model
- [Cobra](https://github.com/spf13/cobra) - CLI framework

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/alexandrelam/openscribe-go/issues)
- **Discussions**: [GitHub Discussions](https://github.com/alexandrelam/openscribe-go/discussions)

---

<div align="center">

**Made with ❤️ for the open-source community**

[⭐ Star this repo](https://github.com/alexandrelam/openscribe-go) | [🐛 Report a bug](https://github.com/alexandrelam/openscribe-go/issues) | [💡 Request a feature](https://github.com/alexandrelam/openscribe-go/issues)

</div>

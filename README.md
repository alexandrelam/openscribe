<div align="center">

<!-- Logo placeholder - Replace with actual logo -->
<img src="https://github.com/user-attachments/assets/877ec94a-782c-4156-b4e4-49885bd010e3" alt="OpenScribe Logo" width="200"/>

# OpenScribe

**The Free, Private, Universal Speech-to-Text App for macOS**

*Transform your voice into text instantly, anywhere, without compromising your privacy*

[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![Python 3.8+](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/)
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
- 11+ languages with auto-detection
- Smart voice detection skips silence for 60-80% performance boost

## ⚡ Quick Start - Get Running in 2 Minutes

### 🖥️ One-Command Installation

```bash
# Clone and install globally
git clone git@github.com:alexandrelam/openscribe.git
cd openscribe
./install.sh
```

### 🎤 Start Using Immediately

```bash
# Launch the app from anywhere
openscribe

# Or use CLI mode
openscribe --cli
```

**That's it!** Hit `Cmd+Shift+R` to start recording, or double-tap `Shift` for quick access.

---

## 🌟 Powerful Features That Set Us Apart

### 🎯 **Smart Recording Modes**
- **Traditional Mode**: Record → Transcribe → Click to paste anywhere
- **Live Streaming**: Speak continuously, text appears in real-time
- **Voice Activity Detection**: Automatically starts/stops with your speech
- **Double-Key Shortcuts**: Double-tap Shift (record) or Control (live mode)

### 🌍 **Universal Language Support**
- **11+ Languages**: English, French, Spanish, German, Italian, Portuguese, and more
- **Auto-Detection**: Automatically detects what language you're speaking
- **Mixed Conversations**: Seamlessly handles switching between languages
- **Accent Friendly**: Trained on diverse voice patterns and accents

### 🔧 **Professional Customization**
- **Multiple AI Models**: Choose between speed (small) and accuracy (large-v3)
- **GPU + CPU Optimized**: Faster-whisper (GPU) or whisper.cpp (CPU) engines
- **Smart Chunking**: Natural speech boundaries prevent mid-word cuts
- **Configurable Hotkeys**: Set your preferred keyboard shortcuts

### 🎨 **Seamless Integration**
- **Works Everywhere**: Slack, Email, Notion, Terminal, any macOS app
- **Intelligent Pasting**: Preserves your clipboard, restores after pasting
- **Background Operation**: Minimal system resources, runs quietly
- **International Keyboards**: Full support for AZERTY, QWERTZ, and more  

## 🚀 Quick Start

### Global Installation (Recommended)

1. **Setup the project:**
   ```bash
   cd speech-to-text
   python3 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   pip install -r requirements.txt
   ```

2. **Install globally:**
   ```bash
   ./install.sh
   ```

3. **Use from anywhere:**
   ```bash
   vibe              # Launch GUI interface
   vibe --cli        # Launch CLI interface  
   vibe-cli          # Direct CLI access
   vibe --help       # Show help information
   vibe --version    # Show version info
   ```

### Development Installation

For development or testing without global installation:

1. **Install dependencies:**
   ```bash
   # On macOS
   brew install python-tk
   
   # On Ubuntu/Debian
   sudo apt-get install python3-tk
   
   pip install -r requirements.txt
   ```

2. **Test components:**
   ```bash
   python test_components.py
   ```

3. **Run locally:**
   ```bash
   python main.py      # GUI version
   python main_cli.py  # CLI version
   ```

## 📖 Usage

### Command Line Interface

After global installation, use these commands from any terminal:

```bash
# Launch GUI interface (default)
vibe

# Launch CLI interface
vibe --cli

# Direct CLI access
vibe-cli

# Show help
vibe --help

# Show version
vibe --version
```

### GUI Interface

1. **Traditional Recording Mode:**
   - Click "🎤 Start Recording" or press Cmd+Shift+R
   - **NEW: Double-tap Shift** for quick activation
   - Speak into your microphone
   - Click "⏹️ Stop Recording" or press Cmd+Shift+R again (or double-tap Shift)
   - Click to insert text into focused input field

2. **Live Streaming Mode:**
   - Click "🔴 Live Mode"
   - **NEW: Double-tap Control** for quick activation
   - Speak continuously
   - Text appears in real-time at cursor position
   - Click "⏹️ Stop Live" to end (or double-tap Control)

3. **Configuration:**
   - Click "Settings" to configure:
     - **Language**: Auto-detection or specific language (English, French, etc.)
     - **Model Size**: Balance between speed and accuracy
     - **VAD Parameters**: Speech detection sensitivity
     - **Paste Behavior**: Clipboard method and timing
     - **Double Key Shortcuts**: Enable/disable and timing (default 500ms)

### CLI Interface

```bash
# Traditional CLI usage
vibe --cli

# Interactive prompts guide you through:
# - Press ENTER to start recording
# - Speak into microphone
# - Press ENTER to stop and transcribe
# - Choose to insert text or copy to clipboard
```

## 🗂️ Project Structure

```
speech-to-text/
├── install.sh                    # Global installation script  
├── uninstall.sh                  # Global uninstallation script
├── download_whisper_models.sh    # Whisper.cpp model download script
├── main.py                       # GUI application entry point
├── main_cli.py                   # CLI version 
├── test_components.py            # Component testing script
├── requirements.txt              # Python dependencies
├── config.json                   # User configuration (auto-generated)
├── src/
│   ├── __init__.py
│   ├── gui.py                   # Main tkinter GUI with settings
│   ├── audio_recorder.py        # VAD-based audio recording  
│   ├── transcription.py         # Multilingual whisper engine with provider support
│   ├── text_inserter.py         # Clipboard-based text insertion
│   ├── vad_chunker.py           # Voice Activity Detection
│   ├── double_key_shortcuts.py  # Double key press detection
│   ├── config.py                # Configuration management
│   └── providers/               # Whisper provider implementations
│       ├── __init__.py          # Provider availability checks
│       ├── base_provider.py     # Abstract base provider class
│       ├── faster_whisper_provider.py  # Default GPU-optimized provider
│       └── whisper_cpp_provider.py     # CPU-optimized CLI provider
└── venv/                        # Virtual environment
```

## 🔄 Management Commands

### Installation Management
```bash
# Install globally (run from project directory)
./install.sh

# Check installation status  
vibe --version

# Uninstall globally
./uninstall.sh                    # Interactive removal
./uninstall.sh --force           # Force removal
```

### Troubleshooting Installation
```bash
# Reinstall if issues occur
./uninstall.sh --force
./install.sh

# Check if vibe is in PATH
which vibe

# Manual verification
ls -la ~/.local/bin/vibe*
```

## 🔧 Dependencies

### Core Dependencies
- **sounddevice**: High-quality audio recording from microphone
- **faster-whisper**: Offline speech recognition with multilingual support (default provider)
- **webrtcvad**: Voice Activity Detection for intelligent chunking
- **pyperclip**: Clipboard operations for text insertion
- **pynput**: Global hotkey support
- **soundfile**: Audio file processing
- **tkinter**: GUI framework (system dependent)

### Additional Dependencies  
- **pyautogui**: Fallback text insertion method
- **numpy**: Audio data processing

### Optional: Whisper.cpp Provider
For CPU-optimized transcription using the whisper.cpp implementation:

```bash
# Install whisper.cpp via Homebrew (macOS)
brew install whisper-cpp

# Download Whisper models (required for whisper.cpp)
# See "Whisper.cpp Model Setup" section below
```

## 🤖 Whisper.cpp Model Setup

The application supports two Whisper providers:
- **faster-whisper** (default): GPU-optimized, models downloaded automatically
- **whisper.cpp**: CPU-optimized, requires manual model download

### Installing Whisper.cpp Provider

1. **Install whisper.cpp:**
   ```bash
   # macOS (Homebrew)
   brew install whisper-cpp
   
   # Linux (build from source)
   git clone https://github.com/ggerganov/whisper.cpp.git
   cd whisper.cpp
   make
   ```

2. **Download Whisper Models:**
   
   The application will automatically look for models in these locations:
   - `~/whisper-models/` (recommended)
   - `/opt/homebrew/share/whisper-cpp/`
   - `~/.whisper-cpp/models/`
   
   **Quick Setup - Automated Model Download:**
   ```bash
   # Use the included download script (recommended)
   ./download_whisper_models.sh
   
   # Or download specific models directly
   ./download_whisper_models.sh small-en base-en
   
   # See available models
   ./download_whisper_models.sh list
   ```
   
   **Manual Download (Alternative):**
   ```bash
   # Create models directory
   mkdir -p ~/whisper-models
   cd ~/whisper-models
   
   # Download English-optimized models (faster)
   curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin
   curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin
   
   # Download multilingual models (supports all languages)
   curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin
   curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
   ```

### Available Model Sizes

| Model | File Size | Speed | Accuracy | Languages |
|-------|-----------|--------|----------|-----------|
| `ggml-tiny.en.bin` | ~39 MB | Fastest | Basic | English only |
| `ggml-base.en.bin` | ~148 MB | Fast | Good | English only |
| `ggml-small.en.bin` | ~488 MB | Medium | Better | English only |
| `ggml-tiny.bin` | ~39 MB | Fastest | Basic | Multilingual |
| `ggml-base.bin` | ~148 MB | Fast | Good | Multilingual |
| `ggml-small.bin` | ~488 MB | Medium | Better | Multilingual |
| `ggml-medium.bin` | ~1.5 GB | Slow | Great | Multilingual |
| `ggml-large-v3.bin` | ~3.1 GB | Slowest | Best | Multilingual |

### Model Download Script Usage

The included `download_whisper_models.sh` script simplifies model management:

```bash
# Interactive mode - shows available models and prompts for selection
./download_whisper_models.sh

# Download specific models
./download_whisper_models.sh small-en base-en tiny

# Download all models (~6GB total)
./download_whisper_models.sh all

# List available models
./download_whisper_models.sh list

# Show currently downloaded models
./download_whisper_models.sh status

# Get help
./download_whisper_models.sh help
```

**Script Features:**
- ✅ Interactive model selection with progress bars
- ✅ Automatic model directory creation (`~/whisper-models`)
- ✅ Overwrite protection (prompts before replacing existing files)
- ✅ Download verification and error handling
- ✅ Compatible with bash 3.2+ (macOS default)
- ✅ Shows model sizes, speeds, and language support

### Manual Model Download URLs

If you prefer manual download, all models are available from the Hugging Face repository:
```bash
# Base URL
https://huggingface.co/ggerganov/whisper.cpp/resolve/main/

# Examples:
curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin
curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin
```

### Using the Whisper.cpp Provider

1. **In GUI**: Go to Settings → Whisper Provider → Select "Whisper.cpp"
2. **Automatic Fallback**: If your requested model isn't found, the app automatically falls back to available models
3. **Model Priority**: The app searches for models in this order:
   - Exact model requested
   - Available English models (if language is English)
   - Available multilingual models
   - Test model (last resort)

### Troubleshooting Whisper.cpp

**"Whisper-cli command not found":**
```bash
# Check if whisper-cli is installed
which whisper-cli

# macOS: Install via Homebrew
brew install whisper-cpp
```

**"No speech detected" with proper audio:**
```bash
# Verify model file exists and is valid
ls -la ~/whisper-models/
file ~/whisper-models/ggml-small.en.bin

# Test with whisper-cli directly
echo "test" | whisper-cli -m ~/whisper-models/ggml-small.en.bin
```

**Model not found errors:**
```bash
# Check current models
ls -la ~/whisper-models/

# Download missing model
cd ~/whisper-models
curl -L -O https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin
```

## 🔒 Privacy & Security

**Complete Privacy-focused Design:**
- **100% Offline Processing**: All transcription happens locally using faster-whisper
- **No Internet Required**: No data sent to external services (Google, OpenAI, etc.)
- **No Audio Storage**: Temporary audio files deleted immediately after processing
- **Clipboard Protection**: Original clipboard content automatically backed up and restored
- **Local Configuration**: All settings stored locally in `config.json`

## 🛠️ Troubleshooting

### Installation Issues
```bash
# If vibe command not found
echo $PATH | grep "local/bin"  # Check if ~/.local/bin is in PATH
export PATH="$HOME/.local/bin:$PATH"  # Add to PATH temporarily
# Add to ~/.bashrc or ~/.zshrc for permanent fix

# If installation fails
./uninstall.sh --force
rm -rf venv/
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
./install.sh
```

### Microphone Issues
- **macOS**: Grant microphone permissions in System Preferences → Security & Privacy → Microphone
- **Windows**: Check microphone permissions in Settings → Privacy → Microphone
- **Linux**: Ensure user is in audio group: `sudo usermod -a -G audio $USER`

### Text Insertion Issues
- **macOS**: Grant accessibility permissions in System Preferences → Security & Privacy → Accessibility
- **Windows**: Run as administrator if clipboard operations fail
- **Linux**: Install xclip for clipboard support: `sudo apt-get install xclip`

### VAD (Voice Activity Detection) Issues
- **Sensitivity**: Adjust VAD aggressiveness in Settings (0=least sensitive, 3=most sensitive)
- **Background Noise**: Use higher aggressiveness (2-3) in noisy environments
- **Quiet Speech**: Use lower aggressiveness (0-1) for quiet or distant speech

### Language Detection Issues
- **Auto-detection Problems**: Switch to specific language in Settings instead of "Automatic Detection"
- **Mixed Languages**: Use "Automatic Detection" mode for conversations switching between languages
- **Model Loading**: First run may be slow as faster-whisper downloads language models

### Performance Issues
- **Slow Transcription**: Switch to "small" model size in Settings for better speed
- **High Accuracy Needed**: Switch to "large-v3" model (requires more RAM and processing time)
- **Memory Issues**: Restart the application if it becomes unresponsive after extended use

## 🚀 Advanced Features Implemented

✅ **Offline Speech Recognition**: Complete privacy with faster-whisper  
✅ **Multiple Language Support**: 11+ languages with auto-detection  
✅ **Real-time Transcription**: Live streaming mode with instant text insertion  
✅ **Voice Activity Detection**: Intelligent chunking for 60-80% performance improvement  
✅ **Advanced Text Insertion**: Clipboard-based with backup/restore  
✅ **Double Key Press Shortcuts**: Quick activation with double Shift/Control  
✅ **Global Command Access**: Install once, use `vibe` from anywhere  
✅ **Comprehensive Settings**: Language, model size, VAD parameters, paste behavior  

## 🔮 Future Enhancements

- [ ] Audio playback for verification before insertion
- [ ] Transcription history and session management
- [ ] System tray integration for background operation
- [ ] Plugin system for custom text processing
- [ ] Batch file transcription mode
- [ ] Real-time translation between languages
- [ ] Voice commands for application control
- [ ] Custom hotkey configuration in GUI (partially implemented with double key shortcuts)

## License

MIT License - see LICENSE file for details.

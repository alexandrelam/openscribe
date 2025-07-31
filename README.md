# Vibe Speech-to-Text

An advanced real-time speech-to-text desktop application with Voice Activity Detection, multilingual support, and intelligent clipboard-based text insertion.

## ✨ Features

### 🎯 Core Capabilities
✅ **Advanced Voice Recognition**: Offline transcription using faster-whisper models  
✅ **11+ Language Support**: Auto-detection or specific language selection (English, French, Spanish, German, etc.)  
✅ **Voice Activity Detection (VAD)**: Intelligent chunking based on natural speech boundaries  
✅ **Real-time Transcription**: Live streaming mode with instant text insertion  
✅ **Smart Text Insertion**: Clipboard-based pasting with backup/restore functionality  

### 🚀 Performance Features
✅ **60-80% Performance Improvement**: VAD skips silence automatically  
✅ **Natural Speech Boundaries**: No mid-word cuts or awkward breaks  
✅ **Configurable Models**: Choose between speed (small) and accuracy (large-v3)  
✅ **Global Command Access**: Install once, use `vibe` from anywhere  

### ⚙️ Advanced Settings
✅ **Comprehensive Configuration**: Language, VAD parameters, paste behavior  
✅ **Real-time Language Display**: Shows detected language and confidence  
✅ **Multiple Paste Methods**: AppleScript (primary) with keyboard fallback  
✅ **Double Key Press Shortcuts**: Quick mode activation with double Shift/Control  
✅ **International Keyboard Support**: Proper handling of French AZERTY, etc.  

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
├── install.sh            # Global installation script  
├── uninstall.sh          # Global uninstallation script
├── main.py               # GUI application entry point
├── main_cli.py           # CLI version 
├── test_components.py    # Component testing script
├── requirements.txt      # Python dependencies
├── config.json           # User configuration (auto-generated)
├── src/
│   ├── __init__.py
│   ├── gui.py           # Main tkinter GUI with settings
│   ├── audio_recorder.py  # VAD-based audio recording  
│   ├── transcription.py   # Multilingual whisper engine
│   ├── text_inserter.py   # Clipboard-based text insertion
│   ├── vad_chunker.py     # Voice Activity Detection
│   ├── double_key_shortcuts.py  # Double key press detection
│   └── config.py          # Configuration management
└── venv/                # Virtual environment
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
- **faster-whisper**: Offline speech recognition with multilingual support
- **webrtcvad**: Voice Activity Detection for intelligent chunking
- **pyperclip**: Clipboard operations for text insertion
- **pynput**: Global hotkey support
- **soundfile**: Audio file processing
- **tkinter**: GUI framework (system dependent)

### Additional Dependencies  
- **pyautogui**: Fallback text insertion method
- **numpy**: Audio data processing

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
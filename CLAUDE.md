# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Installation & Usage

### Global Installation (Recommended)
Install Vibe globally to use the `vibe` command from anywhere:
```bash
./install.sh
```

After installation, you can use these commands from any terminal:
```bash
vibe              # Launch GUI interface (default)
vibe --cli        # Launch CLI interface  
vibe-cli          # Direct CLI access
vibe --help       # Show help information
vibe --version    # Show version info
```

### Uninstallation
To remove the global installation:
```bash
./uninstall.sh                    # Interactive removal
./uninstall.sh --force           # Force removal without confirmation
```

### Development Environment

#### Virtual Environment Setup
For development, always activate the virtual environment:
```bash
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

#### Essential Development Commands
- **Run GUI application**: `python main.py`
- **Run CLI version**: `python main_cli.py` 
- **Test components**: `python test_components.py`
- **Install dependencies**: `pip install -r requirements.txt`

#### Important Notes
- Development execution must use the virtual environment (`venv/`)
- Configuration is automatically saved to `config.json` in the project root
- The application requires microphone and accessibility permissions on macOS

## Architecture Overview

This is a real-time speech-to-text desktop application with both traditional recording and live streaming transcription modes.

### Core Components

**TranscriptionEngine** (`src/transcription.py`)
- Uses faster-whisper with configurable multilingual models
- **Language Support**: 11+ languages including English, French, Spanish, German, etc.
- **Automatic Detection**: Can auto-detect spoken language or use specific language
- **Model Selection**: English-only (`small.en`) for performance or multilingual (`small`) for versatility
- **VAD Integration**: Works with Voice Activity Detection for natural speech boundaries
- Simplified text assembly system for streaming mode (no complex overlap detection needed)

**AudioRecorder** (`src/audio_recorder.py`)
- **VAD-Based Chunking**: Intelligent 1-10 second variable chunks based on speech boundaries
- **60-80% Performance Improvement**: Only processes audio containing actual speech
- Configurable microphone device selection and VAD parameters
- Separate processing threads for real-time performance

**TextInserter** (`src/text_inserter.py`)  
- **Clipboard-Based Pasting**: Instant text insertion via clipboard (replaces slow character typing)
- **Clipboard-Safe**: Backs up and restores user's original clipboard content
- **Dual Methods**: AppleScript paste (primary) with keyboard shortcut fallback
- **Live Pasting**: Real-time pasting during streaming mode
- Thread-safe queuing system with configurable paste intervals

**GUI** (`src/gui.py`)
- Two main modes: Traditional recording vs Live streaming  
- **Enhanced Settings**: Language selection, model size, VAD parameters, paste configuration
- **Real-time Language Display**: Shows detected language and confidence during transcription
- Real-time status updates and visual feedback
- Proper cleanup on app exit

### Data Flow

**Traditional Mode**: User clicks record → Audio captured → Whisper transcription → Clipboard paste on click

**Live Streaming Mode**: VAD detects speech → Natural boundary chunks → Real-time transcription → Live clipboard pasting

### Critical Implementation Details

**VAD-Based Processing**: The system uses Voice Activity Detection for intelligent audio processing:
1. **WebRTC VAD**: Real-time speech boundary detection with configurable aggressiveness
2. **Natural Chunking**: Variable 1-10 second chunks based on actual speech patterns
3. **Silence Skipping**: 60-80% performance improvement by only processing speech
4. **Smart Boundaries**: No mid-word cuts or awkward audio breaks

**Clipboard-Based Text Insertion**: Advanced pasting system for maximum speed and reliability:
1. **Clipboard Backup**: Preserves user's original clipboard content
2. **AppleScript Integration**: Primary method for maximum compatibility
3. **Fallback System**: Keyboard shortcuts if AppleScript fails
4. **Live Pasting**: Real-time insertion during streaming with configurable intervals

**Multilingual Architecture**: Sophisticated language handling system:
1. **Model Selection**: Automatic English-only vs multilingual model switching
2. **Language Detection**: Real-time detection with confidence scoring
3. **Configuration Persistence**: Language settings saved and restored
4. **Legacy Migration**: Automatic upgrade of old language configurations

**Threading Architecture**: Multiple coordinated threads:
- Audio capture thread (sounddevice callback)
- VAD processing thread (webrtcvad analysis)
- Transcription processing thread (serialized)
- Live pasting thread (clipboard-based insertion)
- GUI main thread (user interface)

## Configuration

Config automatically persists to `config.json` with these key settings:

### Audio & Device Settings
- `microphone_device`: Selected audio input device ID
- `hotkey`: Global hotkey (default: "cmd+shift+r")

### Language & Transcription Settings  
- `transcription_language`: Language code ("auto", "en", "fr", etc.)
- `model_size`: Whisper model size ("small", "base", "large-v3")
- `language_detection_enabled`: Show real-time language detection info
- `fallback_language`: Default language when detection fails

### VAD (Voice Activity Detection) Settings
- `vad_aggressiveness`: Speech detection sensitivity (0-3)
- `vad_min_chunk_duration`: Minimum chunk length in seconds
- `vad_max_chunk_duration`: Maximum chunk length in seconds  
- `vad_silence_timeout`: Silence duration before processing chunk

### Text Insertion Settings
- `enable_auto_insert`: Whether to enable click-to-insert mode
- `auto_insert_timeout`: Timeout for click-to-insert mode
- `paste_method`: Paste method ("applescript" or "keyboard")
- `paste_delay`: Delay after copying to clipboard
- `live_paste_interval`: Interval between live paste operations
- `restore_clipboard`: Whether to restore original clipboard content

## Testing Considerations

- Component testing available via `test_components.py`
- **VAD Testing**: Voice Activity Detection requires actual speech input with varying silence patterns
- **Multilingual Testing**: Test language detection and switching with French, English, and other supported languages
- **Clipboard Testing**: Verify clipboard backup/restore functionality across different applications
- **Audio Device Testing**: Test microphone switching with multiple input devices
- **Installation Testing**: Verify global installation and uninstallation scripts work correctly
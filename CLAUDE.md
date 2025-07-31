# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Environment

- **Run GUI**: `python main.py`
- **Run CLI**: `python main_cli.py` 
- **Install deps**: `pip install -r requirements.txt`
- **Virtual env**: `source venv/bin/activate`
- **Format code**: `black *.py src/*.py`
- **Config**: Automatically saved to `config.json`
- **Permissions**: Requires microphone and accessibility permissions on macOS

## Architecture Overview

This is a real-time speech-to-text desktop application with both traditional recording and live streaming transcription modes.

### Core Components

**TranscriptionEngine** (`src/transcription.py`)
- Uses faster-whisper with multilingual models (11+ languages)
- Auto-detect or specific language selection
- VAD integration for natural speech boundaries

**AudioRecorder** (`src/audio_recorder.py`)
- VAD-based chunking (1-10 second variable chunks)
- Only processes audio containing speech (60-80% performance improvement)
- Configurable microphone device selection

**TextInserter** (`src/text_inserter.py`)  
- Clipboard-based text insertion with backup/restore
- AppleScript paste (primary) with keyboard shortcuts fallback
- Live pasting during streaming mode

**RecordingIndicator** (`src/recording_indicator.py`)
- Fixed-position visual indicator for recording states
- Shows Recording (red dot), Live Mode (pulsing), Processing (spinning)
- Configurable position, size, and opacity

**GUI** (`src/gui.py`)
- Traditional recording and live streaming modes
- Settings dialog for all configuration options
- Real-time language display and status updates

**DoubleKeyDetector** (`src/double_key_shortcuts.py`)
- Double Shift → Toggle recording mode
- Double Control → Toggle live mode
- Configurable timing window (default 500ms)

### Data Flow

**Traditional Mode**: User clicks record → Audio captured → Whisper transcription → Clipboard paste on click

**Live Streaming Mode**: VAD detects speech → Natural boundary chunks → Real-time transcription → Live clipboard pasting

### Threading Architecture
- Audio capture thread (sounddevice callback)
- VAD processing thread (webrtcvad analysis)  
- Transcription processing thread (serialized)
- Live pasting thread (clipboard-based insertion)
- GUI main thread (user interface)

## Configuration

Config automatically persists to `config.json`:

### Core Settings
- `microphone_device`: Audio input device ID
- `hotkey`: Global hotkey (default: "cmd+shift+r")
- `transcription_language`: Language code ("auto", "en", "fr", etc.)
- `model_size`: Whisper model size ("small", "base", "large-v3")

### VAD Settings
- `vad_aggressiveness`: Speech detection sensitivity (0-3)
- `vad_min_chunk_duration`: Minimum chunk length in seconds
- `vad_max_chunk_duration`: Maximum chunk length in seconds  

### Text Insertion Settings
- `enable_auto_insert`: Click-to-insert mode
- `paste_method`: "applescript" or "keyboard"
- `live_paste_interval`: Interval between live paste operations

### Recording Indicator Settings
- `show_recording_indicator`: Enable/disable visual indicator
- `indicator_position_x`: Screen X position (pixels)
- `indicator_position_y`: Screen Y position (pixels)
- `indicator_size`: Indicator size (pixels)
- `indicator_opacity`: Transparency (0.0-1.0)

### Double Key Shortcuts Settings
- `double_press_enabled`: Enable/disable double key shortcuts
- `double_press_timeout`: Time window for double press detection (seconds)


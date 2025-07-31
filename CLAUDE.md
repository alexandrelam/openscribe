# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Environment

### Virtual Environment Setup
Always activate the virtual environment before running any commands:
```bash
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

### Essential Commands
- **Run GUI application**: `python main.py`
- **Run CLI version**: `python main_cli.py` 
- **Test components**: `python test_components.py`
- **Install dependencies**: `pip install -r requirements.txt`

### Important Notes
- All Python execution must use the virtual environment (`venv/`)
- Configuration is automatically saved to `config.json` in the project root
- The application requires microphone and accessibility permissions on macOS

## Architecture Overview

This is a real-time speech-to-text desktop application with both traditional recording and live streaming transcription modes.

### Core Components

**TranscriptionEngine** (`src/transcription.py`)
- Uses faster-whisper "small.en" model for offline speech recognition  
- Supports both batch transcription and real-time streaming
- Implements sophisticated text assembly system for streaming mode:
  - Serialized chunk processing to prevent race conditions
  - Sequence-based overlap detection using word-level matching
  - Continuous text buffer that assembles overlapping audio chunks
  - Smart deduplication to prevent repeated words/phrases

**AudioRecorder** (`src/audio_recorder.py`)
- Handles both traditional recording and streaming audio capture
- Streaming mode: 3-second chunks with 1.5-second overlap for continuity
- Configurable microphone device selection
- Separate processing threads for real-time performance

**TextInserter** (`src/text_inserter.py`)  
- Two insertion modes:
  - **Auto-insert**: Click-to-insert after traditional recording
  - **Live typing**: Real-time typing during streaming mode
- Uses AppleScript (macOS) for proper keyboard layout support (handles French AZERTY, etc.)
- Implements text cleaning and preprocessing for international keyboards
- Thread-safe queuing system for live typing

**GUI** (`src/gui.py`)
- Two main modes: Traditional recording vs Live streaming  
- Settings dialog for microphone selection and auto-insert configuration
- Real-time status updates and visual feedback
- Proper cleanup on app exit

### Data Flow

**Traditional Mode**: User clicks record → Audio captured → Whisper transcription → Auto-insert on click

**Live Streaming Mode**: Continuous audio chunks → Real-time transcription → Text assembly → Live typing at cursor

### Critical Implementation Details

**Text Assembly System**: The streaming mode uses a sophisticated text assembly pipeline to handle overlapping audio chunks:
1. Audio chunks processed serially (not parallel) to prevent race conditions
2. Each transcribed chunk compared against existing text buffer using sequence matching
3. Overlapping portions identified and removed to prevent duplication  
4. Proper spacing inserted between chunks
5. Assembled text sent to live typing system

**Keyboard Layout Handling**: Uses AppleScript for text insertion instead of PyAutoGUI to properly support international keyboard layouts (critical for French AZERTY users).

**Threading Architecture**: Multiple coordinated threads:
- Audio capture thread (sounddevice callback)
- Transcription processing thread (serialized)
- Live typing thread (queued text insertion)
- GUI main thread (user interface)

## Configuration

Config automatically persists to `config.json` with these key settings:
- `microphone_device`: Selected audio input device ID
- `enable_auto_insert`: Whether to enable click-to-insert mode  
- `auto_insert_timeout`: Timeout for click-to-insert mode
- `hotkey`: Global hotkey (default: "cmd+shift+r")

## Testing Considerations

- Component testing available via `test_components.py`
- Live streaming mode requires actual speech input for proper testing
- International keyboard layouts (especially French AZERTY) need specific testing
- Audio device switching should be tested with multiple microphones
- Text assembly deduplication needs testing with overlapping speech patterns
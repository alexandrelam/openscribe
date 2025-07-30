# Speech-to-Text MVP

A lightweight desktop application that converts speech to text and inserts it directly into any focused input field.

## Features

✅ **Voice Recording**: Capture audio from system microphone using sounddevice  
✅ **Speech Transcription**: Convert audio to text using Google Speech Recognition  
✅ **Text Insertion**: Automatically insert transcribed text using pyautogui  
✅ **Global Hotkey**: Cmd+Shift+R to start/stop recording  
✅ **Visual Feedback**: Recording status and processing indicators  
✅ **Error Handling**: Microphone access, permissions, and transcription failure handling  
✅ **Cross-platform**: Works on macOS, Windows, and Linux  

## Installation

1. **Clone and setup virtual environment:**
   ```bash
   cd speech-to-text
   python3 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   pip install -r requirements.txt
   ```

2. **Install tkinter (for GUI version):**
   ```bash
   # On macOS
   brew install python-tk
   
   # On Ubuntu/Debian
   sudo apt-get install python3-tk
   
   # On Windows - usually included with Python
   ```

3. **Test components:**
   ```bash
   python test_components.py
   ```

## Usage

### GUI Version (Recommended)
```bash
python main.py
```

- Click "🎤 Start Recording" or press Cmd+Shift+R
- Speak into your microphone
- Click "⏹️ Stop Recording" or press Cmd+Shift+R again
- Text will be automatically inserted into the focused input field

### CLI Version (Testing)
```bash
python main_cli.py
```

- Press ENTER to start recording
- Speak into your microphone  
- Press ENTER again to stop and transcribe
- Choose to insert text or copy to clipboard

## Project Structure

```
speech-to-text/
├── main.py              # GUI application entry point
├── main_cli.py          # CLI version for testing
├── test_components.py   # Component testing script
├── requirements.txt     # Python dependencies
├── src/
│   ├── __init__.py
│   ├── gui.py          # Main tkinter GUI
│   ├── audio_recorder.py  # Microphone recording
│   ├── transcription.py   # Speech-to-text engine
│   ├── text_inserter.py   # System text insertion
│   └── config.py          # Configuration management
└── venv/               # Virtual environment
```

## Dependencies

- **sounddevice**: Audio recording from microphone
- **SpeechRecognition**: Google Speech Recognition API
- **pyautogui**: System-wide text insertion
- **pynput**: Global hotkey support
- **soundfile**: Audio file processing
- **tkinter**: GUI framework (system dependent)

## Privacy & Security

🔒 **Privacy-focused design:**
- Audio processing happens locally when possible
- Google Speech Recognition used for transcription (requires internet)
- No audio data stored permanently
- Temporary audio files deleted immediately after processing

## Troubleshooting

### Microphone Issues
- **macOS**: Grant microphone permissions in System Preferences → Security & Privacy
- **Windows**: Check microphone permissions in Settings → Privacy
- **Linux**: Ensure user is in audio group

### Text Insertion Issues
- **macOS**: Grant accessibility permissions for text insertion
- **Windows**: Run as administrator if needed
- **Linux**: Install xdotool for better compatibility

### Hotkey Issues
- Default hotkey: Cmd+Shift+R (macOS) / Ctrl+Shift+R (Windows/Linux)
- May conflict with system shortcuts - configurable in future versions

## Future Enhancements

- [ ] Offline speech recognition with Whisper
- [ ] Custom hotkey configuration
- [ ] Multiple language support
- [ ] Audio playback for verification
- [ ] Transcription history
- [ ] System tray integration
- [ ] Real-time transcription

## License

MIT License - see LICENSE file for details.
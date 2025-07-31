#!/usr/bin/env python3

import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from src.audio_recorder import AudioRecorder
from src.transcription import TranscriptionEngine
from src.text_inserter import TextInserter
from src.config import Config


def test_components():
    print("🧪 Testing Speech-to-Text Components")
    print("=" * 40)

    # Test Config
    print("✅ Config module loaded")
    config = Config.load()
    print(f"   Default hotkey: {config.hotkey}")

    # Test AudioRecorder
    print("✅ AudioRecorder module loaded")
    recorder = AudioRecorder()
    devices = recorder.get_available_devices()
    input_devices = [d for d in devices if d["max_input_channels"] > 0]
    print(f"   Found {len(input_devices)} input devices")

    # Test TranscriptionEngine
    print("✅ TranscriptionEngine module loaded")
    transcriber = TranscriptionEngine()
    print("   Speech recognition service ready")

    # Test TextInserter
    print("✅ TextInserter module loaded")
    inserter = TextInserter()
    print("   Text insertion service ready")

    print("\n🎉 All components loaded successfully!")
    print("\nNext steps:")
    print("1. For GUI version: Install tkinter support")
    print("   - On macOS: brew install python-tk")
    print("   - Then run: python main.py")
    print("2. For CLI testing: python main_cli.py")


if __name__ == "__main__":
    test_components()

#!/usr/bin/env python3

import os
import sys
import time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from src.audio_recorder import AudioRecorder
from src.transcription import TranscriptionEngine
from src.text_inserter import TextInserter

def main():
    print("🎤 Speech-to-Text MVP (CLI Version)")
    print("=" * 40)
    
    audio_recorder = AudioRecorder()
    transcription_engine = TranscriptionEngine()
    text_inserter = TextInserter()
    
    print("Available audio devices:")
    devices = audio_recorder.get_available_devices()
    for i, device in enumerate(devices):
        if device['max_input_channels'] > 0:
            print(f"  {i}: {device['name']}")
    
    print("\nInstructions:")
    print("- Press ENTER to start recording")
    print("- Press ENTER again to stop recording and transcribe")
    print("- Type 'quit' to exit")
    print()
    
    try:
        while True:
            command = input("Ready (press ENTER to record): ").strip().lower()
            
            if command == 'quit':
                break
            
            if command == '':
                print("🔴 Recording... Speak now!")
                if audio_recorder.start_recording():
                    input("Press ENTER to stop recording...")
                    print("⏹️ Stopping recording...")
                    
                    audio_data = audio_recorder.stop_recording()
                    if audio_data is not None and len(audio_data) > 0:
                        print("🔄 Transcribing...")
                        text = transcription_engine.transcribe_audio(audio_data)
                        
                        if text:
                            print(f"✅ Transcribed: '{text}'")
                            
                            choice = input("Insert text (i) or copy to clipboard (c)? [i/c]: ").strip().lower()
                            if choice == 'c':
                                if text_inserter.copy_to_clipboard(text):
                                    print("📋 Text copied to clipboard!")
                                else:
                                    print("❌ Failed to copy to clipboard")
                            else:
                                print("Positioning cursor... (3 seconds)")
                                time.sleep(3)
                                if text_inserter.insert_text(text):
                                    print("✅ Text inserted!")
                                else:
                                    print("❌ Failed to insert text")
                        else:
                            print("❌ No speech detected or transcription failed")
                    else:
                        print("❌ No audio data recorded")
                else:
                    print("❌ Failed to start recording. Check microphone permissions.")
            
            print()
    
    except KeyboardInterrupt:
        print("\n👋 Goodbye!")
    except Exception as e:
        print(f"❌ Error: {e}")

if __name__ == "__main__":
    main()
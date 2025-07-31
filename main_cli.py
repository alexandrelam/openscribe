#!/usr/bin/env python3

from src.double_key_shortcuts import DoubleKeyDetector
from src.text_inserter import TextInserter
from src.transcription import TranscriptionEngine
from src.audio_recorder import AudioRecorder
from src.sound_notifications import SoundNotifications
from src.config import Config
import os
import sys
import time
import threading
from enum import Enum

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))


class RecordingState(Enum):
    IDLE = "idle"
    RECORDING = "recording"
    PROCESSING = "processing"
    READY_TO_PASTE = "ready_to_paste"


class SpeechToTextCLI:
    def __init__(self):
        self.config = Config.load()
        self.audio_recorder = AudioRecorder()
        self.transcription_engine = TranscriptionEngine()
        self.text_inserter = TextInserter()
        self.double_key_detector = DoubleKeyDetector(double_press_timeout=0.5)
        self.sound_notifications = SoundNotifications(
            enabled=self.config.sound_notifications_enabled,
            volume=self.config.sound_volume
        )

        self.state = RecordingState.IDLE
        self.state_lock = threading.Lock()
        self.last_transcribed_text = None
        self.running = True

        # Set up double key callbacks
        self.double_key_detector.set_callbacks(
            on_double_shift=self._on_double_shift)

    def _print_status(self, message: str):
        """Print status message with current state"""
        state_symbols = {
            RecordingState.IDLE: "⏺️",
            RecordingState.RECORDING: "🔴",
            RecordingState.PROCESSING: "🔄",
            RecordingState.READY_TO_PASTE: "📋"
        }
        symbol = state_symbols.get(self.state, "⏺️")
        print(f"{symbol} {message}")

    def _set_state(self, new_state: RecordingState):
        """Thread-safe state change"""
        try:
            # Try to acquire the lock with a very short timeout
            # If we can't get it, we probably already have it
            acquired = self.state_lock.acquire(blocking=False)

            if self.state != new_state:
                self.state = new_state

            if acquired:
                self.state_lock.release()

        except Exception as e:
            # Fallback - just change the state
            if self.state != new_state:
                self.state = new_state

    def _on_double_shift(self):
        """Handle double-shift key press to toggle recording"""
        with self.state_lock:
            if self.state == RecordingState.IDLE:
                self._start_recording()
            elif self.state == RecordingState.RECORDING:
                self._stop_recording()
            elif self.state == RecordingState.READY_TO_PASTE:
                # Reset to idle if user presses double-shift while waiting to paste
                self.text_inserter.stop_auto_insert_mode()
                self._set_state(RecordingState.IDLE)
                self._print_status(
                    "Ready - Double-press Shift to start recording")

    def _start_recording(self):
        """Start recording audio"""
        try:
            if self.audio_recorder.start_recording():
                self._set_state(RecordingState.RECORDING)
                self.sound_notifications.play_start_recording()
                self._print_status(
                    "Recording... Speak now! (Double-press Shift again to stop)")
            else:
                self._print_status(
                    "❌ Failed to start recording. Check microphone permissions.")
        except Exception as e:
            self._print_status(f"❌ Error starting recording: {e}")

    def _stop_recording(self):
        """Stop recording and start transcription"""
        try:
            self._set_state(RecordingState.PROCESSING)
            self.sound_notifications.play_stop_recording()
            self._print_status("Processing...")

            # Stop recording and get audio data
            audio_data = self.audio_recorder.stop_recording()

            if audio_data is not None and len(audio_data) > 0:
                # Transcribe in background thread to avoid blocking
                threading.Thread(target=self._transcribe_audio,
                                 args=(audio_data,), daemon=True).start()
            else:
                self._print_status("❌ No audio data recorded")
                self._set_state(RecordingState.IDLE)
                self._print_status(
                    "Ready - Double-press Shift to start recording")

        except Exception as e:
            self._print_status(f"❌ Error stopping recording: {e}")
            self._set_state(RecordingState.IDLE)
            self._print_status("Ready - Double-press Shift to start recording")

    def _transcribe_audio(self, audio_data):
        """Transcribe audio data and prepare for pasting"""
        try:
            text = self.transcription_engine.transcribe_audio(audio_data)

            if text and text.strip():
                self.last_transcribed_text = text.strip()
                self.sound_notifications.play_transcription_ready()
                self._print_status(
                    f"✅ Transcribed: '{self.last_transcribed_text}'")

                # Start auto-insert mode for click-to-paste
                success = self.text_inserter.start_auto_insert_mode(
                    self.last_transcribed_text,
                    timeout_seconds=30,  # 30 second timeout
                    on_complete=self._on_paste_complete
                )

                if success:
                    self._set_state(RecordingState.READY_TO_PASTE)
                    self._print_status(
                        "Click anywhere to paste text (30s timeout)")
                else:
                    self._print_status("❌ Failed to prepare text for pasting")
                    self._set_state(RecordingState.IDLE)
            else:
                self._print_status(
                    "❌ No speech detected or transcription failed")
                self._set_state(RecordingState.IDLE)

        except Exception as e:
            self._print_status(f"❌ Transcription error: {e}")
            self._set_state(RecordingState.IDLE)

    def _on_paste_complete(self, success: bool):
        """Handle completion of paste operation"""
        if success:
            self._print_status("✅ Text pasted successfully!")
        else:
            self._print_status("⏰ Paste timeout or failed")

        self._set_state(RecordingState.IDLE)
        self._print_status("Ready - Double-press Shift to start recording")

    def start(self):
        """Start the CLI application"""
        print("🎤 Speech-to-Text MVP (CLI Version)")
        print("=" * 40)

        # Show available audio devices
        print("Available audio devices:")
        devices = self.audio_recorder.get_available_devices()
        for i, device in enumerate(devices):
            if device['max_input_channels'] > 0:
                print(f"  {i}: {device['name']}")

        print("\nInstructions:")
        print("- Double-press Shift to start/stop recording")
        print("- After transcription, click anywhere to paste text")
        print("- Type 'quit' to exit")
        print()

        print(
            "📋 Status indicators: ⏺️ Ready | 🔴 Recording | 🔄 Processing | 📋 Click to paste")
        print("💡 For visual screen indicator, use GUI mode: python main.py")
        print()

        print("⚠️  IMPORTANT: If double-Shift doesn't work, make sure this app has")
        print("   Accessibility permissions in System Preferences > Security & Privacy")
        print()

        # Start the double key detector
        try:
            self.double_key_detector.start()
            self._print_status("Ready - Double-press Shift to start recording")
        except Exception as e:
            print(f"❌ Failed to start keyboard listener: {e}")
            print("⚠️  Make sure this app has Accessibility permissions:")
            print("   System Preferences > Security & Privacy > Privacy > Accessibility")
            print("   Add Terminal (or your terminal app) to the list")
            self._print_status("Keyboard detection failed - check permissions")

        try:
            # Start input handling in background
            self._start_input_thread()

            # Simple main loop
            while self.running:
                time.sleep(0.1)  # Small sleep to prevent high CPU usage

        except KeyboardInterrupt:
            pass
        finally:
            self.cleanup()

    def _start_input_thread(self):
        """Start input handling in a separate thread to avoid blocking keyboard listener"""
        def input_handler():
            try:
                while self.running:
                    try:
                        command = input().strip().lower()

                        if command == 'quit':
                            self.running = False
                            break
                        elif command == 'status':
                            # Hidden command to check current status
                            with self.state_lock:
                                print(f"Current state: {self.state.value}")
                                if self.last_transcribed_text:
                                    print(
                                        f"Last transcribed: '{self.last_transcribed_text}'")
                        elif command == 'help':
                            print("\nAvailable commands:")
                            print("- quit: Exit the application")
                            print("- status: Show current state")
                            print("- help: Show this help")
                            print("- Double-press Shift: Start/Stop recording")

                    except EOFError:
                        # Handle Ctrl+D
                        self.running = False
                        break

            except Exception as e:
                print(f"Input thread error: {e}")
                self.running = False

        input_thread = threading.Thread(target=input_handler, daemon=True)
        input_thread.start()

    def cleanup(self):
        """Clean up resources"""
        print("\n🛑 Shutting down...")
        self.running = False

        # Stop double key detector
        if self.double_key_detector:
            self.double_key_detector.stop()

        # Stop auto-insert mode if active
        if self.text_inserter and self.text_inserter.is_auto_insert_active():
            self.text_inserter.stop_auto_insert_mode()

        # Stop recording if active
        if self.state == RecordingState.RECORDING:
            try:
                self.audio_recorder.stop_recording()
            except:
                pass

        print("👋 Goodbye!")


def main():
    try:
        cli = SpeechToTextCLI()
        cli.start()
    except Exception as e:
        print(f"❌ Error: {e}")


if __name__ == "__main__":
    main()

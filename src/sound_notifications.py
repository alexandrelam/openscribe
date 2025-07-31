import numpy as np
import sounddevice as sd
import threading
from typing import Optional


class SoundNotifications:
    def __init__(
        self, enabled: bool = True, volume: float = 0.5, sample_rate: int = 44100
    ):
        self.enabled = enabled
        self.volume = max(0.0, min(1.0, volume))  # Clamp volume between 0 and 1
        self.sample_rate = sample_rate

    def _generate_tone(self, frequency: float, duration: float) -> np.ndarray:
        """Generate a simple sine wave tone"""
        samples = int(duration * self.sample_rate)
        t = np.linspace(0, duration, samples, False)
        tone = np.sin(2 * np.pi * frequency * t) * self.volume

        # Apply fade in/out to avoid clicks
        fade_samples = int(0.01 * self.sample_rate)  # 10ms fade
        if samples > fade_samples * 2:
            tone[:fade_samples] *= np.linspace(0, 1, fade_samples)
            tone[-fade_samples:] *= np.linspace(1, 0, fade_samples)

        return tone.astype(np.float32)

    def _generate_chord(self, frequencies: list, duration: float) -> np.ndarray:
        """Generate multiple tones played simultaneously"""
        samples = int(duration * self.sample_rate)
        t = np.linspace(0, duration, samples, False)
        chord = np.zeros(samples)

        for freq in frequencies:
            chord += np.sin(2 * np.pi * freq * t)

        # Normalize and apply volume
        chord = chord / len(frequencies) * self.volume

        # Apply fade in/out to avoid clicks
        fade_samples = int(0.01 * self.sample_rate)  # 10ms fade
        if samples > fade_samples * 2:
            chord[:fade_samples] *= np.linspace(0, 1, fade_samples)
            chord[-fade_samples:] *= np.linspace(1, 0, fade_samples)

        return chord.astype(np.float32)

    def _play_sound_async(self, audio: np.ndarray):
        """Play sound in a separate thread to avoid blocking"""

        def play():
            try:
                sd.play(audio, self.sample_rate)
                sd.wait()  # Wait until sound finishes
            except Exception as e:
                # Silently handle sound errors - don't break the app
                pass

        thread = threading.Thread(target=play, daemon=True)
        thread.start()

    def play_start_recording(self):
        """Play sound when recording starts - higher pitched beep"""
        if not self.enabled:
            return

        tone = self._generate_tone(frequency=800, duration=0.2)
        self._play_sound_async(tone)

    def play_stop_recording(self):
        """Play sound when recording stops - lower pitched beep"""
        if not self.enabled:
            return

        tone = self._generate_tone(frequency=400, duration=0.2)
        self._play_sound_async(tone)

    def play_transcription_ready(self):
        """Play sound when transcription is ready - success chime"""
        if not self.enabled:
            return

        # Create an ascending chord progression for success
        chord1 = self._generate_chord([440, 554.37], 0.1)  # A + C#
        chord2 = self._generate_chord([554.37, 659.25], 0.1)  # C# + E
        chord3 = self._generate_chord([659.25, 880], 0.1)  # E + A (octave)

        # Concatenate the chords
        success_sound = np.concatenate([chord1, chord2, chord3])
        self._play_sound_async(success_sound)

    def set_enabled(self, enabled: bool):
        """Enable or disable sound notifications"""
        self.enabled = enabled

    def set_volume(self, volume: float):
        """Set volume level (0.0 to 1.0)"""
        self.volume = max(0.0, min(1.0, volume))

import threading
import queue
import sounddevice as sd
import numpy as np
import time
from typing import Optional, Callable
from .vad_chunker import VADChunker


class AudioRecorder:
    def __init__(
        self,
        sample_rate: int = 16000,
        channels: int = 1,
        device_id: Optional[int] = None,
    ):
        self.sample_rate = sample_rate
        self.channels = channels
        self.device_id = device_id
        self.recording = False
        self.streaming = False
        self.audio_queue = queue.Queue()
        self.stream: Optional[sd.InputStream] = None

        # VAD-based streaming mode configuration
        self.vad_aggressiveness = 2  # 0-3, higher = more aggressive
        self.min_chunk_duration = 1.0  # minimum seconds per chunk
        self.max_chunk_duration = 10.0  # maximum seconds per chunk
        self.silence_timeout = 0.5  # seconds of silence before processing chunk

        self.streaming_callback: Optional[Callable[[np.ndarray], None]] = None
        self.vad_chunker: Optional[VADChunker] = None
        self.audio_buffer = []
        self.buffer_lock = threading.Lock()
        self.streaming_thread: Optional[threading.Thread] = None
        self.streaming_stop_event = threading.Event()

    def _audio_callback(self, indata, frames, time, status):
        if status:
            print(f"Audio callback status: {status}")

        if self.recording:
            self.audio_queue.put(indata.copy())

        if self.streaming:
            with self.buffer_lock:
                self.audio_buffer.append(indata.copy())

    def start_recording(self) -> bool:
        try:
            self.recording = True
            stream_params = {
                "samplerate": self.sample_rate,
                "channels": self.channels,
                "callback": self._audio_callback,
                "dtype": np.float32,
            }

            # Add device parameter if specified
            if self.device_id is not None:
                stream_params["device"] = self.device_id

            self.stream = sd.InputStream(**stream_params)
            self.stream.start()
            return True
        except Exception as e:
            print(f"Failed to start recording with device {self.device_id}: {e}")
            self.recording = False
            return False

    def stop_recording(self) -> Optional[np.ndarray]:
        if not self.recording:
            return None

        self.recording = False
        if self.stream:
            self.stream.stop()
            self.stream.close()
            self.stream = None

        audio_data = []
        while not self.audio_queue.empty():
            audio_data.append(self.audio_queue.get())

        if audio_data:
            return np.concatenate(audio_data, axis=0)
        return None

    def is_recording(self) -> bool:
        return self.recording

    def start_streaming_recording(self, callback: Callable[[np.ndarray], None]) -> bool:
        """Start streaming recording that processes audio chunks using VAD"""
        try:
            self.streaming = True
            self.streaming_callback = callback
            self.audio_buffer = []
            self.streaming_stop_event.clear()

            # Initialize VAD chunker
            self.vad_chunker = VADChunker(
                sample_rate=self.sample_rate,
                aggressiveness=self.vad_aggressiveness,
                min_chunk_duration=self.min_chunk_duration,
                max_chunk_duration=self.max_chunk_duration,
                silence_timeout=self.silence_timeout,
            )
            self.vad_chunker.set_chunk_callback(callback)

            # Start audio stream
            stream_params = {
                "samplerate": self.sample_rate,
                "channels": self.channels,
                "callback": self._audio_callback,
                "dtype": np.float32,
            }

            if self.device_id is not None:
                stream_params["device"] = self.device_id

            self.stream = sd.InputStream(**stream_params)
            self.stream.start()

            # Start streaming processing thread
            self.streaming_thread = threading.Thread(
                target=self._streaming_processor, daemon=True
            )
            self.streaming_thread.start()

            return True
        except Exception as e:
            print(
                f"Failed to start streaming recording with device {self.device_id}: {e}"
            )
            self.streaming = False
            return False

    def stop_streaming_recording(self):
        """Stop streaming recording"""
        if not self.streaming:
            return

        self.streaming = False
        self.streaming_stop_event.set()

        if self.stream:
            self.stream.stop()
            self.stream.close()
            self.stream = None

        if self.streaming_thread and self.streaming_thread.is_alive():
            self.streaming_thread.join(timeout=2.0)

        self.streaming_callback = None

        # Force process any remaining audio and cleanup VAD
        if self.vad_chunker:
            self.vad_chunker.force_chunk()
            self.vad_chunker = None

        self.audio_buffer = []

    def _streaming_processor(self):
        """Process audio using VAD-based chunking for streaming transcription"""
        print("🎤 Starting VAD-based streaming processor")

        # Process interval for checking audio buffer
        process_interval = 0.1  # Check every 100ms
        last_process_time = time.time()

        while not self.streaming_stop_event.is_set():
            current_time = time.time()

            # Check if it's time to process buffered audio
            if current_time - last_process_time >= process_interval:
                with self.buffer_lock:
                    if self.audio_buffer and self.vad_chunker:
                        # Combine all buffered audio
                        combined_audio = np.concatenate(self.audio_buffer, axis=0)

                        # Process through VAD chunker
                        try:
                            self.vad_chunker.process_audio(combined_audio.flatten())
                        except Exception as e:
                            print(f"❌ Error in VAD processing: {e}")

                        # Clear processed buffer
                        self.audio_buffer = []

                last_process_time = current_time

            # Small sleep to prevent excessive CPU usage
            time.sleep(0.05)

    def is_streaming(self) -> bool:
        return self.streaming

    def configure_vad(
        self,
        aggressiveness: int = 2,
        min_chunk_duration: float = 1.0,
        max_chunk_duration: float = 10.0,
        silence_timeout: float = 0.5,
    ):
        """Configure VAD parameters"""
        self.vad_aggressiveness = aggressiveness
        self.min_chunk_duration = min_chunk_duration
        self.max_chunk_duration = max_chunk_duration
        self.silence_timeout = silence_timeout

        print(
            f"🔧 VAD configured: aggressiveness={aggressiveness}, "
            f"chunk_range={min_chunk_duration}-{max_chunk_duration}s, "
            f"silence_timeout={silence_timeout}s"
        )

    def get_vad_stats(self) -> dict:
        """Get current VAD statistics"""
        if self.vad_chunker:
            return self.vad_chunker.get_stats()
        return {}

    @staticmethod
    def get_available_devices():
        return sd.query_devices()

    @staticmethod
    def get_input_devices():
        """Get only input devices that can be used for recording"""
        devices = sd.query_devices()
        input_devices = []
        for i, device in enumerate(devices):
            if device["max_input_channels"] > 0:
                input_devices.append(
                    {
                        "index": i,
                        "name": device["name"],
                        "channels": device["max_input_channels"],
                        "default_samplerate": device["default_samplerate"],
                    }
                )
        return input_devices

    @staticmethod
    def get_available_device_ids():
        """Get list of available input device IDs"""
        try:
            devices = sd.query_devices()
            available_ids = []
            for i, device in enumerate(devices):
                if device["max_input_channels"] > 0:
                    available_ids.append(i)
            return available_ids
        except Exception as e:
            print(f"⚠️ Error querying available devices: {e}")
            return []

    @staticmethod
    def is_device_available(device_id: int) -> bool:
        """Check if a specific device ID is available for recording"""
        try:
            available_ids = AudioRecorder.get_available_device_ids()
            return device_id in available_ids
        except Exception:
            return False

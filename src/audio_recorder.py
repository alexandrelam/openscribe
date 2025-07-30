import threading
import queue
import sounddevice as sd
import numpy as np
from typing import Optional, Callable

class AudioRecorder:
    def __init__(self, sample_rate: int = 16000, channels: int = 1):
        self.sample_rate = sample_rate
        self.channels = channels
        self.recording = False
        self.audio_queue = queue.Queue()
        self.stream: Optional[sd.InputStream] = None
        
    def _audio_callback(self, indata, frames, time, status):
        if status:
            print(f"Audio callback status: {status}")
        if self.recording:
            self.audio_queue.put(indata.copy())
    
    def start_recording(self) -> bool:
        try:
            self.recording = True
            self.stream = sd.InputStream(
                samplerate=self.sample_rate,
                channels=self.channels,
                callback=self._audio_callback,
                dtype=np.float32
            )
            self.stream.start()
            return True
        except Exception as e:
            print(f"Failed to start recording: {e}")
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
    
    @staticmethod
    def get_available_devices():
        return sd.query_devices()
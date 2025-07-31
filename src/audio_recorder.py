import threading
import queue
import sounddevice as sd
import numpy as np
import time
from typing import Optional, Callable

class AudioRecorder:
    def __init__(self, sample_rate: int = 16000, channels: int = 1, device_id: Optional[int] = None):
        self.sample_rate = sample_rate
        self.channels = channels
        self.device_id = device_id
        self.recording = False
        self.streaming = False
        self.audio_queue = queue.Queue()
        self.stream: Optional[sd.InputStream] = None
        
        # Streaming mode configuration
        self.chunk_duration = 3.0  # seconds per chunk
        self.overlap_duration = 1.5  # seconds of overlap
        self.streaming_callback: Optional[Callable[[np.ndarray], None]] = None
        self.chunk_buffer = []
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
                self.chunk_buffer.append(indata.copy())
    
    def start_recording(self) -> bool:
        try:
            self.recording = True
            stream_params = {
                'samplerate': self.sample_rate,
                'channels': self.channels,
                'callback': self._audio_callback,
                'dtype': np.float32
            }
            
            # Add device parameter if specified
            if self.device_id is not None:
                stream_params['device'] = self.device_id
            
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
        """Start streaming recording that processes audio chunks in real-time"""
        try:
            self.streaming = True
            self.streaming_callback = callback
            self.chunk_buffer = []
            self.streaming_stop_event.clear()
            
            # Start audio stream
            stream_params = {
                'samplerate': self.sample_rate,
                'channels': self.channels,
                'callback': self._audio_callback,
                'dtype': np.float32
            }
            
            if self.device_id is not None:
                stream_params['device'] = self.device_id
            
            self.stream = sd.InputStream(**stream_params)
            self.stream.start()
            
            # Start streaming processing thread
            self.streaming_thread = threading.Thread(target=self._streaming_processor, daemon=True)
            self.streaming_thread.start()
            
            return True
        except Exception as e:
            print(f"Failed to start streaming recording with device {self.device_id}: {e}")
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
        self.chunk_buffer = []
    
    def _streaming_processor(self):
        """Process audio chunks in real-time for streaming transcription"""
        chunk_samples = int(self.chunk_duration * self.sample_rate)
        overlap_samples = int(self.overlap_duration * self.sample_rate)
        process_interval = self.chunk_duration - self.overlap_duration  # 1.5 seconds
        
        last_process_time = time.time()
        processed_samples = 0
        
        while not self.streaming_stop_event.is_set():
            current_time = time.time()
            
            # Check if it's time to process a new chunk
            if current_time - last_process_time >= process_interval:
                with self.buffer_lock:
                    if self.chunk_buffer:
                        # Combine all buffered audio
                        combined_audio = np.concatenate(self.chunk_buffer, axis=0)
                        
                        # Check if we have enough audio for a chunk
                        if len(combined_audio) >= chunk_samples:
                            # Extract chunk starting from appropriate position
                            start_sample = max(0, processed_samples - overlap_samples)
                            end_sample = start_sample + chunk_samples
                            
                            if end_sample <= len(combined_audio):
                                chunk = combined_audio[start_sample:end_sample]
                                processed_samples = end_sample - overlap_samples
                                
                                # Process this chunk
                                if self.streaming_callback:
                                    try:
                                        self.streaming_callback(chunk)
                                    except Exception as e:
                                        print(f"Error in streaming callback: {e}")
                                
                                last_process_time = current_time
                            
                            # Clean old buffer data to prevent memory buildup
                            # Keep only recent data needed for next overlap
                            keep_samples = chunk_samples  # Keep enough for next chunk
                            if len(combined_audio) > keep_samples * 2:
                                recent_start = len(combined_audio) - keep_samples
                                self.chunk_buffer = [combined_audio[recent_start:]]
                                processed_samples = max(0, processed_samples - recent_start)
            
            # Small sleep to prevent excessive CPU usage
            time.sleep(0.1)
    
    def is_streaming(self) -> bool:
        return self.streaming
    
    @staticmethod
    def get_available_devices():
        return sd.query_devices()
    
    @staticmethod
    def get_input_devices():
        """Get only input devices that can be used for recording"""
        devices = sd.query_devices()
        input_devices = []
        for i, device in enumerate(devices):
            if device['max_input_channels'] > 0:
                input_devices.append({
                    'index': i,
                    'name': device['name'],
                    'channels': device['max_input_channels'],
                    'default_samplerate': device['default_samplerate']
                })
        return input_devices
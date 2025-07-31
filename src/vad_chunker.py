import webrtcvad
import numpy as np
import time
from typing import Optional, Callable, List
import threading

class VADChunker:
    """Voice Activity Detection based chunking for real-time speech processing"""
    
    def __init__(self, 
                 sample_rate: int = 16000,
                 aggressiveness: int = 2,
                 frame_duration_ms: int = 20,
                 min_chunk_duration: float = 1.0,
                 max_chunk_duration: float = 10.0,
                 silence_timeout: float = 0.5):
        """
        Initialize VAD-based chunker
        
        Args:
            sample_rate: Audio sample rate (must be 8000, 16000, 32000, or 48000)
            aggressiveness: VAD aggressiveness level (0-3, higher = more aggressive)
            frame_duration_ms: Frame duration in ms (10, 20, or 30ms)
            min_chunk_duration: Minimum chunk duration in seconds
            max_chunk_duration: Maximum chunk duration in seconds  
            silence_timeout: Seconds of silence before triggering chunk
        """
        # Validate sample rate for WebRTC VAD
        if sample_rate not in [8000, 16000, 32000, 48000]:
            raise ValueError(f"Sample rate {sample_rate} not supported. Must be 8000, 16000, 32000, or 48000")
        
        # Validate frame duration
        if frame_duration_ms not in [10, 20, 30]:
            raise ValueError(f"Frame duration {frame_duration_ms}ms not supported. Must be 10, 20, or 30ms")
            
        self.sample_rate = sample_rate
        self.aggressiveness = aggressiveness
        self.frame_duration_ms = frame_duration_ms
        self.min_chunk_duration = min_chunk_duration
        self.max_chunk_duration = max_chunk_duration
        self.silence_timeout = silence_timeout
        
        # Calculate frame parameters
        self.frame_samples = int(sample_rate * frame_duration_ms / 1000)
        self.bytes_per_frame = self.frame_samples * 2  # 16-bit audio
        
        # Initialize WebRTC VAD
        self.vad = webrtcvad.Vad(aggressiveness)
        
        # Speech tracking state
        self.is_speaking = False
        self.speech_buffer: List[np.ndarray] = []
        self.last_speech_time = time.time()
        self.chunk_start_time = time.time()
        
        # Callback for processed chunks
        self.chunk_callback: Optional[Callable[[np.ndarray], None]] = None
        
        # Thread safety
        self.buffer_lock = threading.Lock()
        
    def set_chunk_callback(self, callback: Callable[[np.ndarray], None]):
        """Set callback function for processed audio chunks"""
        self.chunk_callback = callback
        
    def process_audio(self, audio_data: np.ndarray) -> bool:
        """
        Process incoming audio data with VAD-based chunking
        
        Args:
            audio_data: Audio samples as float32 numpy array
            
        Returns:
            True if a chunk was processed, False otherwise
        """
        # Convert float32 to int16 for VAD
        audio_int16 = (audio_data * 32767).astype(np.int16)
        
        # Process audio in frames
        frames_processed = 0
        chunk_triggered = False
        
        for i in range(0, len(audio_int16) - self.frame_samples + 1, self.frame_samples):
            frame = audio_int16[i:i + self.frame_samples]
            
            # Convert frame to bytes for VAD
            frame_bytes = frame.tobytes()
            
            # Run VAD on frame
            is_speech = self.vad.is_speech(frame_bytes, self.sample_rate)
            
            # Update speech state and buffer
            with self.buffer_lock:
                if is_speech:
                    if not self.is_speaking:
                        # Speech started
                        self.is_speaking = True
                        self.chunk_start_time = time.time()
                        print(f"🎤 Speech detected - starting new chunk")
                    
                    self.last_speech_time = time.time()
                    # Add corresponding float32 audio to buffer
                    start_sample = i
                    end_sample = i + self.frame_samples
                    self.speech_buffer.append(audio_data[start_sample:end_sample])
                    
                else:
                    # No speech in this frame
                    if self.is_speaking:
                        # Add frame to buffer anyway (short silence gaps are normal)
                        start_sample = i
                        end_sample = i + self.frame_samples
                        self.speech_buffer.append(audio_data[start_sample:end_sample])
                        
                        # Check if silence timeout exceeded
                        silence_duration = time.time() - self.last_speech_time
                        chunk_duration = time.time() - self.chunk_start_time
                        
                        if (silence_duration >= self.silence_timeout and 
                            chunk_duration >= self.min_chunk_duration) or \
                           chunk_duration >= self.max_chunk_duration:
                            
                            # Trigger chunk processing
                            chunk_triggered = self._trigger_chunk()
            
            frames_processed += 1
        
        return chunk_triggered
    
    def _trigger_chunk(self) -> bool:
        """Process accumulated speech buffer as a chunk"""
        if not self.speech_buffer:
            return False
            
        # Combine buffered audio
        chunk_audio = np.concatenate(self.speech_buffer, axis=0)
        chunk_duration = len(chunk_audio) / self.sample_rate
        
        print(f"🔄 Processing speech chunk: {chunk_duration:.1f}s ({len(self.speech_buffer)} frames)")
        
        # Call chunk callback if set
        if self.chunk_callback:
            try:
                self.chunk_callback(chunk_audio)
            except Exception as e:
                print(f"❌ Error in chunk callback: {e}")
        
        # Reset state
        self.is_speaking = False
        self.speech_buffer = []
        
        return True
    
    def force_chunk(self) -> bool:
        """Force processing of current buffer (useful for cleanup)"""
        with self.buffer_lock:
            if self.speech_buffer:
                return self._trigger_chunk()
        return False
    
    def reset(self):
        """Reset VAD state and clear buffers"""
        with self.buffer_lock:
            self.is_speaking = False
            self.speech_buffer = []
            self.last_speech_time = time.time()
            self.chunk_start_time = time.time()
    
    def get_stats(self) -> dict:
        """Get current VAD statistics"""
        with self.buffer_lock:
            current_buffer_duration = 0
            if self.speech_buffer:
                total_samples = sum(len(frame) for frame in self.speech_buffer)
                current_buffer_duration = total_samples / self.sample_rate
                
            return {
                'is_speaking': self.is_speaking,
                'buffer_duration': current_buffer_duration,
                'silence_duration': time.time() - self.last_speech_time if self.is_speaking else 0,
                'chunk_duration': time.time() - self.chunk_start_time if self.is_speaking else 0,
                'aggressiveness': self.aggressiveness,
                'min_chunk_duration': self.min_chunk_duration,
                'max_chunk_duration': self.max_chunk_duration
            }
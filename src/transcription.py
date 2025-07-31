from faster_whisper import WhisperModel
import tempfile
import os
import numpy as np
from typing import Optional, Callable, Set
import soundfile as sf
import threading
import queue

class TranscriptionEngine:
    def __init__(self):
        # Initialize Whisper model with English small version for optimal performance
        print("🔄 Loading Whisper model...")
        self.model = WhisperModel("small.en", device="cpu", compute_type="int8")
        print("✅ Whisper model loaded successfully")
        
        # Streaming transcription state
        self.streaming_active = False
        self.transcription_queue = queue.Queue()
        self.streaming_callback: Optional[Callable[[str], None]] = None
        self.transcription_thread: Optional[threading.Thread] = None
        self.streaming_stop_event = threading.Event()
        
        # Text deduplication for overlapping chunks
        self.recent_words = []  # Track recent words to avoid duplication
        self.words_lock = threading.Lock()
        self.max_recent_words = 20  # Keep track of last 20 words
        
    def transcribe_audio(self, audio_data: np.ndarray, sample_rate: int = 16000) -> Optional[str]:
        try:
            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, audio_data.flatten(), sample_rate)
                temp_filename = temp_file.name
            
            try:
                # Transcribe using faster-whisper
                segments, info = self.model.transcribe(temp_filename, beam_size=5)
                
                print(f"🌍 Detected language: {info.language} (probability: {info.language_probability:.2f})")
                
                # Combine all segments into a single text
                transcribed_text = ""
                segment_count = 0
                for segment in segments:
                    transcribed_text += segment.text.strip() + " "
                    segment_count += 1
                
                transcribed_text = transcribed_text.strip()
                
                if transcribed_text:
                    print(f"🎯 Speech detected ({segment_count} segments): '{transcribed_text}'")
                    return transcribed_text
                else:
                    print("⚠️ No speech detected in audio")
                    return None
                    
            finally:
                if os.path.exists(temp_filename):
                    os.unlink(temp_filename)
                    
        except Exception as e:
            print(f"❌ Transcription error: {e}")
            return None
    
    def start_streaming_transcription(self, callback: Callable[[str], None]):
        """Start streaming transcription mode"""
        self.streaming_active = True
        self.streaming_callback = callback
        self.streaming_stop_event.clear()
        
        with self.words_lock:
            self.recent_words.clear()
        
        # Start the transcription processing thread
        self.transcription_thread = threading.Thread(target=self._streaming_processor, daemon=True)
        self.transcription_thread.start()
    
    def stop_streaming_transcription(self):
        """Stop streaming transcription mode"""
        self.streaming_active = False
        self.streaming_stop_event.set()
        
        if self.transcription_thread and self.transcription_thread.is_alive():
            self.transcription_thread.join(timeout=3.0)
        
        self.streaming_callback = None
        
        with self.words_lock:
            self.recent_words.clear()
        
        # Clear any remaining items in queue
        while not self.transcription_queue.empty():
            try:
                self.transcription_queue.get_nowait()
            except queue.Empty:
                break
    
    def process_audio_chunk(self, audio_data: np.ndarray):
        """Process an audio chunk for streaming transcription"""
        if self.streaming_active:
            self.transcription_queue.put(audio_data.copy())
    
    def _streaming_processor(self):
        """Process queued audio chunks for streaming transcription"""
        while not self.streaming_stop_event.is_set():
            try:
                # Get audio chunk with timeout
                audio_chunk = self.transcription_queue.get(timeout=0.5)
                
                # Transcribe the chunk
                text = self._transcribe_chunk(audio_chunk)
                
                if text:
                    # Apply deduplication
                    new_text = self._deduplicate_text(text)
                    
                    if new_text and self.streaming_callback:
                        self.streaming_callback(new_text)
                
            except queue.Empty:
                continue
            except Exception as e:
                print(f"❌ Streaming transcription error: {e}")
                continue
    
    def _transcribe_chunk(self, audio_data: np.ndarray, sample_rate: int = 16000) -> Optional[str]:
        """Transcribe a single audio chunk with optimized settings for speed"""
        try:
            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, audio_data.flatten(), sample_rate)
                temp_filename = temp_file.name
            
            try:
                # Use faster settings for real-time transcription
                segments, info = self.model.transcribe(
                    temp_filename, 
                    beam_size=1,  # Faster but potentially less accurate
                    best_of=1,    # Single pass for speed
                    temperature=0.0  # Deterministic output
                )
                
                # Combine segments into text
                transcribed_text = ""
                for segment in segments:
                    transcribed_text += segment.text.strip() + " "
                
                transcribed_text = transcribed_text.strip()
                return transcribed_text if transcribed_text else None
                    
            finally:
                if os.path.exists(temp_filename):
                    os.unlink(temp_filename)
                    
        except Exception as e:
            print(f"❌ Chunk transcription error: {e}")
            return None
    
    def _deduplicate_text(self, text: str) -> str:
        """Remove duplicate words from overlapping transcription chunks"""
        if not text:
            return ""
        
        # Split into words and normalize
        words = text.split()
        if not words:
            return ""
        
        new_words = []
        
        with self.words_lock:
            # Convert recent words to lowercase for comparison
            recent_lower = [w.lower() for w in self.recent_words]
            
            # Find where the new text overlaps with recent words
            overlap_found = False
            start_index = 0
            
            # Look for the best overlap point
            for i in range(min(len(words), len(recent_lower))):
                # Check if the first i+1 words of new text match the last i+1 words of recent
                new_start = [w.lower() for w in words[:i+1]]
                recent_end = recent_lower[-(i+1):] if len(recent_lower) >= i+1 else recent_lower
                
                if new_start == recent_end:
                    start_index = i + 1  # Skip the overlapping part
                    overlap_found = True
            
            # If no clear overlap found but we have recent words, be conservative
            if not overlap_found and recent_lower and len(words) > 0:
                # Check if first word is same as last recent word
                if words[0].lower() == recent_lower[-1]:
                    start_index = 1
            
            # Extract only new words
            if start_index < len(words):
                new_words = words[start_index:]
            
            # Update recent words list
            if new_words:
                self.recent_words.extend(new_words)
                
                # Keep only the most recent words to prevent memory buildup
                if len(self.recent_words) > self.max_recent_words:
                    self.recent_words = self.recent_words[-self.max_recent_words:]
        
        return " ".join(new_words) if new_words else ""
    
    def is_streaming_active(self) -> bool:
        """Check if streaming transcription is active"""
        return self.streaming_active
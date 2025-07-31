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
        
        # Text stream assembly system
        self.text_buffer = ""  # Continuous assembled text
        self.buffer_lock = threading.Lock()
        self.pending_chunks = []  # Queue for serialized processing
        self.processing_lock = threading.Lock()  # Ensure one chunk at a time
        
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
        """Start streaming transcription mode with text assembly"""
        self.streaming_active = True
        self.streaming_callback = callback
        self.streaming_stop_event.clear()
        
        # Initialize text buffer system
        with self.buffer_lock:
            self.text_buffer = ""
            self.pending_chunks = []
        
        # Start the serialized transcription processing thread
        self.transcription_thread = threading.Thread(target=self._streaming_processor, daemon=True)
        self.transcription_thread.start()
    
    def stop_streaming_transcription(self):
        """Stop streaming transcription mode"""
        self.streaming_active = False
        self.streaming_stop_event.set()
        
        if self.transcription_thread and self.transcription_thread.is_alive():
            self.transcription_thread.join(timeout=3.0)
        
        self.streaming_callback = None
        
        # Clear text buffer system
        with self.buffer_lock:
            self.text_buffer = ""
            self.pending_chunks = []
        
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
        """Serialized processing of audio chunks with text assembly"""
        while not self.streaming_stop_event.is_set():
            try:
                # Get audio chunk with timeout
                audio_chunk = self.transcription_queue.get(timeout=0.5)
                
                # Process chunk with serialization lock (one at a time)
                with self.processing_lock:
                    # Transcribe the chunk
                    new_text = self._transcribe_chunk(audio_chunk)
                    
                    if new_text:
                        # Assemble with existing text buffer
                        assembled_text = self._assemble_text_chunk(new_text)
                        
                        if assembled_text and self.streaming_callback:
                            self.streaming_callback(assembled_text)
                
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
    
    def _assemble_text_chunk(self, new_text: str) -> str:
        """Assemble new text chunk with existing buffer using sequence-based overlap detection"""
        if not new_text:
            return ""
        
        # Clean and normalize the new text
        cleaned_new = self._preprocess_text(new_text)
        if not cleaned_new:
            return ""
        
        with self.buffer_lock:
            # If buffer is empty, this is the first chunk
            if not self.text_buffer:
                self.text_buffer = cleaned_new
                return cleaned_new
            
            # Find overlap between buffer end and new text start
            overlap_result = self._find_text_overlap(self.text_buffer, cleaned_new)
            
            if overlap_result:
                # Merge texts by removing overlap from new text
                buffer_end, new_start, overlap_length = overlap_result
                
                # Remove overlapping part from new text
                truly_new_text = cleaned_new[overlap_length:].strip()
                
                # Always add space before new text unless buffer ends with punctuation
                if truly_new_text:
                    if not self.text_buffer.endswith((' ', '.', ',', '!', '?', ';', ':')):
                        truly_new_text = ' ' + truly_new_text
                    elif self.text_buffer.endswith(('.', '!', '?')) and truly_new_text and truly_new_text[0].islower():
                        # Add space after sentence-ending punctuation
                        truly_new_text = ' ' + truly_new_text
                
                # Update buffer
                if truly_new_text:
                    self.text_buffer += truly_new_text
                    return truly_new_text
                else:
                    return ""  # All text was duplicate
            else:
                # No overlap found, add with spacing
                spaced_text = cleaned_new
                if not self.text_buffer.endswith(' ') and not cleaned_new.startswith(' '):
                    spaced_text = ' ' + cleaned_new
                
                self.text_buffer += spaced_text
                return spaced_text
    
    def _find_text_overlap(self, buffer_text: str, new_text: str):
        """Find overlapping sequence between buffer end and new text start"""
        if not buffer_text or not new_text:
            return None
        
        # Convert to words for better matching
        buffer_words = buffer_text.split()
        new_words = new_text.split()
        
        if not buffer_words or not new_words:
            return None
        
        # Look for longest overlap (check longer overlaps first)
        max_overlap_len = min(len(buffer_words), len(new_words), 8)  # Limit to 8 words max
        
        for overlap_len in range(max_overlap_len, 0, -1):
            # Get last N words from buffer
            buffer_end = buffer_words[-overlap_len:]
            # Get first N words from new text  
            new_start = new_words[:overlap_len]
            
            # Compare word sequences (case-insensitive)
            buffer_end_lower = [w.lower().strip('.,!?;:') for w in buffer_end]
            new_start_lower = [w.lower().strip('.,!?;:') for w in new_start]
            
            if buffer_end_lower == new_start_lower:
                # Found overlap - calculate character position
                overlap_chars = len(' '.join(new_words[:overlap_len]))
                return (buffer_end, new_start, overlap_chars)
        
        return None
    
    def _preprocess_text(self, text: str) -> str:
        """Clean and preprocess text before deduplication"""
        if not text:
            return ""
        
        # Remove extra whitespace and normalize
        cleaned = ' '.join(text.split())
        
        # Remove or fix common transcription artifacts
        artifacts = [
            ('  ', ' '),      # Double spaces
            (' .', '.'),      # Space before period
            (' ,', ','),      # Space before comma
            (' !', '!'),      # Space before exclamation
            (' ?', '?'),      # Space before question mark
        ]
        
        for old, new in artifacts:
            cleaned = cleaned.replace(old, new)
        
        return cleaned.strip()
    
    def is_streaming_active(self) -> bool:
        """Check if streaming transcription is active"""
        return self.streaming_active
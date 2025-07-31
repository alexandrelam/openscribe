from faster_whisper import WhisperModel
import tempfile
import os
import numpy as np
from typing import Optional, Callable, Set
import soundfile as sf
import threading
import queue

class TranscriptionEngine:
    def __init__(self, language_config: str = "auto", model_size: str = "small"):
        """
        Initialize Whisper model with configurable language support
        
        Args:
            language_config: "auto" for automatic detection, "en"/"fr"/etc for specific language
            model_size: "small", "base", "large-v3" etc.
        """
        self.language_config = language_config
        self.model_size = model_size
        self.current_model_name = None
        
        # Load appropriate model based on language configuration
        self._load_model()
        
        # Language detection settings
        self.detected_language = None
        self.language_confidence = 0.0
        self.language_detection_enabled = language_config == "auto"
        
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
    
    def _load_model(self):
        """Load appropriate Whisper model based on language configuration"""
        # Determine model name based on language config
        if self.language_config == "en":
            # Use English-only model for optimal performance
            model_name = f"{self.model_size}.en"
        else:
            # Use multilingual model for other languages or auto-detection
            model_name = self.model_size
        
        # Only reload if model has changed
        if model_name != self.current_model_name:
            print(f"🔄 Loading Whisper model: {model_name}")
            self.model = WhisperModel(model_name, device="cpu", compute_type="int8")
            self.current_model_name = model_name
            print(f"✅ Whisper model loaded: {model_name}")
    
    def configure_language(self, language_config: str = "auto", model_size: str = "small"):
        """Configure language settings and reload model if necessary"""
        old_language_config = self.language_config
        self.language_config = language_config
        self.model_size = model_size
        self.language_detection_enabled = language_config == "auto"
        
        # Reload model if language configuration changed
        if old_language_config != language_config:
            self._load_model()
            print(f"🌍 Language configuration updated: {language_config}")
    
    def get_language_info(self) -> dict:
        """Get current language detection information"""
        return {
            'configured_language': self.language_config,
            'detected_language': self.detected_language,
            'language_confidence': self.language_confidence,
            'detection_enabled': self.language_detection_enabled,
            'current_model': self.current_model_name
        }
        
    def transcribe_audio(self, audio_data: np.ndarray, sample_rate: int = 16000) -> Optional[str]:
        try:
            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, audio_data.flatten(), sample_rate)
                temp_filename = temp_file.name
            
            try:
                # Prepare transcription parameters
                transcribe_params = {
                    'beam_size': 5,
                    'condition_on_previous_text': False
                }
                
                # Add language parameter if not auto-detecting
                if self.language_config != "auto":
                    transcribe_params['language'] = self.language_config
                
                # Transcribe using faster-whisper
                segments, info = self.model.transcribe(temp_filename, **transcribe_params)
                
                # Update language detection info
                self.detected_language = info.language
                self.language_confidence = info.language_probability
                
                # Log language information
                if self.language_detection_enabled:
                    print(f"🌍 Auto-detected language: {info.language} (confidence: {info.language_probability:.2f})")
                else:
                    print(f"🌍 Using configured language: {self.language_config}")
                
                # Combine all segments into a single text
                transcribed_text = ""
                segment_count = 0
                for segment in segments:
                    transcribed_text += segment.text.strip() + " "
                    segment_count += 1
                
                transcribed_text = transcribed_text.strip()
                
                if transcribed_text:
                    print(f"🎯 Speech transcribed ({segment_count} segments): '{transcribed_text}'")
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
                # Prepare fast transcription parameters
                transcribe_params = {
                    'beam_size': 1,  # Faster but potentially less accurate
                    'best_of': 1,    # Single pass for speed
                    'temperature': 0.0,  # Deterministic output
                    'condition_on_previous_text': False
                }
                
                # Add language parameter if not auto-detecting
                if self.language_config != "auto":
                    transcribe_params['language'] = self.language_config
                
                # Use faster settings for real-time transcription
                segments, info = self.model.transcribe(temp_filename, **transcribe_params)
                
                # Update language detection info (for streaming mode)
                self.detected_language = info.language
                self.language_confidence = info.language_probability
                
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
        """Assemble new text chunk with existing buffer (simplified for VAD-based chunks)"""
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
            
            # Since VAD provides natural speech boundaries, we can simplify assembly
            # Just add proper spacing between chunks
            spaced_text = cleaned_new
            
            # Add space if needed between chunks
            if not self.text_buffer.endswith((' ', '.', ',', '!', '?', ';', ':')):
                spaced_text = ' ' + cleaned_new
            elif self.text_buffer.endswith(('.', '!', '?')) and cleaned_new and cleaned_new[0].islower():
                # Add space after sentence-ending punctuation  
                spaced_text = ' ' + cleaned_new
            
            # Update buffer
            self.text_buffer += spaced_text
            return spaced_text
    
    
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
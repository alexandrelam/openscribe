import tempfile
import os
import numpy as np
from typing import Optional, Callable
import soundfile as sf
import threading
import queue

from .base_provider import BaseWhisperProvider

try:
    from whisper_cpp import Whisper

    WHISPER_CPP_AVAILABLE = True
except ImportError:
    WHISPER_CPP_AVAILABLE = False
    Whisper = None


class WhisperCppProvider(BaseWhisperProvider):
    """
    Whisper.cpp provider implementation.

    This provider uses the whisper.cpp library via Python bindings for
    CPU-optimized Whisper model inference.
    """

    def __init__(
        self,
        language_config: str = "auto",
        model_size: str = "small",
        live_quality_mode: str = "balanced",
        enable_overlap_detection: bool = True,
        debug_text_assembly: bool = False,
    ):
        if not WHISPER_CPP_AVAILABLE:
            raise ImportError(
                "whisper-cpp-python is not installed. "
                "Install it with: pip install whisper-cpp-python"
            )

        self.language_config = language_config
        self.model_size = model_size
        self.current_model_name = None
        self.live_quality_mode = live_quality_mode
        self.enable_overlap_detection = enable_overlap_detection
        self.debug_text_assembly = debug_text_assembly

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
        self.text_buffer = ""
        self.buffer_lock = threading.Lock()
        self.pending_chunks = []
        self.processing_lock = threading.Lock()

        # Context preservation for better accuracy
        self.previous_context = ""
        self.context_length = 50
        self.overlap_buffer = []

    def _load_model(self):
        """Load appropriate Whisper model based on language configuration"""
        # Map model sizes to whisper.cpp model names
        model_mapping = {
            "small": "small",
            "base": "base",
            "large-v3": "large",
            "small.en": "small.en",
            "base.en": "base.en",
        }

        # Determine model name based on language config
        if self.language_config == "en":
            # Use English-only model for optimal performance
            model_key = f"{self.model_size}.en"
        else:
            # Use multilingual model for other languages or auto-detection
            model_key = self.model_size

        model_name = model_mapping.get(model_key, "small")

        # Only reload if model has changed
        if model_name != self.current_model_name:
            print(f"🔄 Loading Whisper.cpp model: {model_name}")
            try:
                # Initialize whisper.cpp model
                self.model = Whisper.from_pretrained(model_name)
                self.current_model_name = model_name
                print(f"✅ Whisper.cpp model loaded: {model_name}")
            except Exception as e:
                print(f"❌ Failed to load Whisper.cpp model: {e}")
                raise

    def configure_language(
        self,
        language_config: str = "auto",
        model_size: str = "small",
        live_quality_mode: str = "balanced",
        enable_overlap_detection: bool = True,
        debug_text_assembly: bool = False,
    ):
        """Configure language settings and reload model if necessary"""
        old_language_config = self.language_config
        self.language_config = language_config
        self.model_size = model_size
        self.live_quality_mode = live_quality_mode
        self.enable_overlap_detection = enable_overlap_detection
        self.debug_text_assembly = debug_text_assembly
        self.language_detection_enabled = language_config == "auto"

        # Reload model if language configuration changed
        if old_language_config != language_config:
            self._load_model()
            print(f"🌍 Language configuration updated: {language_config}")

    def get_language_info(self) -> dict:
        """Get current language detection information"""
        return {
            "configured_language": self.language_config,
            "detected_language": self.detected_language,
            "language_confidence": self.language_confidence,
            "detection_enabled": self.language_detection_enabled,
            "current_model": self.current_model_name,
        }

    def transcribe_audio(
        self, audio_data: np.ndarray, sample_rate: int = 16000
    ) -> Optional[str]:
        try:
            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, audio_data.flatten(), sample_rate)
                temp_filename = temp_file.name

            try:
                # Prepare transcription parameters for whisper.cpp
                transcribe_params = {}

                # Add language parameter if not auto-detecting
                if self.language_config != "auto":
                    transcribe_params["language"] = self.language_config

                # Transcribe using whisper.cpp
                result = self.model.transcribe(temp_filename, **transcribe_params)

                # Extract text from result
                if isinstance(result, dict):
                    transcribed_text = result.get("text", "").strip()
                    # Try to get language info if available
                    self.detected_language = result.get(
                        "language", self.language_config
                    )
                    self.language_confidence = result.get("probability", 0.0)
                elif isinstance(result, str):
                    transcribed_text = result.strip()
                    # Set defaults for language info
                    self.detected_language = (
                        self.language_config if self.language_config != "auto" else "en"
                    )
                    self.language_confidence = 1.0
                else:
                    transcribed_text = str(result).strip()
                    self.detected_language = (
                        self.language_config if self.language_config != "auto" else "en"
                    )
                    self.language_confidence = 1.0

                # Log language information
                if self.language_detection_enabled:
                    print(
                        f"🌍 Detected language: {self.detected_language} (confidence: {self.language_confidence:.2f})"
                    )
                else:
                    print(f"🌍 Using configured language: {self.language_config}")

                if transcribed_text:
                    print(f"🎯 Speech transcribed: '{transcribed_text}'")
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
        self.transcription_thread = threading.Thread(
            target=self._streaming_processor, daemon=True
        )
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
            self.previous_context = ""
            self.overlap_buffer = []

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

    def _transcribe_chunk(
        self, audio_data: np.ndarray, sample_rate: int = 16000
    ) -> Optional[str]:
        """Transcribe a single audio chunk with context preservation for better accuracy"""
        try:
            # Add overlapping audio for better context continuity
            if self.overlap_buffer:
                overlap_audio = np.concatenate(self.overlap_buffer + [audio_data])
            else:
                overlap_audio = audio_data

            # Store last 1 second of current chunk for next overlap
            overlap_samples = int(sample_rate * 1.0)
            if len(audio_data) > overlap_samples:
                self.overlap_buffer = [audio_data[-overlap_samples:]]
            else:
                self.overlap_buffer = [audio_data]

            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, overlap_audio.flatten(), sample_rate)
                temp_filename = temp_file.name

            try:
                # Prepare transcription parameters based on quality mode
                transcribe_params = {}

                # Add language parameter if not auto-detecting
                if self.language_config != "auto":
                    transcribe_params["language"] = self.language_config

                # Note: whisper.cpp may have different parameter names than faster-whisper
                # These parameters might need adjustment based on the actual whisper-cpp-python API

                # Transcribe using whisper.cpp
                result = self.model.transcribe(temp_filename, **transcribe_params)

                # Extract text from result
                if isinstance(result, dict):
                    transcribed_text = result.get("text", "").strip()
                    self.detected_language = result.get(
                        "language", self.detected_language
                    )
                    self.language_confidence = result.get(
                        "probability", self.language_confidence
                    )
                elif isinstance(result, str):
                    transcribed_text = result.strip()
                else:
                    transcribed_text = str(result).strip()

                # Update context for next chunk
                if transcribed_text:
                    self._update_context(transcribed_text)

                return transcribed_text if transcribed_text else None

            finally:
                if os.path.exists(temp_filename):
                    os.unlink(temp_filename)

        except Exception as e:
            print(f"❌ Chunk transcription error: {e}")
            return None

    def _detect_and_remove_overlap(self, previous_text: str, new_text: str) -> str:
        """Detect and remove overlapping text between consecutive chunks"""
        if not previous_text or not new_text:
            return new_text

        # Normalize and split into words
        prev_words = previous_text.strip().lower().split()
        new_words = new_text.strip().split()
        new_words_lower = [w.lower() for w in new_words]

        # Safety limits to prevent over-removal
        min_remaining_words = 2
        max_overlap_percent = 0.7
        max_overlap_words = min(
            len(prev_words), int(len(new_words) * max_overlap_percent), 8
        )

        # Try different overlap lengths (from longest to shortest)
        for overlap_length in range(max_overlap_words, 0, -1):
            # Safety check: ensure we keep minimum remaining words
            if len(new_words) - overlap_length < min_remaining_words:
                continue

            # Get last N words from previous text
            prev_suffix = prev_words[-overlap_length:]
            # Get first N words from new text (case-insensitive)
            new_prefix = new_words_lower[:overlap_length]

            # Check for exact match
            if prev_suffix == new_prefix:
                # Remove overlapped words from new text (preserve original case)
                remaining_words = new_words[overlap_length:]
                deduped_text = " ".join(remaining_words)
                if self.debug_text_assembly:
                    print(
                        f"🔄 Detected {overlap_length}-word overlap, removed: '{' '.join(new_words[:overlap_length])}'"
                    )
                return deduped_text

        # No overlap detected, return full new text
        return new_text

    def _assemble_text_chunk(self, new_text: str) -> str:
        """Assemble new text chunk with existing buffer, removing overlaps from audio buffering"""
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

            # Detect and remove text overlap caused by audio buffering (if enabled)
            if self.enable_overlap_detection:
                if self.debug_text_assembly:
                    print(f"🔍 Before overlap detection:")
                    print(f"   Buffer: '{self.text_buffer[-50:]}'")
                    print(f"   New: '{cleaned_new}'")
                deduped_text = self._detect_and_remove_overlap(
                    self.text_buffer, cleaned_new
                )
                if self.debug_text_assembly:
                    print(f"   After deduplication: '{deduped_text}'")
            else:
                deduped_text = cleaned_new

            # If deduplication removed everything, return empty
            if not deduped_text.strip():
                return ""

            # Add proper spacing between chunks
            needs_space = False

            if self.text_buffer.endswith(" "):
                needs_space = False
                if self.debug_text_assembly:
                    print(
                        f"📝 No space needed (buffer ends with space): '{deduped_text}'"
                    )
            elif deduped_text and deduped_text[0] in ".,!?;:":
                needs_space = False
                if self.debug_text_assembly:
                    print(
                        f"📝 No space needed (starts with punctuation): '{deduped_text}'"
                    )
            else:
                needs_space = True
                if self.debug_text_assembly:
                    print(
                        f"📝 Space needed for word boundary: '{deduped_text}' -> will add space"
                    )

            # Update buffer with proper spacing
            if needs_space:
                self.text_buffer += " " + deduped_text
                result_text = " " + deduped_text
            else:
                self.text_buffer += deduped_text
                result_text = deduped_text

            if self.debug_text_assembly:
                print(f"📋 Final result:")
                print(f"   Updated buffer: '{self.text_buffer[-100:]}'")
                print(f"   Returning for paste: '{repr(result_text)}'")
                print(f"   Result length: {len(result_text)}")

            return result_text

    def _preprocess_text(self, text: str) -> str:
        """Clean and preprocess text before deduplication"""
        if not text:
            return ""

        # Remove extra whitespace and normalize
        cleaned = " ".join(text.split())

        # Remove or fix common transcription artifacts
        artifacts = [
            ("  ", " "),
            (" .", "."),
            (" ,", ","),
            (" !", "!"),
            (" ?", "?"),
        ]

        for old, new in artifacts:
            cleaned = cleaned.replace(old, new)

        return cleaned.strip()

    def _update_context(self, new_text: str):
        """Update previous context for better transcription continuity"""
        if new_text:
            # Combine with existing context
            combined_context = f"{self.previous_context} {new_text}".strip()

            # Keep only the last N characters as context
            if len(combined_context) > self.context_length:
                # Try to cut at word boundary
                words = combined_context.split()
                context_words = []
                char_count = 0

                # Add words from the end until we reach the character limit
                for word in reversed(words):
                    if char_count + len(word) + 1 <= self.context_length:
                        context_words.insert(0, word)
                        char_count += len(word) + 1
                    else:
                        break

                self.previous_context = " ".join(context_words)
            else:
                self.previous_context = combined_context

    def is_streaming_active(self) -> bool:
        """Check if streaming transcription is active"""
        return self.streaming_active

    @property
    def provider_name(self) -> str:
        """Return the name of this provider"""
        return "whisper-cpp"

    @property
    def provider_info(self) -> dict:
        """Return information about this provider"""
        return {
            "name": "Whisper.cpp",
            "description": "CPU-optimized C++ implementation of Whisper",
            "current_model": self.current_model_name,
            "language_config": self.language_config,
            "model_size": self.model_size,
            "live_quality_mode": self.live_quality_mode,
        }

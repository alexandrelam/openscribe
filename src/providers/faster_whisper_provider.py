from faster_whisper import WhisperModel
import tempfile
import os
import numpy as np
from typing import Optional, Callable
import soundfile as sf
import threading
import queue

from .base_provider import BaseWhisperProvider


class FasterWhisperProvider(BaseWhisperProvider):
    """
    Faster Whisper provider implementation.

    This provider uses the faster-whisper library for high-performance
    Whisper model inference with GPU optimization.
    """

    def __init__(
        self,
        language_config: str = "auto",
        model_size: str = "small",
        live_quality_mode: str = "balanced",
        enable_overlap_detection: bool = True,
        debug_text_assembly: bool = False,
    ):
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
        # Determine model name based on language config
        if self.language_config == "en":
            # Use English-only model for optimal performance
            model_name = f"{self.model_size}.en"
        else:
            # Use multilingual model for other languages or auto-detection
            model_name = self.model_size

        # Only reload if model has changed
        if model_name != self.current_model_name:
            print(f"🔄 Loading Faster Whisper model: {model_name}")
            self.model = WhisperModel(model_name, device="cpu", compute_type="int8")
            self.current_model_name = model_name
            print(f"✅ Faster Whisper model loaded: {model_name}")

    def _load_model_async(self, model_name: str, callback: Optional[Callable] = None):
        """Load model asynchronously to prevent GUI blocking"""

        def load_model():
            try:
                print(f"🔄 Loading Faster Whisper model: {model_name}")

                # Clean up old model to free memory
                if hasattr(self, "model") and self.model is not None:
                    print("🧹 Releasing previous model to free memory")
                    del self.model
                    import gc

                    gc.collect()  # Force garbage collection

                model = WhisperModel(model_name, device="cpu", compute_type="int8")
                self.model = model
                self.current_model_name = model_name
                print(f"✅ Faster Whisper model loaded: {model_name}")
                if callback:
                    callback(True, None)
            except Exception as e:
                print(f"❌ Failed to load model {model_name}: {e}")
                if callback:
                    callback(False, str(e))

        threading.Thread(target=load_model, daemon=True).start()

    def cleanup_resources(self):
        """Clean up model resources to free memory"""
        if hasattr(self, "model") and self.model is not None:
            print("🧹 Cleaning up Faster Whisper model resources")
            del self.model
            self.model = None
            self.current_model_name = None
            import gc

            gc.collect()

    def configure_language(
        self,
        language_config: str = "auto",
        model_size: str = "small",
        live_quality_mode: str = "balanced",
        enable_overlap_detection: bool = True,
        debug_text_assembly: bool = False,
        async_loading: bool = False,
        callback: Optional[Callable] = None,
    ):
        """Configure language settings and reload model if necessary"""
        old_language_config = self.language_config
        old_model_size = self.model_size

        self.language_config = language_config
        self.model_size = model_size
        self.live_quality_mode = live_quality_mode
        self.enable_overlap_detection = enable_overlap_detection
        self.debug_text_assembly = debug_text_assembly
        self.language_detection_enabled = language_config == "auto"

        # Determine new model name
        if language_config == "en":
            new_model_name = f"{model_size}.en"
        else:
            new_model_name = model_size

        # Reload model if language/model configuration changed
        if (
            old_language_config != language_config or old_model_size != model_size
        ) and new_model_name != self.current_model_name:
            if async_loading:
                self._load_model_async(new_model_name, callback)
            else:
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
        self,
        audio_data: np.ndarray,
        sample_rate: int = 16000,
        timeout: Optional[float] = None,
    ) -> Optional[str]:
        """Transcribe audio with optional timeout to prevent indefinite blocking"""
        try:
            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, audio_data.flatten(), sample_rate)
                temp_filename = temp_file.name

            try:
                # Prepare transcription parameters
                transcribe_params = {
                    "beam_size": 5,
                    "condition_on_previous_text": False,
                }

                # Add language parameter if not auto-detecting
                if self.language_config != "auto":
                    transcribe_params["language"] = self.language_config

                # Use timeout-aware transcription
                if timeout:
                    segments, info = self._transcribe_with_timeout(
                        temp_filename, transcribe_params, timeout
                    )
                else:
                    # Transcribe using faster-whisper
                    segments, info = self.model.transcribe(
                        temp_filename, **transcribe_params
                    )

                # Update language detection info
                self.detected_language = info.language
                self.language_confidence = info.language_probability

                # Log language information
                if self.language_detection_enabled:
                    print(
                        f"🌍 Auto-detected language: {info.language} (confidence: {info.language_probability:.2f})"
                    )
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
                    print(
                        f"🎯 Speech transcribed ({segment_count} segments): '{transcribed_text}'"
                    )
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

    def _transcribe_with_timeout(
        self, temp_filename: str, transcribe_params: dict, timeout: float
    ):
        """Transcribe audio with timeout using threading"""
        import concurrent.futures

        with concurrent.futures.ThreadPoolExecutor(max_workers=1) as executor:
            future = executor.submit(
                self.model.transcribe, temp_filename, **transcribe_params
            )
            try:
                segments, info = future.result(timeout=timeout)
                return segments, info
            except concurrent.futures.TimeoutError:
                print(f"⏰ Transcription timeout after {timeout} seconds")
                # Cancel the future if possible
                future.cancel()
                raise TimeoutError(f"Transcription timed out after {timeout} seconds")

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
                # Quality-based transcription parameters
                if self.live_quality_mode == "fast":
                    transcribe_params = {
                        "beam_size": 1,
                        "best_of": 1,
                        "temperature": 0.0,
                        "condition_on_previous_text": False,
                        "no_speech_threshold": 0.6,
                    }
                elif self.live_quality_mode == "accurate":
                    transcribe_params = {
                        "beam_size": 5,
                        "best_of": 3,
                        "temperature": 0.0,
                        "condition_on_previous_text": True,
                        "no_speech_threshold": 0.4,
                    }
                else:  # balanced (default)
                    transcribe_params = {
                        "beam_size": 3,
                        "best_of": 1,
                        "temperature": 0.0,
                        "condition_on_previous_text": True,
                        "no_speech_threshold": 0.6,
                    }

                # Add previous context if available
                if self.previous_context:
                    transcribe_params["initial_prompt"] = self.previous_context

                # Add language parameter if not auto-detecting
                if self.language_config != "auto":
                    transcribe_params["language"] = self.language_config

                # Use improved settings for real-time transcription
                segments, info = self.model.transcribe(
                    temp_filename, **transcribe_params
                )

                # Update language detection info (for streaming mode)
                self.detected_language = info.language
                self.language_confidence = info.language_probability

                # Combine segments into text
                transcribed_text = ""
                for segment in segments:
                    transcribed_text += segment.text.strip() + " "

                transcribed_text = transcribed_text.strip()

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
        return "faster-whisper"

    @property
    def provider_info(self) -> dict:
        """Return information about this provider"""
        return {
            "name": "Faster Whisper",
            "description": "Fast, GPU-optimized Whisper implementation",
            "current_model": self.current_model_name,
            "language_config": self.language_config,
            "model_size": self.model_size,
            "live_quality_mode": self.live_quality_mode,
        }

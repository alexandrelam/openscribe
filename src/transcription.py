import numpy as np
from typing import Optional, Callable

from .providers import (
    BaseWhisperProvider,
    FasterWhisperProvider,
    WhisperCppProvider,
    WHISPER_CPP_AVAILABLE,
)


class TranscriptionEngine:
    def __init__(
        self,
        language_config: str = "auto",
        model_size: str = "small",
        live_quality_mode: str = "balanced",
        enable_overlap_detection: bool = True,
        debug_text_assembly: bool = False,
        whisper_provider: str = "faster-whisper",
    ):
        """
        Initialize Whisper model with configurable language support

        Args:
            language_config: "auto" for automatic detection, "en"/"fr"/etc for specific language
            model_size: "small", "base", "large-v3" etc.
            live_quality_mode: "fast", "balanced", or "accurate" for live transcription
            enable_overlap_detection: Whether to detect and remove text overlaps in live mode
            debug_text_assembly: Whether to enable verbose logging for text assembly debugging
            whisper_provider: "faster-whisper" or "whisper-cpp"
        """
        self.whisper_provider = whisper_provider

        # Create the appropriate provider instance
        self.provider = self._create_provider(
            language_config=language_config,
            model_size=model_size,
            live_quality_mode=live_quality_mode,
            enable_overlap_detection=enable_overlap_detection,
            debug_text_assembly=debug_text_assembly,
        )

    def _create_provider(
        self,
        language_config: str,
        model_size: str,
        live_quality_mode: str,
        enable_overlap_detection: bool,
        debug_text_assembly: bool,
    ) -> BaseWhisperProvider:
        """Create the appropriate provider instance based on configuration"""
        provider_params = {
            "language_config": language_config,
            "model_size": model_size,
            "live_quality_mode": live_quality_mode,
            "enable_overlap_detection": enable_overlap_detection,
            "debug_text_assembly": debug_text_assembly,
        }

        if self.whisper_provider == "whisper-cpp":
            if not WHISPER_CPP_AVAILABLE:
                print("⚠️ Whisper.cpp not available, falling back to faster-whisper")
                return FasterWhisperProvider(**provider_params)
            try:
                return WhisperCppProvider(**provider_params)
            except Exception as e:
                print(f"❌ Failed to initialize Whisper.cpp provider: {e}")
                print("⚠️ Falling back to faster-whisper")
                return FasterWhisperProvider(**provider_params)
        elif self.whisper_provider == "faster-whisper":
            return FasterWhisperProvider(**provider_params)
        else:
            print(f"⚠️ Unknown provider '{self.whisper_provider}', using faster-whisper")
            return FasterWhisperProvider(**provider_params)

    def configure_language(
        self,
        language_config: str = "auto",
        model_size: str = "small",
        live_quality_mode: str = "balanced",
        enable_overlap_detection: bool = True,
        debug_text_assembly: bool = False,
    ):
        """Configure language settings and reload model if necessary"""
        return self.provider.configure_language(
            language_config=language_config,
            model_size=model_size,
            live_quality_mode=live_quality_mode,
            enable_overlap_detection=enable_overlap_detection,
            debug_text_assembly=debug_text_assembly,
        )

    def get_language_info(self) -> dict:
        """Get current language detection information"""
        return self.provider.get_language_info()

    def transcribe_audio(
        self, audio_data: np.ndarray, sample_rate: int = 16000
    ) -> Optional[str]:
        """Transcribe audio data and return the transcribed text"""
        return self.provider.transcribe_audio(audio_data, sample_rate)

    def start_streaming_transcription(self, callback: Callable[[str], None]):
        """Start streaming transcription mode with text assembly"""
        return self.provider.start_streaming_transcription(callback)

    def stop_streaming_transcription(self):
        """Stop streaming transcription mode"""
        return self.provider.stop_streaming_transcription()

    def process_audio_chunk(self, audio_data: np.ndarray):
        """Process an audio chunk for streaming transcription"""
        return self.provider.process_audio_chunk(audio_data)

    def is_streaming_active(self) -> bool:
        """Check if streaming transcription is active"""
        return self.provider.is_streaming_active()

    @property
    def provider_name(self) -> str:
        """Return the name of the current provider"""
        return self.provider.provider_name

    @property
    def provider_info(self) -> dict:
        """Return information about the current provider"""
        return self.provider.provider_info

from abc import ABC, abstractmethod
import numpy as np
from typing import Optional, Callable


class BaseWhisperProvider(ABC):
    """
    Abstract base class for Whisper model providers.

    Defines the common interface that all Whisper providers must implement,
    allowing different backends (faster-whisper, whisper.cpp, etc.) to be
    used interchangeably.
    """

    @abstractmethod
    def __init__(
        self,
        language_config: str = "auto",
        model_size: str = "small",
        live_quality_mode: str = "balanced",
        enable_overlap_detection: bool = True,
        debug_text_assembly: bool = False,
    ):
        """
        Initialize the Whisper provider with configuration.

        Args:
            language_config: "auto" for automatic detection, "en"/"fr"/etc for specific language
            model_size: "small", "base", "large-v3" etc.
            live_quality_mode: "fast", "balanced", or "accurate" for live transcription
            enable_overlap_detection: Whether to detect and remove text overlaps in live mode
            debug_text_assembly: Whether to enable verbose logging for text assembly debugging
        """
        pass

    @abstractmethod
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
        pass

    @abstractmethod
    def get_language_info(self) -> dict:
        """Get current language detection information"""
        pass

    @abstractmethod
    def transcribe_audio(
        self,
        audio_data: np.ndarray,
        sample_rate: int = 16000,
        timeout: Optional[float] = None,
    ) -> Optional[str]:
        """
        Transcribe audio data and return the transcribed text.

        Args:
            audio_data: Audio data as numpy array
            sample_rate: Sample rate of the audio data
            timeout: Optional timeout in seconds to prevent indefinite blocking

        Returns:
            Transcribed text or None if no speech detected
        """
        pass

    @abstractmethod
    def start_streaming_transcription(self, callback: Callable[[str], None]):
        """
        Start streaming transcription mode with text assembly.

        Args:
            callback: Function to call with each transcribed text chunk
        """
        pass

    @abstractmethod
    def stop_streaming_transcription(self):
        """Stop streaming transcription mode"""
        pass

    @abstractmethod
    def process_audio_chunk(self, audio_data: np.ndarray):
        """
        Process an audio chunk for streaming transcription.

        Args:
            audio_data: Audio chunk as numpy array
        """
        pass

    @abstractmethod
    def is_streaming_active(self) -> bool:
        """Check if streaming transcription is active"""
        pass

    @property
    @abstractmethod
    def provider_name(self) -> str:
        """Return the name of this provider"""
        pass

    @property
    @abstractmethod
    def provider_info(self) -> dict:
        """Return information about this provider"""
        pass

    def cleanup_resources(self):
        """Clean up provider resources to free memory (optional override)"""
        pass

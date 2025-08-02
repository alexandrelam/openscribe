"""
Whisper model providers package.

This package contains different implementations of Whisper model providers,
allowing users to choose between different backends (faster-whisper, whisper.cpp, etc.)
"""

from .base_provider import BaseWhisperProvider
from .faster_whisper_provider import FasterWhisperProvider

try:
    from .whisper_cpp_provider import WhisperCppProvider

    WHISPER_CPP_AVAILABLE = True
except ImportError:
    WHISPER_CPP_AVAILABLE = False
    WhisperCppProvider = None

__all__ = [
    "BaseWhisperProvider",
    "FasterWhisperProvider",
    "WhisperCppProvider",
    "WHISPER_CPP_AVAILABLE",
]

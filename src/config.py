import os
import json
from dataclasses import dataclass, asdict
from typing import Optional, List


@dataclass
class Config:
    hotkey: str = "alt_r"
    # Legacy field for backward compatibility (will be migrated to transcription_language)
    language: str = "en-US"
    timeout: int = 5
    microphone_preferences: List[int] = None
    always_on_top: bool = True
    enable_auto_insert: bool = True
    auto_insert_timeout: int = 10

    # Language and transcription configuration
    transcription_language: str = "auto"  # "auto", "en", "fr", etc.
    model_size: str = "small"  # "small", "base", "large-v3"
    whisper_provider: str = "faster-whisper"  # "faster-whisper", "whisper-cpp"
    language_detection_enabled: bool = True  # Show detected language info
    fallback_language: str = "en"  # Fallback when auto-detection fails

    # Live transcription quality configuration
    live_quality_mode: str = "balanced"  # "fast", "balanced", "accurate"
    enable_overlap_detection: bool = True  # Enable text overlap detection and removal
    debug_text_assembly: bool = False  # Enable verbose logging for text assembly

    # Transcription timeout configuration
    transcription_timeout: float = (
        30.0  # timeout in seconds for transcription (0 = no timeout)
    )

    # VAD configuration
    vad_aggressiveness: int = 2  # 0-3, higher = more aggressive
    vad_min_chunk_duration: float = 1.0  # minimum seconds per chunk
    vad_max_chunk_duration: float = 10.0  # maximum seconds per chunk
    vad_silence_timeout: float = 0.5  # seconds of silence before processing chunk

    # Text insertion configuration
    paste_method: str = "applescript"  # "applescript" or "keyboard"
    paste_delay: float = 0.05  # seconds to wait after copying to clipboard
    live_paste_interval: float = 0.3  # seconds between live paste operations
    restore_clipboard: bool = True  # whether to restore original clipboard content

    # Recording indicator configuration
    show_recording_indicator: bool = True  # whether to show recording indicator
    indicator_position_x: int = 20  # horizontal position on screen
    indicator_position_y: int = 20  # vertical position on screen
    indicator_size: int = 20  # size of the indicator in pixels
    indicator_opacity: float = 0.9  # transparency of the indicator (0.0-1.0)

    # Double key press shortcuts configuration
    double_press_enabled: bool = True  # enable double key press shortcuts
    double_press_timeout: float = 0.5  # maximum time between presses (seconds)

    # Sound notifications configuration
    sound_notifications_enabled: bool = True  # enable sound notifications
    sound_volume: float = 0.5  # volume level (0.0-1.0)

    CONFIG_FILE = "config.json"

    def __post_init__(self):
        """Initialize default values that can't be set in dataclass field defaults"""
        if self.microphone_preferences is None:
            self.microphone_preferences = []

    def get_preferred_device(self) -> Optional[int]:
        """Get the best available microphone device from preferences"""
        if not self.microphone_preferences:
            return None

        # Import here to avoid circular imports
        from .audio_recorder import AudioRecorder

        try:
            available_device_ids = AudioRecorder.get_available_device_ids()

            # Check preferences in order and return first available one
            for preferred_device_id in self.microphone_preferences:
                if preferred_device_id in available_device_ids:
                    return preferred_device_id

            # If no preferred devices are available, return None (system default)
            return None
        except Exception as e:
            print(f"⚠️ Error resolving preferred microphone device: {e}")
            return None

    @classmethod
    def _validate_config_data(cls, data: dict) -> dict:
        """Validate and sanitize configuration data"""
        validated = {}

        # Define validation rules for each field
        validators = {
            "microphone_preferences": lambda x: isinstance(x, list)
            and all(isinstance(item, int) and item >= -1 for item in x),
            "hotkey": lambda x: isinstance(x, str)
            and len(x) <= 50
            and (
                all(c.isalnum() or c in "+-_<>" for c in x)
                or x in ["alt_l", "alt_r", "shift_l", "shift_r", "ctrl_l", "ctrl_r"]
            ),
            "transcription_language": lambda x: isinstance(x, str)
            and (len(x) <= 10 and (x.isalpha() or x == "auto")),
            "model_size": lambda x: isinstance(x, str)
            and x
            in ["tiny", "base", "small", "medium", "large", "large-v2", "large-v3"],
            "vad_aggressiveness": lambda x: isinstance(x, int) and 0 <= x <= 3,
            "vad_min_chunk_duration": lambda x: isinstance(x, (int, float))
            and 0.1 <= x <= 10.0,
            "vad_max_chunk_duration": lambda x: isinstance(x, (int, float))
            and 1.0 <= x <= 30.0,
            "paste_method": lambda x: isinstance(x, str)
            and x in ["applescript", "keyboard"],
            "whisper_provider": lambda x: isinstance(x, str)
            and x in ["faster-whisper", "whisper-cpp"],
            "transcription_timeout": lambda x: isinstance(x, (int, float))
            and x >= 0
            and x <= 300,  # 0 to 5 minutes max
        }

        for key, value in data.items():
            # Only validate keys we have validators for
            if key in validators:
                try:
                    if validators[key](value):
                        validated[key] = value
                    else:
                        print(f"⚠️ Invalid config value for {key}: {value} (skipping)")
                except Exception:
                    print(f"⚠️ Error validating config key {key} (skipping)")
            else:
                # For keys without validators, pass through (backward compatibility)
                validated[key] = value

        return validated

    @classmethod
    def load(cls) -> "Config":
        if os.path.exists(cls.CONFIG_FILE):
            try:
                with open(cls.CONFIG_FILE, "r", encoding="utf-8") as f:
                    data = json.load(f)

                # Validate configuration data
                validated_data = cls._validate_config_data(data)

                # Handle legacy config migration
                config = cls(**validated_data)
                config._migrate_legacy_settings()
                return config
            except (json.JSONDecodeError, FileNotFoundError) as e:
                print(f"Failed to load config (JSON error): {e}. Using defaults.")
            except Exception as e:
                print(f"Unexpected error loading config: {e}. Using defaults.")
        return cls()

    def _migrate_legacy_settings(self):
        """Migrate legacy language and microphone settings to new format"""
        # Migrate legacy microphone_device to microphone_preferences
        if hasattr(self, "microphone_device") and self.microphone_device is not None:
            if not self.microphone_preferences:  # Only migrate if preferences are empty
                self.microphone_preferences = [self.microphone_device]
                print(
                    f"🔄 Migrated microphone device: {self.microphone_device} → preferences: {self.microphone_preferences}"
                )
            # Remove the old field to avoid confusion
            delattr(self, "microphone_device")
        # If transcription_language is still default but legacy language exists
        if (
            self.transcription_language == "auto"
            and hasattr(self, "language")
            and self.language
            and self.language != "en-US"
        ):

            # Convert common legacy formats
            if self.language.startswith("en"):
                self.transcription_language = "en"
            elif self.language.startswith("fr"):
                self.transcription_language = "fr"
            else:
                # Try to extract language code from locale format
                lang_code = self.language.split("-")[0].lower()
                if lang_code in [
                    "en",
                    "fr",
                    "es",
                    "de",
                    "it",
                    "pt",
                    "ru",
                    "zh",
                    "ja",
                    "ko",
                ]:
                    self.transcription_language = lang_code

            print(
                f"🔄 Migrated language setting: '{self.language}' → '{self.transcription_language}'"
            )

    @staticmethod
    def get_available_languages() -> dict:
        """Get available language options for the GUI"""
        return {
            "auto": "Automatic Detection",
            "en": "English",
            "fr": "French (Français)",
            "es": "Spanish (Español)",
            "de": "German (Deutsch)",
            "it": "Italian (Italiano)",
            "pt": "Portuguese (Português)",
            "ru": "Russian (Русский)",
            "zh": "Chinese (中文)",
            "ja": "Japanese (日本語)",
            "ko": "Korean (한국어)",
        }

    @staticmethod
    def get_available_models() -> dict:
        """Get available model size options"""
        return {
            "small": "Small (39M parameters, fast)",
            "base": "Base (74M parameters, balanced)",
            "large-v3": "Large V3 (1550M parameters, most accurate)",
        }

    @staticmethod
    def get_live_quality_modes() -> dict:
        """Get available live transcription quality modes"""
        return {
            "fast": "Fast (Low latency, basic accuracy)",
            "balanced": "Balanced (Good speed and accuracy)",
            "accurate": "Accurate (Best quality, higher latency)",
        }

    @staticmethod
    def get_available_providers() -> dict:
        """Get available Whisper model providers"""
        return {
            "faster-whisper": "Faster Whisper (Fast, GPU-optimized)",
            "whisper-cpp": "Whisper.cpp (CPU-optimized, CLI-based C++ implementation)",
        }

    def save(self) -> None:
        try:
            with open(self.CONFIG_FILE, "w", encoding="utf-8") as f:
                json.dump(asdict(self), f, indent=2)
        except Exception as e:
            print(f"Failed to save config: {e}")

import os
import json
from dataclasses import dataclass, asdict
from typing import Optional

@dataclass
class Config:
    hotkey: str = "cmd+shift+r"
    # Legacy field for backward compatibility (will be migrated to transcription_language)
    language: str = "en-US"
    timeout: int = 5
    microphone_device: Optional[int] = None
    always_on_top: bool = True
    enable_auto_insert: bool = True
    auto_insert_timeout: int = 10
    
    # Language and transcription configuration
    transcription_language: str = "auto"  # "auto", "en", "fr", etc.
    model_size: str = "small"  # "small", "base", "large-v3"
    language_detection_enabled: bool = True  # Show detected language info
    fallback_language: str = "en"  # Fallback when auto-detection fails
    
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
    
    CONFIG_FILE = "config.json"
    
    @classmethod
    def load(cls) -> 'Config':
        if os.path.exists(cls.CONFIG_FILE):
            try:
                with open(cls.CONFIG_FILE, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                
                # Handle legacy config migration
                config = cls(**data)
                config._migrate_legacy_settings()
                return config
            except Exception as e:
                print(f"Failed to load config: {e}. Using defaults.")
        return cls()
    
    def _migrate_legacy_settings(self):
        """Migrate legacy language settings to new format"""
        # If transcription_language is still default but legacy language exists
        if (self.transcription_language == "auto" and 
            hasattr(self, 'language') and self.language and 
            self.language != "en-US"):
            
            # Convert common legacy formats
            if self.language.startswith("en"):
                self.transcription_language = "en"
            elif self.language.startswith("fr"):
                self.transcription_language = "fr"
            else:
                # Try to extract language code from locale format
                lang_code = self.language.split('-')[0].lower()
                if lang_code in ['en', 'fr', 'es', 'de', 'it', 'pt', 'ru', 'zh', 'ja', 'ko']:
                    self.transcription_language = lang_code
            
            print(f"🔄 Migrated language setting: '{self.language}' → '{self.transcription_language}'")
    
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
            "ko": "Korean (한국어)"
        }
    
    @staticmethod
    def get_available_models() -> dict:
        """Get available model size options"""
        return {
            "small": "Small (39M parameters, fast)",
            "base": "Base (74M parameters, balanced)", 
            "large-v3": "Large V3 (1550M parameters, most accurate)"
        }
    
    def save(self) -> None:
        try:
            with open(self.CONFIG_FILE, 'w', encoding='utf-8') as f:
                json.dump(asdict(self), f, indent=2)
        except Exception as e:
            print(f"Failed to save config: {e}")
import os
import json
from dataclasses import dataclass, asdict
from typing import Optional

@dataclass
class Config:
    hotkey: str = "cmd+shift+r"
    language: str = "en-US"
    timeout: int = 5
    microphone_device: Optional[int] = None
    always_on_top: bool = True
    enable_auto_insert: bool = True
    auto_insert_timeout: int = 10
    
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
    
    CONFIG_FILE = "config.json"
    
    @classmethod
    def load(cls) -> 'Config':
        if os.path.exists(cls.CONFIG_FILE):
            try:
                with open(cls.CONFIG_FILE, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                return cls(**data)
            except Exception as e:
                print(f"Failed to load config: {e}. Using defaults.")
        return cls()
    
    def save(self) -> None:
        try:
            with open(self.CONFIG_FILE, 'w', encoding='utf-8') as f:
                json.dump(asdict(self), f, indent=2)
        except Exception as e:
            print(f"Failed to save config: {e}")
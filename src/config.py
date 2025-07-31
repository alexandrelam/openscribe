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
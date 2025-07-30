import os
from dataclasses import dataclass
from typing import Optional

@dataclass
class Config:
    hotkey: str = "cmd+shift+r"
    language: str = "en-US"
    timeout: int = 5
    microphone_device: Optional[int] = None
    always_on_top: bool = True
    
    @classmethod
    def load(cls) -> 'Config':
        return cls()
    
    def save(self) -> None:
        pass
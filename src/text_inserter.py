import pyautogui
import time
from typing import Optional

class TextInserter:
    def __init__(self):
        pyautogui.FAILSAFE = True
        pyautogui.PAUSE = 0.1
    
    def insert_text(self, text: str) -> bool:
        try:
            time.sleep(0.1)
            pyautogui.typewrite(text)
            return True
        except Exception as e:
            print(f"Failed to insert text: {e}")
            return False
    
    def copy_to_clipboard(self, text: str) -> bool:
        try:
            import pyperclip
            pyperclip.copy(text)
            return True
        except Exception as e:
            print(f"Failed to copy to clipboard: {e}")
            return False
    
    def get_focused_window(self) -> Optional[str]:
        try:
            import pygetwindow as gw
            active_window = gw.getActiveWindow()
            return active_window.title if active_window else None
        except Exception as e:
            print(f"Failed to get focused window: {e}")
            return None
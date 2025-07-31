import pyautogui
import time
import threading
from typing import Optional, Callable

class TextInserter:
    def __init__(self):
        pyautogui.FAILSAFE = True
        pyautogui.PAUSE = 0.1
        self.pending_text: Optional[str] = None
        self.auto_insert_active = False
        self.mouse_listener = None
        self.timeout_timer: Optional[threading.Timer] = None
        self.on_auto_insert_complete: Optional[Callable[[bool], None]] = None
    
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
    
    def start_auto_insert_mode(self, text: str, timeout_seconds: int = 10, 
                              on_complete: Optional[Callable[[bool], None]] = None) -> bool:
        """Start auto-insert mode that will insert text on next click"""
        try:
            if self.auto_insert_active:
                self.stop_auto_insert_mode()
            
            self.pending_text = text
            self.auto_insert_active = True
            self.on_auto_insert_complete = on_complete
            
            # Start mouse listener
            from pynput import mouse
            self.mouse_listener = mouse.Listener(on_click=self._on_mouse_click)
            self.mouse_listener.start()
            
            # Start timeout timer
            self.timeout_timer = threading.Timer(timeout_seconds, self._on_timeout)
            self.timeout_timer.start()
            
            return True
        except Exception as e:
            print(f"Failed to start auto-insert mode: {e}")
            return False
    
    def stop_auto_insert_mode(self):
        """Stop auto-insert mode and cleanup resources"""
        self.auto_insert_active = False
        self.pending_text = None
        
        if self.mouse_listener:
            self.mouse_listener.stop()
            self.mouse_listener = None
        
        if self.timeout_timer:
            self.timeout_timer.cancel()
            self.timeout_timer = None
    
    def _on_mouse_click(self, x, y, button, pressed):
        """Handle mouse click during auto-insert mode"""
        if not self.auto_insert_active or not pressed or not self.pending_text:
            return
        
        try:
            # Wait a moment for the input field to focus
            time.sleep(0.2)
            
            # Insert the pending text
            success = self.insert_text(self.pending_text)
            
            # Stop auto-insert mode
            self.stop_auto_insert_mode()
            
            # Notify completion
            if self.on_auto_insert_complete:
                self.on_auto_insert_complete(success)
                
        except Exception as e:
            print(f"Error during auto-insert: {e}")
            self.stop_auto_insert_mode()
            if self.on_auto_insert_complete:
                self.on_auto_insert_complete(False)
    
    def _on_timeout(self):
        """Handle timeout for auto-insert mode"""
        if self.auto_insert_active:
            self.stop_auto_insert_mode()
            if self.on_auto_insert_complete:
                self.on_auto_insert_complete(False)
    
    def is_auto_insert_active(self) -> bool:
        """Check if auto-insert mode is currently active"""
        return self.auto_insert_active
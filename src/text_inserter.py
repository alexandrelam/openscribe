import pyautogui
import time
import threading
import subprocess
import shlex
from typing import Optional, Callable

class TextInserter:
    def __init__(self):
        pyautogui.FAILSAFE = True
        pyautogui.PAUSE = 0.01  # Faster for live typing
        self.pending_text: Optional[str] = None
        self.auto_insert_active = False
        self.mouse_listener = None
        self.timeout_timer: Optional[threading.Timer] = None
        self.on_auto_insert_complete: Optional[Callable[[bool], None]] = None
        
        # Live typing mode
        self.live_typing_active = False
        self.typing_queue = []
        self.typing_lock = threading.Lock()
        self.typing_thread: Optional[threading.Thread] = None
        self.typing_stop_event = threading.Event()
    
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
    
    def start_live_typing_mode(self):
        """Start live typing mode for real-time text insertion"""
        if self.live_typing_active:
            return
        
        self.live_typing_active = True
        self.typing_stop_event.clear()
        self.typing_queue = []
        
        # Start the typing thread
        self.typing_thread = threading.Thread(target=self._live_typing_processor, daemon=True)
        self.typing_thread.start()
    
    def stop_live_typing_mode(self):
        """Stop live typing mode"""
        if not self.live_typing_active:
            return
        
        self.live_typing_active = False
        self.typing_stop_event.set()
        
        if self.typing_thread and self.typing_thread.is_alive():
            self.typing_thread.join(timeout=2.0)
        
        with self.typing_lock:
            self.typing_queue.clear()
    
    def queue_text_for_live_typing(self, text: str):
        """Queue text for immediate live typing at cursor position"""
        text = text.strip()
        if not self.live_typing_active or not text:
            return
        
        with self.typing_lock:
            # Add space before text if queue is not empty and text doesn't start with punctuation
            if self.typing_queue and not text.startswith((' ', '.', ',', '!', '?', ';', ':')):
                text = ' ' + text
            self.typing_queue.append(text)
    
    def _live_typing_processor(self):
        """Process queued text for live typing"""
        while not self.typing_stop_event.is_set():
            try:
                text_to_type = None
                
                with self.typing_lock:
                    if self.typing_queue:
                        text_to_type = self.typing_queue.pop(0)
                
                if text_to_type:
                    self._type_text_immediately(text_to_type)
                    # Add longer delay after typing to prevent overwhelming the system
                    time.sleep(0.5)  # 500ms delay between text chunks
                else:
                    # Small sleep to prevent excessive CPU usage
                    time.sleep(0.1)
                    
            except Exception as e:
                print(f"Error in live typing processor: {e}")
                time.sleep(0.3)
    
    def _type_text_immediately(self, text: str) -> bool:
        """Type text immediately at current cursor position using native macOS typing"""
        try:
            # Clean and prepare text
            cleaned_text = self._clean_text_for_typing(text)
            if not cleaned_text:
                return True
            
            # Use native macOS typing via AppleScript for proper keyboard layout support
            return self._type_with_applescript(cleaned_text)
            
        except Exception as e:
            print(f"Failed to type text immediately: {e}")
            # Fallback to pyautogui as last resort
            try:
                pyautogui.typewrite(text, interval=0.1)
                return True
            except Exception as e2:
                print(f"All typing methods failed: {e2}")
                return False
    
    def _clean_text_for_typing(self, text: str) -> str:
        """Clean and normalize text before typing"""
        if not text:
            return ""
        
        # Remove any control characters and normalize whitespace
        cleaned = ''.join(char for char in text if ord(char) >= 32 or char in '\n\t')
        
        # Normalize multiple spaces to single space
        cleaned = ' '.join(cleaned.split())
        
        # Remove any problematic characters that might cause issues
        problematic_chars = ['', '', '', '']  # Various control characters
        for char in problematic_chars:
            cleaned = cleaned.replace(char, '')
        
        return cleaned.strip()
    
    def _type_with_applescript(self, text: str) -> bool:
        """Use AppleScript to type text with proper keyboard layout support"""
        try:
            # Escape text for AppleScript (handle quotes and backslashes)
            escaped_text = text.replace('\\', '\\\\').replace('"', '\\"')
            
            # Create AppleScript command to type the text
            applescript = f'''
            tell application "System Events"
                keystroke "{escaped_text}"
            end tell
            '''
            
            # Execute AppleScript via osascript
            result = subprocess.run(
                ['osascript', '-e', applescript],
                capture_output=True,
                text=True,
                timeout=5
            )
            
            if result.returncode == 0:
                return True
            else:
                print(f"AppleScript typing failed: {result.stderr}")
                return False
                
        except subprocess.TimeoutExpired:
            print("AppleScript typing timed out")
            return False
        except Exception as e:
            print(f"AppleScript typing error: {e}")
            return False
    
    def is_live_typing_active(self) -> bool:
        """Check if live typing mode is currently active"""
        return self.live_typing_active
import pyautogui
import pyperclip
import time
import threading
import subprocess
import shlex
from typing import Optional, Callable


class TextInserter:
    def __init__(self):
        pyautogui.FAILSAFE = True
        pyautogui.PAUSE = 0.01
        self.pending_text: Optional[str] = None
        self.auto_insert_active = False
        self.mouse_listener = None
        self.timeout_timer: Optional[threading.Timer] = None
        self.on_auto_insert_complete: Optional[Callable[[bool], None]] = None

        # Live pasting mode (renamed from typing)
        self.live_pasting_active = False
        self.paste_queue = []
        self.paste_lock = threading.Lock()
        self.paste_thread: Optional[threading.Thread] = None
        self.paste_stop_event = threading.Event()

        # Clipboard management
        self._original_clipboard: Optional[str] = None

        # Configuration
        self.paste_method = "applescript"  # "applescript" or "keyboard"
        self.paste_delay = 0.05  # seconds to wait after copying to clipboard
        self.live_paste_interval = 0.3  # seconds between live paste operations
        self.restore_clipboard = True  # whether to restore original clipboard content

    def configure_pasting(
        self,
        paste_method: str = "applescript",
        paste_delay: float = 0.05,
        live_paste_interval: float = 0.3,
        restore_clipboard: bool = True,
    ):
        """Configure text pasting behavior"""
        self.paste_method = paste_method
        self.paste_delay = paste_delay
        self.live_paste_interval = live_paste_interval
        self.restore_clipboard = restore_clipboard

        print(
            f"🔧 Text pasting configured: method={paste_method}, "
            f"delay={paste_delay}s, interval={live_paste_interval}s, "
            f"restore_clipboard={restore_clipboard}"
        )

    def insert_text(self, text: str) -> bool:
        """Insert text using clipboard + paste for speed and reliability"""
        try:
            # Clean text for pasting
            cleaned_text = self._clean_text_for_pasting(text)
            if not cleaned_text:
                return True

            # Use clipboard-safe pasting
            return self._paste_text_safely(cleaned_text)

        except Exception as e:
            print(f"Failed to insert text: {e}")
            return False

    def copy_to_clipboard(self, text: str) -> bool:
        """Copy text to clipboard (public method for external use)"""
        try:
            pyperclip.copy(text)
            return True
        except Exception as e:
            print(f"Failed to copy to clipboard: {e}")
            return False

    def _backup_clipboard(self) -> bool:
        """Backup current clipboard content"""
        try:
            self._original_clipboard = pyperclip.paste()
            return True
        except Exception as e:
            print(f"Failed to backup clipboard: {e}")
            self._original_clipboard = None
            return False

    def _restore_clipboard(self) -> bool:
        """Restore original clipboard content"""
        try:
            if self.restore_clipboard and self._original_clipboard is not None:
                pyperclip.copy(self._original_clipboard)
                self._original_clipboard = None
            return True
        except Exception as e:
            print(f"Failed to restore clipboard: {e}")
            return False

    def _paste_text_safely(self, text: str) -> bool:
        """Paste text using clipboard while preserving original clipboard content"""
        try:
            # Backup current clipboard
            self._backup_clipboard()

            # Copy our text to clipboard
            pyperclip.copy(text)
            time.sleep(self.paste_delay)  # Configurable delay for clipboard to update

            # Paste using preferred method
            success = self._execute_paste()

            # Restore original clipboard (with small delay)
            time.sleep(self.paste_delay * 2)  # Slightly longer delay before restore
            self._restore_clipboard()

            return success

        except Exception as e:
            print(f"Failed to paste text safely: {e}")
            # Try to restore clipboard even if paste failed
            self._restore_clipboard()
            return False

    def _execute_paste(self) -> bool:
        """Execute paste operation using configured method"""
        try:
            if self.paste_method == "applescript":
                # Method 1: AppleScript paste (most reliable on macOS)
                if self._paste_with_applescript():
                    return True
                # Fallback to keyboard shortcut if AppleScript fails
                print("AppleScript paste failed, falling back to keyboard shortcut")
                pyautogui.hotkey("cmd", "v")
                return True

            elif self.paste_method == "keyboard":
                # Method 2: Direct keyboard shortcut
                pyautogui.hotkey("cmd", "v")
                return True

            else:
                print(f"Unknown paste method: {self.paste_method}, using default")
                pyautogui.hotkey("cmd", "v")
                return True

        except Exception as e:
            print(f"Paste operation failed: {e}")
            return False

    def _paste_with_applescript(self) -> bool:
        """Paste using AppleScript for maximum compatibility"""
        try:
            applescript = """
            tell application "System Events"
                keystroke "v" using command down
            end tell
            """

            result = subprocess.run(
                ["osascript", "-e", applescript],
                capture_output=True,
                text=True,
                timeout=2,
            )

            return result.returncode == 0

        except Exception as e:
            print(f"AppleScript paste failed: {e}")
            return False

    def _clean_text_for_pasting(self, text: str) -> str:
        """Clean and normalize text before pasting"""
        if not text:
            return ""

        # Remove control characters but keep newlines and tabs
        cleaned = "".join(char for char in text if ord(char) >= 32 or char in "\n\t")

        # Normalize whitespace but preserve intentional formatting
        lines = cleaned.split("\n")
        cleaned_lines = [" ".join(line.split()) for line in lines]
        cleaned = "\n".join(cleaned_lines)

        return cleaned.strip()

    def get_focused_window(self) -> Optional[str]:
        try:
            import pygetwindow as gw

            active_window = gw.getActiveWindow()
            return active_window.title if active_window else None
        except Exception as e:
            print(f"Failed to get focused window: {e}")
            return None

    def start_auto_insert_mode(
        self,
        text: str,
        timeout_seconds: int = 10,
        on_complete: Optional[Callable[[bool], None]] = None,
    ) -> bool:
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

            # Paste the pending text
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
        """Start live pasting mode for real-time text insertion"""
        if self.live_pasting_active:
            return

        self.live_pasting_active = True
        self.paste_stop_event.clear()
        self.paste_queue = []

        # Start the pasting thread
        self.paste_thread = threading.Thread(
            target=self._live_pasting_processor, daemon=True
        )
        self.paste_thread.start()

        print("🚀 Live pasting mode started - text will be pasted in real-time")

    def stop_live_typing_mode(self):
        """Stop live pasting mode"""
        if not self.live_pasting_active:
            return

        self.live_pasting_active = False
        self.paste_stop_event.set()

        if self.paste_thread and self.paste_thread.is_alive():
            self.paste_thread.join(timeout=2.0)

        with self.paste_lock:
            self.paste_queue.clear()

        print("⏹️ Live pasting mode stopped")

    def queue_text_for_live_typing(self, text: str):
        """Queue text for immediate live pasting at cursor position"""
        text = text.strip()
        if not self.live_pasting_active or not text:
            return

        with self.paste_lock:
            # Add space before text if queue is not empty and text doesn't start with punctuation
            if self.paste_queue and not text.startswith(
                (" ", ".", ",", "!", "?", ";", ":")
            ):
                text = " " + text
            self.paste_queue.append(text)

        print(f"📝 Queued text for pasting: '{text}'")

    def _live_pasting_processor(self):
        """Process queued text for live pasting"""
        while not self.paste_stop_event.is_set():
            try:
                text_to_paste = None

                with self.paste_lock:
                    if self.paste_queue:
                        # For live pasting, we can either paste each chunk individually
                        # or combine multiple small chunks for efficiency
                        text_to_paste = self.paste_queue.pop(0)

                if text_to_paste:
                    success = self._paste_text_safely(text_to_paste)
                    if success:
                        print(f"✅ Pasted: '{text_to_paste}'")
                    else:
                        print(f"❌ Failed to paste: '{text_to_paste}'")

                    # Configurable delay between paste operations
                    time.sleep(self.live_paste_interval)
                else:
                    # Small sleep to prevent excessive CPU usage
                    time.sleep(0.1)

            except Exception as e:
                print(f"Error in live pasting processor: {e}")
                time.sleep(0.3)

    def is_live_typing_active(self) -> bool:
        """Check if live pasting mode is currently active"""
        return self.live_pasting_active

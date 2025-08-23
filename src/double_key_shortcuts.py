import time
import threading
from typing import Callable, Optional
from enum import Enum
from pynput import keyboard


class KeyType(Enum):
    SHIFT = "shift"
    CONTROL = "control"
    ALT = "alt"


class DoubleKeyDetector:
    """Detects double key presses for Shift and Control keys, and single key presses to trigger recording modes."""

    def __init__(self, double_press_timeout: float = 0.5):
        self.double_press_timeout = double_press_timeout
        self.listener: Optional[keyboard.Listener] = None
        self.running = False

        # Callbacks for different actions
        self.on_double_shift: Optional[Callable] = None
        self.on_double_control: Optional[Callable] = None
        self.on_double_alt: Optional[Callable] = None
        self.on_single_key: Optional[Callable] = None

        # Single key configuration
        self.single_key_target: Optional[keyboard.Key] = None

        # State tracking for double presses
        self._shift_state = self._create_key_state()
        self._control_state = self._create_key_state()
        self._alt_state = self._create_key_state()

        # Lock for thread safety
        self._lock = threading.Lock()

    def _create_key_state(self) -> dict:
        """Create initial state for a key"""
        return {"last_press_time": 0, "press_count": 0, "is_pressed": False}

    def set_callbacks(
        self,
        on_double_shift: Callable = None,
        on_double_control: Callable = None,
        on_double_alt: Callable = None,
        on_single_key: Callable = None,
    ):
        """Set callback functions for double press and single key events"""
        self.on_double_shift = on_double_shift
        self.on_double_control = on_double_control
        self.on_double_alt = on_double_alt
        self.on_single_key = on_single_key

    def configure_single_key(self, key_name: str):
        """Configure single key detection for specific keys"""
        single_key_map = {
            "alt_l": keyboard.Key.alt_l,
            "alt_r": keyboard.Key.alt_r,
            "shift_l": keyboard.Key.shift_l,
            "shift_r": keyboard.Key.shift_r,
            "ctrl_l": keyboard.Key.ctrl_l,
            "ctrl_r": keyboard.Key.ctrl_r,
        }

        if key_name in single_key_map:
            self.single_key_target = single_key_map[key_name]
        else:
            self.single_key_target = None

    def start(self):
        """Start listening for double key presses"""
        if self.running:
            return

        self.running = True
        self.listener = keyboard.Listener(
            on_press=self._on_key_press, on_release=self._on_key_release
        )
        self.listener.start()

    def stop(self):
        """Stop listening for double key presses"""
        if not self.running:
            return

        self.running = False
        if self.listener:
            self.listener.stop()
            self.listener = None

    def _on_key_press(self, key):
        """Handle key press events"""
        with self._lock:
            current_time = time.time()

            # Check for right Alt key (right Option key on macOS)
            if key == keyboard.Key.alt_r:
                self._handle_key_press(KeyType.ALT, current_time)

            # Check for Shift key
            elif key in [
                keyboard.Key.shift,
                keyboard.Key.shift_l,
                keyboard.Key.shift_r,
            ]:
                self._handle_key_press(KeyType.SHIFT, current_time)

            # Check for Control key
            elif key in [keyboard.Key.ctrl, keyboard.Key.ctrl_l, keyboard.Key.ctrl_r]:
                self._handle_key_press(KeyType.CONTROL, current_time)

    def _on_key_release(self, key):
        """Handle key release events"""
        with self._lock:
            # Mark keys as released
            if key == keyboard.Key.alt_r:
                self._alt_state["is_pressed"] = False
            elif key in [
                keyboard.Key.shift,
                keyboard.Key.shift_l,
                keyboard.Key.shift_r,
            ]:
                self._shift_state["is_pressed"] = False
            elif key in [keyboard.Key.ctrl, keyboard.Key.ctrl_l, keyboard.Key.ctrl_r]:
                self._control_state["is_pressed"] = False

    def _handle_key_press(self, key_type: KeyType, current_time: float):
        """Handle key press logic for double press detection"""
        if key_type == KeyType.SHIFT:
            state = self._shift_state
        elif key_type == KeyType.CONTROL:
            state = self._control_state
        else:  # KeyType.ALT
            state = self._alt_state

        # If key is already pressed, ignore (holding key down)
        if state["is_pressed"]:
            return

        state["is_pressed"] = True

        # Check if this is within the double press timeout window
        time_since_last = current_time - state["last_press_time"]

        if time_since_last <= self.double_press_timeout:
            # This is the second press - trigger double press action
            state["press_count"] = 0  # Reset count
            state["last_press_time"] = 0  # Reset timer

            # Trigger appropriate callback
            if key_type == KeyType.SHIFT and self.on_double_shift:
                try:
                    self.on_double_shift()
                except Exception as e:
                    print(f"Error in double shift callback: {e}")
            elif key_type == KeyType.CONTROL and self.on_double_control:
                try:
                    self.on_double_control()
                except Exception as e:
                    print(f"Error in double control callback: {e}")
            elif key_type == KeyType.ALT and self.on_double_alt:
                try:
                    self.on_double_alt()
                except Exception as e:
                    print(f"Error in double alt callback: {e}")
        else:
            # This is the first press or timeout has passed
            state["press_count"] = 1
            state["last_press_time"] = current_time

    def configure(self, double_press_timeout: float):
        """Update configuration"""
        self.double_press_timeout = double_press_timeout

import tkinter as tk
import threading
import time
from typing import Optional
from enum import Enum


class IndicatorState(Enum):
    HIDDEN = "hidden"
    RECORDING = "recording"
    LIVE_MODE = "live_mode"
    PROCESSING = "processing"


class RecordingIndicator:
    def __init__(self, position_x: int = 20, position_y: int = 20, size: int = 20, opacity: float = 0.9):
        self.position_x = position_x
        self.position_y = position_y
        self.size = size
        self.opacity = opacity
        
        # State management
        self.current_state = IndicatorState.HIDDEN
        self.is_running = False
        
        # Tkinter window
        self.window: Optional[tk.Toplevel] = None
        self.canvas: Optional[tk.Canvas] = None
        
        # Animation for pulsing effect
        self.animation_thread: Optional[threading.Thread] = None
        self.animation_stop_event = threading.Event()
        self.animation_phase = 0
    
    def start(self, parent_window: tk.Tk):
        """Initialize and start the indicator"""
        if self.is_running:
            return
        
        self.is_running = True
        
        # Create the indicator window
        self._create_window(parent_window)
        
        print("🎯 Fixed position recording indicator started")
        
        # Test visibility by briefly showing the indicator
        self.set_state(IndicatorState.RECORDING)
        parent_window.after(2000, lambda: self.set_state(IndicatorState.HIDDEN) if self.current_state == IndicatorState.RECORDING else None)
    
    def stop(self):
        """Stop the indicator and cleanup"""
        if not self.is_running:
            return
        
        self.is_running = False
        
        # Stop animation
        if self.animation_thread and self.animation_thread.is_alive():
            self.animation_stop_event.set()
            self.animation_thread.join(timeout=1.0)
        
        # Close window
        if self.window:
            try:
                self.window.withdraw()
                self.window.destroy()
            except tk.TclError:
                pass  # Window may already be destroyed
            self.window = None
            self.canvas = None
        
        print("🎯 Recording indicator stopped")
    
    def set_state(self, state: IndicatorState):
        """Update the indicator state"""
        if not self.is_running:
            return
        
        self.current_state = state
        
        if state == IndicatorState.HIDDEN:
            self._hide_window()
        else:
            self._show_window()
            self._update_appearance()
            
            # Start animation for live mode
            if state == IndicatorState.LIVE_MODE and (
                not self.animation_thread or not self.animation_thread.is_alive()
            ):
                self._start_animation()
            elif state != IndicatorState.LIVE_MODE:
                self._stop_animation()
    
    def _create_window(self, parent_window: tk.Tk):
        """Create the fixed position indicator window"""
        self.window = tk.Toplevel(parent_window)
        
        # Configure window properties
        self.window.overrideredirect(True)  # Remove window decorations
        self.window.attributes('-topmost', True)  # Always on top
        
        # Set transparency
        try:
            self.window.attributes('-alpha', self.opacity)
        except tk.TclError:
            print("Warning: Window transparency not supported on this platform")
        
        # Set fixed position and size
        geometry = f"{self.size}x{self.size}+{self.position_x}+{self.position_y}"
        self.window.geometry(geometry)
        
        # Create canvas for drawing
        self.canvas = tk.Canvas(
            self.window,
            width=self.size,
            height=self.size,
            bg='black',  # Black background for transparency
            highlightthickness=0,
            bd=0
        )
        self.canvas.pack()
        
        # Initially hide the window
        self.window.withdraw()
        
        print(f"🎯 Fixed indicator window created at ({self.position_x}, {self.position_y}): {self.size}x{self.size}")
    
    
    def _show_window(self):
        """Show the indicator window"""
        if self.window:
            try:
                self.window.deiconify()
                print(f"🎯 Indicator window shown at ({self.position_x}, {self.position_y})")
            except tk.TclError as e:
                print(f"Error showing indicator window: {e}")
    
    def _hide_window(self):
        """Hide the indicator window"""
        if self.window:
            try:
                self.window.withdraw()
                print("🎯 Indicator window hidden")
            except tk.TclError:
                pass
    
    def _update_appearance(self):
        """Update the visual appearance based on current state"""
        if not self.canvas:
            return
        
        try:
            self.canvas.delete("all")
            
            center = self.size // 2
            radius = (self.size - 4) // 2
            
            if self.current_state == IndicatorState.RECORDING:
                # Red filled circle for recording
                self.canvas.create_oval(
                    center - radius, center - radius,
                    center + radius, center + radius,
                    fill='red', outline='darkred', width=1
                )
            
            elif self.current_state == IndicatorState.LIVE_MODE:
                # Pulsing red circle for live mode (animation handled separately)
                alpha = 0.5 + 0.5 * abs(self.animation_phase)
                color_intensity = int(255 * alpha)
                color = f"#{color_intensity:02x}0000"
                
                self.canvas.create_oval(
                    center - radius, center - radius,
                    center + radius, center + radius,
                    fill=color, outline='red', width=2
                )
                
                # Add inner dot
                inner_radius = radius // 2
                self.canvas.create_oval(
                    center - inner_radius, center - inner_radius,
                    center + inner_radius, center + inner_radius,
                    fill='white', outline=''
                )
            
            elif self.current_state == IndicatorState.PROCESSING:
                # Spinning indicator for processing
                angle = (time.time() * 360) % 360
                
                # Draw spinning arc
                self.canvas.create_arc(
                    center - radius, center - radius,
                    center + radius, center + radius,
                    start=angle, extent=120,
                    outline='orange', width=3, style='arc'
                )
                
                # Center dot
                self.canvas.create_oval(
                    center - 2, center - 2,
                    center + 2, center + 2,
                    fill='orange', outline=''
                )
        
        except tk.TclError:
            pass  # Canvas may be destroyed
    
    def _start_animation(self):
        """Start the animation thread for live mode"""
        self.animation_stop_event.clear()
        self.animation_thread = threading.Thread(target=self._animate, daemon=True)
        self.animation_thread.start()
    
    def _stop_animation(self):
        """Stop the animation thread"""
        if self.animation_thread and self.animation_thread.is_alive():
            self.animation_stop_event.set()
    
    def _animate(self):
        """Animation loop for pulsing effects"""
        while not self.animation_stop_event.is_set():
            if self.current_state == IndicatorState.LIVE_MODE:
                # Pulsing animation
                self.animation_phase = (self.animation_phase + 0.1) % (2 * 3.14159)
                
                if self.window:
                    try:
                        self.window.after_idle(self._update_appearance)
                    except tk.TclError:
                        break
                
                time.sleep(0.05)  # 20 FPS animation
            
            elif self.current_state == IndicatorState.PROCESSING:
                # Continuous update for spinning animation
                if self.window:
                    try:
                        self.window.after_idle(self._update_appearance)
                    except tk.TclError:
                        break
                
                time.sleep(0.03)  # ~33 FPS for smooth spinning
            
            else:
                time.sleep(0.1)  # Slower update when not animating
    
    def configure_appearance(self, position_x: int = None, position_y: int = None, 
                           size: int = None, opacity: float = None):
        """Update appearance configuration"""
        if position_x is not None:
            self.position_x = position_x
        if position_y is not None:
            self.position_y = position_y
        if size is not None:
            self.size = size
            if self.window:
                geometry = f"{self.size}x{self.size}+{self.position_x}+{self.position_y}"
                self.window.geometry(geometry)
                if self.canvas:
                    self.canvas.config(width=self.size, height=self.size)
        if opacity is not None:
            self.opacity = opacity
            if self.window:
                try:
                    self.window.attributes('-alpha', self.opacity)
                except tk.TclError:
                    pass
        
        # Update appearance and position if currently visible
        if self.current_state != IndicatorState.HIDDEN:
            self._update_appearance()
            if self.window:
                geometry = f"{self.size}x{self.size}+{self.position_x}+{self.position_y}"
                self.window.geometry(geometry)
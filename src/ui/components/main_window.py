import tkinter as tk
from tkinter import ttk
from typing import Callable, Optional
from ...config import Config


class MainWindow:
    """
    Main application window component.
    Handles window setup, layout, and basic window management.
    """

    def __init__(self, title: str = "Speech to Text"):
        self.root = tk.Tk()
        self.root.title(title)
        self.root.geometry("300x200")
        self.root.resizable(False, False)

        # Store callbacks for window events
        self.on_closing: Optional[Callable] = None

        self.setup_window()

    def setup_window(self):
        """Setup the main window layout and configuration"""
        # Main frame that will contain all components
        self.main_frame = ttk.Frame(self.root, padding="20")
        self.main_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))

        # Configure grid weights for proper resizing
        self.root.columnconfigure(0, weight=1)
        self.root.rowconfigure(0, weight=1)
        self.main_frame.columnconfigure(0, weight=1)
        self.main_frame.columnconfigure(1, weight=1)

    def set_close_callback(self, callback: Callable):
        """Set the callback function for window close events"""
        self.on_closing = callback
        self.root.protocol("WM_DELETE_WINDOW", callback)

    def get_main_frame(self) -> ttk.Frame:
        """Get the main frame for adding components"""
        return self.main_frame

    def get_root(self) -> tk.Tk:
        """Get the root Tk window"""
        return self.root

    def run(self):
        """Start the main event loop"""
        self.root.mainloop()

    def quit(self):
        """Quit the application"""
        self.root.quit()

    def after(self, delay: int, callback: Callable):
        """Schedule a callback after a delay"""
        self.root.after(delay, callback)

    def update_title(self, title: str):
        """Update the window title"""
        self.root.title(title)

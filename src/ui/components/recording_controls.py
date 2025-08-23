import tkinter as tk
from tkinter import ttk
from typing import Callable, Optional


class RecordingControls:
    """
    Component for recording and live mode control buttons.
    Handles the UI state and callbacks for recording operations.
    """

    def __init__(self, parent_frame: ttk.Frame, row: int = 0):
        self.parent_frame = parent_frame
        self.row = row

        # Button state tracking
        self.recording = False
        self.processing = False
        self.live_mode_active = False

        # Callback functions
        self.on_toggle_recording: Optional[Callable] = None
        self.on_toggle_live_mode: Optional[Callable] = None
        self.on_settings: Optional[Callable] = None
        self.on_quit: Optional[Callable] = None

        self.setup_ui()

    def setup_ui(self):
        """Setup the recording controls UI"""
        # Regular recording button
        self.record_button = ttk.Button(
            self.parent_frame,
            text="🎤 Start Recording",
            command=self._on_record_clicked,
            width=18,
        )
        self.record_button.grid(row=self.row, column=0, pady=(0, 10), padx=(0, 5))

        # Live mode button
        self.live_button = ttk.Button(
            self.parent_frame,
            text="🔴 Live Mode",
            command=self._on_live_clicked,
            width=18,
        )
        self.live_button.grid(row=self.row, column=1, pady=(0, 10), padx=(5, 0))

        # Progress bar (initially hidden)
        self.progress = ttk.Progressbar(self.parent_frame, mode="indeterminate")
        self.progress.grid(
            row=self.row + 1, column=0, columnspan=2, sticky=(tk.W, tk.E), pady=(0, 10)
        )

        # Bottom control buttons
        bottom_row = self.row + 2
        settings_button = ttk.Button(
            self.parent_frame, text="Settings", command=self._on_settings_clicked
        )
        settings_button.grid(row=bottom_row, column=0, sticky=tk.W)

        quit_button = ttk.Button(
            self.parent_frame, text="Quit", command=self._on_quit_clicked
        )
        quit_button.grid(row=bottom_row, column=1, sticky=tk.E)

        # Store the next available row
        self.next_row = bottom_row + 1

    def set_callbacks(
        self,
        on_toggle_recording: Callable = None,
        on_toggle_live_mode: Callable = None,
        on_settings: Callable = None,
        on_quit: Callable = None,
    ):
        """Set callback functions for button events"""
        if on_toggle_recording:
            self.on_toggle_recording = on_toggle_recording
        if on_toggle_live_mode:
            self.on_toggle_live_mode = on_toggle_live_mode
        if on_settings:
            self.on_settings = on_settings
        if on_quit:
            self.on_quit = on_quit

    def _on_record_clicked(self):
        """Handle record button click"""
        if self.on_toggle_recording:
            self.on_toggle_recording()

    def _on_live_clicked(self):
        """Handle live mode button click"""
        if self.on_toggle_live_mode:
            self.on_toggle_live_mode()

    def _on_settings_clicked(self):
        """Handle settings button click"""
        if self.on_settings:
            self.on_settings()

    def _on_quit_clicked(self):
        """Handle quit button click"""
        if self.on_quit:
            self.on_quit()

    def set_recording_state(self, recording: bool, processing: bool = False):
        """Update the recording button state"""
        self.recording = recording
        self.processing = processing

        if processing:
            self.record_button.config(text="Processing...", state="disabled")
            self.progress.start()
        elif recording:
            self.record_button.config(text="⏹️ Stop Recording", style="Accent.TButton")
            self.progress.start()
        else:
            self.record_button.config(
                text="🎤 Start Recording", state="normal", style="TButton"
            )
            self.progress.stop()

    def set_live_mode_state(self, active: bool):
        """Update the live mode button state"""
        self.live_mode_active = active

        if active:
            self.live_button.config(text="⏹️ Stop Live", style="Accent.TButton")
            self.record_button.config(state="disabled")
            self.progress.start()
        else:
            self.live_button.config(text="🔴 Live Mode", style="TButton")
            self.record_button.config(state="normal")
            self.progress.stop()

    def enable_controls(self, enabled: bool = True):
        """Enable or disable all controls"""
        state = "normal" if enabled else "disabled"

        if not self.processing:  # Don't enable record button if processing
            self.record_button.config(state=state)
        if not self.recording:  # Don't enable live button if recording
            self.live_button.config(state=state)

    def get_next_row(self) -> int:
        """Get the next available row after this component"""
        return self.next_row

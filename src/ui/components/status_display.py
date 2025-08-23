import tkinter as tk
from tkinter import ttk
from typing import Optional
from ...config import Config


class StatusDisplay:
    """
    Component for displaying application status and language information.
    Handles status messages, language detection display, and user feedback.
    """

    def __init__(self, parent_frame: ttk.Frame, config: Config, row: int = 0):
        self.parent_frame = parent_frame
        self.config = config
        self.row = row

        # Reference to transcription engine for language info
        self.transcription_engine = None

        self.setup_ui()

    def setup_ui(self):
        """Setup the status display UI components"""
        # Main status label
        self.status_label = ttk.Label(
            self.parent_frame, text="Ready", font=("Arial", 12)
        )
        self.status_label.grid(row=self.row, column=0, columnspan=2, pady=(0, 10))

        # Language info label
        self.language_label = ttk.Label(
            self.parent_frame, text="", font=("Arial", 9), foreground="gray"
        )
        self.language_label.grid(row=self.row + 1, column=0, columnspan=2, pady=(0, 15))

        # Store the next available row
        self.next_row = self.row + 2

    def set_transcription_engine(self, transcription_engine):
        """Set the transcription engine reference for language info"""
        self.transcription_engine = transcription_engine
        self.update_language_display()

    def set_status(
        self, message: str, auto_clear: bool = False, clear_delay: int = 3000
    ):
        """Set the status message"""
        self.status_label.config(text=message)

        if auto_clear:
            self.parent_frame.after(
                clear_delay, lambda: self.status_label.config(text="Ready")
            )

    def set_recording_status(self, source: str = "button"):
        """Set status for recording mode"""
        if source == "double_shift":
            self.set_status("🔴 Recording (Double Shift)... Speak now!")
        elif source == "hotkey":
            self.set_status("🔴 Recording (Cmd+Shift+R)... Speak now!")
        else:
            self.set_status("🔴 Recording... Speak now!")

    def set_live_mode_status(self, source: str = "button"):
        """Set status for live mode"""
        if source == "double_control":
            self.set_status(
                "🔴 Live Mode (Double Ctrl): Speak and text will appear at cursor!"
            )
        else:
            self.set_status("🔴 Live Mode: Speak and text will appear at cursor!")

    def set_processing_status(self):
        """Set status for processing mode"""
        self.set_status("Processing audio...")

    def set_transcription_result(
        self, text: Optional[str], auto_insert_enabled: bool = False
    ):
        """Set status for transcription results"""
        if text:
            if auto_insert_enabled:
                self.set_status(
                    f"📋 Copied to clipboard • 🖱️ Click in input to insert: {text[:20]}...",
                    auto_clear=False,  # Don't auto-clear when waiting for click
                )
            else:
                self.set_status(
                    f"📋 Copied to clipboard: {text[:30]}...", auto_clear=True
                )
        else:
            self.set_status("❌ No speech detected", auto_clear=True)

    def set_live_transcription_status(self, text: str):
        """Set status for live transcription"""
        if len(text) > 30:
            self.set_status(f"🔴 Live: {text[:30]}...")
        else:
            self.set_status(f"🔴 Live: {text}")

    def set_auto_insert_complete(self, success: bool):
        """Set status when auto-insert completes"""
        if success:
            self.set_status("✅ Text inserted!", auto_clear=True, clear_delay=2000)
        else:
            self.set_status(
                "⏰ Auto-insert timed out", auto_clear=True, clear_delay=2000
            )

    def set_timeout_error(self, timeout_seconds: float):
        """Set status for transcription timeout"""
        self.set_status("⏰ Transcription timed out", auto_clear=True)

    def set_error(self, error_message: str):
        """Set status for errors"""
        self.set_status("❌ Error occurred", auto_clear=True)

    def update_language_display(self):
        """Update the language information display"""
        if not self.transcription_engine or not self.config.language_detection_enabled:
            self.language_label.config(text="")
            return

        lang_info = self.transcription_engine.get_language_info()
        configured_lang = lang_info["configured_language"]

        if configured_lang == "auto":
            detected = lang_info.get("detected_language", "none")
            confidence = lang_info.get("language_confidence", 0.0)
            if detected and detected != "none":
                self.language_label.config(
                    text=f"🌍 Auto-detect: {detected} ({confidence:.1%})"
                )
            else:
                self.language_label.config(text="🌍 Auto-detect: waiting for speech...")
        else:
            # Get display name for configured language
            available_languages = Config.get_available_languages()
            display_name = available_languages.get(configured_lang, configured_lang)
            self.language_label.config(text=f"🌍 Language: {display_name}")

    def clear_status(self):
        """Clear the status display"""
        self.set_status("Ready")

    def get_next_row(self) -> int:
        """Get the next available row after this component"""
        return self.next_row

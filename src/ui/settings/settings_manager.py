import tkinter as tk
from tkinter import ttk, messagebox
from typing import Callable
from ...config import Config
from .audio_settings_panel import AudioSettingsPanel
from .language_settings_panel import LanguageSettingsPanel
from .vad_settings_panel import VADSettingsPanel
from .indicator_settings_panel import IndicatorSettingsPanel


class SettingsManager:
    """
    Manages the settings dialog by coordinating multiple settings panels.
    Replaces the original monolithic SettingsDialog class.
    """

    def __init__(
        self, parent, config: Config, on_config_changed: Callable[[Config], None]
    ):
        self.parent = parent
        self.config = config
        self.on_config_changed = on_config_changed
        self.result = None

        self.window = tk.Toplevel(parent)
        self.window.title("Settings")
        self.window.geometry("450x1000")
        self.window.resizable(False, False)
        self.window.transient(parent)
        self.window.grab_set()

        # Center the window
        self.window.update_idletasks()
        x = (self.window.winfo_screenwidth() // 2) - (450 // 2)
        y = (self.window.winfo_screenheight() // 2) - (1000 // 2)
        self.window.geometry(f"450x1000+{x}+{y}")

        # Initialize panels
        self.audio_panel = None
        self.language_panel = None
        self.vad_panel = None
        self.indicator_panel = None

        self.setup_ui()

    def setup_ui(self):
        """Setup the main settings UI with all panels"""
        main_frame = ttk.Frame(self.window, padding="20")
        main_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))

        # Create and setup all settings panels
        current_row = 0

        # Audio Settings Panel
        self.audio_panel = AudioSettingsPanel(main_frame, self.config)
        current_row = self.audio_panel.row_end

        # Language Settings Panel
        self.language_panel = LanguageSettingsPanel(
            main_frame, self.config, current_row
        )
        current_row = self.language_panel.row_end

        # VAD Settings Panel
        self.vad_panel = VADSettingsPanel(main_frame, self.config, current_row)
        current_row = self.vad_panel.row_end

        # Indicator Settings Panel
        self.indicator_panel = IndicatorSettingsPanel(
            main_frame, self.config, current_row
        )
        current_row = self.indicator_panel.row_end

        # Buttons
        button_frame = ttk.Frame(main_frame)
        button_frame.grid(row=current_row, column=0, columnspan=2, pady=(20, 0))

        ttk.Button(button_frame, text="Cancel", command=self.cancel).pack(
            side=tk.RIGHT, padx=(10, 0)
        )
        ttk.Button(button_frame, text="Save", command=self.save).pack(side=tk.RIGHT)

        # Configure grid weights
        self.window.columnconfigure(0, weight=1)
        self.window.rowconfigure(0, weight=1)
        main_frame.columnconfigure(0, weight=1)
        main_frame.columnconfigure(1, weight=1)

    def save(self):
        """Save all settings from all panels"""
        try:
            # Validate all panels first
            if not self._validate_all_panels():
                return

            # Collect values from all panels
            audio_values = self.audio_panel.get_values()
            language_values = self.language_panel.get_values()
            vad_values = self.vad_panel.get_values()
            indicator_values = self.indicator_panel.get_values()

            # Update config with all values
            self._update_config(audio_values)
            self._update_config(language_values)
            self._update_config(vad_values)
            self._update_config(indicator_values)

            # Save config
            self.config.save()

            # Notify parent
            self.on_config_changed(self.config)

            self.result = "saved"
            self.window.destroy()

        except Exception as e:
            messagebox.showerror("Error", f"Failed to save settings: {e}")

    def _validate_all_panels(self) -> bool:
        """Validate all panels and return True if all are valid"""
        panels = [
            self.audio_panel,
            self.language_panel,
            self.vad_panel,
            self.indicator_panel,
        ]

        for panel in panels:
            if not panel.validate():
                return False
        return True

    def _update_config(self, values: dict):
        """Update config object with values from a panel"""
        for key, value in values.items():
            if hasattr(self.config, key):
                setattr(self.config, key, value)

    def cancel(self):
        """Cancel settings dialog"""
        self.result = "cancelled"
        self.window.destroy()

    def show(self) -> str:
        """Show the settings dialog and return the result"""
        self.window.wait_window()
        return self.result or "cancelled"

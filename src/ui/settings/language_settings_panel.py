import tkinter as tk
from tkinter import ttk, messagebox
from typing import Callable, Optional
from ...config import Config


class LanguageSettingsPanel:
    def __init__(self, parent_frame: ttk.Frame, config: Config, row_start: int = 0):
        self.parent_frame = parent_frame
        self.config = config
        self.row_start = row_start

        # Variables for form controls
        self.language_var = tk.StringVar()
        self.model_var = tk.StringVar()
        self.provider_var = tk.StringVar()
        self.quality_var = tk.StringVar()
        self.language_detection_var = tk.BooleanVar(
            value=getattr(config, "language_detection_enabled", True)
        )
        self.transcription_timeout_var = tk.StringVar(
            value=str(getattr(config, "transcription_timeout", 30.0))
        )

        self.setup_ui()

    def setup_ui(self):
        """Setup the language and transcription settings UI components"""
        row = self.row_start

        # Language & Transcription section
        ttk.Label(
            self.parent_frame,
            text="Language & Transcription:",
            font=("Arial", 12, "bold"),
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Label(self.parent_frame, text="Transcription Language:").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        # Language selection dropdown
        language_options = list(Config.get_available_languages().values())
        language_codes = list(Config.get_available_languages().keys())

        self.language_combo = ttk.Combobox(
            self.parent_frame,
            textvariable=self.language_var,
            state="readonly",
            width=25,
        )
        self.language_combo["values"] = language_options
        self.language_combo.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))

        # Set current language selection
        try:
            current_index = language_codes.index(self.config.transcription_language)
            self.language_combo.current(current_index)
        except (ValueError, AttributeError):
            self.language_combo.current(0)  # Default to auto
        row += 1

        ttk.Label(self.parent_frame, text="Model Size:").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        # Model size selection dropdown
        model_options = list(Config.get_available_models().values())
        model_codes = list(Config.get_available_models().keys())

        self.model_combo = ttk.Combobox(
            self.parent_frame, textvariable=self.model_var, state="readonly", width=25
        )
        self.model_combo["values"] = model_options
        self.model_combo.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))

        # Set current model selection
        try:
            current_model_index = model_codes.index(self.config.model_size)
            self.model_combo.current(current_model_index)
        except (ValueError, AttributeError):
            self.model_combo.current(0)  # Default to small
        row += 1

        ttk.Label(self.parent_frame, text="Whisper Provider:").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        # Whisper provider selection dropdown
        provider_options = list(Config.get_available_providers().values())
        provider_codes = list(Config.get_available_providers().keys())

        self.provider_combo = ttk.Combobox(
            self.parent_frame,
            textvariable=self.provider_var,
            state="readonly",
            width=25,
        )
        self.provider_combo["values"] = provider_options
        self.provider_combo.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))

        # Set current provider selection
        try:
            current_provider_index = provider_codes.index(self.config.whisper_provider)
            self.provider_combo.current(current_provider_index)
        except (ValueError, AttributeError):
            self.provider_combo.current(0)  # Default to faster-whisper
        row += 1

        ttk.Label(self.parent_frame, text="Live Quality Mode:").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        # Live quality mode selection dropdown
        quality_options = list(Config.get_live_quality_modes().values())
        quality_codes = list(Config.get_live_quality_modes().keys())

        self.quality_combo = ttk.Combobox(
            self.parent_frame, textvariable=self.quality_var, state="readonly", width=25
        )
        self.quality_combo["values"] = quality_options
        self.quality_combo.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))

        # Set current quality mode selection
        try:
            current_quality_index = quality_codes.index(self.config.live_quality_mode)
            self.quality_combo.current(current_quality_index)
        except (ValueError, AttributeError):
            self.quality_combo.current(1)  # Default to balanced
        row += 1

        # Language detection checkbox
        ttk.Checkbutton(
            self.parent_frame,
            text="Show language detection info",
            variable=self.language_detection_var,
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Label(self.parent_frame, text="Transcription timeout (seconds):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        timeout_spinbox = ttk.Spinbox(
            self.parent_frame,
            from_=5,
            to=300,
            increment=5,
            textvariable=self.transcription_timeout_var,
            width=10,
        )
        timeout_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 20))
        row += 1

        # Store the next available row for other panels
        self.row_end = row

    def get_values(self) -> dict:
        """Get the current values from the form controls"""
        language_codes = list(Config.get_available_languages().keys())
        model_codes = list(Config.get_available_models().keys())
        provider_codes = list(Config.get_available_providers().keys())
        quality_codes = list(Config.get_live_quality_modes().keys())

        values = {}

        # Language selection
        selected_language_index = self.language_combo.current()
        if selected_language_index >= 0:
            values["transcription_language"] = language_codes[selected_language_index]

        # Model selection
        selected_model_index = self.model_combo.current()
        if selected_model_index >= 0:
            values["model_size"] = model_codes[selected_model_index]

        # Provider selection
        selected_provider_index = self.provider_combo.current()
        if selected_provider_index >= 0:
            values["whisper_provider"] = provider_codes[selected_provider_index]

        # Quality mode selection
        selected_quality_index = self.quality_combo.current()
        if selected_quality_index >= 0:
            values["live_quality_mode"] = quality_codes[selected_quality_index]

        values["language_detection_enabled"] = self.language_detection_var.get()
        values["transcription_timeout"] = float(self.transcription_timeout_var.get())

        return values

    def validate(self) -> bool:
        """Validate the current form values"""
        try:
            timeout_value = float(self.transcription_timeout_var.get())
            if timeout_value < 5 or timeout_value > 300:
                messagebox.showerror(
                    "Validation Error",
                    "Transcription timeout must be between 5 and 300 seconds",
                )
                return False
            return True
        except ValueError:
            messagebox.showerror(
                "Validation Error", "Transcription timeout must be a valid number"
            )
            return False

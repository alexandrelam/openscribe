import tkinter as tk
from tkinter import ttk, messagebox
from typing import Callable, Optional
from ...config import Config


class VADSettingsPanel:
    def __init__(self, parent_frame: ttk.Frame, config: Config, row_start: int = 0):
        self.parent_frame = parent_frame
        self.config = config
        self.row_start = row_start

        # Variables for form controls
        self.vad_aggressiveness_var = tk.StringVar(value=str(config.vad_aggressiveness))
        self.vad_min_duration_var = tk.StringVar(
            value=str(config.vad_min_chunk_duration)
        )
        self.vad_max_duration_var = tk.StringVar(
            value=str(config.vad_max_chunk_duration)
        )
        self.vad_silence_timeout_var = tk.StringVar(
            value=str(config.vad_silence_timeout)
        )

        self.setup_ui()

    def setup_ui(self):
        """Setup the VAD settings UI components"""
        row = self.row_start

        # VAD settings section
        ttk.Label(
            self.parent_frame,
            text="Voice Activity Detection:",
            font=("Arial", 12, "bold"),
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Label(self.parent_frame, text="Aggressiveness (0-3):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        aggressiveness_spinbox = ttk.Spinbox(
            self.parent_frame,
            from_=0,
            to=3,
            textvariable=self.vad_aggressiveness_var,
            width=10,
        )
        aggressiveness_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))
        row += 1

        ttk.Label(self.parent_frame, text="Min chunk duration (s):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        min_duration_spinbox = ttk.Spinbox(
            self.parent_frame,
            from_=0.5,
            to=5.0,
            increment=0.1,
            textvariable=self.vad_min_duration_var,
            width=10,
        )
        min_duration_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))
        row += 1

        ttk.Label(self.parent_frame, text="Max chunk duration (s):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        max_duration_spinbox = ttk.Spinbox(
            self.parent_frame,
            from_=5.0,
            to=30.0,
            increment=1.0,
            textvariable=self.vad_max_duration_var,
            width=10,
        )
        max_duration_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))
        row += 1

        ttk.Label(self.parent_frame, text="Silence timeout (s):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        silence_timeout_spinbox = ttk.Spinbox(
            self.parent_frame,
            from_=0.1,
            to=2.0,
            increment=0.1,
            textvariable=self.vad_silence_timeout_var,
            width=10,
        )
        silence_timeout_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 20))
        row += 1

        # Store the next available row for other panels
        self.row_end = row

    def get_values(self) -> dict:
        """Get the current values from the form controls"""
        return {
            "vad_aggressiveness": int(self.vad_aggressiveness_var.get()),
            "vad_min_chunk_duration": float(self.vad_min_duration_var.get()),
            "vad_max_chunk_duration": float(self.vad_max_duration_var.get()),
            "vad_silence_timeout": float(self.vad_silence_timeout_var.get()),
        }

    def validate(self) -> bool:
        """Validate the current form values"""
        try:
            # Validate aggressiveness
            aggressiveness = int(self.vad_aggressiveness_var.get())
            if aggressiveness < 0 or aggressiveness > 3:
                messagebox.showerror(
                    "Validation Error", "VAD aggressiveness must be between 0 and 3"
                )
                return False

            # Validate durations
            min_duration = float(self.vad_min_duration_var.get())
            max_duration = float(self.vad_max_duration_var.get())
            silence_timeout = float(self.vad_silence_timeout_var.get())

            if min_duration < 0.5 or min_duration > 5.0:
                messagebox.showerror(
                    "Validation Error",
                    "Min chunk duration must be between 0.5 and 5.0 seconds",
                )
                return False

            if max_duration < 5.0 or max_duration > 30.0:
                messagebox.showerror(
                    "Validation Error",
                    "Max chunk duration must be between 5.0 and 30.0 seconds",
                )
                return False

            if silence_timeout < 0.1 or silence_timeout > 2.0:
                messagebox.showerror(
                    "Validation Error",
                    "Silence timeout must be between 0.1 and 2.0 seconds",
                )
                return False

            if min_duration >= max_duration:
                messagebox.showerror(
                    "Validation Error",
                    "Min chunk duration must be less than max chunk duration",
                )
                return False

            return True
        except ValueError:
            messagebox.showerror(
                "Validation Error", "VAD settings must be valid numbers"
            )
            return False

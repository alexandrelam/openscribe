import tkinter as tk
from tkinter import ttk, messagebox
from typing import Callable, Optional
from ...config import Config


class IndicatorSettingsPanel:
    def __init__(self, parent_frame: ttk.Frame, config: Config, row_start: int = 0):
        self.parent_frame = parent_frame
        self.config = config
        self.row_start = row_start

        # Variables for form controls
        self.show_indicator_var = tk.BooleanVar(value=config.show_recording_indicator)
        self.indicator_position_x_var = tk.StringVar(
            value=str(config.indicator_position_x)
        )
        self.indicator_position_y_var = tk.StringVar(
            value=str(config.indicator_position_y)
        )
        self.indicator_size_var = tk.StringVar(value=str(config.indicator_size))
        self.indicator_opacity_var = tk.StringVar(value=str(config.indicator_opacity))

        self.setup_ui()

    def setup_ui(self):
        """Setup the recording indicator settings UI components"""
        row = self.row_start

        # Recording Indicator section
        ttk.Label(
            self.parent_frame, text="Recording Indicator:", font=("Arial", 12, "bold")
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Checkbutton(
            self.parent_frame,
            text="Show fixed position recording indicator",
            variable=self.show_indicator_var,
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Label(self.parent_frame, text="Screen position (X, Y):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        position_frame = ttk.Frame(self.parent_frame)
        position_frame.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))

        position_x_spinbox = ttk.Spinbox(
            position_frame,
            from_=0,
            to=1000,
            textvariable=self.indicator_position_x_var,
            width=8,
        )
        position_x_spinbox.pack(side=tk.LEFT, padx=(0, 5))

        position_y_spinbox = ttk.Spinbox(
            position_frame,
            from_=0,
            to=1000,
            textvariable=self.indicator_position_y_var,
            width=8,
        )
        position_y_spinbox.pack(side=tk.LEFT)
        row += 1

        ttk.Label(self.parent_frame, text="Size (pixels):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        size_spinbox = ttk.Spinbox(
            self.parent_frame,
            from_=10,
            to=50,
            textvariable=self.indicator_size_var,
            width=10,
        )
        size_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 5))
        row += 1

        ttk.Label(self.parent_frame, text="Opacity (0.1-1.0):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        opacity_spinbox = ttk.Spinbox(
            self.parent_frame,
            from_=0.1,
            to=1.0,
            increment=0.1,
            textvariable=self.indicator_opacity_var,
            width=10,
        )
        opacity_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 20))
        row += 1

        # Store the next available row for other panels
        self.row_end = row

    def get_values(self) -> dict:
        """Get the current values from the form controls"""
        return {
            "show_recording_indicator": self.show_indicator_var.get(),
            "indicator_position_x": int(self.indicator_position_x_var.get()),
            "indicator_position_y": int(self.indicator_position_y_var.get()),
            "indicator_size": int(self.indicator_size_var.get()),
            "indicator_opacity": float(self.indicator_opacity_var.get()),
        }

    def validate(self) -> bool:
        """Validate the current form values"""
        try:
            # Validate position values
            position_x = int(self.indicator_position_x_var.get())
            position_y = int(self.indicator_position_y_var.get())
            size = int(self.indicator_size_var.get())
            opacity = float(self.indicator_opacity_var.get())

            if position_x < 0 or position_x > 1000:
                messagebox.showerror(
                    "Validation Error",
                    "Indicator X position must be between 0 and 1000",
                )
                return False

            if position_y < 0 or position_y > 1000:
                messagebox.showerror(
                    "Validation Error",
                    "Indicator Y position must be between 0 and 1000",
                )
                return False

            if size < 10 or size > 50:
                messagebox.showerror(
                    "Validation Error",
                    "Indicator size must be between 10 and 50 pixels",
                )
                return False

            if opacity < 0.1 or opacity > 1.0:
                messagebox.showerror(
                    "Validation Error", "Indicator opacity must be between 0.1 and 1.0"
                )
                return False

            return True
        except ValueError:
            messagebox.showerror(
                "Validation Error", "Indicator settings must be valid numbers"
            )
            return False

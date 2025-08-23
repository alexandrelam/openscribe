import tkinter as tk
from tkinter import ttk, messagebox
from typing import Callable, Optional
from ...audio_recorder import AudioRecorder
from ...config import Config


class AudioSettingsPanel:
    def __init__(self, parent_frame: ttk.Frame, config: Config):
        self.parent_frame = parent_frame
        self.config = config
        self.row_start = 0

        # Variables for form controls
        self.device_var = tk.StringVar()
        self.auto_insert_var = tk.BooleanVar(value=config.enable_auto_insert)
        self.timeout_var = tk.StringVar(value=str(config.auto_insert_timeout))

        # Store device indices for mapping
        self.device_indices = []

        self.setup_ui()

    def setup_ui(self):
        """Setup the audio settings UI components"""
        row = self.row_start

        # Microphone section
        ttk.Label(
            self.parent_frame, text="Microphone Device:", font=("Arial", 12, "bold")
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        self.device_combo = ttk.Combobox(
            self.parent_frame, textvariable=self.device_var, state="readonly", width=50
        )
        self.device_combo.grid(
            row=row, column=0, columnspan=2, sticky=(tk.W, tk.E), pady=(0, 20)
        )
        row += 1

        # Load available devices
        self.load_devices()

        # Auto-insert section
        ttk.Label(
            self.parent_frame, text="Auto-Insert Settings:", font=("Arial", 12, "bold")
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Checkbutton(
            self.parent_frame,
            text="Enable auto-insert on click",
            variable=self.auto_insert_var,
        ).grid(row=row, column=0, columnspan=2, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Label(self.parent_frame, text="Auto-insert timeout (seconds):").grid(
            row=row, column=0, sticky=tk.W, pady=(0, 5)
        )

        timeout_spinbox = ttk.Spinbox(
            self.parent_frame, from_=5, to=60, textvariable=self.timeout_var, width=10
        )
        timeout_spinbox.grid(row=row, column=1, sticky=tk.W, pady=(0, 20))
        row += 1

        # Store the next available row for other panels
        self.row_end = row

    def load_devices(self):
        """Load available audio devices"""
        try:
            devices = AudioRecorder.get_input_devices()
            device_options = ["Default (System Default)"]
            device_indices = [None]

            for device in devices:
                device_options.append(f"{device['name']} (Device {device['index']})")
                device_indices.append(device["index"])

            self.device_combo["values"] = device_options
            self.device_indices = device_indices

            # Set current selection
            if self.config.microphone_device is None:
                self.device_combo.current(0)
            else:
                try:
                    index = device_indices.index(self.config.microphone_device)
                    self.device_combo.current(index)
                except ValueError:
                    self.device_combo.current(0)

        except Exception as e:
            messagebox.showerror("Error", f"Failed to load audio devices: {e}")
            self.device_combo["values"] = ["Default (System Default)"]
            self.device_indices = [None]
            self.device_combo.current(0)

    def get_values(self) -> dict:
        """Get the current values from the form controls"""
        selected_index = self.device_combo.current()
        selected_device = (
            self.device_indices[selected_index] if selected_index >= 0 else None
        )

        return {
            "microphone_device": selected_device,
            "enable_auto_insert": self.auto_insert_var.get(),
            "auto_insert_timeout": int(self.timeout_var.get()),
        }

    def validate(self) -> bool:
        """Validate the current form values"""
        try:
            timeout_value = int(self.timeout_var.get())
            if timeout_value < 5 or timeout_value > 60:
                messagebox.showerror(
                    "Validation Error",
                    "Auto-insert timeout must be between 5 and 60 seconds",
                )
                return False
            return True
        except ValueError:
            messagebox.showerror(
                "Validation Error", "Auto-insert timeout must be a valid number"
            )
            return False

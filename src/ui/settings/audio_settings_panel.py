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
        self.auto_insert_var = tk.BooleanVar(value=config.enable_auto_insert)
        self.timeout_var = tk.StringVar(value=str(config.auto_insert_timeout))
        self.add_device_var = tk.StringVar()

        # Store device information for mapping
        self.available_devices = []
        self.device_indices = []

        self.setup_ui()

    def setup_ui(self):
        """Setup the audio settings UI components"""
        row = self.row_start

        # Microphone preferences section
        ttk.Label(
            self.parent_frame,
            text="Microphone Preferences:",
            font=("Arial", 12, "bold"),
        ).grid(row=row, column=0, columnspan=3, sticky=tk.W, pady=(0, 10))
        row += 1

        # Currently active device display
        self.active_device_label = ttk.Label(
            self.parent_frame, text="", foreground="green"
        )
        self.active_device_label.grid(
            row=row, column=0, columnspan=3, sticky=tk.W, pady=(0, 10)
        )
        row += 1

        # Preferences list frame
        list_frame = ttk.Frame(self.parent_frame)
        list_frame.grid(
            row=row, column=0, columnspan=3, sticky=(tk.W, tk.E), pady=(0, 10)
        )

        # Preferences listbox with scrollbar
        listbox_frame = ttk.Frame(list_frame)
        listbox_frame.grid(row=0, column=0, sticky=(tk.W, tk.E))

        self.preferences_listbox = tk.Listbox(listbox_frame, height=6, width=50)
        scrollbar = ttk.Scrollbar(
            listbox_frame, orient="vertical", command=self.preferences_listbox.yview
        )
        self.preferences_listbox.configure(yscrollcommand=scrollbar.set)

        self.preferences_listbox.grid(row=0, column=0, sticky=(tk.W, tk.E))
        scrollbar.grid(row=0, column=1, sticky=(tk.N, tk.S))

        # Control buttons frame
        buttons_frame = ttk.Frame(list_frame)
        buttons_frame.grid(row=0, column=1, sticky=(tk.N), padx=(10, 0))

        ttk.Button(buttons_frame, text="↑ Move Up", command=self.move_up).grid(
            row=0, column=0, pady=(0, 5), sticky=tk.W
        )
        ttk.Button(buttons_frame, text="↓ Move Down", command=self.move_down).grid(
            row=1, column=0, pady=(0, 5), sticky=tk.W
        )
        ttk.Button(buttons_frame, text="Remove", command=self.remove_preference).grid(
            row=2, column=0, pady=(0, 5), sticky=tk.W
        )

        row += 1

        # Add device section
        add_frame = ttk.Frame(self.parent_frame)
        add_frame.grid(
            row=row, column=0, columnspan=3, sticky=(tk.W, tk.E), pady=(0, 20)
        )

        ttk.Label(add_frame, text="Add Device:").grid(
            row=0, column=0, sticky=tk.W, padx=(0, 5)
        )

        self.add_device_combo = ttk.Combobox(
            add_frame, textvariable=self.add_device_var, state="readonly", width=35
        )
        self.add_device_combo.grid(row=0, column=1, sticky=tk.W, padx=(0, 5))

        ttk.Button(add_frame, text="Add", command=self.add_preference).grid(
            row=0, column=2, sticky=tk.W
        )
        row += 1

        # Auto-insert section
        ttk.Label(
            self.parent_frame, text="Auto-Insert Settings:", font=("Arial", 12, "bold")
        ).grid(row=row, column=0, columnspan=3, sticky=tk.W, pady=(0, 10))
        row += 1

        ttk.Checkbutton(
            self.parent_frame,
            text="Enable auto-insert on click",
            variable=self.auto_insert_var,
        ).grid(row=row, column=0, columnspan=3, sticky=tk.W, pady=(0, 10))
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

        # Load data and update UI
        self.load_devices()
        self.update_preferences_display()
        self.update_active_device_display()

    def load_devices(self):
        """Load available audio devices"""
        try:
            self.available_devices = AudioRecorder.get_input_devices()
            device_options = []
            device_indices = []

            for device in self.available_devices:
                device_options.append(f"{device['name']} (Device {device['index']})")
                device_indices.append(device["index"])

            self.add_device_combo["values"] = device_options
            self.device_indices = device_indices

        except Exception as e:
            messagebox.showerror("Error", f"Failed to load audio devices: {e}")
            self.add_device_combo["values"] = []
            self.device_indices = []

    def update_preferences_display(self):
        """Update the preferences listbox with current preferences"""
        self.preferences_listbox.delete(0, tk.END)

        for device_id in self.config.microphone_preferences:
            device_name = self.get_device_name_by_id(device_id)
            if device_name:
                available_indicator = (
                    "✓" if AudioRecorder.is_device_available(device_id) else "✗"
                )
                self.preferences_listbox.insert(
                    tk.END, f"{available_indicator} {device_name} (Device {device_id})"
                )

    def update_active_device_display(self):
        """Update the label showing the currently active device"""
        active_device = self.config.get_preferred_device()
        if active_device is None:
            self.active_device_label.config(
                text="🎤 Currently active: Default (System Default)"
            )
        else:
            device_name = self.get_device_name_by_id(active_device)
            self.active_device_label.config(
                text=f"🎤 Currently active: {device_name} (Device {active_device})"
            )

    def get_device_name_by_id(self, device_id: int) -> Optional[str]:
        """Get device name by device ID"""
        for device in self.available_devices:
            if device["index"] == device_id:
                return device["name"]
        return f"Device {device_id}"

    def add_preference(self):
        """Add a device to the preferences list"""
        selected_index = self.add_device_combo.current()
        if selected_index < 0:
            messagebox.showwarning("Warning", "Please select a device to add")
            return

        device_id = self.device_indices[selected_index]

        # Check if device is already in preferences
        if device_id in self.config.microphone_preferences:
            messagebox.showwarning("Warning", "Device is already in preferences")
            return

        # Add to preferences
        self.config.microphone_preferences.append(device_id)
        self.update_preferences_display()
        self.update_active_device_display()

    def remove_preference(self):
        """Remove selected device from preferences list"""
        selected_index = self.preferences_listbox.curselection()
        if not selected_index:
            messagebox.showwarning(
                "Warning", "Please select a device preference to remove"
            )
            return

        index = selected_index[0]
        if index < len(self.config.microphone_preferences):
            removed_device = self.config.microphone_preferences.pop(index)
            self.update_preferences_display()
            self.update_active_device_display()

    def move_up(self):
        """Move selected preference up in the list (higher priority)"""
        selected_index = self.preferences_listbox.curselection()
        if not selected_index:
            messagebox.showwarning(
                "Warning", "Please select a device preference to move"
            )
            return

        index = selected_index[0]
        if index > 0 and index < len(self.config.microphone_preferences):
            # Swap with previous item
            prefs = self.config.microphone_preferences
            prefs[index], prefs[index - 1] = prefs[index - 1], prefs[index]
            self.update_preferences_display()
            self.update_active_device_display()
            # Maintain selection
            self.preferences_listbox.selection_set(index - 1)

    def move_down(self):
        """Move selected preference down in the list (lower priority)"""
        selected_index = self.preferences_listbox.curselection()
        if not selected_index:
            messagebox.showwarning(
                "Warning", "Please select a device preference to move"
            )
            return

        index = selected_index[0]
        if index < len(self.config.microphone_preferences) - 1:
            # Swap with next item
            prefs = self.config.microphone_preferences
            prefs[index], prefs[index + 1] = prefs[index + 1], prefs[index]
            self.update_preferences_display()
            self.update_active_device_display()
            # Maintain selection
            self.preferences_listbox.selection_set(index + 1)

    def get_values(self) -> dict:
        """Get the current values from the form controls"""
        return {
            "microphone_preferences": self.config.microphone_preferences.copy(),
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

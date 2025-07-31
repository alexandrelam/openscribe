import tkinter as tk
from tkinter import ttk, messagebox
import threading
from typing import Callable, Optional
from .audio_recorder import AudioRecorder
from .transcription import TranscriptionEngine
from .text_inserter import TextInserter
from .config import Config


class SettingsDialog:
    def __init__(self, parent, config: Config, on_config_changed: Callable[[Config], None]):
        self.parent = parent
        self.config = config
        self.on_config_changed = on_config_changed
        self.result = None
        
        self.window = tk.Toplevel(parent)
        self.window.title("Settings")
        self.window.geometry("450x550")
        self.window.resizable(False, False)
        self.window.transient(parent)
        self.window.grab_set()
        
        # Center the window
        self.window.update_idletasks()
        x = (self.window.winfo_screenwidth() // 2) - (450 // 2)
        y = (self.window.winfo_screenheight() // 2) - (550 // 2)
        self.window.geometry(f"450x550+{x}+{y}")
        
        self.setup_ui()
        
    def setup_ui(self):
        main_frame = ttk.Frame(self.window, padding="20")
        main_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))
        
        # Microphone selection
        ttk.Label(main_frame, text="Microphone Device:", font=("Arial", 12, "bold")).grid(
            row=0, column=0, columnspan=2, sticky=tk.W, pady=(0, 10)
        )
        
        self.device_var = tk.StringVar()
        self.device_combo = ttk.Combobox(main_frame, textvariable=self.device_var, state="readonly", width=50)
        self.device_combo.grid(row=1, column=0, columnspan=2, sticky=(tk.W, tk.E), pady=(0, 20))
        
        # Load available devices
        self.load_devices()
        
        # Auto-insert settings
        ttk.Label(main_frame, text="Auto-Insert Settings:", font=("Arial", 12, "bold")).grid(
            row=2, column=0, columnspan=2, sticky=tk.W, pady=(0, 10)
        )
        
        self.auto_insert_var = tk.BooleanVar(value=self.config.enable_auto_insert)
        ttk.Checkbutton(main_frame, text="Enable auto-insert on click", 
                       variable=self.auto_insert_var).grid(
            row=3, column=0, columnspan=2, sticky=tk.W, pady=(0, 10)
        )
        
        ttk.Label(main_frame, text="Auto-insert timeout (seconds):").grid(
            row=4, column=0, sticky=tk.W, pady=(0, 5)
        )
        
        self.timeout_var = tk.StringVar(value=str(self.config.auto_insert_timeout))
        timeout_spinbox = ttk.Spinbox(main_frame, from_=5, to=60, textvariable=self.timeout_var, width=10)
        timeout_spinbox.grid(row=4, column=1, sticky=tk.W, pady=(0, 20))
        
        # Language settings
        ttk.Label(main_frame, text="Language & Transcription:", font=("Arial", 12, "bold")).grid(
            row=5, column=0, columnspan=2, sticky=tk.W, pady=(0, 10)
        )
        
        ttk.Label(main_frame, text="Transcription Language:").grid(
            row=6, column=0, sticky=tk.W, pady=(0, 5)
        )
        
        # Language selection dropdown
        self.language_var = tk.StringVar()
        language_options = list(Config.get_available_languages().values())
        language_codes = list(Config.get_available_languages().keys())
        
        self.language_combo = ttk.Combobox(main_frame, textvariable=self.language_var, state="readonly", width=25)
        self.language_combo['values'] = language_options
        self.language_combo.grid(row=6, column=1, sticky=tk.W, pady=(0, 5))
        
        # Set current language selection
        try:
            current_index = language_codes.index(self.config.transcription_language)
            self.language_combo.current(current_index)
        except (ValueError, AttributeError):
            self.language_combo.current(0)  # Default to auto
        
        ttk.Label(main_frame, text="Model Size:").grid(
            row=7, column=0, sticky=tk.W, pady=(0, 5)
        )
        
        # Model size selection dropdown
        self.model_var = tk.StringVar()
        model_options = list(Config.get_available_models().values())
        model_codes = list(Config.get_available_models().keys())
        
        self.model_combo = ttk.Combobox(main_frame, textvariable=self.model_var, state="readonly", width=25)
        self.model_combo['values'] = model_options
        self.model_combo.grid(row=7, column=1, sticky=tk.W, pady=(0, 5))
        
        # Set current model selection
        try:
            current_model_index = model_codes.index(self.config.model_size)
            self.model_combo.current(current_model_index)
        except (ValueError, AttributeError):
            self.model_combo.current(0)  # Default to small
        
        # Language detection checkbox
        self.language_detection_var = tk.BooleanVar(value=getattr(self.config, 'language_detection_enabled', True))
        ttk.Checkbutton(main_frame, text="Show language detection info", 
                       variable=self.language_detection_var).grid(
            row=8, column=0, columnspan=2, sticky=tk.W, pady=(0, 20)
        )
        
        # VAD settings
        ttk.Label(main_frame, text="Voice Activity Detection:", font=("Arial", 12, "bold")).grid(
            row=9, column=0, columnspan=2, sticky=tk.W, pady=(0, 10)
        )
        
        ttk.Label(main_frame, text="Aggressiveness (0-3):").grid(
            row=10, column=0, sticky=tk.W, pady=(0, 5)
        )
        
        self.vad_aggressiveness_var = tk.StringVar(value=str(self.config.vad_aggressiveness))
        aggressiveness_spinbox = ttk.Spinbox(main_frame, from_=0, to=3, textvariable=self.vad_aggressiveness_var, width=10)
        aggressiveness_spinbox.grid(row=10, column=1, sticky=tk.W, pady=(0, 5))
        
        ttk.Label(main_frame, text="Min chunk duration (s):").grid(
            row=11, column=0, sticky=tk.W, pady=(0, 5)
        )
        
        self.vad_min_duration_var = tk.StringVar(value=str(self.config.vad_min_chunk_duration))
        min_duration_spinbox = ttk.Spinbox(main_frame, from_=0.5, to=5.0, increment=0.1, textvariable=self.vad_min_duration_var, width=10)
        min_duration_spinbox.grid(row=11, column=1, sticky=tk.W, pady=(0, 5))
        
        ttk.Label(main_frame, text="Max chunk duration (s):").grid(
            row=12, column=0, sticky=tk.W, pady=(0, 5)
        )
        
        self.vad_max_duration_var = tk.StringVar(value=str(self.config.vad_max_chunk_duration))
        max_duration_spinbox = ttk.Spinbox(main_frame, from_=5.0, to=30.0, increment=1.0, textvariable=self.vad_max_duration_var, width=10)
        max_duration_spinbox.grid(row=12, column=1, sticky=tk.W, pady=(0, 5))
        
        ttk.Label(main_frame, text="Silence timeout (s):").grid(
            row=13, column=0, sticky=tk.W, pady=(0, 5)
        )
        
        self.vad_silence_timeout_var = tk.StringVar(value=str(self.config.vad_silence_timeout))
        silence_timeout_spinbox = ttk.Spinbox(main_frame, from_=0.1, to=2.0, increment=0.1, textvariable=self.vad_silence_timeout_var, width=10)
        silence_timeout_spinbox.grid(row=13, column=1, sticky=tk.W, pady=(0, 20))
        
        # Buttons
        button_frame = ttk.Frame(main_frame)
        button_frame.grid(row=14, column=0, columnspan=2, pady=(20, 0))
        
        ttk.Button(button_frame, text="Cancel", command=self.cancel).pack(side=tk.RIGHT, padx=(10, 0))
        ttk.Button(button_frame, text="Save", command=self.save).pack(side=tk.RIGHT)
        
        self.window.columnconfigure(0, weight=1)
        self.window.rowconfigure(0, weight=1)
        main_frame.columnconfigure(0, weight=1)
        main_frame.columnconfigure(1, weight=1)
        
    def load_devices(self):
        try:
            devices = AudioRecorder.get_input_devices()
            device_options = ["Default (System Default)"]
            device_indices = [None]
            
            for device in devices:
                device_options.append(f"{device['name']} (Device {device['index']})")
                device_indices.append(device['index'])
            
            self.device_combo['values'] = device_options
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
            self.device_combo['values'] = ["Default (System Default)"]
            self.device_indices = [None]
            self.device_combo.current(0)
    
    def save(self):
        try:
            # Get selected device
            selected_index = self.device_combo.current()
            selected_device = self.device_indices[selected_index] if selected_index >= 0 else None
            
            # Update config
            self.config.microphone_device = selected_device
            self.config.enable_auto_insert = self.auto_insert_var.get()
            self.config.auto_insert_timeout = int(self.timeout_var.get())
            
            # Update language config
            language_codes = list(Config.get_available_languages().keys())
            model_codes = list(Config.get_available_models().keys())
            
            selected_language_index = self.language_combo.current()
            selected_model_index = self.model_combo.current()
            
            if selected_language_index >= 0:
                self.config.transcription_language = language_codes[selected_language_index]
            if selected_model_index >= 0:
                self.config.model_size = model_codes[selected_model_index]
            
            self.config.language_detection_enabled = self.language_detection_var.get()
            
            # Update VAD config
            self.config.vad_aggressiveness = int(self.vad_aggressiveness_var.get())
            self.config.vad_min_chunk_duration = float(self.vad_min_duration_var.get())
            self.config.vad_max_chunk_duration = float(self.vad_max_duration_var.get())
            self.config.vad_silence_timeout = float(self.vad_silence_timeout_var.get())
            
            # Save config
            self.config.save()
            
            # Notify parent
            self.on_config_changed(self.config)
            
            self.result = "saved"
            self.window.destroy()
            
        except Exception as e:
            messagebox.showerror("Error", f"Failed to save settings: {e}")
    
    def cancel(self):
        self.result = "cancelled"
        self.window.destroy()

class SpeechToTextGUI:
    def __init__(self):
        self.root = tk.Tk()
        self.root.title("Speech to Text")
        self.root.geometry("300x200")
        self.root.resizable(False, False)
        
        self.config = Config.load()
        self.audio_recorder = AudioRecorder(device_id=self.config.microphone_device)
        self.transcription_engine = TranscriptionEngine(
            language_config=self.config.transcription_language,
            model_size=self.config.model_size
        )
        self.text_inserter = TextInserter()
        
        # Apply configurations on startup
        self._apply_vad_config()
        self._apply_paste_config()
        self._apply_language_config()
        
        self.recording = False
        self.processing = False
        self.live_mode_active = False
        
        self.setup_ui()
        self.setup_hotkeys()
    
    def setup_ui(self):
        main_frame = ttk.Frame(self.root, padding="20")
        main_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))
        
        self.status_label = ttk.Label(main_frame, text="Ready", font=("Arial", 12))
        self.status_label.grid(row=0, column=0, columnspan=2, pady=(0, 10))
        
        # Language info label
        self.language_label = ttk.Label(main_frame, text="", font=("Arial", 9), foreground="gray")
        self.language_label.grid(row=1, column=0, columnspan=2, pady=(0, 15))
        
        # Update language display
        self._update_language_display()
        
        # Regular recording button
        self.record_button = ttk.Button(
            main_frame, 
            text="🎤 Start Recording", 
            command=self.toggle_recording,
            width=18
        )
        self.record_button.grid(row=2, column=0, pady=(0, 10), padx=(0, 5))
        
        # Live mode button
        self.live_button = ttk.Button(
            main_frame,
            text="🔴 Live Mode",
            command=self.toggle_live_mode,
            width=18
        )
        self.live_button.grid(row=2, column=1, pady=(0, 10), padx=(5, 0))
        
        self.progress = ttk.Progressbar(main_frame, mode='indeterminate')
        self.progress.grid(row=3, column=0, columnspan=2, sticky=(tk.W, tk.E), pady=(0, 10))
        
        settings_button = ttk.Button(main_frame, text="Settings", command=self.show_settings)
        settings_button.grid(row=4, column=0, sticky=tk.W)
        
        quit_button = ttk.Button(main_frame, text="Quit", command=self.quit_app)
        quit_button.grid(row=4, column=1, sticky=tk.E)
        
        self.root.columnconfigure(0, weight=1)
        self.root.rowconfigure(0, weight=1)
        main_frame.columnconfigure(0, weight=1)
        main_frame.columnconfigure(1, weight=1)
    
    def setup_hotkeys(self):
        try:
            from pynput import keyboard
            
            def on_hotkey():
                self.root.after(0, self.toggle_recording)
            
            hotkey = keyboard.HotKey(
                keyboard.HotKey.parse('<cmd>+<shift>+r'),
                on_hotkey
            )
            
            def for_canonical(f):
                return lambda k: f(listener.canonical(k))
            
            listener = keyboard.Listener(
                on_press=for_canonical(hotkey.press),
                on_release=for_canonical(hotkey.release)
            )
            listener.start()
            
        except Exception as e:
            print(f"Failed to setup hotkeys: {e}")
    
    def toggle_recording(self):
        if self.processing or self.live_mode_active:
            return
            
        if not self.recording:
            self.start_recording()
        else:
            self.stop_recording()
    
    def toggle_live_mode(self):
        if self.processing or self.recording:
            return
        
        if not self.live_mode_active:
            self.start_live_mode()
        else:
            self.stop_live_mode()
    
    def start_recording(self):
        if self.audio_recorder.start_recording():
            self.recording = True
            self.record_button.config(text="⏹️ Stop Recording", style="Accent.TButton")
            self.status_label.config(text="🔴 Recording... Speak now!")
            self.progress.start()
        else:
            messagebox.showerror("Error", "Failed to start recording. Check microphone permissions.")
    
    def stop_recording(self):
        if not self.recording:
            return
            
        self.recording = False
        self.progress.stop()
        self.record_button.config(text="Processing...", state="disabled")
        self.status_label.config(text="Processing audio...")
        
        def process_audio():
            try:
                audio_data = self.audio_recorder.stop_recording()
                if audio_data is not None and len(audio_data) > 0:
                    text = self.transcription_engine.transcribe_audio(audio_data)
                    self.root.after(0, lambda: self.handle_transcription(text))
                else:
                    self.root.after(0, lambda: self.handle_transcription(None))
            except Exception as e:
                self.root.after(0, lambda: self.handle_error(str(e)))
        
        self.processing = True
        threading.Thread(target=process_audio, daemon=True).start()
    
    def handle_transcription(self, text: Optional[str]):
        self.processing = False
        self.record_button.config(text="🎤 Start Recording", state="normal", style="TButton")
        
        if text:
            # Always copy to clipboard first
            self.text_inserter.copy_to_clipboard(text)
            
            # Start auto-insert mode if enabled
            if self.config.enable_auto_insert:
                success = self.text_inserter.start_auto_insert_mode(
                    text, 
                    self.config.auto_insert_timeout,
                    self._on_auto_insert_complete
                )
                if success:
                    self.status_label.config(text=f"📋 Copied to clipboard • 🖱️ Click in input to insert: {text[:20]}...")
                else:
                    self.status_label.config(text=f"📋 Copied to clipboard: {text[:30]}...")
            else:
                self.status_label.config(text=f"📋 Copied to clipboard: {text[:30]}...")
        else:
            self.status_label.config(text="❌ No speech detected")
            messagebox.showwarning("No Speech", "No speech was detected or transcription failed.")
        
        # Only reset to "Ready" if auto-insert is not active
        if not self.text_inserter.is_auto_insert_active():
            self.root.after(3000, lambda: self.status_label.config(text="Ready"))
    
    def _on_auto_insert_complete(self, success: bool):
        """Callback when auto-insert mode completes"""
        if success:
            self.status_label.config(text="✅ Text inserted!")
        else:
            self.status_label.config(text="⏰ Auto-insert timed out")
        
        # Reset to "Ready" after a delay
        self.root.after(2000, lambda: self.status_label.config(text="Ready"))
    
    def handle_error(self, error_msg: str):
        self.processing = False
        self.record_button.config(text="🎤 Start Recording", state="normal", style="TButton")
        self.status_label.config(text="❌ Error occurred")
        messagebox.showerror("Error", f"An error occurred: {error_msg}")
        self.root.after(3000, lambda: self.status_label.config(text="Ready"))
    
    def show_settings(self):
        SettingsDialog(self.root, self.config, self.on_config_changed)
    
    def on_config_changed(self, new_config: Config):
        """Handle config changes from settings dialog"""
        self.config = new_config
        
        # Recreate AudioRecorder with new device if it changed
        old_device = getattr(self.audio_recorder, 'device_id', None)
        if old_device != self.config.microphone_device:
            # Only recreate if not currently recording
            if not self.recording:
                self.audio_recorder = AudioRecorder(device_id=self.config.microphone_device)
                self.transcription_engine = TranscriptionEngine(
                    language_config=self.config.transcription_language,
                    model_size=self.config.model_size
                )
                self._apply_vad_config()
                self._apply_paste_config()
                self._apply_language_config()
                messagebox.showinfo("Settings", "All settings updated successfully!")
            else:
                messagebox.showwarning("Settings", "Settings saved. Microphone device will be updated after current recording stops.")
        else:
            # Apply configuration to existing components
            self._apply_vad_config()
            self._apply_paste_config()
            self._apply_language_config()
            messagebox.showinfo("Settings", "Settings saved successfully!")
    
    def _apply_vad_config(self):
        """Apply VAD configuration to the audio recorder"""
        if self.audio_recorder:
            self.audio_recorder.configure_vad(
                aggressiveness=self.config.vad_aggressiveness,
                min_chunk_duration=self.config.vad_min_chunk_duration,
                max_chunk_duration=self.config.vad_max_chunk_duration,
                silence_timeout=self.config.vad_silence_timeout
            )
    
    def _apply_paste_config(self):
        """Apply paste configuration to the text inserter"""
        if self.text_inserter:
            self.text_inserter.configure_pasting(
                paste_method=self.config.paste_method,
                paste_delay=self.config.paste_delay,
                live_paste_interval=self.config.live_paste_interval,
                restore_clipboard=self.config.restore_clipboard
            )
    
    def _apply_language_config(self):
        """Apply language configuration to the transcription engine"""
        if self.transcription_engine:
            self.transcription_engine.configure_language(
                language_config=self.config.transcription_language,
                model_size=self.config.model_size
            )
    
    def _update_language_display(self):
        """Update the language information display"""
        if self.transcription_engine and self.config.language_detection_enabled:
            lang_info = self.transcription_engine.get_language_info()
            configured_lang = lang_info['configured_language']
            
            if configured_lang == "auto":
                detected = lang_info.get('detected_language', 'none')
                confidence = lang_info.get('language_confidence', 0.0)
                if detected and detected != 'none':
                    self.language_label.config(text=f"🌍 Auto-detect: {detected} ({confidence:.1%})")
                else:
                    self.language_label.config(text="🌍 Auto-detect: waiting for speech...")
            else:
                # Get display name for configured language
                available_languages = Config.get_available_languages()
                display_name = available_languages.get(configured_lang, configured_lang)
                self.language_label.config(text=f"🌍 Language: {display_name}")
        else:
            self.language_label.config(text="")
    
    def start_live_mode(self):
        """Start live transcription mode with real-time typing"""
        try:
            # Start transcription engine streaming
            self.transcription_engine.start_streaming_transcription(self.handle_streaming_text)
            
            # Start live typing mode
            self.text_inserter.start_live_typing_mode()
            
            # Start streaming audio recording
            success = self.audio_recorder.start_streaming_recording(self.transcription_engine.process_audio_chunk)
            
            if success:
                self.live_mode_active = True
                self.live_button.config(text="⏹️ Stop Live", style="Accent.TButton")
                self.record_button.config(state="disabled")
                self.status_label.config(text="🔴 Live Mode: Speak and text will appear at cursor!")
                self.progress.start()
            else:
                # Cleanup on failure
                self.transcription_engine.stop_streaming_transcription()
                self.text_inserter.stop_live_typing_mode()
                messagebox.showerror("Error", "Failed to start live mode. Check microphone permissions.")
                
        except Exception as e:
            messagebox.showerror("Error", f"Failed to start live mode: {e}")
    
    def stop_live_mode(self):
        """Stop live transcription mode"""
        if not self.live_mode_active:
            return
        
        try:
            # Stop all streaming processes
            self.audio_recorder.stop_streaming_recording()
            self.transcription_engine.stop_streaming_transcription()
            self.text_inserter.stop_live_typing_mode()
            
            self.live_mode_active = False
            self.live_button.config(text="🔴 Live Mode", style="TButton")
            self.record_button.config(state="normal")
            self.status_label.config(text="Ready")
            self.progress.stop()
            
        except Exception as e:
            print(f"Error stopping live mode: {e}")
    
    def handle_streaming_text(self, text: str):
        """Handle real-time transcribed text from streaming engine"""
        if text and self.live_mode_active:
            # Queue text for live typing
            self.text_inserter.queue_text_for_live_typing(text)
            
            # Update status to show what's being transcribed
            self.root.after(0, lambda: self.status_label.config(
                text=f"🔴 Live: {text[:30]}..." if len(text) > 30 else f"🔴 Live: {text}"
            ))
            
            # Update language display with latest detection info
            if self.config.language_detection_enabled:
                self.root.after(0, self._update_language_display)
    
    def quit_app(self):
        if self.recording:
            self.audio_recorder.stop_recording()
        if self.live_mode_active:
            self.stop_live_mode()
        self.root.quit()
    
    def run(self):
        self.root.protocol("WM_DELETE_WINDOW", self.quit_app)
        self.root.mainloop()
import threading
from typing import Optional
from tkinter import messagebox

from .audio_recorder import AudioRecorder
from .transcription import TranscriptionEngine
from .text_inserter import TextInserter
from .config import Config
from .recording_indicator import RecordingIndicator, IndicatorState
from .double_key_shortcuts import DoubleKeyDetector

from .ui.components.main_window import MainWindow
from .ui.components.status_display import StatusDisplay
from .ui.components.recording_controls import RecordingControls
from .ui.settings.settings_manager import SettingsManager


class SpeechToTextGUI:
    """
    Refactored main GUI class using modular components.
    Orchestrates the interaction between UI components and business logic.
    """

    def __init__(self):
        # Initialize core components
        self.config = Config.load()
        self._init_business_logic()

        # Initialize UI components
        self.main_window = MainWindow("Speech to Text")
        main_frame = self.main_window.get_main_frame()

        # Setup UI components in order
        self.status_display = StatusDisplay(main_frame, self.config, row=0)
        next_row = self.status_display.get_next_row()

        self.recording_controls = RecordingControls(main_frame, row=next_row)

        # Wire up callbacks
        self._setup_callbacks()

        # Apply initial configurations
        self._apply_initial_configuration()

        # State tracking
        self.recording = False
        self.processing = False
        self.live_mode_active = False
        self.current_transcription_thread = None
        self.transcription_cancelled = False

    def _init_business_logic(self):
        """Initialize business logic components"""
        self.audio_recorder = AudioRecorder(device_id=self.config.microphone_device)
        self.transcription_engine = TranscriptionEngine(
            language_config=self.config.transcription_language,
            model_size=self.config.model_size,
            live_quality_mode=self.config.live_quality_mode,
            enable_overlap_detection=self.config.enable_overlap_detection,
            debug_text_assembly=self.config.debug_text_assembly,
            whisper_provider=self.config.whisper_provider,
        )
        self.text_inserter = TextInserter()

        # Initialize recording indicator
        self.recording_indicator = RecordingIndicator(
            position_x=self.config.indicator_position_x,
            position_y=self.config.indicator_position_y,
            size=self.config.indicator_size,
            opacity=self.config.indicator_opacity,
        )

        # Initialize double key detector
        self.double_key_detector = DoubleKeyDetector(
            double_press_timeout=self.config.double_press_timeout
        )

    def _setup_callbacks(self):
        """Setup all component callbacks"""
        # Main window callbacks
        self.main_window.set_close_callback(self.quit_app)

        # Recording controls callbacks
        self.recording_controls.set_callbacks(
            on_toggle_recording=self.toggle_recording,
            on_toggle_live_mode=self.toggle_live_mode,
            on_settings=self.show_settings,
            on_quit=self.quit_app,
        )

        # Status display callback
        self.status_display.set_transcription_engine(self.transcription_engine)

    def _apply_initial_configuration(self):
        """Apply configurations on startup"""
        self._apply_vad_config()
        self._apply_paste_config()
        self._apply_language_config()
        self._apply_shortcuts_config()

        # Start recording indicator if enabled
        if self.config.show_recording_indicator:
            self.recording_indicator.start(self.main_window.get_root())

    def toggle_recording(self):
        """Toggle recording mode"""
        self.toggle_recording_with_source("button")

    def toggle_live_mode(self):
        """Toggle live mode"""
        self.toggle_live_mode_with_source("button")

    def toggle_recording_with_source(self, source: str = "button"):
        """Toggle recording with source information for status messages"""
        if self.processing or self.live_mode_active:
            return

        if not self.recording:
            self.start_recording(source)
        else:
            self.stop_recording(source)

    def toggle_live_mode_with_source(self, source: str = "button"):
        """Toggle live mode with source information for status messages"""
        if self.processing or self.recording:
            return

        if not self.live_mode_active:
            self.start_live_mode(source)
        else:
            self.stop_live_mode(source)

    def start_recording(self, source: str = "button"):
        """Start recording mode"""
        if self.audio_recorder.start_recording():
            self.recording = True
            self.recording_controls.set_recording_state(recording=True)
            self.status_display.set_recording_status(source)

            # Update indicator state
            if self.config.show_recording_indicator:
                self.recording_indicator.set_state(IndicatorState.RECORDING)
        else:
            messagebox.showerror(
                "Error", "Failed to start recording. Check microphone permissions."
            )

    def stop_recording(self, source: str = "button"):
        """Stop recording mode"""
        if not self.recording:
            return

        # Cancel any ongoing transcription
        self.transcription_cancelled = True
        if (
            self.current_transcription_thread
            and self.current_transcription_thread.is_alive()
        ):
            # Give the thread a moment to notice the cancellation
            self.current_transcription_thread.join(timeout=1.0)

        self.recording = False
        self.transcription_cancelled = False
        self.recording_controls.set_recording_state(recording=False, processing=True)
        self.status_display.set_processing_status()

        # Update indicator to processing state
        if self.config.show_recording_indicator:
            self.recording_indicator.set_state(IndicatorState.PROCESSING)

        self.processing = True
        self.current_transcription_thread = threading.Thread(
            target=self._process_audio_thread, daemon=True
        )
        self.current_transcription_thread.start()

    def _process_audio_thread(self):
        """Process audio in a separate thread"""
        try:
            if self.transcription_cancelled:
                self.main_window.after(0, lambda: self.handle_transcription(None))
                return

            audio_data = self.audio_recorder.stop_recording()
            if audio_data is not None and len(audio_data) > 0:
                # Use timeout from configuration (0 means no timeout)
                timeout = (
                    self.config.transcription_timeout
                    if self.config.transcription_timeout > 0
                    else None
                )

                if self.transcription_cancelled:
                    self.main_window.after(0, lambda: self.handle_transcription(None))
                    return

                text = self.transcription_engine.transcribe_audio(
                    audio_data, timeout=timeout
                )

                if not self.transcription_cancelled:
                    self.main_window.after(0, lambda: self.handle_transcription(text))
            else:
                self.main_window.after(0, lambda: self.handle_transcription(None))
        except TimeoutError as e:
            print(f"⏰ Transcription timeout: {e}")
            self.main_window.after(0, lambda: self.handle_timeout_error())
        except Exception as e:
            if not self.transcription_cancelled:
                self.main_window.after(0, lambda: self.handle_error(str(e)))

    def handle_transcription(self, text: Optional[str]):
        """Handle transcription results"""
        self.processing = False
        self.recording_controls.set_recording_state(recording=False, processing=False)

        # Hide indicator when processing is complete
        if self.config.show_recording_indicator:
            self.recording_indicator.set_state(IndicatorState.HIDDEN)

        if text:
            # Always copy to clipboard first
            self.text_inserter.copy_to_clipboard(text)

            # Start auto-insert mode if enabled
            if self.config.enable_auto_insert:
                success = self.text_inserter.start_auto_insert_mode(
                    text, self.config.auto_insert_timeout, self._on_auto_insert_complete
                )
                self.status_display.set_transcription_result(
                    text, auto_insert_enabled=success
                )
            else:
                self.status_display.set_transcription_result(
                    text, auto_insert_enabled=False
                )
        else:
            self.status_display.set_transcription_result(None)
            messagebox.showwarning(
                "No Speech", "No speech was detected or transcription failed."
            )

        # Only reset to "Ready" if auto-insert is not active
        if not self.text_inserter.is_auto_insert_active():
            self.main_window.after(3000, lambda: self.status_display.clear_status())

    def _on_auto_insert_complete(self, success: bool):
        """Callback when auto-insert mode completes"""
        self.status_display.set_auto_insert_complete(success)

    def handle_timeout_error(self):
        """Handle transcription timeout errors"""
        self.processing = False
        self.recording_controls.set_recording_state(recording=False, processing=False)

        # Hide indicator when timeout occurs
        if self.config.show_recording_indicator:
            self.recording_indicator.set_state(IndicatorState.HIDDEN)

        self.status_display.set_timeout_error(self.config.transcription_timeout)
        messagebox.showwarning(
            "Timeout",
            f"Transcription timed out after {self.config.transcription_timeout} seconds. "
            "Try using a smaller model size in settings or increase the timeout.",
        )

    def handle_error(self, error_msg: str):
        """Handle general errors"""
        self.processing = False
        self.recording_controls.set_recording_state(recording=False, processing=False)

        # Hide indicator when error occurs
        if self.config.show_recording_indicator:
            self.recording_indicator.set_state(IndicatorState.HIDDEN)

        self.status_display.set_error(error_msg)
        messagebox.showerror("Error", f"An error occurred: {error_msg}")

    def start_live_mode(self, source: str = "button"):
        """Start live transcription mode with real-time typing"""
        try:
            # Start transcription engine streaming
            self.transcription_engine.start_streaming_transcription(
                self.handle_streaming_text
            )

            # Start live typing mode
            self.text_inserter.start_live_typing_mode()

            # Start streaming audio recording
            success = self.audio_recorder.start_streaming_recording(
                self.transcription_engine.process_audio_chunk
            )

            if success:
                self.live_mode_active = True
                self.recording_controls.set_live_mode_state(active=True)
                self.status_display.set_live_mode_status(source)

                # Update indicator to live mode state
                if self.config.show_recording_indicator:
                    self.recording_indicator.set_state(IndicatorState.LIVE_MODE)
            else:
                # Cleanup on failure
                self.transcription_engine.stop_streaming_transcription()
                self.text_inserter.stop_live_typing_mode()
                messagebox.showerror(
                    "Error", "Failed to start live mode. Check microphone permissions."
                )

        except Exception as e:
            messagebox.showerror("Error", f"Failed to start live mode: {e}")

    def stop_live_mode(self, source: str = "button"):
        """Stop live transcription mode"""
        if not self.live_mode_active:
            return

        try:
            # Stop all streaming processes
            self.audio_recorder.stop_streaming_recording()
            self.transcription_engine.stop_streaming_transcription()
            self.text_inserter.stop_live_typing_mode()

            self.live_mode_active = False
            self.recording_controls.set_live_mode_state(active=False)
            self.status_display.clear_status()

            # Hide indicator when live mode stops
            if self.config.show_recording_indicator:
                self.recording_indicator.set_state(IndicatorState.HIDDEN)

        except Exception as e:
            print(f"Error stopping live mode: {e}")

    def handle_streaming_text(self, text: str):
        """Handle real-time transcribed text from streaming engine"""
        if text and self.live_mode_active:
            # Queue text for live typing
            self.text_inserter.queue_text_for_live_typing(text)

            # Update status to show what's being transcribed
            self.main_window.after(
                0, lambda: self.status_display.set_live_transcription_status(text)
            )

            # Update language display with latest detection info
            if self.config.language_detection_enabled:
                self.main_window.after(0, self.status_display.update_language_display)

    def show_settings(self):
        """Show the settings dialog"""
        settings_manager = SettingsManager(
            self.main_window.get_root(), self.config, self.on_config_changed
        )
        settings_manager.show()

    def on_config_changed(self, new_config: Config):
        """Handle config changes from settings dialog"""
        self.config = new_config

        # Recreate AudioRecorder with new device if it changed
        old_device = getattr(self.audio_recorder, "device_id", None)
        if old_device != self.config.microphone_device:
            # Only recreate if not currently recording
            if not self.recording:
                self.audio_recorder = AudioRecorder(
                    device_id=self.config.microphone_device
                )
                self.transcription_engine = TranscriptionEngine(
                    language_config=self.config.transcription_language,
                    model_size=self.config.model_size,
                    live_quality_mode=self.config.live_quality_mode,
                    enable_overlap_detection=self.config.enable_overlap_detection,
                    debug_text_assembly=self.config.debug_text_assembly,
                    whisper_provider=self.config.whisper_provider,
                )
                self._apply_vad_config()
                self._apply_paste_config()
                self._apply_language_config_async()
                self._apply_indicator_config()
                self._apply_shortcuts_config()
                messagebox.showinfo("Settings", "All settings updated successfully!")
            else:
                messagebox.showwarning(
                    "Settings",
                    "Settings saved. Microphone device will be updated after current recording stops.",
                )
        else:
            # Apply configuration to existing components
            self._apply_vad_config()
            self._apply_paste_config()
            self._apply_language_config_async()
            self._apply_indicator_config()
            self._apply_shortcuts_config()
            messagebox.showinfo("Settings", "Settings saved successfully!")

    def _apply_vad_config(self):
        """Apply VAD configuration to the audio recorder"""
        if self.audio_recorder:
            self.audio_recorder.configure_vad(
                aggressiveness=self.config.vad_aggressiveness,
                min_chunk_duration=self.config.vad_min_chunk_duration,
                max_chunk_duration=self.config.vad_max_chunk_duration,
                silence_timeout=self.config.vad_silence_timeout,
            )

    def _apply_paste_config(self):
        """Apply paste configuration to the text inserter"""
        if self.text_inserter:
            self.text_inserter.configure_pasting(
                paste_method=self.config.paste_method,
                paste_delay=self.config.paste_delay,
                live_paste_interval=self.config.live_paste_interval,
                restore_clipboard=self.config.restore_clipboard,
            )

    def _apply_language_config(self):
        """Apply language configuration to the transcription engine"""
        if self.transcription_engine:
            self.transcription_engine.configure_language(
                language_config=self.config.transcription_language,
                model_size=self.config.model_size,
                live_quality_mode=self.config.live_quality_mode,
                enable_overlap_detection=self.config.enable_overlap_detection,
                debug_text_assembly=self.config.debug_text_assembly,
            )

    def _apply_language_config_async(self):
        """Apply language configuration asynchronously to prevent GUI blocking"""
        if self.transcription_engine:

            def on_model_loaded(success: bool, error: Optional[str]):
                if success:
                    print("✅ Model loaded successfully")
                else:
                    print(f"❌ Model loading failed: {error}")
                    self.main_window.after(
                        0,
                        lambda: messagebox.showerror(
                            "Model Loading Error", f"Failed to load model: {error}"
                        ),
                    )

            self.transcription_engine.configure_language(
                language_config=self.config.transcription_language,
                model_size=self.config.model_size,
                live_quality_mode=self.config.live_quality_mode,
                enable_overlap_detection=self.config.enable_overlap_detection,
                debug_text_assembly=self.config.debug_text_assembly,
                async_loading=True,
                callback=on_model_loaded,
            )

    def _apply_indicator_config(self):
        """Apply indicator configuration"""
        if hasattr(self, "recording_indicator"):
            # Update indicator appearance
            self.recording_indicator.configure_appearance(
                position_x=self.config.indicator_position_x,
                position_y=self.config.indicator_position_y,
                size=self.config.indicator_size,
                opacity=self.config.indicator_opacity,
            )

            # Start or stop indicator based on config
            if self.config.show_recording_indicator:
                if not self.recording_indicator.is_running:
                    self.recording_indicator.start(self.main_window.get_root())
            else:
                if self.recording_indicator.is_running:
                    self.recording_indicator.stop()

    def _apply_shortcuts_config(self):
        """Apply shortcuts configuration"""
        if hasattr(self, "double_key_detector"):
            # Update double key detector timeout
            self.double_key_detector.configure(
                double_press_timeout=self.config.double_press_timeout
            )

            # Start or stop double key detection based on config
            if self.config.double_press_enabled:
                if not self.double_key_detector.running:
                    self._setup_double_key_shortcuts()
            else:
                if self.double_key_detector.running:
                    self.double_key_detector.stop()

    def _setup_double_key_shortcuts(self):
        """Setup double key press shortcuts"""
        try:
            # Set callback functions for double key presses
            self.double_key_detector.set_callbacks(
                on_double_shift=lambda: self.main_window.after(
                    0, lambda: self.toggle_recording_with_source("double_shift")
                ),
                on_double_control=lambda: self.main_window.after(
                    0, lambda: self.toggle_live_mode_with_source("double_control")
                ),
            )

            # Start the double key detector
            self.double_key_detector.start()

        except Exception as e:
            print(f"Failed to setup double key shortcuts: {e}")

    def setup_all_shortcuts(self):
        """Setup both traditional hotkeys and double key press shortcuts"""
        # Setup traditional hotkey (cmd+shift+r)
        self._setup_traditional_hotkey()

        # Setup double key press shortcuts if enabled
        if self.config.double_press_enabled:
            self._setup_double_key_shortcuts()

    def _setup_traditional_hotkey(self):
        """Setup the configurable hotkey"""
        try:
            from pynput import keyboard

            def on_hotkey():
                self.main_window.after(
                    0, lambda: self.toggle_recording_with_source("hotkey")
                )

            # Check if this is a single key or a key combination
            hotkey_str = self.config.hotkey

            # Single key hotkeys (like alt_r, shift_l, etc.)
            single_key_map = {
                "alt_l": keyboard.Key.alt_l,
                "alt_r": keyboard.Key.alt_r,
                "shift_l": keyboard.Key.shift_l,
                "shift_r": keyboard.Key.shift_r,
                "ctrl_l": keyboard.Key.ctrl_l,
                "ctrl_r": keyboard.Key.ctrl_r,
            }

            if hotkey_str in single_key_map:
                # Handle single key press
                target_key = single_key_map[hotkey_str]

                def on_press(key):
                    if key == target_key:
                        on_hotkey()

                def on_release(key):
                    pass  # We only care about press events for single keys

                self.hotkey_listener = keyboard.Listener(
                    on_press=on_press,
                    on_release=on_release,
                )
                self.hotkey_listener.start()

            else:
                # Handle key combinations (existing logic)
                hotkey = keyboard.HotKey(keyboard.HotKey.parse(hotkey_str), on_hotkey)

                # Create listener first
                self.hotkey_listener = keyboard.Listener()

                def for_canonical(f):
                    return lambda k: f(self.hotkey_listener.canonical(k))

                # Now recreate with proper callbacks
                self.hotkey_listener = keyboard.Listener(
                    on_press=for_canonical(hotkey.press),
                    on_release=for_canonical(hotkey.release),
                )
                self.hotkey_listener.start()

        except Exception as e:
            print(f"Failed to setup hotkey: {e}")

    def quit_app(self):
        """Quit the application"""
        if self.recording:
            self.audio_recorder.stop_recording()
        if self.live_mode_active:
            self.stop_live_mode()

        # Stop recording indicator
        if hasattr(self, "recording_indicator"):
            self.recording_indicator.stop()

        # Stop double key detector
        if hasattr(self, "double_key_detector"):
            self.double_key_detector.stop()

        # Stop traditional hotkey listener
        if hasattr(self, "hotkey_listener"):
            self.hotkey_listener.stop()

        # Clean up transcription engine resources
        if hasattr(self, "transcription_engine"):
            self.transcription_engine.cleanup_resources()

        self.main_window.quit()

    def run(self):
        """Start the application"""
        self.setup_all_shortcuts()
        self.main_window.run()

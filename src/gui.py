import tkinter as tk
from tkinter import ttk, messagebox
import threading
from typing import Callable, Optional
from .audio_recorder import AudioRecorder
from .transcription import TranscriptionEngine
from .text_inserter import TextInserter
from .config import Config

class SpeechToTextGUI:
    def __init__(self):
        self.root = tk.Tk()
        self.root.title("Speech to Text")
        self.root.geometry("300x200")
        self.root.resizable(False, False)
        
        self.config = Config.load()
        self.audio_recorder = AudioRecorder()
        self.transcription_engine = TranscriptionEngine()
        self.text_inserter = TextInserter()
        
        self.recording = False
        self.processing = False
        
        self.setup_ui()
        self.setup_hotkeys()
    
    def setup_ui(self):
        main_frame = ttk.Frame(self.root, padding="20")
        main_frame.grid(row=0, column=0, sticky=(tk.W, tk.E, tk.N, tk.S))
        
        self.status_label = ttk.Label(main_frame, text="Ready", font=("Arial", 12))
        self.status_label.grid(row=0, column=0, columnspan=2, pady=(0, 20))
        
        self.record_button = ttk.Button(
            main_frame, 
            text="🎤 Start Recording", 
            command=self.toggle_recording,
            width=20
        )
        self.record_button.grid(row=1, column=0, columnspan=2, pady=(0, 10))
        
        self.progress = ttk.Progressbar(main_frame, mode='indeterminate')
        self.progress.grid(row=2, column=0, columnspan=2, sticky=(tk.W, tk.E), pady=(0, 10))
        
        settings_button = ttk.Button(main_frame, text="Settings", command=self.show_settings)
        settings_button.grid(row=3, column=0, sticky=tk.W)
        
        quit_button = ttk.Button(main_frame, text="Quit", command=self.quit_app)
        quit_button.grid(row=3, column=1, sticky=tk.E)
        
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
        if self.processing:
            return
            
        if not self.recording:
            self.start_recording()
        else:
            self.stop_recording()
    
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
            if self.text_inserter.insert_text(text):
                self.status_label.config(text=f"✅ Inserted: {text[:30]}...")
            else:
                self.text_inserter.copy_to_clipboard(text)
                self.status_label.config(text=f"📋 Copied to clipboard: {text[:30]}...")
        else:
            self.status_label.config(text="❌ No speech detected")
            messagebox.showwarning("No Speech", "No speech was detected or transcription failed.")
        
        self.root.after(3000, lambda: self.status_label.config(text="Ready"))
    
    def handle_error(self, error_msg: str):
        self.processing = False
        self.record_button.config(text="🎤 Start Recording", state="normal", style="TButton")
        self.status_label.config(text="❌ Error occurred")
        messagebox.showerror("Error", f"An error occurred: {error_msg}")
        self.root.after(3000, lambda: self.status_label.config(text="Ready"))
    
    def show_settings(self):
        messagebox.showinfo("Settings", "Settings panel coming soon!")
    
    def quit_app(self):
        if self.recording:
            self.audio_recorder.stop_recording()
        self.root.quit()
    
    def run(self):
        self.root.protocol("WM_DELETE_WINDOW", self.quit_app)
        self.root.mainloop()
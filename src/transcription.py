import speech_recognition as sr
import tempfile
import os
import numpy as np
from typing import Optional
import soundfile as sf

class TranscriptionEngine:
    def __init__(self):
        self.recognizer = sr.Recognizer()
        
    def transcribe_audio(self, audio_data: np.ndarray, sample_rate: int = 16000) -> Optional[str]:
        try:
            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, audio_data.flatten(), sample_rate)
                temp_filename = temp_file.name
            
            try:
                with sr.AudioFile(temp_filename) as source:
                    audio = self.recognizer.record(source)
                
                text = self.recognizer.recognize_google(audio)
                return text
            
            except sr.UnknownValueError:
                print("Could not understand audio")
                return None
            except sr.RequestError as e:
                print(f"Speech recognition service error: {e}")
                return None
            finally:
                if os.path.exists(temp_filename):
                    os.unlink(temp_filename)
                    
        except Exception as e:
            print(f"Transcription error: {e}")
            return None
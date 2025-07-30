from faster_whisper import WhisperModel
import tempfile
import os
import numpy as np
from typing import Optional
import soundfile as sf

class TranscriptionEngine:
    def __init__(self):
        # Initialize Whisper model with English small version for optimal performance
        print("🔄 Loading Whisper model...")
        self.model = WhisperModel("small.en", device="cpu", compute_type="int8")
        print("✅ Whisper model loaded successfully")
        
    def transcribe_audio(self, audio_data: np.ndarray, sample_rate: int = 16000) -> Optional[str]:
        try:
            with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as temp_file:
                sf.write(temp_file.name, audio_data.flatten(), sample_rate)
                temp_filename = temp_file.name
            
            try:
                # Transcribe using faster-whisper
                segments, info = self.model.transcribe(temp_filename, beam_size=5)
                
                print(f"🌍 Detected language: {info.language} (probability: {info.language_probability:.2f})")
                
                # Combine all segments into a single text
                transcribed_text = ""
                segment_count = 0
                for segment in segments:
                    transcribed_text += segment.text.strip() + " "
                    segment_count += 1
                
                transcribed_text = transcribed_text.strip()
                
                if transcribed_text:
                    print(f"🎯 Speech detected ({segment_count} segments): '{transcribed_text}'")
                    return transcribed_text
                else:
                    print("⚠️ No speech detected in audio")
                    return None
                    
            finally:
                if os.path.exists(temp_filename):
                    os.unlink(temp_filename)
                    
        except Exception as e:
            print(f"❌ Transcription error: {e}")
            return None
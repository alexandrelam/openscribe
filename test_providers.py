#!/usr/bin/env python3
"""
Test script to verify provider availability and error handling.
"""


def test_provider_availability():
    print("Testing provider availability and error handling...")

    try:
        # Test basic imports
        from src.config import Config

        print("✅ Config import successful")

        # Test provider availability
        try:
            from src.providers import WHISPER_CPP_AVAILABLE, FasterWhisperProvider

            print(
                f"✅ Providers import successful - Whisper.cpp available: {WHISPER_CPP_AVAILABLE}"
            )
        except ImportError as e:
            print(f"❌ Providers import failed: {e}")
            return False

        # Test creating FasterWhisperProvider
        try:
            from src.providers import FasterWhisperProvider

            print("✅ FasterWhisperProvider can be imported")
        except ImportError as e:
            print(f"❌ FasterWhisperProvider import failed: {e}")

        # Test creating WhisperCppProvider
        try:
            from src.providers import WhisperCppProvider

            if WHISPER_CPP_AVAILABLE:
                print("✅ WhisperCppProvider available - whisper-cli command found")
            else:
                print(
                    "⚠️ WhisperCppProvider import successful but whisper-cli command not available"
                )
        except ImportError as e:
            print(
                f"⚠️ WhisperCppProvider import failed (expected if whisper-cli not installed): {e}"
            )

        # Test TranscriptionEngine
        try:
            from src.transcription import TranscriptionEngine

            print("✅ TranscriptionEngine import successful")

            # Test with faster-whisper provider
            engine = TranscriptionEngine(whisper_provider="faster-whisper")
            print("✅ TranscriptionEngine with faster-whisper created successfully")
            print(f"   Provider: {engine.provider_name}")
            print(f"   Provider info: {engine.provider_info}")

        except Exception as e:
            print(f"❌ TranscriptionEngine test failed: {e}")

        # Test with whisper-cpp provider
        try:
            engine_cpp = TranscriptionEngine(whisper_provider="whisper-cpp")
            print("✅ TranscriptionEngine with whisper-cpp created successfully")
            print(f"   Provider: {engine_cpp.provider_name}")
            print(f"   Provider info: {engine_cpp.provider_info}")
        except Exception as e:
            print(
                f"⚠️ TranscriptionEngine with whisper-cpp failed (expected if whisper-cli not available): {e}"
            )

        # Test Config providers
        try:
            providers = Config.get_available_providers()
            print(f"✅ Available providers: {list(providers.keys())}")
        except Exception as e:
            print(f"❌ Config.get_available_providers() failed: {e}")

        print("\n🎉 Provider test completed!")
        return True

    except Exception as e:
        print(f"❌ Fatal error during provider test: {e}")
        return False


if __name__ == "__main__":
    test_provider_availability()

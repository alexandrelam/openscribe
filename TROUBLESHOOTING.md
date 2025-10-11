# OpenScribe Troubleshooting Guide

This guide provides solutions to common issues you might encounter while using OpenScribe.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Setup Issues](#setup-issues)
- [Permission Issues](#permission-issues)
- [Recording Issues](#recording-issues)
- [Transcription Issues](#transcription-issues)
- [Hotkey Issues](#hotkey-issues)
- [Performance Issues](#performance-issues)
- [General Debugging](#general-debugging)

---

## Installation Issues

### whisper-cpp Installation Fails

**Problem**: `brew install whisper-cpp` fails or returns an error

**Solutions**:
1. Update Homebrew first:
   ```bash
   brew update
   brew upgrade
   ```

2. If whisper-cpp is not found:
   ```bash
   # Add the tap if needed
   brew tap ggerganov/ggerganov
   brew install whisper-cpp
   ```

3. Verify installation:
   ```bash
   which whisper-cli
   whisper-cli --version
   ```

### Build Errors

**Problem**: `make build` fails with compilation errors

**Solutions**:
1. Check Go version (requires 1.21+):
   ```bash
   go version
   ```

2. Update Go if needed:
   ```bash
   brew upgrade go
   ```

3. Clean and rebuild:
   ```bash
   make clean
   make deps
   make build
   ```

4. Check for C compiler (needed for CGo):
   ```bash
   xcode-select --install
   ```

### Cannot Find openscribe Command

**Problem**: Terminal says `command not found: openscribe`

**Solutions**:
1. Check if binary exists:
   ```bash
   ls -la bin/openscribe
   ```

2. Add to PATH or use absolute path:
   ```bash
   # Run from project directory
   ./bin/openscribe

   # Or install to GOPATH
   make install
   ```

3. Verify GOPATH is in your PATH:
   ```bash
   echo $GOPATH
   echo $PATH
   # GOPATH/bin should be in PATH
   ```

---

## Setup Issues

### "No Whisper models found!"

**Problem**: Running `openscribe start` shows error about missing models

**Solution**:
```bash
# Run setup to download default model
openscribe setup

# Or download a specific model
openscribe models download small
```

### Model Download Fails

**Problem**: Model download fails with network error or timeout

**Solutions**:
1. Check internet connection

2. Try downloading manually:
   ```bash
   # Download to the correct directory
   cd ~/Library/Application\ Support/openscribe/models/

   # Download model (example for small model)
   curl -L -o ggml-small.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin
   ```

3. Verify download:
   ```bash
   openscribe models list
   ```

4. Check disk space:
   ```bash
   df -h ~
   # Models range from 75MB (tiny) to 3GB (large)
   ```

### Model Download Incomplete or Corrupted

**Problem**: Model download appears to complete but transcription fails

**Solutions**:
1. Remove corrupted model:
   ```bash
   rm ~/Library/Application\ Support/openscribe/models/ggml-*.bin
   ```

2. Download again:
   ```bash
   openscribe models download small
   ```

3. Check file size matches expected:
   - tiny: ~75MB
   - base: ~145MB
   - small: ~500MB
   - medium: ~1.5GB
   - large: ~3GB

---

## Permission Issues

### Microphone Permission Denied

**Problem**: OpenScribe cannot access microphone

**Solutions**:
1. Grant microphone permission:
   - Open **System Preferences** → **Security & Privacy** → **Privacy** → **Microphone**
   - Look for `openscribe` or your terminal app (Terminal.app, iTerm2, etc.)
   - Check the box to enable microphone access

2. If openscribe is not in the list:
   - Run `openscribe start` to trigger the permission prompt
   - Click "OK" when macOS asks for microphone access

3. Restart OpenScribe after granting permission

### Accessibility Permission Denied

**Problem**: Text auto-paste doesn't work, or you see "Accessibility permissions required"

**Solutions**:
1. Grant accessibility permission:
   - Open **System Preferences** → **Security & Privacy** → **Privacy** → **Accessibility**
   - Click the lock icon (bottom left) and enter your password
   - Click the "+" button
   - Navigate to and add your terminal application or the `openscribe` binary
   - Check the box to enable it

2. Alternative: Add from terminal:
   ```bash
   # This will prompt you to grant permission
   openscribe start
   ```

3. Restart OpenScribe after granting permission

4. If auto-paste still doesn't work:
   - Use `--no-paste` flag to disable auto-paste:
     ```bash
     openscribe start --no-paste
     ```
   - Copy text manually from terminal output

### "Operation not permitted" Errors

**Problem**: Various "operation not permitted" errors

**Solutions**:
1. Grant **Full Disk Access** (may be needed on newer macOS versions):
   - Open **System Preferences** → **Security & Privacy** → **Privacy** → **Full Disk Access**
   - Add your terminal application

2. Check System Integrity Protection (SIP) status:
   ```bash
   csrutil status
   ```
   - SIP should be enabled for security
   - If disabled, re-enable it

---

## Recording Issues

### No Microphone Detected

**Problem**: `openscribe config --list-microphones` shows no devices

**Solutions**:
1. Check if microphone is connected (for external mics)

2. Verify microphone works in other apps (e.g., QuickTime Player)

3. Check System Preferences:
   - Open **System Preferences** → **Sound** → **Input**
   - Verify microphone is listed and working (input level bars should move)

4. Restart OpenScribe and try again

### Wrong Microphone Selected

**Problem**: OpenScribe is using the wrong microphone

**Solutions**:
1. List available microphones:
   ```bash
   openscribe config --list-microphones
   ```

2. Set the correct microphone:
   ```bash
   openscribe config --set-microphone "MacBook Pro Microphone"
   ```

3. Or override on command line:
   ```bash
   openscribe start --microphone "USB Microphone"
   ```

### Recording Audio is Distorted or Silent

**Problem**: Recording produces no sound or distorted audio

**Solutions**:
1. Check microphone input level:
   - Open **System Preferences** → **Sound** → **Input**
   - Adjust input volume (should be in middle range, not maxed out)

2. Test microphone in verbose mode:
   ```bash
   openscribe start --verbose
   ```
   - Check console output for audio level indicators

3. Try a different microphone:
   ```bash
   openscribe config --list-microphones
   openscribe config --set-microphone "Different Mic"
   ```

4. Check for background noise/interference:
   - Move away from fans, AC units, etc.
   - Use an external microphone if built-in mic is poor quality

### Recording Cuts Off Early

**Problem**: Recording stops before you finish speaking

**Solutions**:
1. Make sure you're double-pressing the hotkey to stop, not single-pressing

2. Check hotkey configuration:
   ```bash
   openscribe config --show
   ```

3. Try a different hotkey to avoid conflicts:
   ```bash
   openscribe config --list-hotkeys
   openscribe config --set-hotkey "Left Shift"
   ```

---

## Transcription Issues

### Transcription is Inaccurate

**Problem**: Transcribed text has many errors or doesn't match what you said

**Solutions**:
1. Use a larger, more accurate model:
   ```bash
   openscribe config --set-model medium
   # or
   openscribe start --model large
   ```

2. Specify your language explicitly:
   ```bash
   openscribe config --set-language en
   # or
   openscribe start --language es
   ```

3. Improve recording quality:
   - Speak clearly and at moderate pace
   - Reduce background noise
   - Move closer to microphone
   - Use an external microphone if possible

4. Check audio levels in System Preferences

5. Try recording a longer sample (Whisper works better with more context)

### Transcription is Slow

**Problem**: Transcription takes a long time to complete

**Solutions**:
1. Use a smaller, faster model:
   ```bash
   openscribe config --set-model base
   # or
   openscribe start --model tiny
   ```

2. Close resource-intensive applications

3. First transcription is always slower (model loading):
   - Subsequent transcriptions will be faster
   - Keep OpenScribe running if doing multiple transcriptions

4. Check system resources:
   ```bash
   top -o cpu
   ```

5. Consider hardware limitations:
   - Larger models require more RAM and CPU
   - Older Macs may struggle with medium/large models

### Transcription Fails / Returns Empty Text

**Problem**: Transcription completes but produces no text or fails with error

**Solutions**:
1. Check model file integrity:
   ```bash
   openscribe models list
   # Re-download if needed
   openscribe models download small
   ```

2. Run in verbose mode to see detailed error:
   ```bash
   openscribe start --verbose
   ```

3. Verify whisper-cli works:
   ```bash
   whisper-cli --version
   ```

4. Check logs for errors:
   ```bash
   openscribe logs show
   ```

5. Try a different model:
   ```bash
   openscribe start --model base
   ```

6. Make sure you actually spoke during recording:
   - Recording silent audio will produce empty transcription
   - Verify audio feedback sounds played (start and stop beeps)

### Wrong Language Detected

**Problem**: Whisper transcribes in wrong language (e.g., Spanish instead of English)

**Solutions**:
1. Specify language explicitly:
   ```bash
   openscribe config --set-language en
   # or
   openscribe start --language en
   ```

2. Speak more clearly in the target language

3. Use a larger model (better at language detection):
   ```bash
   openscribe config --set-model medium
   ```

---

## Hotkey Issues

### Hotkey Not Detected

**Problem**: Double-pressing hotkey does nothing

**Solutions**:
1. **Most common**: Grant Accessibility permissions
   - See [Accessibility Permission Denied](#accessibility-permission-denied) section above

2. Check hotkey configuration:
   ```bash
   openscribe config --show
   ```

3. Try a different hotkey:
   ```bash
   openscribe config --list-hotkeys
   openscribe config --set-hotkey "Right Command"
   ```

4. Check for hotkey conflicts:
   - Another app might be using the same hotkey
   - Try disabling other hotkey tools temporarily
   - System shortcuts take precedence over app shortcuts

5. Verify OpenScribe is running:
   ```bash
   openscribe start
   # Terminal should show "Ready! Press hotkey to start recording..."
   ```

6. Check double-press timing:
   - Press twice quickly (within 500ms)
   - Don't press too fast or too slow

### Hotkey Triggers Other Actions

**Problem**: Hotkey also triggers system shortcuts or other apps

**Solutions**:
1. Change OpenScribe hotkey:
   ```bash
   openscribe config --set-hotkey "Left Shift"
   ```

2. Disable conflicting system shortcuts:
   - Open **System Preferences** → **Keyboard** → **Shortcuts**
   - Disable shortcuts that conflict with your chosen hotkey

3. Close or disable other hotkey applications temporarily

### Single Press Triggers Recording

**Problem**: Single hotkey press starts recording instead of requiring double-press

**Solutions**:
1. This should not happen - report as a bug if it does

2. Workaround: Use a less commonly pressed key:
   ```bash
   openscribe config --set-hotkey "Right Shift"
   ```

---

## Performance Issues

### High CPU Usage

**Problem**: OpenScribe uses too much CPU

**Solutions**:
1. Use a smaller model:
   ```bash
   openscribe config --set-model tiny
   ```

2. CPU usage is normal during transcription:
   - Whisper is CPU-intensive
   - Usage should drop after transcription completes

3. Close other applications during transcription

4. Don't record excessively long audio clips:
   - Keep recordings under 60 seconds for best performance

### High Memory Usage

**Problem**: OpenScribe uses too much RAM

**Solutions**:
1. Use a smaller model:
   - tiny: ~500MB RAM
   - base: ~800MB RAM
   - small: ~1.2GB RAM
   - medium: ~2.5GB RAM
   - large: ~4GB RAM

2. Quit and restart OpenScribe periodically

3. Close other memory-intensive applications

### Application Crashes

**Problem**: OpenScribe crashes or hangs

**Solutions**:
1. Run in verbose mode to get error details:
   ```bash
   openscribe start --verbose
   ```

2. Check logs:
   ```bash
   tail -f ~/Library/Logs/openscribe/transcriptions.log
   ```

3. Check system logs:
   ```bash
   log show --predicate 'process == "openscribe"' --last 10m
   ```

4. Try with minimal settings:
   ```bash
   openscribe start --model tiny --no-paste
   ```

5. Update to latest version

6. Report the issue with:
   - macOS version
   - OpenScribe version (`openscribe version`)
   - Steps to reproduce
   - Error messages from verbose mode

---

## General Debugging

### Enable Verbose Logging

Get detailed debug output:
```bash
openscribe start --verbose
```

This will:
- Show detailed status messages
- Display audio processing information
- Keep temporary WAV files for inspection
- Show transcription timing details

### Check Configuration

View current configuration:
```bash
openscribe config --show
```

Open configuration file in editor:
```bash
openscribe config --open
```

Or view configuration file directly:
```bash
cat ~/Library/Application\ Support/openscribe/config.yaml
```

### View Transcription History

Check past transcriptions:
```bash
openscribe logs show

# Show last 10 transcriptions
openscribe logs show -n 10
```

Log file location:
```bash
tail -f ~/Library/Logs/openscribe/transcriptions.log
```

### Test Individual Components

Test audio feedback:
```bash
openscribe config --test-sounds
```

Test microphone listing:
```bash
openscribe config --list-microphones
```

List downloaded models:
```bash
openscribe models list
```

### Clean Installation

If all else fails, start fresh:

1. Stop OpenScribe
2. Remove all data:
   ```bash
   rm -rf ~/Library/Application\ Support/openscribe/
   rm -rf ~/Library/Caches/openscribe/
   rm -rf ~/Library/Logs/openscribe/
   ```
3. Reinstall:
   ```bash
   make clean
   make build
   make install
   ```
4. Run setup again:
   ```bash
   openscribe setup
   ```

### Check System Requirements

Verify your system meets requirements:

```bash
# macOS version (10.15+)
sw_vers

# Go version (1.21+)
go version

# whisper-cpp installed
which whisper-cli
whisper-cli --version

# Disk space (need at least 1GB for small model)
df -h ~

# Available RAM
vm_stat | head -n 10
```

---

## Getting Help

If you've tried the solutions above and still have issues:

1. **Check existing issues**: [GitHub Issues](https://github.com/alexandrelam/openscribe-go/issues)
2. **Search discussions**: [GitHub Discussions](https://github.com/alexandrelam/openscribe-go/discussions)
3. **Open a new issue**: Include:
   - macOS version (`sw_vers`)
   - OpenScribe version (`openscribe version`)
   - Steps to reproduce the problem
   - Error messages (run with `--verbose`)
   - Output of `openscribe config --show`
   - Output of `openscribe models list`

---

## Common Error Messages

| Error Message | Likely Cause | Solution |
|---------------|--------------|----------|
| "No Whisper models found" | Model not downloaded | Run `openscribe setup` |
| "whisper-cpp is not installed" | Missing dependency | Run `brew install whisper-cpp` |
| "Microphone permission denied" | No microphone access | Grant permission in System Preferences |
| "Accessibility permission required" | No accessibility access | Grant permission in System Preferences |
| "Failed to initialize hotkey listener" | Accessibility permission or hotkey conflict | Check permissions and try different hotkey |
| "Model file not found" | Corrupted or incomplete download | Re-download model with `openscribe models download` |
| "Failed to transcribe audio" | Model error or invalid audio | Try different model or check audio recording |
| "No microphones available" | Microphone not detected | Check System Preferences → Sound → Input |

---

## Tips for Best Results

1. **Use quality microphone**: Built-in mics work, but external USB mics are better
2. **Quiet environment**: Reduce background noise for better accuracy
3. **Speak clearly**: Moderate pace, clear enunciation
4. **Right model for your needs**:
   - Quick notes: tiny or base
   - General use: small (recommended)
   - Important documents: medium or large
5. **Specify language**: Auto-detect works well, but explicit language is more accurate
6. **Keep recordings focused**: 10-30 seconds per recording is ideal
7. **Grant all permissions**: Both microphone and accessibility are required for full functionality

---

**Still having issues?** Open an issue on [GitHub](https://github.com/alexandrelam/openscribe-go/issues) with details about your problem.

# Audio Gain Control - Technical Design

## Overview

This design document outlines the technical approach for implementing audio level monitoring and automatic gain control in OpenScribe to improve transcription quality.

## Architecture

### Component Structure

```
internal/audio/
├── recorder.go          # Existing: Recording functionality (Modified)
├── levels.go            # New: Audio level measurement
├── gain.go              # New: Gain control and normalization
├── wav.go               # Existing: WAV file handling (Modified)
└── interface.go         # Existing: Audio interfaces (Modified if needed)

internal/config/
└── config.go            # Modified: Add gain control config

internal/cli/
└── start.go             # Modified: Display level warnings
```

### Data Flow

```
Recording Phase:
┌─────────────┐
│ Microphone  │
└──────┬──────┘
       │ PCM samples ([]byte)
       ▼
┌─────────────────────┐
│ Recorder.Start()    │
│ - Capture audio     │
│ - Store in buffer   │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Recorder.Stop()     │
│ - Return raw audio  │
└──────┬──────────────┘
       │
       ▼
Post-Recording Phase:
┌─────────────────────┐
│ AnalyzeLevel()      │
│ - Calculate RMS     │
│ - Convert to dB     │
│ - Return metrics    │
└──────┬──────────────┘
       │
       ▼
    ┌──┴──┐
    │ dB  │ < threshold?
    └──┬──┘
       │ Yes
       ▼
┌─────────────────────┐
│ ApplyGain()         │
│ - Normalize audio   │
│ - Prevent clipping  │
│ - Return adjusted   │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ SaveWAV()           │
│ - Write to disk     │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Whisper Transcribe  │
└─────────────────────┘
```

## Implementation Details

### 1. Audio Level Measurement (`internal/audio/levels.go`)

#### RMS Calculation
Root Mean Square provides an accurate measure of audio signal strength:

```go
// AudioLevelMetrics contains audio level analysis results
type AudioLevelMetrics struct {
    RMS          float64  // Root Mean Square value
    DecibelsFS   float64  // Decibels relative to full scale (dBFS)
    PeakAmplitude int16   // Maximum absolute sample value
    SampleRate   uint32   // Sample rate of audio
    Duration     float64  // Duration in seconds
}

// AnalyzeLevel calculates audio level metrics from PCM data
func AnalyzeLevel(audioData []byte, sampleRate uint32) AudioLevelMetrics
```

**Algorithm:**
1. Convert byte array to int16 samples (16-bit PCM)
2. Calculate sum of squares: Σ(sample²) / n
3. Take square root: RMS = √(mean of squares)
4. Convert to dBFS: 20 × log₁₀(RMS / 32768)
   - 32768 is max value for 16-bit signed integer
   - Result is negative (0 dBFS is maximum, -∞ is silence)

**Why dBFS (Decibels Full Scale)?**
- Logarithmic scale matches human hearing perception
- Standard in audio engineering
- Intuitive threshold values:
  - 0 dBFS: Maximum possible level (clipping)
  - -6 dBFS: Very loud, high quality
  - -20 dBFS: Good speaking level (target)
  - -40 dBFS: Quiet but usable (minimum)
  - -60 dBFS: Very quiet, poor quality
  - -∞ dBFS: Silence

### 2. Gain Control (`internal/audio/gain.go`)

#### Normalization Algorithm

```go
// GainControlConfig defines gain control parameters
type GainControlConfig struct {
    Enabled          bool    // Enable/disable gain control
    TargetLevelDB    float64 // Target level in dBFS (e.g., -20.0)
    MinThresholdDB   float64 // Minimum acceptable level (e.g., -40.0)
    MaxGainDB        float64 // Maximum gain to apply (e.g., 20.0)
    PreventClipping  bool    // Reduce gain if clipping would occur
}

// ApplyGain normalizes audio to target level
func ApplyGain(audioData []byte, currentLevel AudioLevelMetrics, config GainControlConfig) ([]byte, float64)
```

**Algorithm:**
1. **Calculate required gain:**
   ```
   requiredGainDB = targetDB - currentDB
   ```
   Example: If current is -50 dBFS and target is -20 dBFS, gain = 30 dB

2. **Apply gain limit:**
   ```
   actualGainDB = min(requiredGainDB, maxGainDB)
   ```
   Prevents excessive amplification of very quiet audio

3. **Convert dB to linear gain:**
   ```
   linearGain = 10^(actualGainDB / 20)
   ```
   Example: 20 dB gain = 10^(20/20) = 10x amplification

4. **Apply to samples:**
   ```
   For each sample in audioData:
       newSample = sample × linearGain
       if newSample > 32767:  newSample = 32767   # Prevent positive clipping
       if newSample < -32768: newSample = -32768  # Prevent negative clipping
   ```

5. **Return adjusted audio and actual gain applied**

#### Clipping Prevention

When `PreventClipping` is enabled:
1. Find peak sample value in audio
2. Calculate potential peak after gain: `peakAfterGain = peak × linearGain`
3. If `peakAfterGain > 32767`:
   - Reduce gain: `linearGain = 32767 / peak`
   - This ensures peak exactly reaches but doesn't exceed maximum

### 3. Configuration (`internal/config/config.go`)

Add to `Config` struct:

```go
type Config struct {
    // ... existing fields ...

    // Audio gain control settings
    AutoGain           bool    `yaml:"auto_gain"`            // default: true
    TargetLevelDB      float64 `yaml:"target_level_db"`      // default: -20.0
    MinThresholdDB     float64 `yaml:"min_threshold_db"`     // default: -40.0
    MaxGainDB          float64 `yaml:"max_gain_db"`          // default: 20.0
    ShowAudioLevels    bool    `yaml:"show_audio_levels"`    // default: false (verbose only)
}
```

Default values chosen based on audio engineering best practices:
- **-20 dBFS target**: Good speech level, leaves headroom for peaks
- **-40 dBFS threshold**: Below this, quality degrades significantly
- **20 dB max gain**: Reasonable limit, prevents over-amplification of noise

### 4. Integration (`internal/cli/start.go`)

Modify recording workflow in `hotkeyCallback`:

```go
// After recorder.Stop() returns audioData:

// 1. Analyze audio levels
levelMetrics := audio.AnalyzeLevel(audioData, recorder.GetSampleRate())

// 2. Log metrics (if verbose)
if cfg.Verbose {
    fmt.Printf("Audio level: %.1f dBFS (peak: %d)\n",
        levelMetrics.DecibelsFS, levelMetrics.PeakAmplitude)
}

// 3. Check if gain control needed
if cfg.AutoGain && levelMetrics.DecibelsFS < cfg.MinThresholdDB {
    fmt.Printf("⚠️  Low audio level detected (%.1f dBFS), applying gain...\n",
        levelMetrics.DecibelsFS)

    gainConfig := audio.GainControlConfig{
        Enabled:         true,
        TargetLevelDB:   cfg.TargetLevelDB,
        MinThresholdDB:  cfg.MinThresholdDB,
        MaxGainDB:       cfg.MaxGainDB,
        PreventClipping: true,
    }

    audioData, gainApplied := audio.ApplyGain(audioData, levelMetrics, gainConfig)

    if cfg.Verbose {
        fmt.Printf("Applied gain: +%.1f dB\n", gainApplied)
    }
}

// 4. Proceed with SaveWAV and transcription as before
```

## Testing Strategy

### Unit Tests

1. **levels_test.go**
   - Test RMS calculation with known waveforms (sine waves)
   - Test dB conversion accuracy
   - Test edge cases: silence, maximum amplitude
   - Test peak detection

2. **gain_test.go**
   - Test linear gain calculation
   - Test clipping prevention
   - Test gain limiting (max gain)
   - Test that target level is achieved
   - Test that silent audio is handled gracefully

### Integration Tests

1. **Record and process known audio files:**
   - Quiet audio (should trigger gain)
   - Normal audio (should pass through)
   - Loud audio (should not be modified)
   - Silent audio (should be handled gracefully)

2. **Transcription quality tests:**
   - Compare transcription accuracy with/without gain control
   - Use whisper-cli on processed vs. unprocessed audio

### Manual Testing

1. Test with various microphone input levels (via system settings)
2. Test with different microphone hardware
3. Test with different room environments (quiet/noisy)
4. Verify no quality degradation for normal-level audio

## Performance Considerations

### Time Complexity
- **Level analysis**: O(n) where n = number of samples
  - Single pass through audio data
  - Mathematical operations are constant time

- **Gain application**: O(n) where n = number of samples
  - Single pass to apply gain
  - Simple multiplication per sample

### Memory
- **In-place modification**: Modify existing audio buffer when possible
- **No significant overhead**: Only storing metrics struct (few bytes)

### Impact on Recording Flow
- Analysis and gain control happen **after** recording stops
- No impact on real-time recording performance
- Processing time: ~1-5ms for typical 30-second recording (negligible)

## Error Handling

1. **Division by zero**: If all samples are zero (silence), RMS = 0, log₁₀(0) = -∞
   - Handle: Set dBFS = -120.0 (effective silence)

2. **Invalid audio data**: Empty or odd-length byte array
   - Handle: Return error, don't attempt processing

3. **Gain overflow**: Ensure gain doesn't cause integer overflow
   - Handle: Clamp samples to valid range [-32768, 32767]

4. **Configuration validation**: Invalid threshold values
   - Handle: Use default values, log warning

## Future Enhancements (Out of Scope)

1. **Real-time level monitoring**: Display live meter during recording
2. **Adaptive gain**: Adjust gain dynamically during recording
3. **Noise gate**: Suppress background noise below threshold
4. **Compression**: Dynamic range compression for consistent levels
5. **Per-microphone profiles**: Remember optimal settings per device

## References

- Audio level measurement: https://en.wikipedia.org/wiki/Root_mean_square#Average_electrical_power
- dBFS standard: https://en.wikipedia.org/wiki/DBFS
- Audio normalization: https://en.wikipedia.org/wiki/Audio_normalization
- PCM format: https://en.wikipedia.org/wiki/Pulse-code_modulation

# Audio Level Monitoring Specification

## Overview

Audio level monitoring provides measurement and analysis of audio signal strength to ensure adequate input quality for transcription.

## ADDED Requirements

### Requirement: Calculate RMS (Root Mean Square) of Audio Signal

The system SHALL calculate the RMS value of recorded audio to measure average signal strength.

#### Scenario: Calculate RMS from 16-bit PCM audio data

**Given** a byte array containing 16-bit PCM audio samples
**When** the RMS calculation is performed
**Then** the system shall:
- Convert the byte array to int16 samples
- Calculate the sum of squared sample values
- Divide by the number of samples
- Return the square root of the result
- Handle edge case of empty or silent audio gracefully

#### Scenario: Convert RMS to decibels full scale (dBFS)

**Given** an RMS value calculated from audio samples
**When** converting to dBFS
**Then** the system shall:
- Use the formula: dBFS = 20 × log₁₀(RMS / 32768)
- Return a negative value (0 dBFS is maximum, -∞ is silence)
- Handle silence by returning -120.0 dBFS instead of -∞
- Ensure the result is in the range [-120.0, 0.0]

### Requirement: Track Peak Amplitude

The system SHALL identify the peak sample amplitude in recorded audio.

#### Scenario: Find maximum absolute sample value

**Given** recorded audio containing int16 PCM samples
**When** analyzing peak amplitude
**Then** the system shall:
- Iterate through all samples
- Track the maximum absolute value (considering both positive and negative)
- Return the peak as an int16 value
- Use peak to detect potential clipping (peak == 32767)

### Requirement: Provide Audio Level Metrics

The system SHALL return comprehensive audio level metrics for analysis and logging.

#### Scenario: Return structured audio metrics

**Given** completed audio level analysis
**When** metrics are requested
**Then** the system shall return:
- RMS value (float64)
- Decibels full scale value (float64)
- Peak amplitude (int16)
- Sample rate (uint32)
- Audio duration in seconds (float64)

### Requirement: Log Audio Levels in Verbose Mode

The system SHALL display audio level information when verbose mode is enabled.

#### Scenario: Display audio level after recording

**Given** verbose mode is enabled in configuration
**When** recording stops and audio is analyzed
**Then** the system shall:
- Display the dBFS level rounded to 1 decimal place
- Display the peak amplitude value
- Format output as: "Audio level: -X.X dBFS (peak: XXXXX)"
- Display this before transcription begins

### Requirement: Warn Users of Low Audio Levels

The system SHALL alert users when recorded audio levels are below quality threshold.

#### Scenario: Display warning for low audio

**Given** audio level analysis shows dBFS below minimum threshold (default -40 dB)
**When** auto-gain is disabled or before gain is applied
**Then** the system shall:
- Display a warning message with warning emoji (⚠️)
- Include the measured dBFS value
- Suggest the user may want to increase microphone input level
- Continue with transcription attempt (don't fail)

#### Scenario: No warning for adequate audio levels

**Given** audio level analysis shows dBFS at or above minimum threshold
**When** displaying recording results
**Then** the system shall not display any warning about audio levels

### Requirement: Handle Edge Cases in Audio Analysis

The system SHALL gracefully handle unusual or invalid audio data.

#### Scenario: Analyze completely silent audio

**Given** audio data where all samples are zero
**When** calculating audio levels
**Then** the system shall:
- Return RMS value of 0.0
- Return dBFS value of -120.0 (not -∞)
- Return peak amplitude of 0
- Not crash or throw errors

#### Scenario: Handle empty audio data

**Given** an empty byte array (zero length)
**When** attempting to analyze audio levels
**Then** the system shall:
- Return an error indicating invalid input
- Not attempt calculations that would divide by zero
- Provide clear error message

#### Scenario: Handle odd-length byte arrays

**Given** a byte array with odd number of bytes (invalid for 16-bit samples)
**When** attempting to analyze audio levels
**Then** the system shall:
- Either truncate to even length and analyze, or
- Return an error indicating invalid input format
- Document the behavior clearly

## Cross-References

- Related to: **audio-gain-control** - Uses level metrics to determine if gain needed
- Related to: **audio-recording** - Receives raw audio data from recorder
- Related to: **configuration** - Uses threshold values from config

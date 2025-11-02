# Audio Gain Control Specification

## Overview

Automatic gain control normalizes audio levels to an optimal range for transcription, improving quality when microphone input is too quiet.

## ADDED Requirements

### Requirement: Calculate Required Gain

The system SHALL determine the gain adjustment needed to reach target audio level.

#### Scenario: Calculate gain from current and target levels

**Given** current audio level is -50 dBFS and target level is -20 dBFS
**When** calculating required gain
**Then** the system shall:
- Calculate: requiredGain = targetDB - currentDB
- Return +30 dB in this example
- Support both positive gain (boost) and negative gain (attenuation)
- Handle edge case where current level is already at or above target

#### Scenario: Apply maximum gain limit

**Given** required gain exceeds configured maximum (e.g., 20 dB)
**When** determining actual gain to apply
**Then** the system shall:
- Cap the gain at the maximum value
- Log a warning if verbose mode is enabled
- Continue with limited gain rather than failing

### Requirement: Convert Decibel Gain to Linear Multiplier

The system SHALL convert logarithmic dB gain values to linear amplification factors.

#### Scenario: Convert dB gain to linear gain

**Given** a gain value in decibels
**When** converting to linear multiplier
**Then** the system shall:
- Use the formula: linearGain = 10^(gainDB / 20)
- For +20 dB, return 10.0x multiplier
- For +6 dB, return ~2.0x multiplier
- For 0 dB, return 1.0x multiplier (no change)
- For -6 dB, return ~0.5x multiplier

### Requirement: Apply Gain to Audio Samples

The system SHALL amplify audio samples by the calculated linear gain factor.

#### Scenario: Multiply samples by linear gain

**Given** audio data as int16 PCM samples and linear gain factor
**When** applying gain
**Then** the system shall:
- Convert byte array to int16 samples
- Multiply each sample by the linear gain factor
- Round the result to nearest integer
- Convert back to byte array
- Preserve sample rate and channel count

### Requirement: Prevent Audio Clipping

The system SHALL ensure amplified audio does not exceed valid sample range.

#### Scenario: Clamp samples to valid range

**Given** sample values after gain application
**When** any sample exceeds valid int16 range [-32768, 32767]
**Then** the system shall:
- Clamp positive values > 32767 to 32767
- Clamp negative values < -32768 to -32768
- Count number of clipped samples
- Log warning if clipping occurs and verbose mode enabled

#### Scenario: Reduce gain to prevent clipping

**Given** clipping prevention is enabled (default)
**When** calculated gain would cause peak to exceed 32767
**Then** the system shall:
- Find current peak amplitude
- Calculate: safeGain = 32767 / peak
- Use the lesser of calculated gain and safe gain
- Ensure peak reaches exactly 32767 without exceeding
- Log adjusted gain value if verbose mode enabled

### Requirement: Skip Gain for Adequate Audio Levels

The system SHALL only apply gain when audio is below quality threshold.

#### Scenario: Bypass gain for normal audio levels

**Given** recorded audio has level at or above minimum threshold (e.g., -40 dBFS)
**When** auto-gain processing runs
**Then** the system shall:
- Skip gain calculation and application
- Return original audio data unmodified
- Log "Audio level adequate, no gain needed" if verbose mode enabled

#### Scenario: Apply gain for low audio levels

**Given** recorded audio has level below minimum threshold
**When** auto-gain processing runs
**Then** the system shall:
- Calculate and apply appropriate gain
- Display message: "⚠️ Low audio level detected (X.X dBFS), applying gain..."
- Log applied gain amount if verbose mode enabled
- Proceed with boosted audio

### Requirement: Provide Gain Control Configuration

The system SHALL allow users to configure gain control behavior.

#### Scenario: User enables/disables auto-gain

**Given** configuration file or command-line flag
**When** user sets `auto_gain: false`
**Then** the system shall:
- Skip all gain processing
- Still measure and log audio levels
- Display warnings for low levels but not adjust them

#### Scenario: User configures target level

**Given** configuration file with `target_level_db: -18.0`
**When** gain control processes audio
**Then** the system shall normalize audio toward -18 dBFS instead of default -20 dBFS

#### Scenario: User configures minimum threshold

**Given** configuration file with `min_threshold_db: -35.0`
**When** determining if gain is needed
**Then** the system shall only apply gain for audio below -35 dBFS

#### Scenario: User configures maximum gain

**Given** configuration file with `max_gain_db: 15.0`
**When** calculating gain to apply
**Then** the system shall limit gain to +15 dB even if more is needed

### Requirement: Report Gain Applied

The system SHALL inform users about gain adjustments made to their audio.

#### Scenario: Display gain information in verbose mode

**Given** gain was applied to audio and verbose mode is enabled
**When** displaying post-processing results
**Then** the system shall:
- Show message: "Applied gain: +X.X dB"
- Show original and final audio levels
- Format consistently with 1 decimal place

#### Scenario: Display gain information for significant adjustments

**Given** gain >= 10 dB was applied (even if not verbose mode)
**When** displaying post-processing results
**Then** the system shall inform user that significant gain was applied to improve transcription quality

### Requirement: Maintain Audio Quality

The system SHALL ensure gain processing does not degrade audio quality.

#### Scenario: Preserve bit depth and sample rate

**Given** input audio is 16-bit PCM at 16kHz
**When** gain is applied
**Then** the system shall:
- Maintain 16-bit depth
- Maintain 16kHz sample rate
- Not introduce quantization beyond necessary rounding
- Preserve mono channel configuration

#### Scenario: Avoid excessive noise amplification

**Given** audio is extremely quiet (e.g., -70 dBFS)
**When** maximum gain limit would still result in poor quality
**Then** the system shall:
- Apply up to max gain limit only
- Log warning that audio may still be too quiet
- Suggest user increase microphone input level at system settings

### Requirement: Handle Gain Control Errors

The system SHALL gracefully handle errors during gain processing.

#### Scenario: Handle invalid gain values

**Given** configuration contains invalid gain values (e.g., target > 0 dBFS)
**When** initializing gain control
**Then** the system shall:
- Log error about invalid configuration
- Fall back to default values
- Continue operation without crashing

#### Scenario: Handle gain processing failure

**Given** an error occurs during gain application (e.g., memory allocation)
**When** gain processing fails
**Then** the system shall:
- Log the error
- Return original audio unmodified
- Continue with transcription attempt
- Not crash or exit the application

## MODIFIED Requirements

### Requirement: Process Audio Before Transcription

The system SHALL insert gain control step after recording, before saving WAV file.

#### Scenario: Apply gain control in recording workflow

**Given** recording has stopped and audio data is captured
**When** preparing audio for transcription
**Then** the system shall:
1. Stop recorder and retrieve audio bytes
2. **NEW:** Analyze audio levels using level monitoring
3. **NEW:** Apply gain control if needed and enabled
4. Save processed audio to WAV file
5. Pass WAV file to transcription
6. Continue with existing workflow

## Cross-References

- Depends on: **audio-level-monitoring** - Uses metrics to determine gain needed
- Related to: **audio-recording** - Modifies audio captured from recorder
- Related to: **configuration** - Uses gain control settings from config
- Related to: **transcription** - Improved audio quality enhances transcription accuracy

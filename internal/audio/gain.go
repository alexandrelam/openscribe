package audio

import (
	"encoding/binary"
	"fmt"
	"math"
)

// GainControlConfig defines parameters for automatic audio gain control.
// These settings control how audio normalization is performed to improve
// transcription quality for quiet recordings.
type GainControlConfig struct {
	Enabled         bool    // Enable/disable gain control
	TargetLevelDB   float64 // Target level in dBFS (e.g., -20.0)
	MinThresholdDB  float64 // Minimum acceptable level (e.g., -40.0)
	MaxGainDB       float64 // Maximum gain to apply (e.g., 20.0)
	PreventClipping bool    // Reduce gain if clipping would occur
}

// GainResult contains information about gain control processing.
type GainResult struct {
	GainAppliedDB    float64 // Actual gain applied in dB
	OriginalLevelDB  float64 // Original audio level in dBFS
	ResultingLevelDB float64 // Resulting audio level after gain
	WasLimited       bool    // Whether gain was limited (by max gain or clipping prevention)
}

// ProcessAudioGain analyzes audio and applies automatic gain control if needed.
//
// This is the main entry point for gain control. It analyzes the audio level,
// determines if gain is needed, and applies it while preventing clipping.
//
// Parameters:
//   - audioData: Raw PCM audio as byte slice (16-bit little-endian samples)
//   - metrics: Audio level metrics from AnalyzeLevel()
//   - config: Gain control configuration
//
// Returns:
//   - Processed audio data (or original if no gain needed)
//   - GainResult with details about processing
//   - error if processing fails
func ProcessAudioGain(audioData []byte, metrics AudioLevelMetrics, config GainControlConfig) ([]byte, GainResult, error) {
	result := GainResult{
		GainAppliedDB:    0,
		OriginalLevelDB:  metrics.DecibelsFS,
		ResultingLevelDB: metrics.DecibelsFS,
		WasLimited:       false,
	}

	// Check if gain control is enabled
	if !config.Enabled {
		return audioData, result, nil
	}

	// Check if audio level is below threshold (needs gain)
	if metrics.DecibelsFS >= config.MinThresholdDB {
		// Audio level is acceptable, no gain needed
		return audioData, result, nil
	}

	// Calculate required gain
	gainDB := CalculateGain(metrics.DecibelsFS, config.TargetLevelDB, config.MaxGainDB)
	result.GainAppliedDB = gainDB

	// Check if gain was limited
	requiredGain := config.TargetLevelDB - metrics.DecibelsFS
	if gainDB < requiredGain {
		result.WasLimited = true
	}

	// Apply gain to audio
	processedAudio, err := ApplyGain(audioData, gainDB, config.PreventClipping)
	if err != nil {
		return audioData, result, err
	}

	// Calculate resulting level
	result.ResultingLevelDB = metrics.DecibelsFS + gainDB

	return processedAudio, result, nil
}

// CalculateGain determines how much gain should be applied to reach the target level.
//
// Parameters:
//   - currentDB: Current audio level in dBFS
//   - targetDB: Target level in dBFS
//   - maxGainDB: Maximum gain allowed
//
// Returns:
//   - Gain to apply in dB (limited to maxGainDB)
func CalculateGain(currentDB, targetDB, maxGainDB float64) float64 {
	// Calculate required gain to reach target
	requiredGain := targetDB - currentDB

	// Limit to maximum gain
	if requiredGain > maxGainDB {
		return maxGainDB
	}

	// Don't allow negative gain (we only boost, never attenuate)
	if requiredGain < 0 {
		return 0
	}

	return requiredGain
}

// DBToLinear converts decibel gain to linear multiplier.
//
// Formula: linear = 10^(dB / 20)
//
// Examples:
//   - 0 dB = 1.0x (no change)
//   - 6 dB = 2.0x (double)
//   - 20 dB = 10.0x (10 times louder)
func DBToLinear(gainDB float64) float64 {
	return math.Pow(10, gainDB/20.0)
}

// ApplyGain applies gain to audio data and prevents clipping.
//
// Parameters:
//   - audioData: Raw PCM audio as byte slice (16-bit little-endian samples)
//   - gainDB: Gain to apply in decibels
//   - preventClipping: If true, reduce gain to prevent clipping
//
// Returns:
//   - Processed audio data
//   - error if processing fails
func ApplyGain(audioData []byte, gainDB float64, preventClipping bool) ([]byte, error) {
	// Validate input
	if len(audioData) == 0 {
		return audioData, fmt.Errorf("audio data is empty")
	}
	if len(audioData)%2 != 0 {
		return audioData, fmt.Errorf("audio data has odd length, expected 16-bit samples")
	}

	// Convert dB to linear gain
	linearGain := DBToLinear(gainDB)

	// Convert byte array to int16 samples
	numSamples := len(audioData) / 2
	samples := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(audioData[i*2 : i*2+2]))
	}

	// If clipping prevention is enabled, find peak and adjust gain
	if preventClipping {
		var peakAmplitude int16
		for _, sample := range samples {
			absSample := sample
			if absSample < 0 {
				absSample = -absSample
			}
			if absSample > peakAmplitude {
				peakAmplitude = absSample
			}
		}

		// Calculate what the peak would be after applying gain
		projectedPeak := float64(peakAmplitude) * linearGain

		// If it would clip, reduce gain to just reach max
		if projectedPeak > 32767.0 {
			// Calculate safe gain that brings peak to exactly 32767
			safeGain := 32767.0 / float64(peakAmplitude)
			if safeGain < linearGain {
				linearGain = safeGain
			}
		}
	}

	// Apply gain to all samples
	outputData := make([]byte, len(audioData))
	for i := 0; i < numSamples; i++ {
		// Multiply by gain
		gained := float64(samples[i]) * linearGain

		// Clamp BEFORE converting to int16 to avoid overflow
		if gained > 32767.0 {
			gained = 32767.0
		} else if gained < -32768.0 {
			gained = -32768.0
		}

		// Round to nearest integer
		newSample := int16(math.Round(gained))

		// Write back to byte array
		binary.LittleEndian.PutUint16(outputData[i*2:], uint16(newSample))
	}

	return outputData, nil
}

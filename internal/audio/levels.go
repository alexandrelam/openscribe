package audio

import (
	"encoding/binary"
	"fmt"
	"math"
)

// AudioLevelMetrics contains audio level analysis results.
// It provides detailed information about audio signal strength for
// quality assessment and automatic gain control.
type AudioLevelMetrics struct {
	RMS           float64 // Root Mean Square value of audio signal
	DecibelsFS    float64 // Decibels relative to full scale (dBFS), range: [-120, 0]
	PeakAmplitude int16   // Maximum absolute sample value found in audio
	SampleRate    uint32  // Sample rate in Hz (e.g., 16000)
	Duration      float64 // Duration of audio in seconds
}

// AnalyzeLevel calculates comprehensive audio level metrics from PCM data.
//
// This function analyzes raw PCM audio data to determine its signal strength
// and quality characteristics. The audio data should be 16-bit little-endian
// PCM samples (the standard format used by OpenScribe).
//
// Parameters:
//   - audioData: Raw PCM audio as byte slice (16-bit little-endian samples)
//   - sampleRate: Sample rate in Hz (typically 16000 for OpenScribe)
//
// Returns:
//   - AudioLevelMetrics with calculated RMS, dBFS, peak, and duration
//   - error if audio data is invalid or empty
//
// The returned dBFS value indicates audio quality:
//   - 0 dBFS: Maximum possible level (clipping/distortion)
//   - -6 dBFS: Very loud, high quality
//   - -20 dBFS: Good speaking level (target for speech)
//   - -40 dBFS: Quiet but usable (minimum threshold)
//   - -60 dBFS: Very quiet, poor quality
//   - -120 dBFS: Effectively silent
func AnalyzeLevel(audioData []byte, sampleRate uint32) (AudioLevelMetrics, error) {
	// Validate input
	if len(audioData) == 0 {
		return AudioLevelMetrics{}, fmt.Errorf("audio data is empty")
	}
	if len(audioData)%2 != 0 {
		return AudioLevelMetrics{}, fmt.Errorf("audio data has odd length, expected 16-bit samples")
	}
	if sampleRate == 0 {
		return AudioLevelMetrics{}, fmt.Errorf("invalid sample rate: 0")
	}

	// Convert byte array to int16 samples
	numSamples := len(audioData) / 2
	samples := make([]int16, numSamples)
	for i := 0; i < numSamples; i++ {
		samples[i] = int16(binary.LittleEndian.Uint16(audioData[i*2 : i*2+2]))
	}

	// Calculate duration
	duration := float64(numSamples) / float64(sampleRate)

	// Calculate RMS (Root Mean Square)
	var sumSquares float64
	var peakAmplitude int16

	for _, sample := range samples {
		// Accumulate sum of squares for RMS
		sumSquares += float64(sample) * float64(sample)

		// Track peak amplitude
		absSample := sample
		if absSample < 0 {
			absSample = -absSample
		}
		if absSample > peakAmplitude {
			peakAmplitude = absSample
		}
	}

	// Calculate RMS
	meanSquare := sumSquares / float64(numSamples)
	rms := math.Sqrt(meanSquare)

	// Convert RMS to dBFS (Decibels Full Scale)
	// Formula: dBFS = 20 * log10(RMS / 32768)
	// 32768 is the maximum value for 16-bit signed integer
	var decibelsFS float64
	if rms > 0 {
		decibelsFS = 20 * math.Log10(rms/32768.0)
	} else {
		// Handle silence: use -120 dBFS instead of -∞
		decibelsFS = -120.0
	}

	return AudioLevelMetrics{
		RMS:           rms,
		DecibelsFS:    decibelsFS,
		PeakAmplitude: peakAmplitude,
		SampleRate:    sampleRate,
		Duration:      duration,
	}, nil
}

package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestCalculateGain(t *testing.T) {
	tests := []struct {
		name        string
		currentDB   float64
		targetDB    float64
		maxGainDB   float64
		expectedDB  float64
	}{
		{
			name:       "Normal boost needed",
			currentDB:  -50.0,
			targetDB:   -20.0,
			maxGainDB:  30.0,
			expectedDB: 30.0, // -20 - (-50) = 30 dB
		},
		{
			name:       "Gain limited by max",
			currentDB:  -60.0,
			targetDB:   -20.0,
			maxGainDB:  20.0,
			expectedDB: 20.0, // Would need 40 dB, but limited to 20
		},
		{
			name:       "No gain needed",
			currentDB:  -15.0,
			targetDB:   -20.0,
			maxGainDB:  30.0,
			expectedDB: 0.0, // Already above target, don't attenuate
		},
		{
			name:       "Small boost",
			currentDB:  -25.0,
			targetDB:   -20.0,
			maxGainDB:  30.0,
			expectedDB: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateGain(tt.currentDB, tt.targetDB, tt.maxGainDB)
			if math.Abs(got-tt.expectedDB) > 0.01 {
				t.Errorf("CalculateGain() = %.2f, want %.2f", got, tt.expectedDB)
			}
		})
	}
}

func TestDBToLinear(t *testing.T) {
	tests := []struct {
		name      string
		gainDB    float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "0 dB = 1x",
			gainDB:    0.0,
			expected:  1.0,
			tolerance: 0.01,
		},
		{
			name:      "6 dB ≈ 2x",
			gainDB:    6.0,
			expected:  2.0,
			tolerance: 0.01,
		},
		{
			name:      "20 dB = 10x",
			gainDB:    20.0,
			expected:  10.0,
			tolerance: 0.01,
		},
		{
			name:      "40 dB = 100x",
			gainDB:    40.0,
			expected:  100.0,
			tolerance: 0.1,
		},
		{
			name:      "-6 dB ≈ 0.5x",
			gainDB:    -6.0,
			expected:  0.5,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DBToLinear(tt.gainDB)
			if math.Abs(got-tt.expected) > tt.tolerance {
				t.Errorf("DBToLinear(%.1f) = %.3f, want %.3f", tt.gainDB, got, tt.expected)
			}
		})
	}
}

func TestApplyGain_BasicFunctionality(t *testing.T) {
	// Create audio with known value
	numSamples := 100
	audioData := make([]byte, numSamples*2)
	originalValue := int16(1000)

	for i := 0; i < numSamples; i++ {
		binary.LittleEndian.PutUint16(audioData[i*2:], uint16(originalValue))
	}

	// Apply 6 dB gain (should double the amplitude)
	gainDB := 6.0
	result, err := ApplyGain(audioData, gainDB, false)
	if err != nil {
		t.Fatalf("ApplyGain() error = %v", err)
	}

	// Check result
	resultValue := int16(binary.LittleEndian.Uint16(result[0:2]))
	expectedValue := int16(math.Round(float64(originalValue) * DBToLinear(gainDB)))

	// Allow small rounding error (within 1%)
	tolerance := float64(expectedValue) * 0.01
	if math.Abs(float64(resultValue-expectedValue)) > tolerance {
		t.Errorf("ApplyGain() resulted in %d, want ~%d (tolerance: %.0f)", resultValue, expectedValue, tolerance)
	}
}

func TestApplyGain_ClippingPrevention(t *testing.T) {
	// Create audio with high amplitude
	numSamples := 100
	audioData := make([]byte, numSamples*2)
	originalValue := int16(20000)

	for i := 0; i < numSamples; i++ {
		binary.LittleEndian.PutUint16(audioData[i*2:], uint16(originalValue))
	}

	// Apply 20 dB gain (10x) which would cause clipping without prevention
	gainDB := 20.0
	result, err := ApplyGain(audioData, gainDB, true)
	if err != nil {
		t.Fatalf("ApplyGain() error = %v", err)
	}

	// Check that no sample exceeds max value
	numResultSamples := len(result) / 2
	for i := 0; i < numResultSamples; i++ {
		sample := int16(binary.LittleEndian.Uint16(result[i*2 : i*2+2]))
		if sample > 32767 || sample < -32768 {
			t.Errorf("Sample %d = %d, exceeds valid range [-32768, 32767]", i, sample)
		}
	}

	// The peak should be at or very close to 32767
	peakSample := int16(binary.LittleEndian.Uint16(result[0:2]))
	if peakSample < 32700 {
		t.Errorf("Clipping prevention too conservative: peak = %d, want ~32767", peakSample)
	}
}

func TestApplyGain_NoClippingPrevention(t *testing.T) {
	// Create audio that will clip without prevention
	numSamples := 100
	audioData := make([]byte, numSamples*2)
	originalValue := int16(20000)

	for i := 0; i < numSamples; i++ {
		binary.LittleEndian.PutUint16(audioData[i*2:], uint16(originalValue))
	}

	// Apply 20 dB gain (10x) without prevention - will clip
	// 20000 * 10 = 200000, which exceeds 32767, so it should be clipped
	gainDB := 20.0
	result, err := ApplyGain(audioData, gainDB, false)
	if err != nil {
		t.Fatalf("ApplyGain() error = %v", err)
	}

	// Should be clipped at 32767 (the clamping in ApplyGain prevents overflow)
	sample := int16(binary.LittleEndian.Uint16(result[0:2]))
	if sample != 32767 {
		t.Errorf("ApplyGain() without prevention = %d, want 32767 (clipped by safety clamp)", sample)
	}
}

func TestApplyGain_NegativeSamples(t *testing.T) {
	// Create audio with negative samples
	numSamples := 50
	audioData := make([]byte, numSamples*2)

	// Alternating positive and negative
	for i := 0; i < numSamples; i++ {
		var value int16
		if i%2 == 0 {
			value = 1000
		} else {
			value = -1000
		}
		binary.LittleEndian.PutUint16(audioData[i*2:], uint16(value))
	}

	// Apply gain
	gainDB := 6.0
	linearGain := DBToLinear(gainDB)
	result, err := ApplyGain(audioData, gainDB, false)
	if err != nil {
		t.Fatalf("ApplyGain() error = %v", err)
	}

	// Check both positive and negative samples were gained
	positiveSample := int16(binary.LittleEndian.Uint16(result[0:2]))
	negativeSample := int16(binary.LittleEndian.Uint16(result[2:4]))

	expectedPositive := int16(math.Round(1000 * linearGain))
	expectedNegative := int16(math.Round(-1000 * linearGain))

	// Allow 1% tolerance
	tolerancePos := math.Abs(float64(expectedPositive) * 0.01)
	toleranceNeg := math.Abs(float64(expectedNegative) * 0.01)

	if math.Abs(float64(positiveSample-expectedPositive)) > tolerancePos {
		t.Errorf("Positive sample = %d, want ~%d", positiveSample, expectedPositive)
	}
	if math.Abs(float64(negativeSample-expectedNegative)) > toleranceNeg {
		t.Errorf("Negative sample = %d, want ~%d", negativeSample, expectedNegative)
	}
}

func TestApplyGain_InvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		audioData []byte
		gainDB    float64
		wantErr   bool
	}{
		{
			name:      "Empty data",
			audioData: []byte{},
			gainDB:    6.0,
			wantErr:   true,
		},
		{
			name:      "Odd length",
			audioData: []byte{1, 2, 3},
			gainDB:    6.0,
			wantErr:   true,
		},
		{
			name:      "Valid input",
			audioData: make([]byte, 100),
			gainDB:    6.0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ApplyGain(tt.audioData, tt.gainDB, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyGain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessAudioGain_Disabled(t *testing.T) {
	// Create quiet audio
	numSamples := 100
	audioData := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ {
		binary.LittleEndian.PutUint16(audioData[i*2:], 100) // Very quiet
	}

	metrics := AudioLevelMetrics{
		DecibelsFS: -50.0,
	}

	config := GainControlConfig{
		Enabled:        false, // Disabled
		TargetLevelDB:  -20.0,
		MinThresholdDB: -40.0,
	}

	result, gainResult, err := ProcessAudioGain(audioData, metrics, config)
	if err != nil {
		t.Fatalf("ProcessAudioGain() error = %v", err)
	}

	// Should return unchanged audio
	if len(result) != len(audioData) {
		t.Errorf("Audio length changed: got %d, want %d", len(result), len(audioData))
	}

	// No gain should be applied
	if gainResult.GainAppliedDB != 0 {
		t.Errorf("Gain applied when disabled: %.1f dB", gainResult.GainAppliedDB)
	}
}

func TestProcessAudioGain_AboveThreshold(t *testing.T) {
	// Create audio above threshold
	numSamples := 100
	audioData := make([]byte, numSamples*2)

	metrics := AudioLevelMetrics{
		DecibelsFS: -30.0, // Above threshold of -40
	}

	config := GainControlConfig{
		Enabled:        true,
		TargetLevelDB:  -20.0,
		MinThresholdDB: -40.0,
		MaxGainDB:      20.0,
	}

	_, gainResult, err := ProcessAudioGain(audioData, metrics, config)
	if err != nil {
		t.Fatalf("ProcessAudioGain() error = %v", err)
	}

	// No gain should be needed
	if gainResult.GainAppliedDB != 0 {
		t.Errorf("Gain applied when above threshold: %.1f dB", gainResult.GainAppliedDB)
	}
}

func TestProcessAudioGain_BelowThreshold(t *testing.T) {
	// Create quiet audio
	numSamples := 100
	audioData := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ {
		binary.LittleEndian.PutUint16(audioData[i*2:], 500)
	}

	metrics := AudioLevelMetrics{
		DecibelsFS: -50.0, // Below threshold of -40
	}

	config := GainControlConfig{
		Enabled:         true,
		TargetLevelDB:   -20.0,
		MinThresholdDB:  -40.0,
		MaxGainDB:       40.0,
		PreventClipping: true,
	}

	result, gainResult, err := ProcessAudioGain(audioData, metrics, config)
	if err != nil {
		t.Fatalf("ProcessAudioGain() error = %v", err)
	}

	// Gain should be applied
	expectedGain := -20.0 - (-50.0) // 30 dB
	if math.Abs(gainResult.GainAppliedDB-expectedGain) > 0.1 {
		t.Errorf("Gain applied = %.1f dB, want %.1f dB", gainResult.GainAppliedDB, expectedGain)
	}

	// Audio should be modified
	if len(result) != len(audioData) {
		t.Errorf("Result length = %d, want %d", len(result), len(audioData))
	}

	// Samples should be larger
	originalSample := int16(binary.LittleEndian.Uint16(audioData[0:2]))
	resultSample := int16(binary.LittleEndian.Uint16(result[0:2]))

	if resultSample <= originalSample {
		t.Errorf("Result sample (%d) not larger than original (%d)", resultSample, originalSample)
	}
}

func TestProcessAudioGain_GainLimited(t *testing.T) {
	// Create very quiet audio
	numSamples := 100
	audioData := make([]byte, numSamples*2)

	metrics := AudioLevelMetrics{
		DecibelsFS: -70.0, // Very quiet
	}

	config := GainControlConfig{
		Enabled:        true,
		TargetLevelDB:  -20.0,
		MinThresholdDB: -40.0,
		MaxGainDB:      20.0, // Limited to 20 dB
	}

	_, gainResult, err := ProcessAudioGain(audioData, metrics, config)
	if err != nil {
		t.Fatalf("ProcessAudioGain() error = %v", err)
	}

	// Gain should be limited to max
	if gainResult.GainAppliedDB != 20.0 {
		t.Errorf("Gain applied = %.1f dB, want 20.0 dB (limited)", gainResult.GainAppliedDB)
	}

	// WasLimited flag should be set
	if !gainResult.WasLimited {
		t.Error("WasLimited should be true when gain is limited")
	}
}

func TestProcessAudioGain_ResultLevel(t *testing.T) {
	numSamples := 100
	audioData := make([]byte, numSamples*2)

	originalLevel := -45.0
	targetLevel := -20.0

	metrics := AudioLevelMetrics{
		DecibelsFS: originalLevel,
	}

	config := GainControlConfig{
		Enabled:        true,
		TargetLevelDB:  targetLevel,
		MinThresholdDB: -40.0,
		MaxGainDB:      40.0,
	}

	_, gainResult, err := ProcessAudioGain(audioData, metrics, config)
	if err != nil {
		t.Fatalf("ProcessAudioGain() error = %v", err)
	}

	// Check original level
	if gainResult.OriginalLevelDB != originalLevel {
		t.Errorf("OriginalLevelDB = %.1f, want %.1f", gainResult.OriginalLevelDB, originalLevel)
	}

	// Check resulting level
	expectedResult := originalLevel + gainResult.GainAppliedDB
	if math.Abs(gainResult.ResultingLevelDB-expectedResult) > 0.1 {
		t.Errorf("ResultingLevelDB = %.1f, want %.1f", gainResult.ResultingLevelDB, expectedResult)
	}

	// Should be close to target
	if math.Abs(gainResult.ResultingLevelDB-targetLevel) > 0.1 {
		t.Errorf("ResultingLevelDB = %.1f, want ~%.1f (target)", gainResult.ResultingLevelDB, targetLevel)
	}
}

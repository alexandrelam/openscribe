package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestAnalyzeLevel_Silence(t *testing.T) {
	// Create silent audio (all zeros)
	numSamples := 1000
	audioData := make([]byte, numSamples*2)

	metrics, err := AnalyzeLevel(audioData, 16000)
	if err != nil {
		t.Fatalf("AnalyzeLevel failed: %v", err)
	}

	if metrics.RMS != 0 {
		t.Errorf("Expected RMS = 0 for silence, got %f", metrics.RMS)
	}
	if metrics.DecibelsFS != -120.0 {
		t.Errorf("Expected dBFS = -120.0 for silence, got %f", metrics.DecibelsFS)
	}
	if metrics.PeakAmplitude != 0 {
		t.Errorf("Expected peak = 0 for silence, got %d", metrics.PeakAmplitude)
	}
}

func TestAnalyzeLevel_MaxAmplitude(t *testing.T) {
	// Create audio at maximum amplitude (32767)
	numSamples := 1000
	audioData := make([]byte, numSamples*2)

	for i := 0; i < numSamples; i++ {
		binary.LittleEndian.PutUint16(audioData[i*2:], uint16(32767))
	}

	metrics, err := AnalyzeLevel(audioData, 16000)
	if err != nil {
		t.Fatalf("AnalyzeLevel failed: %v", err)
	}

	// At max amplitude, RMS should be 32767
	if metrics.RMS != 32767.0 {
		t.Errorf("Expected RMS = 32767 for max amplitude, got %f", metrics.RMS)
	}

	// dBFS should be 0 (maximum level)
	if math.Abs(metrics.DecibelsFS-0.0) > 0.01 {
		t.Errorf("Expected dBFS ≈ 0.0 for max amplitude, got %f", metrics.DecibelsFS)
	}

	// Peak should be 32767
	if metrics.PeakAmplitude != 32767 {
		t.Errorf("Expected peak = 32767, got %d", metrics.PeakAmplitude)
	}
}

func TestAnalyzeLevel_SineWave(t *testing.T) {
	// Create a sine wave with known RMS
	// For a sine wave with amplitude A, RMS = A / sqrt(2) ≈ A * 0.707
	sampleRate := uint32(16000)
	duration := 1.0 // 1 second
	frequency := 440.0
	amplitude := 10000.0

	numSamples := int(float64(sampleRate) * duration)
	audioData := make([]byte, numSamples*2)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		sample := int16(amplitude * math.Sin(2*math.Pi*frequency*t))
		binary.LittleEndian.PutUint16(audioData[i*2:], uint16(sample))
	}

	metrics, err := AnalyzeLevel(audioData, sampleRate)
	if err != nil {
		t.Fatalf("AnalyzeLevel failed: %v", err)
	}

	// Expected RMS for sine wave: amplitude / sqrt(2)
	expectedRMS := amplitude / math.Sqrt(2)
	tolerance := amplitude * 0.01 // 1% tolerance

	if math.Abs(metrics.RMS-expectedRMS) > tolerance {
		t.Errorf("Expected RMS ≈ %.2f, got %.2f", expectedRMS, metrics.RMS)
	}

	// Check duration calculation
	if math.Abs(metrics.Duration-duration) > 0.001 {
		t.Errorf("Expected duration = %.3f s, got %.3f s", duration, metrics.Duration)
	}

	// Peak should be close to amplitude (within 1 sample)
	if math.Abs(float64(metrics.PeakAmplitude)-amplitude) > 1 {
		t.Errorf("Expected peak ≈ %.0f, got %d", amplitude, metrics.PeakAmplitude)
	}
}

func TestAnalyzeLevel_DBFSConversion(t *testing.T) {
	tests := []struct {
		name        string
		amplitude   int16
		expectedDBFS float64
		tolerance   float64
	}{
		{
			name:         "Half amplitude",
			amplitude:    16384, // Half of 32768
			expectedDBFS: -6.02, // 20 * log10(0.5) ≈ -6.02 dB
			tolerance:    0.1,
		},
		{
			name:         "Quarter amplitude",
			amplitude:    8192,
			expectedDBFS: -12.04, // 20 * log10(0.25) ≈ -12.04 dB
			tolerance:    0.1,
		},
		{
			name:         "Very quiet",
			amplitude:    328, // 1% of max
			expectedDBFS: -40.0,
			tolerance:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create audio with constant amplitude
			numSamples := 1000
			audioData := make([]byte, numSamples*2)

			for i := 0; i < numSamples; i++ {
				binary.LittleEndian.PutUint16(audioData[i*2:], uint16(tt.amplitude))
			}

			metrics, err := AnalyzeLevel(audioData, 16000)
			if err != nil {
				t.Fatalf("AnalyzeLevel failed: %v", err)
			}

			if math.Abs(metrics.DecibelsFS-tt.expectedDBFS) > tt.tolerance {
				t.Errorf("Expected dBFS ≈ %.2f, got %.2f", tt.expectedDBFS, metrics.DecibelsFS)
			}
		})
	}
}

func TestAnalyzeLevel_PeakDetection(t *testing.T) {
	// Create audio with known peaks
	numSamples := 1000
	audioData := make([]byte, numSamples*2)

	// Fill with low amplitude
	for i := 0; i < numSamples; i++ {
		binary.LittleEndian.PutUint16(audioData[i*2:], 100)
	}

	// Insert a peak in the middle
	peakValue := int16(25000)
	midpoint := numSamples / 2
	binary.LittleEndian.PutUint16(audioData[midpoint*2:], uint16(peakValue))

	metrics, err := AnalyzeLevel(audioData, 16000)
	if err != nil {
		t.Fatalf("AnalyzeLevel failed: %v", err)
	}

	if metrics.PeakAmplitude != peakValue {
		t.Errorf("Expected peak = %d, got %d", peakValue, metrics.PeakAmplitude)
	}
}

func TestAnalyzeLevel_NegativeSamples(t *testing.T) {
	// Test with negative samples (signed int16)
	numSamples := 100
	audioData := make([]byte, numSamples*2)

	// Create alternating positive and negative samples
	for i := 0; i < numSamples; i++ {
		var sample int16
		if i%2 == 0 {
			sample = 10000
		} else {
			sample = -10000
		}
		binary.LittleEndian.PutUint16(audioData[i*2:], uint16(sample))
	}

	metrics, err := AnalyzeLevel(audioData, 16000)
	if err != nil {
		t.Fatalf("AnalyzeLevel failed: %v", err)
	}

	// Peak should be absolute value
	if metrics.PeakAmplitude != 10000 {
		t.Errorf("Expected peak = 10000 (absolute), got %d", metrics.PeakAmplitude)
	}

	// RMS should be 10000 (all samples have same magnitude)
	if metrics.RMS != 10000.0 {
		t.Errorf("Expected RMS = 10000, got %f", metrics.RMS)
	}
}

func TestAnalyzeLevel_InvalidInput(t *testing.T) {
	tests := []struct {
		name       string
		audioData  []byte
		sampleRate uint32
		wantErr    bool
	}{
		{
			name:       "Empty data",
			audioData:  []byte{},
			sampleRate: 16000,
			wantErr:    true,
		},
		{
			name:       "Odd length",
			audioData:  []byte{1, 2, 3},
			sampleRate: 16000,
			wantErr:    true,
		},
		{
			name:       "Zero sample rate",
			audioData:  make([]byte, 100),
			sampleRate: 0,
			wantErr:    true,
		},
		{
			name:       "Valid input",
			audioData:  make([]byte, 100),
			sampleRate: 16000,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AnalyzeLevel(tt.audioData, tt.sampleRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAnalyzeLevel_DurationCalculation(t *testing.T) {
	tests := []struct {
		sampleRate      uint32
		durationSeconds float64
	}{
		{16000, 1.0},
		{16000, 0.5},
		{16000, 30.0},
		{44100, 1.0},
	}

	for _, tt := range tests {
		numSamples := int(float64(tt.sampleRate) * tt.durationSeconds)
		audioData := make([]byte, numSamples*2)

		metrics, err := AnalyzeLevel(audioData, tt.sampleRate)
		if err != nil {
			t.Fatalf("AnalyzeLevel failed: %v", err)
		}

		if math.Abs(metrics.Duration-tt.durationSeconds) > 0.001 {
			t.Errorf("Expected duration = %.3f s, got %.3f s",
				tt.durationSeconds, metrics.Duration)
		}
	}
}

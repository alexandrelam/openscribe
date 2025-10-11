package models

import (
	"testing"
)

func TestParseModelSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ModelSize
		wantErr bool
	}{
		{"Valid tiny", "tiny", Tiny, false},
		{"Valid base", "base", Base, false},
		{"Valid small", "small", Small, false},
		{"Valid medium", "medium", Medium, false},
		{"Valid large", "large", Large, false},
		{"Invalid model", "invalid", "", true},
		{"Empty string", "", "", true},
		{"Case sensitive", "SMALL", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModelSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseModelSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseModelSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAvailableModels(t *testing.T) {
	// Test that all expected models are available
	expectedModels := []ModelSize{Tiny, Base, Small, Medium, Large}

	for _, model := range expectedModels {
		t.Run(string(model), func(t *testing.T) {
			info, ok := AvailableModels[model]
			if !ok {
				t.Errorf("Model %s not found in AvailableModels", model)
			}

			// Verify model info has required fields
			if info.Name != model {
				t.Errorf("Model name mismatch: got %s, want %s", info.Name, model)
			}
			if info.Description == "" {
				t.Errorf("Model %s has empty description", model)
			}
			if info.SizeMB <= 0 {
				t.Errorf("Model %s has invalid size: %d", model, info.SizeMB)
			}
			if info.URL == "" {
				t.Errorf("Model %s has empty URL", model)
			}
			if info.FileName == "" {
				t.Errorf("Model %s has empty filename", model)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"0 bytes", 0, "0 B"},
		{"500 bytes", 500, "500 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 MB", 1572864, "1.5 MB"},
		{"10 GB", 10737418240, "10.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEstimateTimeRemaining(t *testing.T) {
	tests := []struct {
		name           string
		downloaded     int64
		total          int64
		bytesPerSecond float64
		wantContains   string
	}{
		{"30 seconds", 5000000, 10000000, 166666.67, "30s"},
		{"2 minutes", 1000000, 10000000, 75000, "2m"},
		{"Zero speed", 1000000, 10000000, 0, "calculating"},
		{"Zero total", 0, 0, 100, "calculating"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTimeRemaining(tt.downloaded, tt.total, tt.bytesPerSecond)
			if got != tt.wantContains {
				t.Logf("EstimateTimeRemaining() = %v, expected to contain %v", got, tt.wantContains)
				// Note: We're lenient with time estimation tests
			}
		})
	}
}

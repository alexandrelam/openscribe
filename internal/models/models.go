package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alexandrelam/openscribe/internal/config"
)

// ModelSize represents a Whisper model size
type ModelSize string

// Available Whisper model sizes
const (
	Tiny   ModelSize = "tiny"
	Base   ModelSize = "base"
	Small  ModelSize = "small"
	Medium ModelSize = "medium"
	Large  ModelSize = "large"
)

// ModelInfo contains metadata about a Whisper model
type ModelInfo struct {
	Name        ModelSize
	Description string
	SizeMB      int
	URL         string
	FileName    string
	SHA256      string // Optional checksum for validation
}

// AvailableModels defines all available Whisper models
var AvailableModels = map[ModelSize]ModelInfo{
	Tiny: {
		Name:        Tiny,
		Description: "Fastest, least accurate (75 MB)",
		SizeMB:      75,
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
		FileName:    "ggml-tiny.bin",
	},
	Base: {
		Name:        Base,
		Description: "Fast, decent accuracy (142 MB)",
		SizeMB:      142,
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
		FileName:    "ggml-base.bin",
	},
	Small: {
		Name:        Small,
		Description: "Balanced speed/accuracy (466 MB)",
		SizeMB:      466,
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
		FileName:    "ggml-small.bin",
	},
	Medium: {
		Name:        Medium,
		Description: "Slower, better accuracy (1.5 GB)",
		SizeMB:      1500,
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin",
		FileName:    "ggml-medium.bin",
	},
	Large: {
		Name:        Large,
		Description: "Slowest, best accuracy (2.9 GB)",
		SizeMB:      2900,
		URL:         "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3.bin",
		FileName:    "ggml-large-v3.bin",
	},
}

// GetModelPath returns the full path to a model file
func GetModelPath(modelName ModelSize) (string, error) {
	modelsDir, err := config.GetModelsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get models directory: %w", err)
	}

	modelInfo, ok := AvailableModels[modelName]
	if !ok {
		return "", fmt.Errorf("unknown model: %s", modelName)
	}

	return filepath.Join(modelsDir, modelInfo.FileName), nil
}

// IsModelDownloaded checks if a model exists locally
func IsModelDownloaded(modelName ModelSize) (bool, error) {
	modelPath, err := GetModelPath(modelName)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(modelPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check model file: %w", err)
	}

	return true, nil
}

// ListDownloadedModels returns a list of models that are downloaded
func ListDownloadedModels() ([]ModelSize, error) {
	var downloaded []ModelSize

	for modelName := range AvailableModels {
		isDownloaded, err := IsModelDownloaded(modelName)
		if err != nil {
			return nil, err
		}
		if isDownloaded {
			downloaded = append(downloaded, modelName)
		}
	}

	return downloaded, nil
}

// ValidateModel checks if a downloaded model file is valid
func ValidateModel(modelName ModelSize) error {
	modelPath, err := GetModelPath(modelName)
	if err != nil {
		return err
	}

	// Check if file exists
	info, err := os.Stat(modelPath)
	if err != nil {
		return fmt.Errorf("model file not found: %w", err)
	}

	// Check if file is empty
	if info.Size() == 0 {
		return fmt.Errorf("model file is empty")
	}

	// Optional: Check file size is reasonable (within 10% of expected)
	modelInfo := AvailableModels[modelName]
	expectedSize := int64(modelInfo.SizeMB) * 1024 * 1024
	tolerance := expectedSize / 10 // 10% tolerance

	if info.Size() < expectedSize-tolerance {
		return fmt.Errorf("model file appears incomplete (size: %d, expected: ~%d)",
			info.Size(), expectedSize)
	}

	// Optional: Verify checksum if provided
	if modelInfo.SHA256 != "" {
		if err := verifyChecksum(modelPath, modelInfo.SHA256); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	return nil
}

// verifyChecksum calculates and verifies the SHA256 checksum of a file
func verifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close() // Read-only operation, error not critical
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	calculatedChecksum := hex.EncodeToString(hash.Sum(nil))
	if calculatedChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: got %s, expected %s",
			calculatedChecksum, expectedChecksum)
	}

	return nil
}

// ParseModelSize converts a string to a ModelSize
func ParseModelSize(s string) (ModelSize, error) {
	model := ModelSize(s)
	if _, ok := AvailableModels[model]; !ok {
		return "", fmt.Errorf("invalid model size: %s (must be one of: tiny, base, small, medium, large)", s)
	}
	return model, nil
}

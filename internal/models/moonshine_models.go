package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandrelam/openscribe/internal/config"
)

// MoonshineModelSize represents a Moonshine model variant
type MoonshineModelSize string

// Available Moonshine model sizes
const (
	MoonshineTiny           MoonshineModelSize = "tiny"
	MoonshineBase           MoonshineModelSize = "base"
	MoonshineSmallStreaming MoonshineModelSize = "small-streaming"
)

// MoonshineModelInfo contains metadata about a Moonshine model
type MoonshineModelInfo struct {
	Name        MoonshineModelSize
	Description string
	// RequiredFiles are the files that must be present in the model directory
	RequiredFiles []string
	// BaseURL is the remote directory URL where model files are hosted
	BaseURL string
}

// AvailableMoonshineModels defines all available Moonshine models.
var AvailableMoonshineModels = map[MoonshineModelSize]MoonshineModelInfo{
	MoonshineTiny: {
		Name:          MoonshineTiny,
		Description:   "Fastest, smallest (~42 MB)",
		RequiredFiles: []string{"encoder_model.ort", "decoder_model_merged.ort", "tokenizer.bin"},
		BaseURL:       "https://download.moonshine.ai/model/tiny-en/quantized/tiny-en/",
	},
	MoonshineBase: {
		Name:          MoonshineBase,
		Description:   "Better accuracy, larger (~80 MB)",
		RequiredFiles: []string{"encoder_model.ort", "decoder_model_merged.ort", "tokenizer.bin"},
		BaseURL:       "https://download.moonshine.ai/model/base-en/quantized/base-en/",
	},
	MoonshineSmallStreaming: {
		Name:          MoonshineSmallStreaming,
		Description:   "Best English accuracy (~157 MB)",
		RequiredFiles: []string{"adapter.ort", "cross_kv.ort", "decoder_kv.ort", "encoder.ort", "frontend.ort", "streaming_config.json", "tokenizer.bin"},
		BaseURL:       "https://download.moonshine.ai/model/small-streaming-en/quantized/",
	},
}

// GetMoonshineModelDir returns the directory for a moonshine model
func GetMoonshineModelDir(modelName MoonshineModelSize) (string, error) {
	modelsDir, err := config.GetModelsDir()
	if err != nil {
		return "", fmt.Errorf("failed to get models directory: %w", err)
	}
	return filepath.Join(modelsDir, "moonshine", string(modelName)), nil
}

// IsMoonshineModelDownloaded checks if all files for a moonshine model are present
func IsMoonshineModelDownloaded(modelName MoonshineModelSize) (bool, error) {
	info, ok := AvailableMoonshineModels[modelName]
	if !ok {
		return false, fmt.Errorf("unknown moonshine model: %s", modelName)
	}

	modelDir, err := GetMoonshineModelDir(modelName)
	if err != nil {
		return false, err
	}

	for _, f := range info.RequiredFiles {
		path := filepath.Join(modelDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, fmt.Errorf("failed to check model file %s: %w", f, err)
		}
	}
	return true, nil
}

// ListDownloadedMoonshineModels returns moonshine models that are fully downloaded
func ListDownloadedMoonshineModels() ([]MoonshineModelSize, error) {
	var downloaded []MoonshineModelSize
	for name := range AvailableMoonshineModels {
		ok, err := IsMoonshineModelDownloaded(name)
		if err != nil {
			return nil, err
		}
		if ok {
			downloaded = append(downloaded, name)
		}
	}
	return downloaded, nil
}

// ParseMoonshineModelSize converts a string to a MoonshineModelSize
func ParseMoonshineModelSize(s string) (MoonshineModelSize, error) {
	model := MoonshineModelSize(s)
	if _, ok := AvailableMoonshineModels[model]; !ok {
		validModels := make([]string, 0, len(AvailableMoonshineModels))
		for k := range AvailableMoonshineModels {
			validModels = append(validModels, string(k))
		}
		return "", fmt.Errorf("invalid moonshine model size: %s (must be one of: %s)", s, strings.Join(validModels, ", "))
	}
	return model, nil
}

// DownloadMoonshineModel downloads model files directly from download.moonshine.ai
func DownloadMoonshineModel(modelName MoonshineModelSize, progress ProgressCallback) error {
	info, ok := AvailableMoonshineModels[modelName]
	if !ok {
		return fmt.Errorf("unknown moonshine model: %s", modelName)
	}

	modelDir, err := GetMoonshineModelDir(modelName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create moonshine model directory: %w", err)
	}

	for _, fileName := range info.RequiredFiles {
		url := info.BaseURL + fileName
		destPath := filepath.Join(modelDir, fileName)
		if err := downloadFile(url, destPath, progress); err != nil {
			return fmt.Errorf("failed to download %s: %w", fileName, err)
		}
	}

	return nil
}

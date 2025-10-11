package config

import (
	"os"
	"path/filepath"
)

// GetAppSupportDir returns ~/Library/Application Support/openscribe/
func GetAppSupportDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "openscribe"), nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	appSupport, err := GetAppSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appSupport, "config.yaml"), nil
}

// GetModelsDir returns the models directory path
func GetModelsDir() (string, error) {
	appSupport, err := GetAppSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appSupport, "models"), nil
}

// GetCacheDir returns ~/Library/Caches/openscribe/
func GetCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Caches", "openscribe"), nil
}

// GetLogsDir returns ~/Library/Logs/openscribe/
func GetLogsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", "openscribe"), nil
}

// GetTranscriptionLogPath returns the path to the transcription log file
func GetTranscriptionLogPath() (string, error) {
	logsDir, err := GetLogsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(logsDir, "transcriptions.log"), nil
}

// EnsureDirectories creates all necessary directories if they don't exist
func EnsureDirectories() error {
	// Get all directory paths
	appSupport, err := GetAppSupportDir()
	if err != nil {
		return err
	}

	modelsDir, err := GetModelsDir()
	if err != nil {
		return err
	}

	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	logsDir, err := GetLogsDir()
	if err != nil {
		return err
	}

	// Create directories with appropriate permissions (0755)
	dirs := []string{appSupport, modelsDir, cacheDir, logsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

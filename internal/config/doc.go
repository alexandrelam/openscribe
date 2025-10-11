// Package config manages OpenScribe's configuration file and system paths.
//
// This package provides:
//   - Configuration file reading and writing (YAML format)
//   - Default configuration values
//   - macOS standard paths for application data:
//   - ~/Library/Application Support/openscribe/ (config and models)
//   - ~/Library/Caches/openscribe/ (temporary files)
//   - ~/Library/Logs/openscribe/ (log files)
//   - Directory creation and validation
//
// The configuration file (config.yaml) stores user preferences including:
//   - Microphone selection
//   - Whisper model preference
//   - Language settings
//   - Hotkey configuration
//   - Audio feedback settings
//
// Example usage:
//
//	// Load configuration
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Modify and save
//	cfg.Microphone = "MacBook Pro Microphone"
//	cfg.Model = "small"
//	if err := cfg.Save(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get standard paths
//	modelsDir := config.GetModelsDir()
//	cacheDir := config.GetCacheDir()
//	logsDir := config.GetLogsDir()
package config

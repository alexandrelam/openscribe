// Package config provides configuration management for OpenScribe.
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Microphone is the selected audio input device (LEGACY - for backward compatibility)
	// Deprecated: Use PreferredMicrophones instead
	Microphone string `yaml:"microphone,omitempty"`

	// PreferredMicrophones is an ordered list of preferred microphone device names
	// The first available device in the list will be selected
	// If empty, falls back to Microphone field or system default
	PreferredMicrophones []string `yaml:"preferred_microphones,omitempty"`

	// Model is the Whisper model to use (tiny, base, small, medium, large)
	Model string `yaml:"model"`

	// Language is the target language for transcription (empty = auto-detect)
	Language string `yaml:"language"`

	// Hotkey is the keyboard shortcut for activation (default: Right Option)
	Hotkey string `yaml:"hotkey"`

	// AutoPaste determines whether to automatically paste transcribed text
	AutoPaste bool `yaml:"auto_paste"`

	// AudioFeedback determines whether to play sounds on state changes
	AudioFeedback bool `yaml:"audio_feedback"`

	// Verbose enables detailed debug output
	Verbose bool `yaml:"verbose"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Microphone:           "",         // Empty means use system default (legacy)
		PreferredMicrophones: []string{}, // Empty means use system default
		Model:                "small",
		Language:             "", // Empty means auto-detect
		Hotkey:               "Right Option",
		AutoPaste:            true,
		AudioFeedback:        true,
		Verbose:              false,
	}
}

// Load reads the configuration from disk, creating it with defaults if it doesn't exist
func Load() (*Config, error) {
	// Ensure directories exist first
	if err := EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Check if config file exists
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		// Config doesn't exist, create with defaults
		cfg := DefaultConfig()
		if saveErr := cfg.Save(); saveErr != nil {
			return nil, fmt.Errorf("failed to save default config: %w", saveErr)
		}
		return cfg, nil
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Run migration to handle legacy config
	cfg.migrate()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration in %s: %w\n\nYou can reset to defaults by running:\n  rm %s\n  openscribe config --show", configPath, err, configPath)
	}

	return cfg, nil
}

// migrate handles backward compatibility by auto-migrating legacy Microphone field
// to PreferredMicrophones array. This ensures seamless upgrade for existing users.
func (c *Config) migrate() {
	// Auto-migrate: If new field is empty but old field is set, populate new field
	if len(c.PreferredMicrophones) == 0 && c.Microphone != "" {
		c.PreferredMicrophones = []string{c.Microphone}
		log.Printf("[CONFIG] Migrated legacy 'microphone' field to 'preferred_microphones': %s", c.Microphone)

		// Optionally save migrated config immediately
		if err := c.Save(); err != nil {
			log.Printf("[CONFIG] Warning: Failed to save migrated config: %v", err)
		}
	}
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure parent directory exists
	if dirErr := EnsureDirectories(); dirErr != nil {
		return fmt.Errorf("failed to create directories: %w", dirErr)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with appropriate permissions (0644)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	// Validate preferred microphones
	seen := make(map[string]bool)
	for i, mic := range c.PreferredMicrophones {
		trimmed := strings.TrimSpace(mic)
		if trimmed == "" {
			return fmt.Errorf("preferred_microphones[%d] cannot be empty", i)
		}
		lowerMic := strings.ToLower(trimmed)
		if seen[lowerMic] {
			return fmt.Errorf("duplicate preferred microphone: %s", trimmed)
		}
		seen[lowerMic] = true
	}

	// Validate model
	validModels := map[string]bool{
		"tiny":   true,
		"base":   true,
		"small":  true,
		"medium": true,
		"large":  true,
	}
	if c.Model != "" && !validModels[c.Model] {
		return fmt.Errorf("invalid model: %s (must be one of: tiny, base, small, medium, large)", c.Model)
	}

	// Validate hotkey (basic validation - can be expanded)
	if c.Hotkey == "" {
		return fmt.Errorf("hotkey cannot be empty")
	}

	// Validate hotkey is a known key
	validHotkeys := map[string]bool{
		"Left Option":   true,
		"Right Option":  true,
		"Left Shift":    true,
		"Right Shift":   true,
		"Left Command":  true,
		"Right Command": true,
		"Left Control":  true,
		"Right Control": true,
	}
	if !validHotkeys[c.Hotkey] {
		return fmt.Errorf("invalid hotkey: %s (must be one of: Left Option, Right Option, Left Shift, Right Shift, Left Command, Right Command, Left Control, Right Control)", c.Hotkey)
	}

	// Note: We don't validate language codes as Whisper supports many languages
	// and we don't want to restrict users to a predefined list

	return nil
}

// String returns a formatted string representation of the config
func (c *Config) String() string {
	configPath, _ := GetConfigPath()
	modelsDir, _ := GetModelsDir()
	cacheDir, _ := GetCacheDir()
	logsDir, _ := GetLogsDir()

	microphone := c.Microphone
	if microphone == "" {
		microphone = "(system default)"
	}

	// Format preferred microphones list
	var preferredMics string
	if len(c.PreferredMicrophones) == 0 {
		preferredMics = "(none - using system default)"
	} else {
		preferredMics = "\n"
		for i, mic := range c.PreferredMicrophones {
			preferredMics += fmt.Sprintf("    %d. %s\n", i+1, mic)
		}
	}

	language := c.Language
	if language == "" {
		language = "auto-detect"
	}

	return fmt.Sprintf(`Current Configuration:

Settings:
  Microphone:      %s (legacy)
  Preferred Mics:  %s
  Model:           %s
  Language:        %s
  Hotkey:          %s
  Auto-paste:      %t
  Audio Feedback:  %t
  Verbose:         %t

Paths:
  Config:          %s
  Models:          %s
  Cache:           %s
  Logs:            %s
`,
		microphone,
		preferredMics,
		c.Model,
		language,
		c.Hotkey,
		c.AutoPaste,
		c.AudioFeedback,
		c.Verbose,
		configPath,
		modelsDir,
		cacheDir,
		logsDir,
	)
}

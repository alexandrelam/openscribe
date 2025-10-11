package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	// Microphone is the selected audio input device
	Microphone string `yaml:"microphone"`

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
		Microphone:    "", // Empty means use system default
		Model:         "small",
		Language:      "", // Empty means auto-detect
		Hotkey:        "Right Option",
		AutoPaste:     true,
		AudioFeedback: true,
		Verbose:       false,
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
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config doesn't exist, create with defaults
		cfg := DefaultConfig()
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
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

	return cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure parent directory exists
	if err := EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
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

	language := c.Language
	if language == "" {
		language = "auto-detect"
	}

	return fmt.Sprintf(`Current Configuration:

Settings:
  Microphone:      %s
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

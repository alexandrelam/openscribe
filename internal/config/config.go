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

	// Hotkey is the keyboard shortcut for activation (LEGACY - for backward compatibility)
	// Deprecated: Use Triggers instead
	Hotkey string `yaml:"hotkey,omitempty"`

	// Triggers is an array of keyboard/mouse triggers for activation
	// Examples: "Right Option", "Forward Button", "Back Button"
	// Any trigger can be double-pressed to start/stop recording
	Triggers []string `yaml:"triggers,omitempty"`

	// AutoPaste determines whether to automatically paste transcribed text
	AutoPaste bool `yaml:"auto_paste"`

	// AudioFeedback determines whether to play sounds on state changes
	AudioFeedback bool `yaml:"audio_feedback"`

	// Backend selects the transcription engine ("whisper", "moonshine", or "openai")
	Backend string `yaml:"backend"`

	// MoonshineModel is the Moonshine model to use (tiny, base)
	MoonshineModel string `yaml:"moonshine_model,omitempty"`

	// OpenAIAPIKey is the API key for OpenAI cloud transcription
	OpenAIAPIKey string `yaml:"openai_api_key,omitempty"`

	// OpenAIModel is the OpenAI model to use for transcription (e.g., "gpt-4o-transcribe", "whisper-1")
	OpenAIModel string `yaml:"openai_model,omitempty"`

	// Verbose enables detailed debug output
	Verbose bool `yaml:"verbose"`

	// Audio gain control settings
	// AutoGain enables automatic audio level normalization to improve transcription quality
	AutoGain bool `yaml:"auto_gain"`

	// TargetLevelDB is the target audio level in dBFS (e.g., -20.0)
	// This is the level that quiet audio will be boosted to
	TargetLevelDB float64 `yaml:"target_level_db"`

	// MinThresholdDB is the minimum acceptable audio level in dBFS (e.g., -40.0)
	// Audio below this level will trigger gain control (if AutoGain is enabled)
	MinThresholdDB float64 `yaml:"min_threshold_db"`

	// MaxGainDB is the maximum gain to apply in dB (e.g., 20.0)
	// This prevents excessive amplification of very quiet audio
	MaxGainDB float64 `yaml:"max_gain_db"`

	// ShowAudioLevels displays audio level information for all recordings
	// When false, levels are only shown in verbose mode
	ShowAudioLevels bool `yaml:"show_audio_levels"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		Microphone:           "",         // Empty means use system default (legacy)
		PreferredMicrophones: []string{}, // Empty means use system default
		Model:                "small",
		Language:             "",          // Empty means auto-detect
		Hotkey:               "",          // Legacy field (deprecated)
		Triggers:             []string{"Right Option"},
		AutoPaste:            true,
		AudioFeedback:        true,
		Backend:              "whisper",
		MoonshineModel:       "",
		Verbose:              false,
		AutoGain:             true,   // Enable automatic gain control by default
		TargetLevelDB:        -18.0,  // Optimal speech level for transcription (-18 dBFS)
		MinThresholdDB:       -35.0,  // Below this is considered too quiet for good transcription
		MaxGainDB:            25.0,   // Maximum 25 dB of gain (allows recovery from -43 dBFS)
		ShowAudioLevels:      false,  // Only show in verbose mode by default
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

// migrate handles backward compatibility by auto-migrating legacy fields
// to their new equivalents. This ensures seamless upgrade for existing users.
func (c *Config) migrate() {
	needsSave := false

	// Auto-migrate: Microphone → PreferredMicrophones
	if len(c.PreferredMicrophones) == 0 && c.Microphone != "" {
		c.PreferredMicrophones = []string{c.Microphone}
		log.Printf("[CONFIG] Migrated legacy 'microphone' field to 'preferred_microphones': %s", c.Microphone)
		needsSave = true
	}

	// Auto-migrate: Hotkey → Triggers
	if len(c.Triggers) == 0 && c.Hotkey != "" {
		c.Triggers = []string{c.Hotkey}
		log.Printf("[CONFIG] Migrated legacy 'hotkey' field to 'triggers': %s", c.Hotkey)
		needsSave = true
	}

	// Auto-migrate: Add gain control defaults if missing (zero values)
	// This handles configs created before gain control was added
	if c.TargetLevelDB == 0 && c.MinThresholdDB == 0 && c.MaxGainDB == 0 {
		defaults := DefaultConfig()
		c.TargetLevelDB = defaults.TargetLevelDB
		c.MinThresholdDB = defaults.MinThresholdDB
		c.MaxGainDB = defaults.MaxGainDB
		log.Printf("[CONFIG] Migrated gain control settings to defaults (target: %.1f dBFS, threshold: %.1f dBFS, max gain: %.1f dB)",
			c.TargetLevelDB, c.MinThresholdDB, c.MaxGainDB)
		needsSave = true
	}

	// Save migrated config if any migrations occurred
	if needsSave {
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

	// Validate backend
	validBackends := map[string]bool{
		"":          true,
		"whisper":   true,
		"moonshine": true,
		"openai":    true,
	}
	if !validBackends[c.Backend] {
		return fmt.Errorf("invalid backend: %s (must be one of: whisper, moonshine, openai)", c.Backend)
	}

	// Validate OpenAI backend requirements
	if c.Backend == "openai" && c.OpenAIAPIKey == "" {
		return fmt.Errorf("openai backend requires openai_api_key to be set. Use: openscribe config --set-openai-api-key <key>")
	}

	// Validate model (only enforce whisper model names when backend is whisper)
	if c.Backend == "" || c.Backend == "whisper" {
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
	}

	// Validate moonshine model if backend is moonshine
	if c.Backend == "moonshine" && c.MoonshineModel != "" {
		validMoonshineModels := map[string]bool{
			"tiny":            true,
			"base":            true,
			"small-streaming":  true,
			"medium-streaming": true,
		}
		if !validMoonshineModels[c.MoonshineModel] {
			return fmt.Errorf("invalid moonshine_model: %s (must be one of: tiny, base, small-streaming, medium-streaming)", c.MoonshineModel)
		}
	}

	// Validate triggers
	if len(c.Triggers) == 0 {
		return fmt.Errorf("triggers cannot be empty - at least one trigger is required")
	}

	// Valid trigger names (keyboard modifiers + mouse buttons)
	validTriggers := map[string]bool{
		"Left Option":    true,
		"Right Option":   true,
		"Left Shift":     true,
		"Right Shift":    true,
		"Left Command":   true,
		"Right Command":  true,
		"Left Control":   true,
		"Right Control":  true,
		"Forward Button": true,
		"Back Button":    true,
	}

	// Check each trigger
	seenTriggers := make(map[string]bool)
	for i, trigger := range c.Triggers {
		trimmed := strings.TrimSpace(trigger)
		if trimmed == "" {
			return fmt.Errorf("triggers[%d] cannot be empty", i)
		}

		// Check for duplicates (case-insensitive)
		lowerTrigger := strings.ToLower(trimmed)
		if seenTriggers[lowerTrigger] {
			return fmt.Errorf("duplicate trigger: %s", trimmed)
		}
		seenTriggers[lowerTrigger] = true

		// Validate trigger name
		if !validTriggers[trimmed] {
			return fmt.Errorf("invalid trigger: %s (must be one of: Left Option, Right Option, Left Shift, Right Shift, Left Command, Right Command, Left Control, Right Control, Forward Button, Back Button)", trimmed)
		}
	}

	// Warn if both legacy Hotkey and new Triggers are set
	if c.Hotkey != "" && len(c.Triggers) > 0 {
		log.Printf("[CONFIG] Warning: Both 'hotkey' (legacy) and 'triggers' are set. Using 'triggers' field.")
	}

	// Note: We don't validate language codes as Whisper supports many languages
	// and we don't want to restrict users to a predefined list

	// Validate audio gain control settings
	if c.TargetLevelDB > 0 {
		return fmt.Errorf("target_level_db must be negative (dBFS scale, 0 = max level)")
	}
	if c.MinThresholdDB > 0 {
		return fmt.Errorf("min_threshold_db must be negative (dBFS scale, 0 = max level)")
	}
	if c.TargetLevelDB < c.MinThresholdDB {
		return fmt.Errorf("target_level_db (%.1f) must be greater than min_threshold_db (%.1f)", c.TargetLevelDB, c.MinThresholdDB)
	}
	if c.MaxGainDB < 0 {
		return fmt.Errorf("max_gain_db must be positive (gain amount)")
	}
	if c.MaxGainDB > 40 {
		return fmt.Errorf("max_gain_db is too high (%.1f dB), maximum recommended is 40 dB", c.MaxGainDB)
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

	// Format triggers list
	var triggers string
	if len(c.Triggers) == 0 {
		triggers = "(none configured)"
	} else {
		triggers = "\n"
		for i, trigger := range c.Triggers {
			triggers += fmt.Sprintf("    %d. %s\n", i+1, trigger)
		}
	}

	// Show legacy hotkey if present
	var hotkeyDisplay string
	if c.Hotkey != "" {
		hotkeyDisplay = fmt.Sprintf("  Hotkey (legacy):  %s\n", c.Hotkey)
	}

	backend := c.Backend
	if backend == "" {
		backend = "whisper"
	}

	// Show moonshine model if relevant
	var moonshineDisplay string
	if c.Backend == "moonshine" {
		mm := c.MoonshineModel
		if mm == "" {
			mm = "tiny"
		}
		moonshineDisplay = fmt.Sprintf("\n  Moonshine Model: %s", mm)
	}

	// Show OpenAI settings if relevant
	var openaiDisplay string
	if c.Backend == "openai" {
		om := c.OpenAIModel
		if om == "" {
			om = "gpt-4o-transcribe"
		}
		keyDisplay := "(not set)"
		if c.OpenAIAPIKey != "" {
			keyDisplay = c.OpenAIAPIKey[:7] + "..." + c.OpenAIAPIKey[len(c.OpenAIAPIKey)-4:]
		}
		openaiDisplay = fmt.Sprintf("\n  OpenAI Model:    %s\n  OpenAI API Key:  %s", om, keyDisplay)
	}

	return fmt.Sprintf(`Current Configuration:

Settings:
  Backend:         %s%s%s
  Microphone:      %s (legacy)
  Preferred Mics:  %s
  Model:           %s
  Language:        %s
  Triggers:        %s%s  Auto-paste:      %t
  Audio Feedback:  %t
  Verbose:         %t

Audio Gain Control:
  Auto Gain:       %t
  Target Level:    %.1f dBFS
  Min Threshold:   %.1f dBFS
  Max Gain:        %.1f dB
  Show Levels:     %t

Paths:
  Config:          %s
  Models:          %s
  Cache:           %s
  Logs:            %s
`,
		backend,
		moonshineDisplay,
		openaiDisplay,
		microphone,
		preferredMics,
		c.Model,
		language,
		triggers,
		hotkeyDisplay,
		c.AutoPaste,
		c.AudioFeedback,
		c.Verbose,
		c.AutoGain,
		c.TargetLevelDB,
		c.MinThresholdDB,
		c.MaxGainDB,
		c.ShowAudioLevels,
		configPath,
		modelsDir,
		cacheDir,
		logsDir,
	)
}

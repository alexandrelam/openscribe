package config

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Microphone", cfg.Microphone, ""},
		{"Model", cfg.Model, "small"},
		{"Language", cfg.Language, ""},
		{"Hotkey", cfg.Hotkey, "Right Option"},
		{"AutoPaste", cfg.AutoPaste, true},
		{"AudioFeedback", cfg.AudioFeedback, true},
		{"Verbose", cfg.Verbose, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("DefaultConfig().%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestLoad_CreatesDefaultConfig(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Load should create config with defaults if it doesn't exist
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify defaults
	if cfg.Model != "small" {
		t.Errorf("Load() default model = %v, want small", cfg.Model)
	}
	if cfg.Hotkey != "Right Option" {
		t.Errorf("Load() default hotkey = %v, want Right Option", cfg.Hotkey)
	}

	// Verify config file was created
	configPath, _ := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Load() did not create config file")
	}
}

func TestLoad_ReadsExistingConfig(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Create a config with custom values
	original := &Config{
		Microphone:    "Test Mic",
		Model:         "base",
		Language:      "fr",
		Hotkey:        "Control+Shift+R",
		AutoPaste:     false,
		AudioFeedback: false,
		Verbose:       true,
	}

	// Save it
	err := original.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load it back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify all fields match
	if loaded.Microphone != original.Microphone {
		t.Errorf("Microphone = %v, want %v", loaded.Microphone, original.Microphone)
	}
	if loaded.Model != original.Model {
		t.Errorf("Model = %v, want %v", loaded.Model, original.Model)
	}
	if loaded.Language != original.Language {
		t.Errorf("Language = %v, want %v", loaded.Language, original.Language)
	}
	if loaded.Hotkey != original.Hotkey {
		t.Errorf("Hotkey = %v, want %v", loaded.Hotkey, original.Hotkey)
	}
	if loaded.AutoPaste != original.AutoPaste {
		t.Errorf("AutoPaste = %v, want %v", loaded.AutoPaste, original.AutoPaste)
	}
	if loaded.AudioFeedback != original.AudioFeedback {
		t.Errorf("AudioFeedback = %v, want %v", loaded.AudioFeedback, original.AudioFeedback)
	}
	if loaded.Verbose != original.Verbose {
		t.Errorf("Verbose = %v, want %v", loaded.Verbose, original.Verbose)
	}
}

func TestSave(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg := &Config{
		Microphone:    "Test Microphone",
		Model:         "large",
		Language:      "es",
		Hotkey:        "Command+R",
		AutoPaste:     false,
		AudioFeedback: true,
		Verbose:       true,
	}

	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	configPath, _ := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Save() did not create config file")
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	content := string(data)
	expectedFields := []string{
		"microphone: Test Microphone",
		"model: large",
		"language: es",
		"hotkey: Command+R",
		"auto_paste: false",
		"audio_feedback: true",
		"verbose: true",
	}

	for _, field := range expectedFields {
		if !strings.Contains(content, field) {
			t.Errorf("Saved config missing expected field: %s", field)
		}
	}
}

func TestSave_CreatesDirectories(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Don't create directories beforehand
	cfg := DefaultConfig()

	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify config directory was created
	appSupport, _ := GetAppSupportDir()
	if _, err := os.Stat(appSupport); os.IsNotExist(err) {
		t.Error("Save() did not create parent directories")
	}
}

func TestValidate_ValidModels(t *testing.T) {
	validModels := []string{"tiny", "base", "small", "medium", "large"}

	for _, model := range validModels {
		t.Run(model, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Model = model

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Validate() with model %s error = %v, want nil", model, err)
			}
		})
	}
}

func TestValidate_InvalidModel(t *testing.T) {
	invalidModels := []string{"invalid", "extra-large", "xl", "SMALL", "tiny-en"}

	for _, model := range invalidModels {
		t.Run(model, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Model = model

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Validate() with model %s error = nil, want error", model)
			}
			if !strings.Contains(err.Error(), "invalid model") {
				t.Errorf("Validate() error = %v, want error containing 'invalid model'", err)
			}
		})
	}
}

func TestValidate_EmptyModel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Model = ""

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() with empty model error = %v, want nil (empty should be valid)", err)
	}
}

func TestValidate_EmptyHotkey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Hotkey = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() with empty hotkey error = nil, want error")
	}
	if !strings.Contains(err.Error(), "hotkey cannot be empty") {
		t.Errorf("Validate() error = %v, want error containing 'hotkey cannot be empty'", err)
	}
}

func TestValidate_EmptyMicrophoneAndLanguage(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Microphone = ""
	cfg.Language = ""

	// Empty microphone and language should be valid (defaults to system default / auto-detect)
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() with empty microphone/language error = %v, want nil", err)
	}
}

func TestString_FormatsCorrectly(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg := &Config{
		Microphone:    "Built-in Microphone",
		Model:         "base",
		Language:      "en",
		Hotkey:        "Left Option",
		AutoPaste:     true,
		AudioFeedback: false,
		Verbose:       true,
	}

	output := cfg.String()

	// Check for expected content
	expectedStrings := []string{
		"Current Configuration:",
		"Built-in Microphone",
		"base",
		"en",
		"Left Option",
		"true",  // AutoPaste
		"false", // AudioFeedback
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("String() output missing expected string: %s\nGot:\n%s", expected, output)
		}
	}
}

func TestString_DefaultDisplays(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg := DefaultConfig()
	output := cfg.String()

	// Empty microphone should show as "(system default)"
	if !strings.Contains(output, "(system default)") {
		t.Error("String() should show '(system default)' for empty microphone")
	}

	// Empty language should show as "auto-detect"
	if !strings.Contains(output, "auto-detect") {
		t.Error("String() should show 'auto-detect' for empty language")
	}
}

func TestRoundTrip_SaveAndLoad(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Create config with all fields set
	original := &Config{
		Microphone:    "USB Microphone",
		Model:         "medium",
		Language:      "de",
		Hotkey:        "Shift+Command+Space",
		AutoPaste:     false,
		AudioFeedback: true,
		Verbose:       false,
	}

	// Save
	err := original.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Compare all fields
	if loaded.Microphone != original.Microphone {
		t.Errorf("RoundTrip: Microphone = %v, want %v", loaded.Microphone, original.Microphone)
	}
	if loaded.Model != original.Model {
		t.Errorf("RoundTrip: Model = %v, want %v", loaded.Model, original.Model)
	}
	if loaded.Language != original.Language {
		t.Errorf("RoundTrip: Language = %v, want %v", loaded.Language, original.Language)
	}
	if loaded.Hotkey != original.Hotkey {
		t.Errorf("RoundTrip: Hotkey = %v, want %v", loaded.Hotkey, original.Hotkey)
	}
	if loaded.AutoPaste != original.AutoPaste {
		t.Errorf("RoundTrip: AutoPaste = %v, want %v", loaded.AutoPaste, original.AutoPaste)
	}
	if loaded.AudioFeedback != original.AudioFeedback {
		t.Errorf("RoundTrip: AudioFeedback = %v, want %v", loaded.AudioFeedback, original.AudioFeedback)
	}
	if loaded.Verbose != original.Verbose {
		t.Errorf("RoundTrip: Verbose = %v, want %v", loaded.Verbose, original.Verbose)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	cfg := DefaultConfig()
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	configPath, _ := GetConfigPath()
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Check file permissions (0644)
	if info.Mode().Perm() != 0644 {
		t.Errorf("Config file permissions = %o, want 0644", info.Mode().Perm())
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Create directories
	err := EnsureDirectories()
	if err != nil {
		t.Fatalf("EnsureDirectories() error = %v", err)
	}

	// Write invalid YAML to config file
	configPath, _ := GetConfigPath()
	invalidYAML := []byte("this is not: [valid yaml")
	err = os.WriteFile(configPath, invalidYAML, 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}

	// Load should return an error
	_, err = Load()
	if err == nil {
		t.Error("Load() with invalid YAML should return error")
	}
	if !strings.Contains(err.Error(), "failed to parse config file") {
		t.Errorf("Load() error = %v, want error containing 'failed to parse config file'", err)
	}
}

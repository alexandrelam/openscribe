package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAppSupportDir(t *testing.T) {
	// Create a temporary home directory for testing
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	got, err := GetAppSupportDir()
	if err != nil {
		t.Fatalf("GetAppSupportDir() error = %v", err)
	}

	want := filepath.Join(tempHome, "Library", "Application Support", "openscribe")
	if got != want {
		t.Errorf("GetAppSupportDir() = %v, want %v", got, want)
	}
}

func TestGetConfigPath(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	got, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	want := filepath.Join(tempHome, "Library", "Application Support", "openscribe", "config.yaml")
	if got != want {
		t.Errorf("GetConfigPath() = %v, want %v", got, want)
	}
}

func TestGetModelsDir(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	got, err := GetModelsDir()
	if err != nil {
		t.Fatalf("GetModelsDir() error = %v", err)
	}

	want := filepath.Join(tempHome, "Library", "Application Support", "openscribe", "models")
	if got != want {
		t.Errorf("GetModelsDir() = %v, want %v", got, want)
	}
}

func TestGetCacheDir(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	got, err := GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir() error = %v", err)
	}

	want := filepath.Join(tempHome, "Library", "Caches", "openscribe")
	if got != want {
		t.Errorf("GetCacheDir() = %v, want %v", got, want)
	}
}

func TestGetLogsDir(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	got, err := GetLogsDir()
	if err != nil {
		t.Fatalf("GetLogsDir() error = %v", err)
	}

	want := filepath.Join(tempHome, "Library", "Logs", "openscribe")
	if got != want {
		t.Errorf("GetLogsDir() = %v, want %v", got, want)
	}
}

func TestGetTranscriptionLogPath(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	got, err := GetTranscriptionLogPath()
	if err != nil {
		t.Fatalf("GetTranscriptionLogPath() error = %v", err)
	}

	want := filepath.Join(tempHome, "Library", "Logs", "openscribe", "transcriptions.log")
	if got != want {
		t.Errorf("GetTranscriptionLogPath() = %v, want %v", got, want)
	}
}

func TestEnsureDirectories(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Ensure directories are created
	err := EnsureDirectories()
	if err != nil {
		t.Fatalf("EnsureDirectories() error = %v", err)
	}

	// Verify all directories were created
	dirs := []struct {
		name   string
		getter func() (string, error)
	}{
		{"AppSupport", GetAppSupportDir},
		{"Models", GetModelsDir},
		{"Cache", GetCacheDir},
		{"Logs", GetLogsDir},
	}

	for _, dir := range dirs {
		path, err := dir.getter()
		if err != nil {
			t.Fatalf("%s getter error = %v", dir.name, err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Directory %s not created at %s: %v", dir.name, path, err)
			continue
		}

		if !info.IsDir() {
			t.Errorf("%s at %s is not a directory", dir.name, path)
		}

		// Check permissions (0755)
		if info.Mode().Perm() != 0755 {
			t.Errorf("%s permissions = %o, want 0755", dir.name, info.Mode().Perm())
		}
	}
}

func TestEnsureDirectoriesIdempotent(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// Call EnsureDirectories multiple times - should not error
	for i := 0; i < 3; i++ {
		err := EnsureDirectories()
		if err != nil {
			t.Fatalf("EnsureDirectories() call %d error = %v", i+1, err)
		}
	}
}

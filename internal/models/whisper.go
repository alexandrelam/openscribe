package models

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alexandrelam/openscribe/internal/config"
)

const (
	WhisperCppRepo    = "https://github.com/ggerganov/whisper.cpp.git"
	WhisperCppVersion = "master" // Can pin to a specific version later
)

// GetWhisperCppDir returns the directory where whisper.cpp is stored
func GetWhisperCppDir() (string, error) {
	appSupport, err := config.GetAppSupportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appSupport, "whisper.cpp"), nil
}

// GetWhisperCppBinaryPath returns the path to the whisper.cpp main executable
func GetWhisperCppBinaryPath() (string, error) {
	whisperDir, err := GetWhisperCppDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(whisperDir, "main"), nil
}

// IsWhisperCppInstalled checks if whisper.cpp is installed and compiled
func IsWhisperCppInstalled() (bool, error) {
	binaryPath, err := GetWhisperCppBinaryPath()
	if err != nil {
		return false, err
	}

	// Check if binary exists and is executable
	info, err := os.Stat(binaryPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check whisper.cpp binary: %w", err)
	}

	// Check if it's executable
	if info.Mode()&0111 == 0 {
		return false, fmt.Errorf("whisper.cpp binary is not executable")
	}

	return true, nil
}

// DownloadWhisperCpp clones the whisper.cpp repository
func DownloadWhisperCpp() error {
	whisperDir, err := GetWhisperCppDir()
	if err != nil {
		return err
	}

	// Check if directory already exists
	if _, err := os.Stat(whisperDir); err == nil {
		return fmt.Errorf("whisper.cpp directory already exists at %s", whisperDir)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(whisperDir), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", WhisperCppVersion, WhisperCppRepo, whisperDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone whisper.cpp: %w", err)
	}

	return nil
}

// CompileWhisperCpp compiles whisper.cpp from source
func CompileWhisperCpp() error {
	whisperDir, err := GetWhisperCppDir()
	if err != nil {
		return err
	}

	// Check if directory exists
	if _, err := os.Stat(whisperDir); os.IsNotExist(err) {
		return fmt.Errorf("whisper.cpp directory not found, run download first")
	}

	// Determine make command and flags based on architecture
	makeCmd := "make"
	makeArgs := []string{}

	// On macOS with Apple Silicon, we might want to enable CoreML or Metal
	if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			// Apple Silicon - can use Metal acceleration
			// For now, we'll use the default build
			// In the future, could add: makeArgs = append(makeArgs, "WHISPER_COREML=1")
		}
	}

	// Run make
	cmd := exec.Command(makeCmd, makeArgs...)
	cmd.Dir = whisperDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compile whisper.cpp: %w", err)
	}

	// Verify the binary was created
	binaryPath, err := GetWhisperCppBinaryPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("compilation succeeded but binary not found at %s", binaryPath)
	}

	return nil
}

// SetupWhisperCpp downloads and compiles whisper.cpp if not already installed
func SetupWhisperCpp() error {
	// Check if already installed
	installed, err := IsWhisperCppInstalled()
	if err != nil {
		return err
	}

	if installed {
		return nil // Already installed
	}

	// Check if directory exists (partially installed)
	whisperDir, err := GetWhisperCppDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(whisperDir); err == nil {
		// Directory exists but binary doesn't, try to compile
		fmt.Println("Found whisper.cpp source, attempting to compile...")
		return CompileWhisperCpp()
	}

	// Need to download first
	if err := DownloadWhisperCpp(); err != nil {
		return err
	}

	// Then compile
	return CompileWhisperCpp()
}

// CheckDependencies checks if required system dependencies are installed
func CheckDependencies() error {
	// Check for git
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed (required for downloading whisper.cpp)")
	}

	// Check for make
	if _, err := exec.LookPath("make"); err != nil {
		return fmt.Errorf("make is not installed (required for compiling whisper.cpp)")
	}

	// Check for C++ compiler
	compilers := []string{"clang++", "g++", "c++"}
	found := false
	for _, compiler := range compilers {
		if _, err := exec.LookPath(compiler); err == nil {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("C++ compiler not found (clang++ or g++ required for compiling whisper.cpp)")
	}

	return nil
}

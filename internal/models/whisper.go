package models

import (
	"fmt"
	"os/exec"
)

// GetWhisperCppBinaryPath returns the path to the whisper-cli executable
func GetWhisperCppBinaryPath() (string, error) {
	path, err := exec.LookPath("whisper-cli")
	if err != nil {
		return "", fmt.Errorf("whisper-cli not found in PATH")
	}
	return path, nil
}

// IsWhisperCppInstalled checks if whisper-cli is installed (via Homebrew or otherwise)
func IsWhisperCppInstalled() (bool, error) {
	_, err := exec.LookPath("whisper-cli")
	if err != nil {
		return false, nil
	}
	return true, nil
}

// CheckHomebrew checks if Homebrew is installed
func CheckHomebrew() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew is not installed. Install it from: https://brew.sh")
	}
	return nil
}

// SetupWhisperCpp checks if whisper-cli is installed and provides guidance
func SetupWhisperCpp() error {
	installed, err := IsWhisperCppInstalled()
	if err != nil {
		return err
	}

	if installed {
		return nil // Already installed
	}

	// Not installed - provide helpful error message
	return fmt.Errorf("whisper-cpp is not installed")
}

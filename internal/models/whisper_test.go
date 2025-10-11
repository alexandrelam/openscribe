package models

import (
	"os/exec"
	"testing"
)

func TestIsWhisperCppInstalled(t *testing.T) {
	installed, err := IsWhisperCppInstalled()
	if err != nil {
		t.Errorf("IsWhisperCppInstalled() returned unexpected error: %v", err)
	}

	// Check the actual system state
	_, lookupErr := exec.LookPath("whisper-cli")
	expectedInstalled := (lookupErr == nil)

	if installed != expectedInstalled {
		t.Errorf("IsWhisperCppInstalled() = %v, but exec.LookPath returned %v", installed, lookupErr)
	}
}

func TestGetWhisperCppBinaryPath(t *testing.T) {
	path, err := GetWhisperCppBinaryPath()

	// Check if whisper-cli is actually installed
	_, lookupErr := exec.LookPath("whisper-cli")

	if lookupErr == nil {
		// whisper-cli is installed, should succeed
		if err != nil {
			t.Errorf("GetWhisperCppBinaryPath() failed but whisper-cli is installed: %v", err)
		}
		if path == "" {
			t.Error("GetWhisperCppBinaryPath() returned empty path when whisper-cli is installed")
		}

		// Verify the path is correct
		expectedPath, _ := exec.LookPath("whisper-cli")
		if path != expectedPath {
			t.Errorf("GetWhisperCppBinaryPath() = %v, expected %v", path, expectedPath)
		}
	} else {
		// whisper-cli is not installed, should fail
		if err == nil {
			t.Error("GetWhisperCppBinaryPath() succeeded but whisper-cli is not installed")
		}
		if path != "" {
			t.Errorf("GetWhisperCppBinaryPath() returned path %v but should be empty when not installed", path)
		}
	}
}

func TestCheckHomebrew(t *testing.T) {
	err := CheckHomebrew()

	// Check the actual system state
	_, lookupErr := exec.LookPath("brew")

	if lookupErr == nil {
		// brew is installed, should succeed
		if err != nil {
			t.Errorf("CheckHomebrew() failed but brew is installed: %v", err)
		}
	} else {
		// brew is not installed, should fail
		if err == nil {
			t.Error("CheckHomebrew() succeeded but brew is not installed")
		}
		// Check error message is helpful
		if err.Error() == "" {
			t.Error("CheckHomebrew() returned empty error message")
		}
	}
}

func TestSetupWhisperCpp(t *testing.T) {
	err := SetupWhisperCpp()

	// Check the actual system state
	installed, _ := IsWhisperCppInstalled()

	if installed {
		// whisper-cli is installed, should succeed
		if err != nil {
			t.Errorf("SetupWhisperCpp() failed but whisper-cli is installed: %v", err)
		}
	} else {
		// whisper-cli is not installed, should return error
		if err == nil {
			t.Error("SetupWhisperCpp() succeeded but whisper-cli is not installed")
		}
	}
}

func TestSetupWhisperCpp_Integration(t *testing.T) {
	// This test verifies the complete flow
	installed, err := IsWhisperCppInstalled()
	if err != nil {
		t.Fatalf("Failed to check installation status: %v", err)
	}

	if !installed {
		t.Skip("Skipping integration test: whisper-cli not installed")
	}

	// If installed, verify we can get the path
	path, err := GetWhisperCppBinaryPath()
	if err != nil {
		t.Errorf("Failed to get binary path when whisper-cli is installed: %v", err)
	}

	if path == "" {
		t.Error("Got empty path when whisper-cli is installed")
	}

	// Verify setup succeeds
	err = SetupWhisperCpp()
	if err != nil {
		t.Errorf("SetupWhisperCpp() failed when whisper-cli is installed: %v", err)
	}
}

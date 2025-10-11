package audio

import (
	"fmt"
	"os"
	"testing"
)

// Integration tests that require real hardware
// These tests are skipped in CI by default
// Run with: go test -tags=integration

func TestListMicrophones_Integration(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	devices, err := ListMicrophones()
	if err != nil {
		t.Fatalf("Failed to list microphones: %v", err)
	}

	if len(devices) == 0 {
		t.Skip("No microphones available on this system")
	}

	// Verify each device has required fields
	for i, device := range devices {
		if device.ID == "" {
			t.Errorf("Device %d has empty ID", i)
		}
		if device.Name == "" {
			t.Errorf("Device %d has empty Name", i)
		}
	}

	// At least one device should be marked as default
	hasDefault := false
	for _, device := range devices {
		if device.IsDefault {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Log("Warning: No device marked as default")
	}
}

func TestGetDefaultMicrophone_Integration(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	device, err := GetDefaultMicrophone()
	if err != nil {
		t.Fatalf("Failed to get default microphone: %v", err)
	}

	if device == nil {
		t.Fatal("GetDefaultMicrophone returned nil device")
	}

	if device.Name == "" {
		t.Error("Default device has empty name")
	}
}

func TestFindMicrophoneByName_Integration(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// First get a list of available devices
	devices, err := ListMicrophones()
	if err != nil {
		t.Fatalf("Failed to list microphones: %v", err)
	}

	if len(devices) == 0 {
		t.Skip("No microphones available on this system")
	}

	// Try to find the first device by name
	testDevice := devices[0]
	found, err := FindMicrophoneByName(testDevice.Name)
	if err != nil {
		t.Fatalf("Failed to find microphone by name '%s': %v", testDevice.Name, err)
	}

	if found.Name != testDevice.Name {
		t.Errorf("Expected device name '%s', got '%s'", testDevice.Name, found.Name)
	}

	// Try to find a non-existent device
	_, err = FindMicrophoneByName("NonExistentMicrophone12345")
	if err == nil {
		t.Error("Expected error when searching for non-existent microphone, got nil")
	}
}

func TestFindMicrophoneByNameOrIndex_Integration(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// First get a list of available devices
	devices, err := ListMicrophones()
	if err != nil {
		t.Fatalf("Failed to list microphones: %v", err)
	}

	if len(devices) == 0 {
		t.Skip("No microphones available on this system")
	}

	t.Run("FindByValidIndex", func(t *testing.T) {
		// Test finding by valid index (1-based)
		for i := 0; i < len(devices); i++ {
			indexStr := fmt.Sprintf("%d", i+1)
			found, err := FindMicrophoneByNameOrIndex(indexStr)
			if err != nil {
				t.Fatalf("Failed to find microphone by index '%s': %v", indexStr, err)
			}

			if found.Name != devices[i].Name {
				t.Errorf("Expected device name '%s', got '%s'", devices[i].Name, found.Name)
			}
		}
	})

	t.Run("FindByName", func(t *testing.T) {
		// Test finding by name
		testDevice := devices[0]
		found, err := FindMicrophoneByNameOrIndex(testDevice.Name)
		if err != nil {
			t.Fatalf("Failed to find microphone by name '%s': %v", testDevice.Name, err)
		}

		if found.Name != testDevice.Name {
			t.Errorf("Expected device name '%s', got '%s'", testDevice.Name, found.Name)
		}
	})

	t.Run("InvalidIndex_Zero", func(t *testing.T) {
		// Test with index 0 (should be out of range since we use 1-based indexing)
		_, err := FindMicrophoneByNameOrIndex("0")
		if err == nil {
			t.Error("Expected error when using index 0, got nil")
		}
	})

	t.Run("InvalidIndex_Negative", func(t *testing.T) {
		// Test with negative index
		_, err := FindMicrophoneByNameOrIndex("-1")
		if err == nil {
			t.Error("Expected error when using negative index, got nil")
		}
	})

	t.Run("InvalidIndex_TooLarge", func(t *testing.T) {
		// Test with index beyond available devices
		tooLargeIndex := fmt.Sprintf("%d", len(devices)+10)
		_, err := FindMicrophoneByNameOrIndex(tooLargeIndex)
		if err == nil {
			t.Error("Expected error when using too large index, got nil")
		}
	})

	t.Run("InvalidName", func(t *testing.T) {
		// Test with non-existent name
		_, err := FindMicrophoneByNameOrIndex("NonExistentMicrophone12345")
		if err == nil {
			t.Error("Expected error when searching for non-existent microphone, got nil")
		}
	})

	t.Run("NameThatLooksLikeNumber", func(t *testing.T) {
		// If a device name is literally a number that's out of range,
		// it should fail on index check first
		_, err := FindMicrophoneByNameOrIndex("999")
		if err == nil {
			t.Error("Expected error when using out of range number, got nil")
		}
	})
}

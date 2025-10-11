package audio

import (
	"testing"
)

func TestListMicrophones(t *testing.T) {
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

func TestGetDefaultMicrophone(t *testing.T) {
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

func TestFindMicrophoneByName(t *testing.T) {
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

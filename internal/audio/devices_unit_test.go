package audio

import (
	"fmt"
	"testing"
)

// Helper to create mock devices for testing
func createMockDevices() []Device {
	return []Device{
		{ID: "0", Name: "MacBook Pro Microphone", IsDefault: true},
		{ID: "1", Name: "External USB Mic", IsDefault: false},
		{ID: "2", Name: "Bluetooth Headset", IsDefault: false},
	}
}

// Helper to create mock DeviceInfo list
func createMockDeviceInfos() []DeviceInfo {
	return []DeviceInfo{
		NewMockDeviceInfo("MacBook Pro Microphone", true),
		NewMockDeviceInfo("External USB Mic", false),
		NewMockDeviceInfo("Bluetooth Headset", false),
	}
}

func TestListMicrophonesWithEnumerator(t *testing.T) {
	mockInfos := createMockDeviceInfos()
	mockEnum := CreateMockEnumerator(mockInfos, nil)

	devices, err := listMicrophonesWithEnumerator(mockEnum)
	if err != nil {
		t.Fatalf("Failed to list microphones: %v", err)
	}

	if len(devices) != 3 {
		t.Errorf("Expected 3 devices, got %d", len(devices))
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

	// Verify first device is default
	if !devices[0].IsDefault {
		t.Error("Expected first device to be default")
	}

	// Verify device names
	expectedNames := []string{"MacBook Pro Microphone", "External USB Mic", "Bluetooth Headset"}
	for i, expected := range expectedNames {
		if devices[i].Name != expected {
			t.Errorf("Device %d: expected name '%s', got '%s'", i, expected, devices[i].Name)
		}
	}
}

func TestListMicrophonesWithEnumerator_Error(t *testing.T) {
	mockEnum := CreateMockEnumerator(nil, fmt.Errorf("device enumeration failed"))

	_, err := listMicrophonesWithEnumerator(mockEnum)
	if err == nil {
		t.Error("Expected error when enumeration fails, got nil")
	}
}

func TestListMicrophonesWithEnumerator_NoDevices(t *testing.T) {
	mockEnum := CreateMockEnumerator([]DeviceInfo{}, nil)

	_, err := listMicrophonesWithEnumerator(mockEnum)
	if err == nil {
		t.Error("Expected error when no devices found, got nil")
	}
}

func TestGetDefaultMicrophoneFromList(t *testing.T) {
	devices := createMockDevices()

	device, err := getDefaultMicrophoneFromList(devices)
	if err != nil {
		t.Fatalf("Failed to get default microphone: %v", err)
	}

	if device == nil {
		t.Fatal("GetDefaultMicrophone returned nil device")
	}

	if device.Name != "MacBook Pro Microphone" {
		t.Errorf("Expected default device 'MacBook Pro Microphone', got '%s'", device.Name)
	}

	if !device.IsDefault {
		t.Error("Expected device to be marked as default")
	}
}

func TestGetDefaultMicrophoneFromList_NoDefault(t *testing.T) {
	// Create devices with no default
	devices := []Device{
		{ID: "0", Name: "Mic 1", IsDefault: false},
		{ID: "1", Name: "Mic 2", IsDefault: false},
	}

	device, err := getDefaultMicrophoneFromList(devices)
	if err != nil {
		t.Fatalf("Failed to get default microphone: %v", err)
	}

	// Should return first device when no default
	if device.Name != "Mic 1" {
		t.Errorf("Expected first device 'Mic 1', got '%s'", device.Name)
	}
}

func TestGetDefaultMicrophoneFromList_Empty(t *testing.T) {
	devices := []Device{}

	_, err := getDefaultMicrophoneFromList(devices)
	if err == nil {
		t.Error("Expected error when no devices available, got nil")
	}
}

func TestFindMicrophoneByNameInList(t *testing.T) {
	devices := createMockDevices()

	testCases := []struct {
		name          string
		searchName    string
		expectError   bool
		expectedName  string
	}{
		{"FindFirst", "MacBook Pro Microphone", false, "MacBook Pro Microphone"},
		{"FindSecond", "External USB Mic", false, "External USB Mic"},
		{"FindThird", "Bluetooth Headset", false, "Bluetooth Headset"},
		{"NotFound", "NonExistent Mic", true, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found, err := findMicrophoneByNameInList(devices, tc.searchName)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if found.Name != tc.expectedName {
					t.Errorf("Expected name '%s', got '%s'", tc.expectedName, found.Name)
				}
			}
		})
	}
}

func TestFindMicrophoneByNameOrIndexInList(t *testing.T) {
	devices := createMockDevices()

	t.Run("FindByValidIndex", func(t *testing.T) {
		testCases := []struct {
			index        string
			expectedName string
		}{
			{"1", "MacBook Pro Microphone"},
			{"2", "External USB Mic"},
			{"3", "Bluetooth Headset"},
		}

		for _, tc := range testCases {
			found, err := findMicrophoneByNameOrIndexInList(devices, tc.index)
			if err != nil {
				t.Fatalf("Failed to find microphone by index '%s': %v", tc.index, err)
			}

			if found.Name != tc.expectedName {
				t.Errorf("Expected device name '%s', got '%s'", tc.expectedName, found.Name)
			}
		}
	})

	t.Run("FindByName", func(t *testing.T) {
		found, err := findMicrophoneByNameOrIndexInList(devices, "External USB Mic")
		if err != nil {
			t.Fatalf("Failed to find microphone by name: %v", err)
		}

		if found.Name != "External USB Mic" {
			t.Errorf("Expected 'External USB Mic', got '%s'", found.Name)
		}
	})

	t.Run("InvalidIndex_Zero", func(t *testing.T) {
		_, err := findMicrophoneByNameOrIndexInList(devices, "0")
		if err == nil {
			t.Error("Expected error when using index 0, got nil")
		}
	})

	t.Run("InvalidIndex_Negative", func(t *testing.T) {
		_, err := findMicrophoneByNameOrIndexInList(devices, "-1")
		if err == nil {
			t.Error("Expected error when using negative index, got nil")
		}
	})

	t.Run("InvalidIndex_TooLarge", func(t *testing.T) {
		_, err := findMicrophoneByNameOrIndexInList(devices, "10")
		if err == nil {
			t.Error("Expected error when using too large index, got nil")
		}
	})

	t.Run("InvalidName", func(t *testing.T) {
		_, err := findMicrophoneByNameOrIndexInList(devices, "NonExistent Mic")
		if err == nil {
			t.Error("Expected error when searching for non-existent microphone, got nil")
		}
	})

	t.Run("OutOfRangeNumber", func(t *testing.T) {
		_, err := findMicrophoneByNameOrIndexInList(devices, "999")
		if err == nil {
			t.Error("Expected error when using out of range number, got nil")
		}
	})
}

func TestFindMicrophoneByNameOrIndexInList_EdgeCases(t *testing.T) {
	devices := createMockDevices()

	t.Run("EmptyString", func(t *testing.T) {
		_, err := findMicrophoneByNameOrIndexInList(devices, "")
		if err == nil {
			t.Error("Expected error when using empty string, got nil")
		}
	})

	t.Run("Whitespace", func(t *testing.T) {
		_, err := findMicrophoneByNameOrIndexInList(devices, "   ")
		if err == nil {
			t.Error("Expected error when using whitespace, got nil")
		}
	})
}

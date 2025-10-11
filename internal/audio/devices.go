package audio

import (
	"fmt"
	"strconv"

	"github.com/gen2brain/malgo"
)

// Device represents an audio input device
type Device struct {
	ID         string
	Name       string
	IsDefault  bool
	SampleRate uint32
	Channels   uint32
}

// ListMicrophones returns a list of all available audio input devices
func ListMicrophones() ([]Device, error) {
	ctx, err := NewMalgoContext()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio context: %w", err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	return listMicrophonesWithEnumerator(ctx)
}

// listMicrophonesWithEnumerator is an internal function that accepts a DeviceEnumerator
// This allows for dependency injection in tests
func listMicrophonesWithEnumerator(enumerator DeviceEnumerator) ([]Device, error) {
	// Get capture devices (microphones)
	infos, err := enumerator.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate devices: %w", err)
	}

	if len(infos) == 0 {
		return nil, fmt.Errorf("no audio input devices found")
	}

	devices := make([]Device, 0, len(infos))
	for i, info := range infos {
		device := Device{
			ID:        fmt.Sprintf("%d", i),
			Name:      info.Name(),
			IsDefault: info.IsDefault() == 1,
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// GetDefaultMicrophone returns the system's default audio input device
func GetDefaultMicrophone() (*Device, error) {
	devices, err := ListMicrophones()
	if err != nil {
		return nil, err
	}

	return getDefaultMicrophoneFromList(devices)
}

// getDefaultMicrophoneFromList is an internal helper for testing
func getDefaultMicrophoneFromList(devices []Device) (*Device, error) {
	for _, device := range devices {
		if device.IsDefault {
			return &device, nil
		}
	}

	// If no default found, return the first device
	if len(devices) > 0 {
		return &devices[0], nil
	}

	return nil, fmt.Errorf("no default microphone found")
}

// FindMicrophoneByName searches for a microphone by name
func FindMicrophoneByName(name string) (*Device, error) {
	devices, err := ListMicrophones()
	if err != nil {
		return nil, err
	}

	return findMicrophoneByNameInList(devices, name)
}

// findMicrophoneByNameInList is an internal helper for testing
func findMicrophoneByNameInList(devices []Device, name string) (*Device, error) {
	for _, device := range devices {
		if device.Name == name {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("microphone not found: %s", name)
}

// FindMicrophoneByNameOrIndex searches for a microphone by name or by index number (1-based)
// If the input is a valid number, it will try to find the device by index.
// Otherwise, it will search by name.
// Examples: "1", "2", "MacBook Pro Microphone"
func FindMicrophoneByNameOrIndex(nameOrIndex string) (*Device, error) {
	devices, err := ListMicrophones()
	if err != nil {
		return nil, err
	}

	return findMicrophoneByNameOrIndexInList(devices, nameOrIndex)
}

// findMicrophoneByNameOrIndexInList is an internal helper for testing
func findMicrophoneByNameOrIndexInList(devices []Device, nameOrIndex string) (*Device, error) {
	// Try to parse as integer first
	if idx, err := strconv.Atoi(nameOrIndex); err == nil {
		// Convert from 1-based to 0-based index
		idx = idx - 1

		if idx >= 0 && idx < len(devices) {
			return &devices[idx], nil
		}

		return nil, fmt.Errorf("microphone index %d is out of range (valid range: 1-%d)", idx+1, len(devices))
	}

	// Fall back to searching by name
	for _, device := range devices {
		if device.Name == nameOrIndex {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("microphone not found: %s", nameOrIndex)
}

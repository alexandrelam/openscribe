package audio

import (
	"fmt"

	"github.com/gen2brain/malgo"
)

// Device represents an audio input device
type Device struct {
	ID          string
	Name        string
	IsDefault   bool
	SampleRate  uint32
	Channels    uint32
}

// ListMicrophones returns a list of all available audio input devices
func ListMicrophones() ([]Device, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audio context: %w", err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	// Get capture devices (microphones)
	infos, err := ctx.Devices(malgo.Capture)
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
			IsDefault: info.IsDefault == 1,
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

	for _, device := range devices {
		if device.Name == name {
			return &device, nil
		}
	}

	return nil, fmt.Errorf("microphone not found: %s", name)
}

//go:build !darwin
// +build !darwin

package audio

// noopOutputMuter is a no-op implementation for unsupported platforms
type noopOutputMuter struct{}

// newPlatformOutputMuter creates a no-op output muter for unsupported platforms
func newPlatformOutputMuter() (OutputMuter, error) {
	return &noopOutputMuter{}, nil
}

// Mute does nothing on unsupported platforms
func (m *noopOutputMuter) Mute() error {
	return nil
}

// Restore does nothing on unsupported platforms
func (m *noopOutputMuter) Restore() error {
	return nil
}

// IsMuted always reports false on unsupported platforms
func (m *noopOutputMuter) IsMuted() bool {
	return false
}

// Close does nothing on unsupported platforms
func (m *noopOutputMuter) Close() error {
	return nil
}

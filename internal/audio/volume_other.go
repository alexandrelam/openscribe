//go:build !darwin
// +build !darwin

package audio

// Compile-time check that the stub muter satisfies the interface
var _ OutputMuter = (*noopOutputMuter)(nil)

// noopOutputMuter is a no-op implementation for unsupported platforms
type noopOutputMuter struct{}

// newPlatformOutputMuter creates a no-op output muter for unsupported platforms
func newPlatformOutputMuter() (OutputMuter, error) {
	return &noopOutputMuter{}, nil
}

// Prepare does nothing on unsupported platforms
func (m *noopOutputMuter) Prepare() {}

// SupportsVolumeControl always reports false on unsupported platforms
func (m *noopOutputMuter) SupportsVolumeControl() bool {
	return false
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

// Method always reports no mechanism on unsupported platforms
func (m *noopOutputMuter) Method() string {
	return ""
}

// Describe reports that muting is unavailable on unsupported platforms
func (m *noopOutputMuter) Describe() string {
	return "output muting is not supported on this platform"
}

// Close does nothing on unsupported platforms
func (m *noopOutputMuter) Close() error {
	return nil
}

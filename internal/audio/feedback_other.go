//go:build !darwin
// +build !darwin

package audio

import "fmt"

// noopFeedback is a no-op implementation for unsupported platforms
type noopFeedback struct{}

// newPlatformFeedback creates a no-op feedback instance for unsupported platforms
func newPlatformFeedback() (Feedback, error) {
	return &noopFeedback{}, fmt.Errorf("audio feedback is not supported on this platform")
}

// PlayStartSound does nothing on unsupported platforms
func (f *noopFeedback) PlayStartSound() error {
	return nil
}

// PlayStopSound does nothing on unsupported platforms
func (f *noopFeedback) PlayStopSound() error {
	return nil
}

// PlayCompleteSound does nothing on unsupported platforms
func (f *noopFeedback) PlayCompleteSound() error {
	return nil
}

// Close does nothing on unsupported platforms
func (f *noopFeedback) Close() error {
	return nil
}

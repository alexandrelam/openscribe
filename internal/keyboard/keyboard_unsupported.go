//go:build !darwin
// +build !darwin

package keyboard

import "fmt"

type unsupportedKeyboard struct{}

func newKeyboard() (Keyboard, error) {
	return nil, fmt.Errorf("keyboard simulation is only supported on macOS")
}

// CheckPermissions always returns an error on unsupported platforms
func (k *unsupportedKeyboard) CheckPermissions() error {
	return fmt.Errorf("keyboard simulation is only supported on macOS")
}

// TypeText always returns an error on unsupported platforms
func (k *unsupportedKeyboard) TypeText(text string) error {
	return fmt.Errorf("keyboard simulation is only supported on macOS")
}

// Close does nothing on unsupported platforms
func (k *unsupportedKeyboard) Close() error {
	return nil
}

// RequestPermissions does nothing on unsupported platforms
func RequestPermissions() {
	// No-op
}

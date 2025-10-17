// Package keyboard provides keyboard simulation functionality for pasting text
package keyboard

// Keyboard provides an interface for simulating keyboard input
type Keyboard interface {
	// PasteText pastes the given text at the current cursor position using clipboard
	PasteText(text string) error

	// CheckPermissions verifies that the necessary permissions are granted
	CheckPermissions() error

	// Close cleans up any resources
	Close() error
}

// New creates a new Keyboard instance for the current platform
func New() (Keyboard, error) {
	return newKeyboard()
}

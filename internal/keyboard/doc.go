// Package keyboard provides keyboard simulation for text input on macOS.
//
// This package handles:
//   - Direct text injection using macOS CGEvent APIs
//   - Unicode character support
//   - Accessibility permission checking
//   - Platform-specific implementations (macOS only)
//
// The keyboard simulation uses CGEventCreateKeyboardEvent and
// CGEventKeyboardSetUnicodeString to simulate typing text character-by-character
// at the current cursor position. This is NOT clipboard-based - it directly
// simulates keyboard input events.
//
// Requirements:
//   - macOS Accessibility permissions must be granted
//   - Application must be added to System Preferences > Security & Privacy > Accessibility
//
// Note: Text is typed with a small delay (2ms) between characters for reliability
// across different applications.
//
// Example usage:
//
//	// Initialize keyboard simulator
//	sim, err := keyboard.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sim.Close()
//
//	// Check permissions
//	if !sim.HasPermissions() {
//	    if err := sim.RequestPermissions(); err != nil {
//	        log.Fatal("Please grant Accessibility permissions")
//	    }
//	}
//
//	// Type text at cursor
//	if err := sim.TypeText("Hello, world!"); err != nil {
//	    log.Fatal(err)
//	}
package keyboard

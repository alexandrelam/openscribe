// Package hotkey provides global hotkey detection for macOS.
//
// This package handles:
//   - Global hotkey registration using macOS Carbon Event Manager
//   - Double-press detection with configurable time window
//   - Support for modifier keys (Option, Shift, Command, Control)
//   - Platform-specific implementations (macOS only)
//
// The hotkey listener runs in a separate goroutine and uses a callback
// mechanism to notify when the configured hotkey double-press is detected.
//
// Supported hotkeys:
//   - Left Option / Right Option
//   - Left Shift / Right Shift
//   - Left Command / Right Command
//   - Left Control / Right Control
//
// Requirements:
//   - macOS Accessibility permissions must be granted
//   - Only works on macOS (Darwin platform)
//
// Example usage:
//
//	// Create listener with callback
//	listener, err := hotkey.NewListener("Right Option", func() {
//	    fmt.Println("Hotkey detected!")
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer listener.Stop()
//
//	// Start listening
//	if err := listener.Start(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Wait for events
//	select {}
package hotkey

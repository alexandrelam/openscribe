//go:build !darwin
// +build !darwin

package hotkey

import "fmt"

// startEventMonitor is a stub for non-Darwin platforms
func (l *Listener) startEventMonitor() error {
	return fmt.Errorf("hotkey monitoring is not supported on this platform")
}

// stopEventMonitor is a stub for non-Darwin platforms
func (l *Listener) stopEventMonitor() {
	// No-op on unsupported platforms
}

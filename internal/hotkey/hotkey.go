package hotkey

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// KeyCode represents a keyboard key code
type KeyCode uint16

// Common macOS key codes (from Carbon framework)
const (
	KeyRightOption KeyCode = 0x3D
	KeyLeftOption  KeyCode = 0x3A
	KeyRightShift  KeyCode = 0x3C
	KeyLeftShift   KeyCode = 0x38
	KeyRightCmd    KeyCode = 0x36
	KeyLeftCmd     KeyCode = 0x37
	KeyRightCtrl   KeyCode = 0x3E
	KeyLeftCtrl    KeyCode = 0x3B
)

// KeyNameMap maps key names to their codes
var KeyNameMap = map[string]KeyCode{
	"Right Option": KeyRightOption,
	"Left Option":  KeyLeftOption,
	"Right Shift":  KeyRightShift,
	"Left Shift":   KeyLeftShift,
	"Right Cmd":    KeyRightCmd,
	"Left Cmd":     KeyLeftCmd,
	"Right Ctrl":   KeyRightCtrl,
	"Left Ctrl":    KeyLeftCtrl,
}

// KeyCodeToName maps key codes back to names
var KeyCodeToName = map[KeyCode]string{
	KeyRightOption: "Right Option",
	KeyLeftOption:  "Left Option",
	KeyRightShift:  "Right Shift",
	KeyLeftShift:   "Left Shift",
	KeyRightCmd:    "Right Cmd",
	KeyLeftCmd:     "Left Cmd",
	KeyRightCtrl:   "Right Ctrl",
	KeyLeftCtrl:    "Left Ctrl",
}

// Listener listens for global hotkey events
type Listener struct {
	keyCode          KeyCode
	doublePressDelay time.Duration
	callback         func()

	mu            sync.Mutex
	lastPressTime time.Time
	pressCount    int

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewListener creates a new hotkey listener
func NewListener(keyName string, callback func()) (*Listener, error) {
	keyCode, ok := KeyNameMap[keyName]
	if !ok {
		return nil, fmt.Errorf("unknown key name: %s", keyName)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Listener{
		keyCode:          keyCode,
		doublePressDelay: 500 * time.Millisecond, // 500ms window for double-press
		callback:         callback,
		ctx:              ctx,
		cancel:           cancel,
	}, nil
}

// Start begins listening for hotkey events
func (l *Listener) Start() error {
	l.wg.Add(1)
	go l.eventLoop()

	// Start the platform-specific event monitoring
	return l.startEventMonitor()
}

// Stop stops listening for hotkey events
func (l *Listener) Stop() {
	l.cancel()
	l.stopEventMonitor()
	l.wg.Wait()
}

// eventLoop processes key events and detects double-presses
func (l *Listener) eventLoop() {
	defer l.wg.Done()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			l.checkPressTimeout()
		}
	}
}

// handleKeyPress processes a key press event
func (l *Listener) handleKeyPress() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Check if this is within the double-press window
	if now.Sub(l.lastPressTime) <= l.doublePressDelay {
		l.pressCount++
		if l.pressCount >= 2 {
			// Double-press detected!
			l.pressCount = 0
			l.lastPressTime = time.Time{} // Reset

			// Call the callback in a goroutine to avoid blocking
			go l.callback()
		}
	} else {
		// First press or timeout, reset counter
		l.pressCount = 1
		l.lastPressTime = now
	}
}

// checkPressTimeout resets the press count if the timeout has elapsed
func (l *Listener) checkPressTimeout() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.lastPressTime.IsZero() && time.Since(l.lastPressTime) > l.doublePressDelay {
		l.pressCount = 0
		l.lastPressTime = time.Time{}
	}
}

// GetAvailableKeys returns a list of available key names
func GetAvailableKeys() []string {
	keys := make([]string, 0, len(KeyNameMap))
	for name := range KeyNameMap {
		keys = append(keys, name)
	}
	return keys
}

// ValidateKeyName checks if a key name is valid
func ValidateKeyName(keyName string) error {
	if _, ok := KeyNameMap[keyName]; !ok {
		return fmt.Errorf("invalid key name: %s", keyName)
	}
	return nil
}

// Package hotkey provides global hotkey monitoring and double-press detection.
package hotkey

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// KeyCode represents a keyboard key code or synthetic mouse button code
type KeyCode uint32

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
	// Mouse buttons (synthetic codes, mapped in platform-specific code)
	ButtonForward KeyCode = 0x10001 // Mouse Forward button (button 4)
	ButtonBack    KeyCode = 0x10002 // Mouse Back button (button 3)
)

// KeyNameMap maps key names to their codes
var KeyNameMap = map[string]KeyCode{
	"Right Option":   KeyRightOption,
	"Left Option":    KeyLeftOption,
	"Right Shift":    KeyRightShift,
	"Left Shift":     KeyLeftShift,
	"Right Command":  KeyRightCmd,
	"Left Command":   KeyLeftCmd,
	"Right Control":  KeyRightCtrl,
	"Left Control":   KeyLeftCtrl,
	"Forward Button": ButtonForward,
	"Back Button":    ButtonBack,
}

// KeyCodeToName maps key codes back to names
var KeyCodeToName = map[KeyCode]string{
	KeyRightOption: "Right Option",
	KeyLeftOption:  "Left Option",
	KeyRightShift:  "Right Shift",
	KeyLeftShift:   "Left Shift",
	KeyRightCmd:    "Right Command",
	KeyLeftCmd:     "Left Command",
	KeyRightCtrl:   "Right Control",
	KeyLeftCtrl:    "Left Control",
	ButtonForward:  "Forward Button",
	ButtonBack:     "Back Button",
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

// MultiListener manages multiple hotkey listeners
type MultiListener struct {
	listeners []*Listener
	callback  func()
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewMultiListener creates a new multi-trigger listener
func NewMultiListener(triggerNames []string, callback func()) (*MultiListener, error) {
	if len(triggerNames) == 0 {
		return nil, fmt.Errorf("at least one trigger name is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	ml := &MultiListener{
		listeners: make([]*Listener, 0, len(triggerNames)),
		callback:  callback,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Create a listener for each trigger
	for _, triggerName := range triggerNames {
		listener, err := NewListener(triggerName, callback)
		if err != nil {
			// Cleanup any listeners we've already created
			for _, l := range ml.listeners {
				l.Stop()
			}
			cancel()
			return nil, fmt.Errorf("failed to create listener for trigger %q: %w", triggerName, err)
		}
		ml.listeners = append(ml.listeners, listener)
	}

	return ml, nil
}

// Start begins listening for all configured triggers
func (ml *MultiListener) Start() error {
	// Start all listeners
	for i, listener := range ml.listeners {
		if err := listener.Start(); err != nil {
			// If any listener fails to start, stop all previously started listeners
			for j := 0; j < i; j++ {
				ml.listeners[j].Stop()
			}
			return fmt.Errorf("failed to start listener %d: %w", i, err)
		}
	}
	return nil
}

// Stop stops all trigger listeners
func (ml *MultiListener) Stop() {
	ml.cancel()
	for _, listener := range ml.listeners {
		listener.Stop()
	}
}

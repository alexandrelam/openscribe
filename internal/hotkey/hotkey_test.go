package hotkey

import (
	"sync"
	"testing"
	"time"
)

func TestKeyNameMap(t *testing.T) {
	tests := []struct {
		name     string
		keyName  string
		expected KeyCode
		exists   bool
	}{
		{"Right Option exists", "Right Option", KeyRightOption, true},
		{"Left Option exists", "Left Option", KeyLeftOption, true},
		{"Right Shift exists", "Right Shift", KeyRightShift, true},
		{"Left Shift exists", "Left Shift", KeyLeftShift, true},
		{"Right Cmd exists", "Right Cmd", KeyRightCmd, true},
		{"Left Cmd exists", "Left Cmd", KeyLeftCmd, true},
		{"Right Ctrl exists", "Right Ctrl", KeyRightCtrl, true},
		{"Left Ctrl exists", "Left Ctrl", KeyLeftCtrl, true},
		{"Invalid key doesn't exist", "Invalid Key", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyCode, exists := KeyNameMap[tt.keyName]
			if exists != tt.exists {
				t.Errorf("KeyNameMap[%q] existence = %v, want %v", tt.keyName, exists, tt.exists)
			}
			if exists && keyCode != tt.expected {
				t.Errorf("KeyNameMap[%q] = %v, want %v", tt.keyName, keyCode, tt.expected)
			}
		})
	}
}

func TestKeyCodeToName(t *testing.T) {
	tests := []struct {
		name     string
		keyCode  KeyCode
		expected string
		exists   bool
	}{
		{"Right Option code maps to name", KeyRightOption, "Right Option", true},
		{"Left Option code maps to name", KeyLeftOption, "Left Option", true},
		{"Right Shift code maps to name", KeyRightShift, "Right Shift", true},
		{"Left Shift code maps to name", KeyLeftShift, "Left Shift", true},
		{"Invalid code doesn't exist", KeyCode(0x99), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyName, exists := KeyCodeToName[tt.keyCode]
			if exists != tt.exists {
				t.Errorf("KeyCodeToName[%v] existence = %v, want %v", tt.keyCode, exists, tt.exists)
			}
			if exists && keyName != tt.expected {
				t.Errorf("KeyCodeToName[%v] = %q, want %q", tt.keyCode, keyName, tt.expected)
			}
		})
	}
}

func TestNewListener(t *testing.T) {
	tests := []struct {
		name      string
		keyName   string
		wantError bool
	}{
		{"Valid key name", "Right Option", false},
		{"Another valid key", "Left Cmd", false},
		{"Invalid key name", "Invalid Key", true},
		{"Empty key name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callback := func() {
				// Callback for listener creation
			}

			listener, err := NewListener(tt.keyName, callback)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewListener(%q) expected error, got nil", tt.keyName)
				}
				return
			}

			if err != nil {
				t.Errorf("NewListener(%q) unexpected error: %v", tt.keyName, err)
				return
			}

			if listener == nil {
				t.Errorf("NewListener(%q) returned nil listener", tt.keyName)
				return
			}

			// Verify listener properties
			expectedKeyCode := KeyNameMap[tt.keyName]
			if listener.keyCode != expectedKeyCode {
				t.Errorf("listener.keyCode = %v, want %v", listener.keyCode, expectedKeyCode)
			}

			if listener.doublePressDelay != 500*time.Millisecond {
				t.Errorf("listener.doublePressDelay = %v, want 500ms", listener.doublePressDelay)
			}

			if listener.callback == nil {
				t.Errorf("listener.callback is nil")
			}

			// Don't call Stop() since we didn't call Start()
			// Just cancel the context to clean up
			listener.cancel()
		})
	}
}

func TestHandleKeyPress_SinglePress(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	listener, err := NewListener("Right Option", callback)
	if err != nil {
		t.Fatalf("NewListener() error: %v", err)
	}
	defer listener.cancel()

	// Single press should not trigger callback
	listener.handleKeyPress()

	// Wait a bit to ensure callback isn't called
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callbackCount
	mu.Unlock()

	if count != 0 {
		t.Errorf("Single key press triggered callback, count = %d, want 0", count)
	}
}

func TestHandleKeyPress_DoublePress(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	listener, err := NewListener("Right Option", callback)
	if err != nil {
		t.Fatalf("NewListener() error: %v", err)
	}
	defer listener.cancel()

	// Double press within window should trigger callback
	listener.handleKeyPress()
	time.Sleep(100 * time.Millisecond) // Within 500ms window
	listener.handleKeyPress()

	// Wait for callback to execute
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callbackCount
	mu.Unlock()

	if count != 1 {
		t.Errorf("Double key press callback count = %d, want 1", count)
	}
}

func TestHandleKeyPress_TwoDoublePresses(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	listener, err := NewListener("Right Option", callback)
	if err != nil {
		t.Fatalf("NewListener() error: %v", err)
	}
	defer listener.cancel()

	// First double press
	listener.handleKeyPress()
	time.Sleep(100 * time.Millisecond)
	listener.handleKeyPress()

	// Wait for callback
	time.Sleep(50 * time.Millisecond)

	// Second double press
	listener.handleKeyPress()
	time.Sleep(100 * time.Millisecond)
	listener.handleKeyPress()

	// Wait for callback
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callbackCount
	mu.Unlock()

	if count != 2 {
		t.Errorf("Two double key presses callback count = %d, want 2", count)
	}
}

func TestHandleKeyPress_PressesOutsideWindow(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	listener, err := NewListener("Right Option", callback)
	if err != nil {
		t.Fatalf("NewListener() error: %v", err)
	}
	defer listener.cancel()

	// Two presses outside the 500ms window should not trigger callback
	listener.handleKeyPress()
	time.Sleep(600 * time.Millisecond) // Outside window
	listener.handleKeyPress()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callbackCount
	mu.Unlock()

	if count != 0 {
		t.Errorf("Key presses outside window triggered callback, count = %d, want 0", count)
	}
}

func TestHandleKeyPress_TriplePress(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	listener, err := NewListener("Right Option", callback)
	if err != nil {
		t.Fatalf("NewListener() error: %v", err)
	}
	defer listener.cancel()

	// Triple press within window should only trigger callback once (after second press)
	listener.handleKeyPress()
	time.Sleep(100 * time.Millisecond)
	listener.handleKeyPress()
	time.Sleep(100 * time.Millisecond)
	listener.handleKeyPress()

	// Wait for callbacks to execute
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callbackCount
	mu.Unlock()

	if count != 1 {
		t.Errorf("Triple key press callback count = %d, want 1", count)
	}
}

func TestCheckPressTimeout(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	listener, err := NewListener("Right Option", callback)
	if err != nil {
		t.Fatalf("NewListener() error: %v", err)
	}
	defer listener.cancel()

	// First press
	listener.handleKeyPress()

	// Verify press count is 1
	listener.mu.Lock()
	if listener.pressCount != 1 {
		t.Errorf("After first press, pressCount = %d, want 1", listener.pressCount)
	}
	listener.mu.Unlock()

	// Wait for timeout
	time.Sleep(600 * time.Millisecond)

	// Manually call checkPressTimeout
	listener.checkPressTimeout()

	// Verify press count is reset
	listener.mu.Lock()
	pressCount := listener.pressCount
	listener.mu.Unlock()

	if pressCount != 0 {
		t.Errorf("After timeout, pressCount = %d, want 0", pressCount)
	}

	// Verify callback was not called
	mu.Lock()
	count := callbackCount
	mu.Unlock()

	if count != 0 {
		t.Errorf("Timeout triggered callback, count = %d, want 0", count)
	}
}

func TestGetAvailableKeys(t *testing.T) {
	keys := GetAvailableKeys()

	if len(keys) == 0 {
		t.Error("GetAvailableKeys() returned empty slice")
	}

	// Should have exactly 8 keys
	if len(keys) != 8 {
		t.Errorf("GetAvailableKeys() returned %d keys, want 8", len(keys))
	}

	// Verify all keys are in KeyNameMap
	for _, key := range keys {
		if _, exists := KeyNameMap[key]; !exists {
			t.Errorf("GetAvailableKeys() returned key %q not in KeyNameMap", key)
		}
	}
}

func TestValidateKeyName(t *testing.T) {
	tests := []struct {
		name      string
		keyName   string
		wantError bool
	}{
		{"Valid key", "Right Option", false},
		{"Another valid key", "Left Ctrl", false},
		{"Invalid key", "Invalid Key", true},
		{"Empty key", "", true},
		{"Random string", "foobar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeyName(tt.keyName)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateKeyName(%q) expected error, got nil", tt.keyName)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateKeyName(%q) unexpected error: %v", tt.keyName, err)
				}
			}
		})
	}
}

func TestListenerConcurrency(t *testing.T) {
	// Test that multiple concurrent key presses are handled safely
	callbackCount := 0
	var mu sync.Mutex

	callback := func() {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	listener, err := NewListener("Right Option", callback)
	if err != nil {
		t.Fatalf("NewListener() error: %v", err)
	}
	defer listener.cancel()

	// Simulate concurrent key presses
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			listener.handleKeyPress()
			time.Sleep(50 * time.Millisecond)
			listener.handleKeyPress()
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Should have triggered callback multiple times without panicking
	mu.Lock()
	count := callbackCount
	mu.Unlock()

	// Just verify we didn't panic and callback was called
	if count == 0 {
		t.Error("Concurrent key presses didn't trigger any callbacks")
	}

	t.Logf("Concurrent test triggered %d callbacks", count)
}

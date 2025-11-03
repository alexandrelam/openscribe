//go:build darwin
// +build darwin

package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics -framework ApplicationServices

#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>
#import <ApplicationServices/ApplicationServices.h>

// Global variable to store the Go callback with keycode parameter
extern void goHotkeyCallback(uint32_t keyCode);

// Global variables for event handling
static CFMachPortRef gEventTap = NULL;
static CFRunLoopSourceRef gRunLoopSource = NULL;
static CFRunLoopRef gRunLoop = NULL;

// Array to store multiple target key codes (max 16 triggers)
#define MAX_TARGET_KEYS 16
static uint32_t gTargetKeyCodes[MAX_TARGET_KEYS];
static int gTargetKeyCount = 0;

// Check if a keycode is in the target list
static bool isTargetKeyCode(uint32_t keyCode) {
    for (int i = 0; i < gTargetKeyCount; i++) {
        if (gTargetKeyCodes[i] == keyCode) {
            return true;
        }
    }
    return false;
}

// Event tap callback for monitoring keyboard and mouse events
static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    // Handle tap disabled event
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        if (gEventTap != NULL) {
            CGEventTapEnable(gEventTap, true);
        }
        return event;
    }

    // Handle mouse button events
    if (type == kCGEventOtherMouseDown) {
        int64_t buttonNumber = CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);

        // Map physical mouse button numbers to synthetic key codes
        uint32_t syntheticKeyCode = 0;
        if (buttonNumber == 3) {
            syntheticKeyCode = 0x10002; // Back Button
        } else if (buttonNumber == 4) {
            syntheticKeyCode = 0x10001; // Forward Button
        }

        // If this is one of our target buttons, trigger the callback
        if (syntheticKeyCode != 0 && isTargetKeyCode(syntheticKeyCode)) {
            goHotkeyCallback(syntheticKeyCode);
        }
    }

    // Handle keyboard modifier events
    if (type == kCGEventFlagsChanged) {
        // Get the key code from the event
        int64_t keyCode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);

        // Get the current flags to determine if key was pressed or released
        CGEventFlags flags = CGEventGetFlags(event);

        // Check if this is one of our target keys and if it was just pressed
        // We detect a press by checking if any modifier flag is set
        // (when released, the flags will be clear)
        if (isTargetKeyCode((uint32_t)keyCode)) {
            // Determine if key was pressed by checking relevant modifier flags
            bool isPressed = false;

            // Map key codes to their corresponding modifier flags
            switch (keyCode) {
                case 0x3D: // Right Option
                case 0x3A: // Left Option
                    isPressed = (flags & kCGEventFlagMaskAlternate) != 0;
                    break;
                case 0x3C: // Right Shift
                case 0x38: // Left Shift
                    isPressed = (flags & kCGEventFlagMaskShift) != 0;
                    break;
                case 0x36: // Right Command
                case 0x37: // Left Command
                    isPressed = (flags & kCGEventFlagMaskCommand) != 0;
                    break;
                case 0x3E: // Right Control
                case 0x3B: // Left Control
                    isPressed = (flags & kCGEventFlagMaskControl) != 0;
                    break;
            }

            // Only trigger callback on key press (not release)
            if (isPressed) {
                goHotkeyCallback((uint32_t)keyCode);
            }
        }
    }

    // Pass through the event
    return event;
}

// Check if we have accessibility permissions
static int checkAccessibilityPermissions() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @NO};
    Boolean isTrusted = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    return isTrusted ? 1 : 0;
}

// Request accessibility permissions with prompt
static void requestAccessibilityPermissions() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

// Add a keycode to the list of monitored keys
static int addKeyCode(uint32_t keyCode) {
    // Check if already in the list
    if (isTargetKeyCode(keyCode)) {
        return 0; // Already added
    }

    // Check if we have space
    if (gTargetKeyCount >= MAX_TARGET_KEYS) {
        return -4; // Too many keys
    }

    // Add to the list
    gTargetKeyCodes[gTargetKeyCount] = keyCode;
    gTargetKeyCount++;

    return 0;
}

// Initialize the event tap (called once for all keys)
static int initializeEventTap() {
    // Check accessibility permissions first
    if (checkAccessibilityPermissions() == 0) {
        return -1; // No accessibility permissions
    }

    // Only create event tap if not already created
    if (gEventTap != NULL) {
        return 0; // Already initialized
    }

    // Create an event tap to monitor flags changed events (for modifier keys)
    // and mouse button events (for mouse triggers)
    CGEventMask eventMask = CGEventMaskBit(kCGEventFlagsChanged) |
                            CGEventMaskBit(kCGEventOtherMouseDown);

    gEventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        eventMask,
        eventTapCallback,
        NULL
    );

    if (gEventTap == NULL) {
        return -2; // Failed to create event tap
    }

    // Create a run loop source and add it to the current run loop
    gRunLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, gEventTap, 0);
    if (gRunLoopSource == NULL) {
        CFRelease(gEventTap);
        gEventTap = NULL;
        return -3; // Failed to create run loop source
    }

    return 0;
}

// Unregister all hotkeys and clean up
static void unregisterHotkeys() {
    if (gRunLoopSource != NULL && gRunLoop != NULL) {
        CFRunLoopRemoveSource(gRunLoop, gRunLoopSource, kCFRunLoopCommonModes);
        CFRelease(gRunLoopSource);
        gRunLoopSource = NULL;
    }

    if (gEventTap != NULL) {
        CGEventTapEnable(gEventTap, false);
        CFRelease(gEventTap);
        gEventTap = NULL;
    }

    // Clear the keycode list
    gTargetKeyCount = 0;
    for (int i = 0; i < MAX_TARGET_KEYS; i++) {
        gTargetKeyCodes[i] = 0;
    }
}

// Start the event loop in a separate thread
static void* runEventLoop(void* arg) {
    @autoreleasepool {
        gRunLoop = CFRunLoopGetCurrent();

        // Add the run loop source if we have one
        if (gRunLoopSource != NULL) {
            CFRunLoopAddSource(gRunLoop, gRunLoopSource, kCFRunLoopCommonModes);

            // Enable the event tap
            if (gEventTap != NULL) {
                CGEventTapEnable(gEventTap, true);
            }
        }

        CFRunLoopRun();
    }
    return NULL;
}

// Stop the event loop
static void stopEventLoop() {
    if (gRunLoop != NULL) {
        CFRunLoopStop(gRunLoop);
        gRunLoop = NULL;
    }
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"sync"
)

// Global map to store listeners by keycode for the C callback
var (
	listenerMap   = make(map[KeyCode]*Listener)
	listenerMutex sync.RWMutex
	eventLoopOnce sync.Once
)

//export goHotkeyCallback
func goHotkeyCallback(keyCode C.uint32_t) {
	listenerMutex.RLock()
	listener := listenerMap[KeyCode(keyCode)]
	listenerMutex.RUnlock()

	if listener != nil {
		listener.handleKeyPress()
	}
}

// startEventMonitor starts monitoring for hotkey events (macOS-specific)
func (l *Listener) startEventMonitor() error {
	// Initialize event tap once for all listeners
	var initErr error
	eventLoopOnce.Do(func() {
		// Lock the OS thread for Carbon/Cocoa APIs
		runtime.LockOSThread()

		// Initialize the event tap
		result := C.initializeEventTap()
		if result == -1 {
			// Request accessibility permissions
			C.requestAccessibilityPermissions()
			initErr = fmt.Errorf("accessibility permissions required: please grant permissions in System Preferences > Security & Privacy > Privacy > Accessibility, then restart OpenScribe")
			return
		} else if result == -2 {
			initErr = fmt.Errorf("failed to create event tap for hotkey monitoring")
			return
		} else if result == -3 {
			initErr = fmt.Errorf("failed to create run loop source for hotkey monitoring")
			return
		} else if result != 0 {
			initErr = fmt.Errorf("failed to initialize event tap (error code: %d)", result)
			return
		}

		// Start the event loop in a goroutine
		go func() {
			runtime.LockOSThread()
			C.runEventLoop(nil)
		}()
	})

	if initErr != nil {
		return initErr
	}

	// Add this keycode to the monitored list
	result := C.addKeyCode(C.uint32_t(l.keyCode))
	if result == -4 {
		return fmt.Errorf("too many triggers configured (maximum %d)", 16)
	} else if result != 0 {
		return fmt.Errorf("failed to add trigger (error code: %d)", result)
	}

	// Register this listener in the global map
	listenerMutex.Lock()
	listenerMap[l.keyCode] = l
	listenerMutex.Unlock()

	return nil
}

// stopEventMonitor stops monitoring for hotkey events (macOS-specific)
func (l *Listener) stopEventMonitor() {
	// Remove this listener from the map
	listenerMutex.Lock()
	delete(listenerMap, l.keyCode)
	isEmpty := len(listenerMap) == 0
	listenerMutex.Unlock()

	// If this was the last listener, clean up the event tap
	if isEmpty {
		C.stopEventLoop()
		C.unregisterHotkeys()
		runtime.UnlockOSThread()
	}
}

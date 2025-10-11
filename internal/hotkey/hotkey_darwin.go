//go:build darwin
// +build darwin

package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreGraphics -framework ApplicationServices

#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>
#import <ApplicationServices/ApplicationServices.h>

// Global variable to store the Go callback
extern void goHotkeyCallback(void);

// Global variables for event handling
static CFMachPortRef gEventTap = NULL;
static CFRunLoopSourceRef gRunLoopSource = NULL;
static CFRunLoopRef gRunLoop = NULL;
static uint16_t gTargetKeyCode = 0;

// Event tap callback for monitoring keyboard events
static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    // Handle tap disabled event
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        if (gEventTap != NULL) {
            CGEventTapEnable(gEventTap, true);
        }
        return event;
    }

    // Only process key down events for flags changed (modifier keys)
    if (type == kCGEventFlagsChanged) {
        // Get the key code from the event
        int64_t keyCode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);

        // Get the current flags to determine if key was pressed or released
        CGEventFlags flags = CGEventGetFlags(event);

        // Check if this is our target key and if it was just pressed
        // We detect a press by checking if any modifier flag is set
        // (when released, the flags will be clear)
        if (keyCode == gTargetKeyCode) {
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
                goHotkeyCallback();
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

// Register a hotkey with the system using CGEventTap
static int registerHotkey(uint16_t keyCode) {
    // Check accessibility permissions first
    if (checkAccessibilityPermissions() == 0) {
        return -1; // No accessibility permissions
    }

    // Store the target key code
    gTargetKeyCode = keyCode;

    // Create an event tap to monitor flags changed events (for modifier keys)
    CGEventMask eventMask = CGEventMaskBit(kCGEventFlagsChanged);

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

// Unregister the hotkey
static void unregisterHotkey() {
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

    gTargetKeyCode = 0;
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
)

// Global reference to the current listener for the C callback
var currentListener *Listener

//export goHotkeyCallback
func goHotkeyCallback() {
	if currentListener != nil {
		currentListener.handleKeyPress()
	}
}

// startEventMonitor starts monitoring for hotkey events (macOS-specific)
func (l *Listener) startEventMonitor() error {
	// Lock the OS thread for Carbon/Cocoa APIs
	runtime.LockOSThread()

	// Set the current listener so the C callback can find it
	currentListener = l

	// Register the hotkey
	result := C.registerHotkey(C.uint16_t(l.keyCode))
	if result == -1 {
		// Request accessibility permissions
		C.requestAccessibilityPermissions()
		return fmt.Errorf("accessibility permissions required: please grant permissions in System Preferences > Security & Privacy > Privacy > Accessibility, then restart OpenScribe")
	} else if result == -2 {
		return fmt.Errorf("failed to create event tap for hotkey monitoring")
	} else if result == -3 {
		return fmt.Errorf("failed to create run loop source for hotkey monitoring")
	} else if result != 0 {
		return fmt.Errorf("failed to register hotkey (error code: %d)", result)
	}

	// Start the event loop in a goroutine
	go func() {
		runtime.LockOSThread()
		C.runEventLoop(nil)
	}()

	return nil
}

// stopEventMonitor stops monitoring for hotkey events (macOS-specific)
func (l *Listener) stopEventMonitor() {
	C.stopEventLoop()
	C.unregisterHotkey()
	currentListener = nil
	runtime.UnlockOSThread()
}

//go:build darwin
// +build darwin

package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon

#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>

// Global variable to store the Go callback
extern void goHotkeyCallback(void);

// Global variables for event handling
static EventHotKeyRef gHotKeyRef = NULL;
static EventHandlerRef gHandlerRef = NULL;
static CFRunLoopRef gRunLoop = NULL;

// Event handler for key events
static OSStatus keyEventHandler(EventHandlerCallRef nextHandler, EventRef theEvent, void* userData) {
    EventHotKeyID hotKeyID;
    GetEventParameter(theEvent, kEventParamDirectObject, typeEventHotKeyID, NULL, sizeof(hotKeyID), NULL, &hotKeyID);

    // Call the Go callback
    goHotkeyCallback();

    return noErr;
}

// Register a hotkey with the system
static int registerHotkey(UInt32 keyCode) {
    EventTypeSpec eventType;
    eventType.eventClass = kEventClassKeyboard;
    eventType.eventKind = kEventHotKeyPressed;

    // Install the event handler
    OSStatus status = InstallEventHandler(GetApplicationEventTarget(),
                                         &keyEventHandler,
                                         1,
                                         &eventType,
                                         NULL,
                                         &gHandlerRef);
    if (status != noErr) {
        return -1;
    }

    // Register the hotkey (no modifiers, just the key itself)
    EventHotKeyID hotKeyID;
    hotKeyID.signature = 'htk1';
    hotKeyID.id = 1;

    status = RegisterEventHotKey(keyCode,
                                0,  // No modifiers
                                hotKeyID,
                                GetApplicationEventTarget(),
                                0,
                                &gHotKeyRef);
    if (status != noErr) {
        RemoveEventHandler(gHandlerRef);
        gHandlerRef = NULL;
        return -2;
    }

    return 0;
}

// Unregister the hotkey
static void unregisterHotkey() {
    if (gHotKeyRef != NULL) {
        UnregisterEventHotKey(gHotKeyRef);
        gHotKeyRef = NULL;
    }
    if (gHandlerRef != NULL) {
        RemoveEventHandler(gHandlerRef);
        gHandlerRef = NULL;
    }
}

// Start the event loop in a separate thread
static void* runEventLoop(void* arg) {
    @autoreleasepool {
        gRunLoop = CFRunLoopGetCurrent();
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
	result := C.registerHotkey(C.UInt32(l.keyCode))
	if result != 0 {
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

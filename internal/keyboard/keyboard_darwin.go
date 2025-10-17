//go:build darwin
// +build darwin

package keyboard

/*
#cgo CFLAGS: -x objective-c -fmodules -fblocks
#cgo LDFLAGS: -framework CoreGraphics -framework ApplicationServices -framework AppKit

#import <CoreGraphics/CoreGraphics.h>
#import <ApplicationServices/ApplicationServices.h>
#import <AppKit/AppKit.h>

// Check if we have accessibility permissions
static int checkAccessibilityPermissions() {
    // Check if the process is trusted to use accessibility features
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @NO};
    Boolean isTrusted = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    return isTrusted ? 1 : 0;
}

// Request accessibility permissions with prompt
static void requestAccessibilityPermissions() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

// Get current clipboard contents
static char* getClipboardContents() {
    @autoreleasepool {
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        NSString *contents = [pasteboard stringForType:NSPasteboardTypeString];

        if (contents == nil) {
            return NULL;
        }

        // Convert to C string - caller must free
        const char *cStr = [contents UTF8String];
        return strdup(cStr);
    }
}

// Set clipboard contents
static void setClipboardContents(const char *text) {
    @autoreleasepool {
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        NSString *nsText = [NSString stringWithUTF8String:text];
        [pasteboard setString:nsText forType:NSPasteboardTypeString];
    }
}

// Simulate Command+V keypress
static void simulateCommandV() {
    // Key code for 'v' is 9
    CGEventRef keyDownCmd = CGEventCreateKeyboardEvent(NULL, 55, true);  // Command key down (55 = left command)
    CGEventRef keyDownV = CGEventCreateKeyboardEvent(NULL, 9, true);     // V key down
    CGEventRef keyUpV = CGEventCreateKeyboardEvent(NULL, 9, false);      // V key up
    CGEventRef keyUpCmd = CGEventCreateKeyboardEvent(NULL, 55, false);   // Command key up

    // Set Command modifier flag on the V key events
    CGEventSetFlags(keyDownV, kCGEventFlagMaskCommand);
    CGEventSetFlags(keyUpV, kCGEventFlagMaskCommand);

    // Post events in sequence
    CGEventPost(kCGHIDEventTap, keyDownCmd);
    CGEventPost(kCGHIDEventTap, keyDownV);
    CGEventPost(kCGHIDEventTap, keyUpV);
    CGEventPost(kCGHIDEventTap, keyUpCmd);

    // Release events
    if (keyDownCmd) CFRelease(keyDownCmd);
    if (keyDownV) CFRelease(keyDownV);
    if (keyUpV) CFRelease(keyUpV);
    if (keyUpCmd) CFRelease(keyUpCmd);
}
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

type macKeyboard struct{}

func newKeyboard() (Keyboard, error) {
	return &macKeyboard{}, nil
}

// CheckPermissions verifies that accessibility permissions are granted
func (k *macKeyboard) CheckPermissions() error {
	if C.checkAccessibilityPermissions() == 0 {
		return fmt.Errorf("accessibility permissions not granted")
	}
	return nil
}

// RequestPermissions prompts the user to grant accessibility permissions
func RequestPermissions() {
	C.requestAccessibilityPermissions()
}

// PasteText pastes the given text at the current cursor position using clipboard and Command+V
// It preserves the original clipboard contents by saving and restoring them
func (k *macKeyboard) PasteText(text string) error {
	// Check permissions first
	if err := k.CheckPermissions(); err != nil {
		return fmt.Errorf("cannot paste text: %w", err)
	}

	// Save current clipboard contents
	originalClipboard := C.getClipboardContents()
	var savedClipboard string
	if originalClipboard != nil {
		savedClipboard = C.GoString(originalClipboard)
		C.free(unsafe.Pointer(originalClipboard))
	}

	// Set clipboard to our text
	cText := C.CString(text)
	C.setClipboardContents(cText)
	C.free(unsafe.Pointer(cText))

	// Small delay to ensure clipboard is set
	time.Sleep(10 * time.Millisecond)

	// Simulate Command+V
	C.simulateCommandV()

	// Small delay to ensure paste completes
	time.Sleep(50 * time.Millisecond)

	// Restore original clipboard contents
	if savedClipboard != "" {
		cSaved := C.CString(savedClipboard)
		C.setClipboardContents(cSaved)
		C.free(unsafe.Pointer(cSaved))
	}

	return nil
}

// Close cleans up any resources (nothing needed for CGEvent/NSPasteboard)
func (k *macKeyboard) Close() error {
	return nil
}

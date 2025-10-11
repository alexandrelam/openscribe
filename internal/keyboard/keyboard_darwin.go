// +build darwin

package keyboard

/*
#cgo CFLAGS: -x objective-c -fmodules -fblocks
#cgo LDFLAGS: -framework CoreGraphics -framework ApplicationServices

#import <CoreGraphics/CoreGraphics.h>
#import <ApplicationServices/ApplicationServices.h>

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

// Type a single Unicode character at the cursor position
static void typeUnicodeChar(UniChar ch) {
    CGEventRef keyDown = CGEventCreateKeyboardEvent(NULL, 0, true);
    CGEventRef keyUp = CGEventCreateKeyboardEvent(NULL, 0, false);

    if (keyDown && keyUp) {
        CGEventKeyboardSetUnicodeString(keyDown, 1, &ch);
        CGEventKeyboardSetUnicodeString(keyUp, 1, &ch);

        CGEventPost(kCGHIDEventTap, keyDown);
        CGEventPost(kCGHIDEventTap, keyUp);

        CFRelease(keyDown);
        CFRelease(keyUp);
    }
}
*/
import "C"
import (
	"fmt"
	"time"
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

// TypeText simulates typing the given text at the current cursor position
func (k *macKeyboard) TypeText(text string) error {
	// Check permissions first
	if err := k.CheckPermissions(); err != nil {
		return fmt.Errorf("cannot type text: %w", err)
	}

	// Type each character with a small delay to ensure reliability
	for _, r := range text {
		C.typeUnicodeChar(C.UniChar(r))
		// Small delay between characters (2ms) for reliability
		time.Sleep(2 * time.Millisecond)
	}

	return nil
}

// Close cleans up any resources (nothing needed for CGEvent)
func (k *macKeyboard) Close() error {
	return nil
}

//go:build darwin
// +build darwin

package audio

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// Play a system sound by name
static void playSystemSound(const char* soundName) {
    @autoreleasepool {
        NSSound *sound = [NSSound soundNamed:[NSString stringWithUTF8String:soundName]];
        if (sound != nil) {
            [sound play];
        }
    }
}

// Play a sound file from disk
static int playSoundFile(const char* filePath) {
    @autoreleasepool {
        NSString *path = [NSString stringWithUTF8String:filePath];
        NSSound *sound = [[NSSound alloc] initWithContentsOfFile:path byReference:NO];
        if (sound != nil) {
            [sound play];
            return 0;
        }
        return -1;
    }
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// darwinFeedback implements audio feedback using macOS NSSound
type darwinFeedback struct {
	enabled bool
}

// newPlatformFeedback creates a new macOS audio feedback instance
func newPlatformFeedback() (Feedback, error) {
	return &darwinFeedback{
		enabled: true,
	}, nil
}

// PlayStartSound plays the sound when recording starts
// Uses "Tink" system sound (a short, ascending beep)
func (f *darwinFeedback) PlayStartSound() error {
	if !f.enabled {
		return nil
	}

	soundName := C.CString("Tink")
	defer C.free(unsafe.Pointer(soundName))

	C.playSystemSound(soundName)
	return nil
}

// PlayStopSound plays the sound when recording stops
// Uses "Pop" system sound (a short, neutral beep)
func (f *darwinFeedback) PlayStopSound() error {
	if !f.enabled {
		return nil
	}

	soundName := C.CString("Pop")
	defer C.free(unsafe.Pointer(soundName))

	C.playSystemSound(soundName)
	return nil
}

// PlayCompleteSound plays the sound when transcription completes
// Uses "Glass" system sound (a pleasant "ding" sound)
func (f *darwinFeedback) PlayCompleteSound() error {
	if !f.enabled {
		return nil
	}

	soundName := C.CString("Glass")
	defer C.free(unsafe.Pointer(soundName))

	C.playSystemSound(soundName)
	return nil
}

// Close releases any resources
func (f *darwinFeedback) Close() error {
	return nil
}

// Disable turns off audio feedback
func (f *darwinFeedback) Disable() {
	f.enabled = false
}

// Enable turns on audio feedback
func (f *darwinFeedback) Enable() {
	f.enabled = true
}

// playSound is a helper function to play a specific system sound
func playSound(soundName string) error {
	cName := C.CString(soundName)
	defer C.free(unsafe.Pointer(cName))

	C.playSystemSound(cName)
	return nil
}

// ListSystemSounds returns a list of available macOS system sounds
// This is useful for testing and configuration
func ListSystemSounds() []string {
	return []string{
		"Basso",     // Deep boom
		"Blow",      // Whoosh
		"Bottle",    // Pop
		"Frog",      // Ribbit
		"Funk",      // Funky beat
		"Glass",     // Pleasant ding (used for complete)
		"Hero",      // Triumphant
		"Morse",     // Beep beep
		"Ping",      // Network ping
		"Pop",       // Short pop (used for stop)
		"Purr",      // Soft purr
		"Sosumi",    // Classic Mac sound
		"Submarine", // Sonar ping
		"Tink",      // Short ascending beep (used for start)
	}
}

// PlaySoundByName plays a specific system sound by name
// Useful for testing different sounds
func PlaySoundByName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.playSystemSound(cName)
	return nil
}

// TestAllSounds plays all available system sounds for testing
func TestAllSounds() error {
	sounds := ListSystemSounds()
	for _, sound := range sounds {
		fmt.Printf("Playing sound: %s\n", sound)
		if err := PlaySoundByName(sound); err != nil {
			return fmt.Errorf("failed to play sound %s: %w", sound, err)
		}
		// Small delay between sounds (handled by caller if needed)
	}
	return nil
}

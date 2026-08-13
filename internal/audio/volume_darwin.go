//go:build darwin
// +build darwin

package audio

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreAudio -framework CoreFoundation

#import <CoreAudio/CoreAudio.h>
#import <CoreFoundation/CoreFoundation.h>

// The "main" element of a property is element 0. Spelling it out avoids having
// to pick between kAudioObjectPropertyElementMain (macOS 12+) and the
// deprecated kAudioObjectPropertyElementMaster, which differ only in name.
#define OS_ELEMENT_MAIN ((AudioObjectPropertyElement)0)

// The virtual main volume is the control the volume keys and the menu bar
// slider drive. Devices that expose no per-device volume (AirPlay, many
// Bluetooth and display outputs) are still adjustable through it. Declared
// here rather than pulled from <AudioToolbox/AudioServices.h> so the code does
// not have to straddle the VirtualMaster/VirtualMain SDK rename.
#define OS_VIRTUAL_MAIN_VOLUME ((AudioObjectPropertySelector)'vmvc')

// Return codes for the helpers below
#define OS_MUTE_OK             0
#define OS_MUTE_NO_DEVICE     -1
#define OS_MUTE_UNSUPPORTED   -2
#define OS_MUTE_SET_FAILED    -3
#define OS_MUTE_READ_FAILED   -4

// How the device was silenced, so Restore can undo the matching thing
#define OS_METHOD_NONE     0
#define OS_METHOD_MUTE     1
#define OS_METHOD_VIRTUAL  2
#define OS_METHOD_VOLUME   3

#define OS_MAX_VOLUME_CHANNELS 2

// MuteState records everything needed to put the output device back exactly
// the way it was found.
typedef struct {
    AudioDeviceID deviceID;
    int           method;
    UInt32        prevMute;
    Float32       prevVirtualVolume;
    Float32       prevVolume[OS_MAX_VOLUME_CHANNELS];
    UInt32        volumeChannels[OS_MAX_VOLUME_CHANNELS];
    int           volumeChannelCount;
    OSStatus      lastStatus;
} MuteState;

static OSStatus defaultOutputDevice(AudioDeviceID *outID) {
    AudioObjectPropertyAddress addr = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        OS_ELEMENT_MAIN
    };
    UInt32 size = sizeof(AudioDeviceID);
    return AudioObjectGetPropertyData(kAudioObjectSystemObject, &addr, 0, NULL, &size, outID);
}

static int propertySettable(AudioDeviceID dev, AudioObjectPropertyAddress *addr) {
    Boolean settable = 0;
    if (!AudioObjectHasProperty(dev, addr)) {
        return 0;
    }
    if (AudioObjectIsPropertySettable(dev, addr, &settable) != noErr) {
        return 0;
    }
    return settable ? 1 : 0;
}

// Copy the name of the current default output device into buf. Returns 0 on
// success. Used purely for diagnostics.
static int openscribeOutputDeviceName(char *buf, int bufLen) {
    AudioDeviceID dev = kAudioObjectUnknown;
    if (defaultOutputDevice(&dev) != noErr || dev == kAudioObjectUnknown) {
        return OS_MUTE_NO_DEVICE;
    }

    AudioObjectPropertyAddress addr = {
        kAudioObjectPropertyName,
        kAudioObjectPropertyScopeGlobal,
        OS_ELEMENT_MAIN
    };

    CFStringRef name = NULL;
    UInt32 size = sizeof(name);
    if (AudioObjectGetPropertyData(dev, &addr, 0, NULL, &size, &name) != noErr || name == NULL) {
        return OS_MUTE_READ_FAILED;
    }

    Boolean ok = CFStringGetCString(name, buf, bufLen, kCFStringEncodingUTF8);
    CFRelease(name);
    return ok ? OS_MUTE_OK : OS_MUTE_READ_FAILED;
}

// Mute the current default output device, recording its previous state.
//
// Three mechanisms are tried in order of how faithfully they can be undone:
// the device's own mute switch, the system-wide virtual main volume, and
// finally the per-device volume scalar.
static int openscribeMuteOutput(MuteState *state) {
    AudioDeviceID dev = kAudioObjectUnknown;
    OSStatus status = defaultOutputDevice(&dev);
    if (status != noErr || dev == kAudioObjectUnknown) {
        state->lastStatus = status;
        return OS_MUTE_NO_DEVICE;
    }

    state->deviceID = dev;
    state->method = OS_METHOD_NONE;
    state->volumeChannelCount = 0;
    state->lastStatus = noErr;

    // 1. The device exposes a settable mute switch.
    AudioObjectPropertyAddress muteAddr = {
        kAudioDevicePropertyMute,
        kAudioDevicePropertyScopeOutput,
        OS_ELEMENT_MAIN
    };

    if (propertySettable(dev, &muteAddr)) {
        UInt32 size = sizeof(UInt32);
        status = AudioObjectGetPropertyData(dev, &muteAddr, 0, NULL, &size, &state->prevMute);
        if (status == noErr) {
            UInt32 on = 1;
            status = AudioObjectSetPropertyData(dev, &muteAddr, 0, NULL, sizeof(on), &on);
            if (status != noErr) {
                state->lastStatus = status;
                return OS_MUTE_SET_FAILED;
            }
            state->method = OS_METHOD_MUTE;
            return OS_MUTE_OK;
        }
        state->lastStatus = status;
        // Fall through and try the volume controls instead
    }

    // 2. The system-wide virtual main volume, i.e. what the volume keys drive.
    AudioObjectPropertyAddress virtualAddr = {
        OS_VIRTUAL_MAIN_VOLUME,
        kAudioDevicePropertyScopeOutput,
        OS_ELEMENT_MAIN
    };

    if (propertySettable(dev, &virtualAddr)) {
        UInt32 size = sizeof(Float32);
        status = AudioObjectGetPropertyData(dev, &virtualAddr, 0, NULL, &size,
                                            &state->prevVirtualVolume);
        if (status == noErr) {
            Float32 silent = 0.0f;
            status = AudioObjectSetPropertyData(dev, &virtualAddr, 0, NULL,
                                                sizeof(silent), &silent);
            if (status != noErr) {
                state->lastStatus = status;
                return OS_MUTE_SET_FAILED;
            }
            state->method = OS_METHOD_VIRTUAL;
            return OS_MUTE_OK;
        }
        state->lastStatus = status;
    }

    // 3. The per-device volume scalar: the main element first, then the
    // preferred stereo channel pair (many devices only expose per-channel).
    AudioObjectPropertyAddress volAddr = {
        kAudioDevicePropertyVolumeScalar,
        kAudioDevicePropertyScopeOutput,
        OS_ELEMENT_MAIN
    };

    if (propertySettable(dev, &volAddr)) {
        state->volumeChannels[0] = OS_ELEMENT_MAIN;
        state->volumeChannelCount = 1;
    } else {
        AudioObjectPropertyAddress chAddr = {
            kAudioDevicePropertyPreferredChannelsForStereo,
            kAudioDevicePropertyScopeOutput,
            OS_ELEMENT_MAIN
        };
        UInt32 channels[OS_MAX_VOLUME_CHANNELS] = {1, 2};
        UInt32 size = sizeof(channels);
        if (AudioObjectHasProperty(dev, &chAddr)) {
            status = AudioObjectGetPropertyData(dev, &chAddr, 0, NULL, &size, channels);
            if (status != noErr) {
                state->lastStatus = status;
                return OS_MUTE_UNSUPPORTED;
            }
        }

        for (int i = 0; i < OS_MAX_VOLUME_CHANNELS; i++) {
            volAddr.mElement = channels[i];
            if (!propertySettable(dev, &volAddr)) {
                continue;
            }
            state->volumeChannels[state->volumeChannelCount] = channels[i];
            state->volumeChannelCount++;
        }

        if (state->volumeChannelCount == 0) {
            return OS_MUTE_UNSUPPORTED;
        }
    }

    // Read the current level for every channel before touching anything, so a
    // partial failure never leaves us without a way back.
    for (int i = 0; i < state->volumeChannelCount; i++) {
        volAddr.mElement = state->volumeChannels[i];
        UInt32 size = sizeof(Float32);
        status = AudioObjectGetPropertyData(dev, &volAddr, 0, NULL, &size, &state->prevVolume[i]);
        if (status != noErr) {
            state->lastStatus = status;
            state->volumeChannelCount = 0;
            return OS_MUTE_READ_FAILED;
        }
    }

    state->method = OS_METHOD_VOLUME;

    for (int i = 0; i < state->volumeChannelCount; i++) {
        volAddr.mElement = state->volumeChannels[i];
        Float32 silent = 0.0f;
        status = AudioObjectSetPropertyData(dev, &volAddr, 0, NULL, sizeof(silent), &silent);
        if (status != noErr) {
            state->lastStatus = status;
            return OS_MUTE_SET_FAILED;
        }
    }

    return OS_MUTE_OK;
}

// Restore the device recorded in state to its previous mute/volume level.
static int openscribeRestoreOutput(MuteState *state) {
    OSStatus status = noErr;
    state->lastStatus = noErr;

    if (state->method == OS_METHOD_MUTE) {
        AudioObjectPropertyAddress muteAddr = {
            kAudioDevicePropertyMute,
            kAudioDevicePropertyScopeOutput,
            OS_ELEMENT_MAIN
        };
        UInt32 previous = state->prevMute;
        status = AudioObjectSetPropertyData(state->deviceID, &muteAddr, 0, NULL,
                                            sizeof(previous), &previous);
    } else if (state->method == OS_METHOD_VIRTUAL) {
        AudioObjectPropertyAddress virtualAddr = {
            OS_VIRTUAL_MAIN_VOLUME,
            kAudioDevicePropertyScopeOutput,
            OS_ELEMENT_MAIN
        };
        Float32 previous = state->prevVirtualVolume;
        status = AudioObjectSetPropertyData(state->deviceID, &virtualAddr, 0, NULL,
                                            sizeof(previous), &previous);
    } else if (state->method == OS_METHOD_VOLUME) {
        AudioObjectPropertyAddress volAddr = {
            kAudioDevicePropertyVolumeScalar,
            kAudioDevicePropertyScopeOutput,
            OS_ELEMENT_MAIN
        };
        for (int i = 0; i < state->volumeChannelCount; i++) {
            volAddr.mElement = state->volumeChannels[i];
            Float32 previous = state->prevVolume[i];
            OSStatus one = AudioObjectSetPropertyData(state->deviceID, &volAddr, 0, NULL,
                                                      sizeof(previous), &previous);
            if (one != noErr) {
                status = one;
            }
        }
    }

    state->method = OS_METHOD_NONE;

    if (status != noErr) {
        state->lastStatus = status;
        return OS_MUTE_SET_FAILED;
    }
    return OS_MUTE_OK;
}
*/
import "C"
import (
	"fmt"
	"sync"
)

// Return codes from the C helpers above. Kept in sync with the OS_MUTE_*
// defines in the preamble.
const (
	muteOK          = 0
	muteNoDevice    = -1
	muteUnsupported = -2
	muteSetFailed   = -3
	muteReadFailed  = -4
)

// Mechanisms the C code can use, matching the OS_METHOD_* defines
const (
	methodNone    = 0
	methodMute    = 1
	methodVirtual = 2
	methodVolume  = 3
)

// Compile-time check that the macOS muter satisfies the interface. Without it
// a missing method only shows up at the call site in newPlatformOutputMuter.
var _ OutputMuter = (*darwinOutputMuter)(nil)

// darwinOutputMuter mutes the default output device using CoreAudio, falling
// back to AppleScript for devices CoreAudio exposes no writable control for.
type darwinOutputMuter struct {
	mu       sync.Mutex
	state    C.MuteState
	muted    bool
	fallback *appleScriptVolume
	method   string
}

// newPlatformOutputMuter creates a new macOS output muter
func newPlatformOutputMuter() (OutputMuter, error) {
	return &darwinOutputMuter{}, nil
}

// muteError turns a C return code into a descriptive Go error
func muteError(code C.int, state *C.MuteState) error {
	status := int32(state.lastStatus)

	switch int(code) {
	case muteNoDevice:
		return fmt.Errorf("no default audio output device (CoreAudio status %d)", status)
	case muteUnsupported:
		return fmt.Errorf("%s exposes no mute or volume control", outputDeviceLabel())
	case muteReadFailed:
		return fmt.Errorf("failed to read current output volume (CoreAudio status %d)", status)
	case muteSetFailed:
		return fmt.Errorf("failed to change output volume (CoreAudio status %d)", status)
	default:
		return fmt.Errorf("unexpected CoreAudio error (code %d, status %d)", int(code), status)
	}
}

// outputDeviceName returns the name of the current default output device
func outputDeviceName() (string, error) {
	buf := make([]C.char, 256)
	if code := C.openscribeOutputDeviceName(&buf[0], C.int(len(buf))); int(code) != muteOK {
		return "", fmt.Errorf("failed to read output device name (code %d)", int(code))
	}
	return C.GoString(&buf[0]), nil
}

// outputDeviceLabel names the output device for error messages, degrading to a
// generic phrase when the name cannot be read
func outputDeviceLabel() string {
	name, err := outputDeviceName()
	if err != nil || name == "" {
		return "the output device"
	}
	return fmt.Sprintf("output device %q", name)
}

// methodName describes a C mechanism for verbose output
func methodName(method int) string {
	switch method {
	case methodMute:
		return "device mute switch"
	case methodVirtual:
		return "system volume"
	case methodVolume:
		return "device volume"
	default:
		return "none"
	}
}

// Mute silences the current default output device, remembering its prior state
func (m *darwinOutputMuter) Mute() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.muted {
		return nil
	}

	code := C.openscribeMuteOutput(&m.state)
	if int(code) == muteOK {
		m.method = methodName(int(m.state.method))
		m.muted = true
		return nil
	}

	// CoreAudio found nothing writable on this device. AppleScript's `set
	// volume` drives the same control as the volume keys and reaches some
	// devices (AirPlay, various Bluetooth outputs) that expose no CoreAudio
	// property, so it is worth a try before giving up.
	if int(code) != muteUnsupported {
		return muteError(code, &m.state)
	}

	fallback, err := muteViaAppleScript()
	if err != nil {
		return fmt.Errorf("%v; AppleScript fallback also failed: %w", muteError(code, &m.state), err)
	}

	m.fallback = fallback
	m.method = "AppleScript"
	m.muted = true
	return nil
}

// Restore puts the previously muted device back to its original state.
// The device recorded at Mute time is targeted, not the current default, so
// switching outputs mid-recording still restores the right one.
func (m *darwinOutputMuter) Restore() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.muted {
		return nil
	}

	// Clear state first: if the device went away (headphones unplugged) there
	// is nothing left to restore, and retrying forever helps no one.
	m.muted = false
	m.method = ""

	if m.fallback != nil {
		fallback := m.fallback
		m.fallback = nil
		return fallback.restore()
	}

	if code := C.openscribeRestoreOutput(&m.state); int(code) != muteOK {
		return muteError(code, &m.state)
	}

	return nil
}

// IsMuted reports whether this muter currently holds a mute
func (m *darwinOutputMuter) IsMuted() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.muted
}

// Method names the mechanism currently holding the mute, empty when not muted
func (m *darwinOutputMuter) Method() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.method
}

// Describe reports the default output device, for diagnostics
func (m *darwinOutputMuter) Describe() string {
	name, err := outputDeviceName()
	if err != nil {
		return fmt.Sprintf("unknown output device (%v)", err)
	}
	return name
}

// Close restores the device if still muted and releases resources
func (m *darwinOutputMuter) Close() error {
	return m.Restore()
}

//go:build darwin
// +build darwin

package audio

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreAudio

#import <CoreAudio/CoreAudio.h>

// The "main" element of a property is element 0. Spelling it out avoids having
// to pick between kAudioObjectPropertyElementMain (macOS 12+) and the
// deprecated kAudioObjectPropertyElementMaster, which differ only in name.
#define OS_ELEMENT_MAIN ((AudioObjectPropertyElement)0)

// Return codes for the helpers below
#define OS_MUTE_OK             0
#define OS_MUTE_NO_DEVICE     -1
#define OS_MUTE_UNSUPPORTED   -2
#define OS_MUTE_SET_FAILED    -3
#define OS_MUTE_READ_FAILED   -4

#define OS_MAX_VOLUME_CHANNELS 2

// MuteState records everything needed to put the output device back exactly
// the way it was found.
typedef struct {
    AudioDeviceID deviceID;
    int           usedVolume;   // 1 when the volume-scalar fallback was used
    UInt32        prevMute;
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

// Mute the current default output device, recording its previous state.
static int openscribeMuteOutput(MuteState *state) {
    AudioDeviceID dev = kAudioObjectUnknown;
    OSStatus status = defaultOutputDevice(&dev);
    if (status != noErr || dev == kAudioObjectUnknown) {
        state->lastStatus = status;
        return OS_MUTE_NO_DEVICE;
    }

    state->deviceID = dev;
    state->usedVolume = 0;
    state->volumeChannelCount = 0;
    state->lastStatus = noErr;

    // Preferred path: the device exposes a settable master mute switch.
    AudioObjectPropertyAddress muteAddr = {
        kAudioDevicePropertyMute,
        kAudioDevicePropertyScopeOutput,
        OS_ELEMENT_MAIN
    };

    if (propertySettable(dev, &muteAddr)) {
        UInt32 size = sizeof(UInt32);
        status = AudioObjectGetPropertyData(dev, &muteAddr, 0, NULL, &size, &state->prevMute);
        if (status != noErr) {
            state->lastStatus = status;
            return OS_MUTE_READ_FAILED;
        }

        UInt32 on = 1;
        status = AudioObjectSetPropertyData(dev, &muteAddr, 0, NULL, sizeof(on), &on);
        if (status != noErr) {
            state->lastStatus = status;
            return OS_MUTE_SET_FAILED;
        }
        return OS_MUTE_OK;
    }

    // Fallback: drive the volume scalar to zero. Try the master element first,
    // then the preferred stereo channel pair (many USB/HDMI devices only
    // expose per-channel volume).
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

    state->usedVolume = 1;

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

    if (state->usedVolume) {
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
    } else {
        AudioObjectPropertyAddress muteAddr = {
            kAudioDevicePropertyMute,
            kAudioDevicePropertyScopeOutput,
            OS_ELEMENT_MAIN
        };
        UInt32 previous = state->prevMute;
        status = AudioObjectSetPropertyData(state->deviceID, &muteAddr, 0, NULL,
                                            sizeof(previous), &previous);
    }

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

// darwinOutputMuter mutes the default output device using CoreAudio
type darwinOutputMuter struct {
	mu    sync.Mutex
	state C.MuteState
	muted bool
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
		return fmt.Errorf("output device does not support mute or volume control")
	case muteReadFailed:
		return fmt.Errorf("failed to read current output volume (CoreAudio status %d)", status)
	case muteSetFailed:
		return fmt.Errorf("failed to change output volume (CoreAudio status %d)", status)
	default:
		return fmt.Errorf("unexpected CoreAudio error (code %d, status %d)", int(code), status)
	}
}

// Mute silences the current default output device, remembering its prior state
func (m *darwinOutputMuter) Mute() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.muted {
		return nil
	}

	if code := C.openscribeMuteOutput(&m.state); int(code) != muteOK {
		return muteError(code, &m.state)
	}

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

// Close restores the device if still muted and releases resources
func (m *darwinOutputMuter) Close() error {
	return m.Restore()
}

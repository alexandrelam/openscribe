package audio

// OutputMuter mutes the system's default audio output device while recording
// is in progress and restores it to its previous state afterwards.
//
// Implementations remember the state of the device that was muted, so a
// Restore always targets the same device even if the user switches outputs
// mid-recording.
type OutputMuter interface {
	// Prepare samples state that playing a sound would disturb. Call it
	// before any feedback sound, otherwise that sound makes the output device
	// look busy to the playback detection.
	Prepare()

	// SupportsVolumeControl reports whether the operating system can change
	// the output device's level at all. Some outputs — notably monitors that
	// keep volume in their own hardware — cannot be controlled, and are
	// silenced by pausing playback instead.
	SupportsVolumeControl() bool

	// Mute silences the current default output device.
	// It is a no-op if the device is already muted by this instance.
	Mute() error

	// Restore returns the previously muted device to its original state.
	// It is a no-op if Mute was never called or already restored.
	Restore() error

	// IsMuted reports whether this instance currently holds a mute.
	IsMuted() bool

	// Method names the mechanism currently holding the mute (for diagnostics),
	// and is empty when nothing is muted.
	Method() string

	// Describe names the audio output device that would be muted.
	Describe() string

	// Close releases any resources, restoring the device first if needed.
	Close() error
}

// NewOutputMuter creates a new platform-specific output muter
func NewOutputMuter() (OutputMuter, error) {
	return newPlatformOutputMuter()
}

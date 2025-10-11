package audio

// Feedback provides audio feedback for state changes during recording and transcription.
// This is a placeholder interface that will be implemented per-platform.
type Feedback interface {
	// PlayStartSound plays the sound when recording starts
	PlayStartSound() error

	// PlayStopSound plays the sound when recording stops
	PlayStopSound() error

	// PlayCompleteSound plays the sound when transcription completes
	PlayCompleteSound() error

	// Close releases any resources used by the feedback system
	Close() error
}

// NewFeedback creates a new platform-specific audio feedback instance
func NewFeedback() (Feedback, error) {
	return newPlatformFeedback()
}

package audio

import "testing"

// TestNewOutputMuter verifies a muter can be created and starts unmuted.
func TestNewOutputMuter(t *testing.T) {
	muter, err := NewOutputMuter()
	if err != nil {
		t.Fatalf("NewOutputMuter() returned error: %v", err)
	}
	if muter == nil {
		t.Fatal("NewOutputMuter() returned nil muter")
	}
	if muter.IsMuted() {
		t.Error("expected a freshly created muter to report IsMuted() = false")
	}
	if muter.Method() != "" {
		t.Errorf("expected Method() = \"\" before muting, got %q", muter.Method())
	}
	if muter.Describe() == "" {
		t.Error("expected Describe() to return a non-empty description")
	}

	// Prepare samples playback state and must be safe to call repeatedly,
	// including before anything has ever been muted
	muter.Prepare()
	muter.Prepare()
	if muter.IsMuted() {
		t.Error("expected Prepare() not to mute anything")
	}

	// SupportsVolumeControl is a read-only probe and must not change state
	muter.SupportsVolumeControl()
	if muter.IsMuted() {
		t.Error("expected SupportsVolumeControl() not to mute anything")
	}
}

// TestRestoreWithoutMute verifies Restore is a safe no-op when nothing was
// muted, which is what the deferred cleanup in `openscribe start` relies on
// when the user never records.
func TestRestoreWithoutMute(t *testing.T) {
	muter, err := NewOutputMuter()
	if err != nil {
		t.Fatalf("NewOutputMuter() returned error: %v", err)
	}

	tests := []struct {
		name string
		call func() error
	}{
		{name: "first restore", call: muter.Restore},
		{name: "repeated restore", call: muter.Restore},
		{name: "close", call: muter.Close},
		{name: "repeated close", call: muter.Close},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if muter.IsMuted() {
				t.Error("expected IsMuted() = false after restore/close")
			}
		})
	}
}

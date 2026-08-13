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

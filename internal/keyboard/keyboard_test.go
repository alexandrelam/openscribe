package keyboard

import (
	"runtime"
	"testing"
)

func TestNew(t *testing.T) {
	kb, err := New()

	if runtime.GOOS == "darwin" {
		// On macOS, should create successfully
		if err != nil {
			t.Fatalf("New() failed on darwin: %v", err)
		}
		if kb == nil {
			t.Fatal("New() returned nil keyboard on darwin")
		}

		// Test Close
		if err := kb.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	} else {
		// On non-macOS, should return error
		if err == nil {
			t.Fatal("New() should fail on non-darwin platforms")
		}
		if kb != nil {
			t.Fatal("New() should return nil keyboard on non-darwin platforms")
		}
	}
}

func TestCheckPermissions(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-darwin platform")
	}

	kb, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() {
		if err := kb.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	}()

	// Check permissions - may or may not be granted in test environment
	err = kb.CheckPermissions()
	// We don't assert the result as permissions depend on system config
	// Just verify it doesn't crash
	t.Logf("CheckPermissions() returned: %v", err)
}

func TestPasteText_EmptyString(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-darwin platform")
	}

	kb, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() {
		if err := kb.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	}()

	// Test with empty string (should not crash)
	// Note: This will fail if permissions aren't granted, which is expected
	err = kb.PasteText("")
	if err != nil && err.Error() != "cannot paste text: accessibility permissions not granted" {
		t.Errorf("PasteText(\"\") failed unexpectedly: %v", err)
	}
}

func TestPasteText_Unicode(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-darwin platform")
	}

	kb, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer func() {
		if err := kb.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	}()

	// Test with Unicode characters
	testStrings := []string{
		"Hello, World!",
		"Hello 世界",
		"Привет мир",
		"مرحبا بالعالم",
		"🚀 Test 123",
	}

	for _, str := range testStrings {
		// Note: This will fail if permissions aren't granted, which is expected
		err := kb.PasteText(str)
		if err != nil && err.Error() != "cannot paste text: accessibility permissions not granted" {
			t.Errorf("PasteText(%q) failed unexpectedly: %v", str, err)
		}
	}
}

func TestClose_Multiple(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-darwin platform")
	}

	kb, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test multiple closes (should not crash)
	if err := kb.Close(); err != nil {
		t.Errorf("First Close() failed: %v", err)
	}

	if err := kb.Close(); err != nil {
		t.Errorf("Second Close() failed: %v", err)
	}
}

func TestRequestPermissions(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test on non-darwin platform")
	}

	// RequestPermissions should not crash
	// We can't test if it actually prompts without user interaction
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RequestPermissions() panicked: %v", r)
		}
	}()

	RequestPermissions()
}

// Benchmark tests
func BenchmarkPasteText(b *testing.B) {
	if runtime.GOOS != "darwin" {
		b.Skip("Skipping macOS-specific benchmark on non-darwin platform")
	}

	kb, err := New()
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	defer func() {
		if err := kb.Close(); err != nil {
			b.Errorf("Close() failed: %v", err)
		}
	}()

	// Check if permissions are granted
	if err := kb.CheckPermissions(); err != nil {
		b.Skip("Skipping benchmark - accessibility permissions not granted")
	}

	testText := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = kb.PasteText(testText)
	}
}

func BenchmarkCheckPermissions(b *testing.B) {
	if runtime.GOOS != "darwin" {
		b.Skip("Skipping macOS-specific benchmark on non-darwin platform")
	}

	kb, err := New()
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	defer func() {
		if err := kb.Close(); err != nil {
			b.Errorf("Close() failed: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = kb.CheckPermissions()
	}
}

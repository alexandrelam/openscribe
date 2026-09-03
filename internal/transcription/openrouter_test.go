package transcription

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrelam/openscribe/internal/config"
)

// writeTestAudio writes a small stand-in audio file and returns its path.
// The transport tests only care that the bytes round-trip, so a real WAV
// is unnecessary and keeps the tests hermetic.
func writeTestAudio(t *testing.T) (string, []byte) {
	t.Helper()
	data := []byte("RIFF\x00\x00\x00\x00WAVEfmt fake-audio-bytes")
	path := filepath.Join(t.TempDir(), "recording.wav")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write test audio: %v", err)
	}
	return path, data
}

// newTestTranscriber returns a transcriber pointed at srv.
func newTestTranscriber(t *testing.T, srv *httptest.Server, model string) *OpenRouterTranscriber {
	t.Helper()
	tr, err := NewOpenRouterTranscriber("test-key", model)
	if err != nil {
		t.Fatalf("NewOpenRouterTranscriber() error = %v", err)
	}
	tr.baseURL = srv.URL
	return tr
}

func TestNewOpenRouterTranscriber(t *testing.T) {
	t.Run("Empty API key is rejected", func(t *testing.T) {
		if _, err := NewOpenRouterTranscriber("", "some/model"); err == nil {
			t.Error("NewOpenRouterTranscriber() with empty key should return an error")
		}
	})

	t.Run("Empty model falls back to default", func(t *testing.T) {
		tr, err := NewOpenRouterTranscriber("test-key", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tr.model != config.DefaultOpenRouterModel {
			t.Errorf("model = %q, want %q", tr.model, config.DefaultOpenRouterModel)
		}
	})

	t.Run("Explicit model is preserved", func(t *testing.T) {
		tr, err := NewOpenRouterTranscriber("test-key", "vendor/other-model")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tr.model != "vendor/other-model" {
			t.Errorf("model = %q, want vendor/other-model", tr.model)
		}
	})
}

func TestOpenRouterTranscribeFile_RequestShape(t *testing.T) {
	audioPath, audioData := writeTestAudio(t)

	var gotMethod, gotAuth, gotContentType, gotReferer, gotTitle string
	var gotBody openRouterRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotReferer = r.Header.Get("HTTP-Referer")
		gotTitle = r.Header.Get("X-OpenRouter-Title")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"text":"hello"}`))
	}))
	defer srv.Close()

	tr := newTestTranscriber(t, srv, "microsoft/mai-transcribe-2")
	if _, err := tr.TranscribeFile(audioPath, Options{Language: "en"}); err != nil {
		t.Fatalf("TranscribeFile() error = %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer test-key")
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotReferer == "" || gotTitle == "" {
		t.Errorf("attribution headers missing: referer=%q title=%q", gotReferer, gotTitle)
	}
	if gotBody.Model != "microsoft/mai-transcribe-2" {
		t.Errorf("model = %q, want microsoft/mai-transcribe-2", gotBody.Model)
	}
	if gotBody.InputAudio.Format != "wav" {
		t.Errorf("input_audio.format = %q, want wav", gotBody.InputAudio.Format)
	}
	if gotBody.Language != "en" {
		t.Errorf("language = %q, want en", gotBody.Language)
	}
	if strings.HasPrefix(gotBody.InputAudio.Data, "data:") {
		t.Error("input_audio.data must be raw base64, not a data URI")
	}
	decoded, err := base64.StdEncoding.DecodeString(gotBody.InputAudio.Data)
	if err != nil {
		t.Fatalf("input_audio.data is not valid base64: %v", err)
	}
	if string(decoded) != string(audioData) {
		t.Errorf("decoded audio does not round-trip: got %q, want %q", decoded, audioData)
	}
}

func TestOpenRouterTranscribeFile_LanguageOmittedWhenAutoDetect(t *testing.T) {
	audioPath, _ := writeTestAudio(t)

	var raw map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"text":"hello"}`))
	}))
	defer srv.Close()

	tr := newTestTranscriber(t, srv, "")
	if _, err := tr.TranscribeFile(audioPath, Options{Language: ""}); err != nil {
		t.Fatalf("TranscribeFile() error = %v", err)
	}

	if _, present := raw["language"]; present {
		t.Error("language key should be omitted when auto-detecting")
	}
}

func TestOpenRouterTranscribeFile_Responses(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         string
		optsLanguage string
		wantErr      bool
		errContains  string
		wantText     string
		wantLanguage string
		wantDuration float64
	}{
		{
			name:         "Plain json response uses requested language and usage.seconds",
			status:       http.StatusOK,
			body:         `{"text":"hello there","usage":{"seconds":3.5}}`,
			optsLanguage: "fr",
			wantText:     "hello there",
			wantLanguage: "fr",
			wantDuration: 3.5,
		},
		{
			name:         "Verbose fields are honored when present",
			status:       http.StatusOK,
			body:         `{"text":"bonjour","language":"fr","duration":2.0}`,
			optsLanguage: "",
			wantText:     "bonjour",
			wantLanguage: "fr",
			wantDuration: 2.0,
		},
		{
			name:        "Empty text is an error",
			status:      http.StatusOK,
			body:        `{"text":""}`,
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "Error envelope on HTTP 200 is surfaced",
			status:      http.StatusOK,
			body:        `{"error":{"message":"provider unavailable"}}`,
			wantErr:     true,
			errContains: "provider unavailable",
		},
		{
			name:        "HTTP 401 mentions the key hint",
			status:      http.StatusUnauthorized,
			body:        `{"error":{"message":"No auth credentials found"}}`,
			wantErr:     true,
			errContains: "sk-or-",
		},
		{
			name:        "HTTP 500 reports the status",
			status:      http.StatusInternalServerError,
			body:        `upstream exploded`,
			wantErr:     true,
			errContains: "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audioPath, _ := writeTestAudio(t)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			tr := newTestTranscriber(t, srv, "")
			result, err := tr.TranscribeFile(audioPath, Options{Language: tt.optsLanguage})

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got result %+v", result)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", result.Text, tt.wantText)
			}
			if result.Language != tt.wantLanguage {
				t.Errorf("Language = %q, want %q", result.Language, tt.wantLanguage)
			}
			if result.Duration != tt.wantDuration {
				t.Errorf("Duration = %v, want %v", result.Duration, tt.wantDuration)
			}
		})
	}
}

func TestOpenRouterTranscribeFile_MissingFileMakesNoRequest(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	tr := newTestTranscriber(t, srv, "")
	if _, err := tr.TranscribeFile("/nonexistent/recording.wav", Options{}); err == nil {
		t.Error("expected an error for a missing audio file")
	}
	if called {
		t.Error("no HTTP request should be made when the audio file cannot be read")
	}
}

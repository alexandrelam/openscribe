package transcription

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexandrelam/openscribe/internal/config"
)

// openRouterEndpoint is the OpenRouter speech-to-text endpoint.
const openRouterEndpoint = "https://openrouter.ai/api/v1/audio/transcriptions"

// openRouterTimeout bounds a single transcription request. Dictation clips are
// short, so a hung connection should surface quickly rather than wedge the app.
const openRouterTimeout = 2 * time.Minute

// OpenRouterTranscriber handles speech-to-text transcription using the OpenRouter API.
type OpenRouterTranscriber struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// openRouterInputAudio carries the base64-encoded audio and its container format.
type openRouterInputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

// openRouterRequest is the JSON body sent to the OpenRouter transcription API.
type openRouterRequest struct {
	Model          string               `json:"model"`
	InputAudio     openRouterInputAudio `json:"input_audio"`
	Language       string               `json:"language,omitempty"`
	ResponseFormat string               `json:"response_format,omitempty"`
}

// openRouterResponse represents the JSON response from the OpenRouter transcription API.
// Language and Duration are only returned by providers that support "verbose_json";
// they are parsed opportunistically and fall back when absent.
type openRouterResponse struct {
	Text     string  `json:"text"`
	Language string  `json:"language"`
	Duration float64 `json:"duration"`
	Usage    struct {
		Seconds float64 `json:"seconds"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewOpenRouterTranscriber creates a new OpenRouter-based transcriber.
func NewOpenRouterTranscriber(apiKey, model string) (*OpenRouterTranscriber, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenRouter API key is required")
	}
	if model == "" {
		model = config.DefaultOpenRouterModel
	}
	return &OpenRouterTranscriber{
		apiKey:  apiKey,
		model:   model,
		baseURL: openRouterEndpoint,
		client:  &http.Client{Timeout: openRouterTimeout},
	}, nil
}

// TranscribeFile transcribes an audio file using the OpenRouter API and returns the text.
func (t *OpenRouterTranscriber) TranscribeFile(audioPath string, opts Options) (*Result, error) {
	// Recordings are short push-to-talk clips (16 kHz mono 16-bit, ~32 KB/s), so
	// reading the whole file to base64-encode it stays comfortably small.
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	reqBody := openRouterRequest{
		Model: t.model,
		InputAudio: openRouterInputAudio{
			Data:   base64.StdEncoding.EncodeToString(audioData),
			Format: "wav",
		},
		Language:       opts.Language, // empty means auto-detect
		ResponseFormat: "json",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := http.NewRequest("POST", t.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")
	// Optional attribution headers, used by OpenRouter to identify the caller.
	req.Header.Set("HTTP-Referer", "https://github.com/alexandrelam/openscribe")
	req.Header.Set("X-OpenRouter-Title", "OpenScribe")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenRouter API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp openRouterResponse
	parseErr := json.Unmarshal(respBody, &apiResp)

	if resp.StatusCode != http.StatusOK {
		return nil, t.apiError(resp.StatusCode, respBody, parseErr, &apiResp)
	}

	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse OpenRouter response: %w", parseErr)
	}

	// OpenRouter can return HTTP 200 carrying an error envelope for
	// provider-level failures, so check it before trusting the text.
	if apiResp.Error != nil && apiResp.Error.Message != "" {
		return nil, fmt.Errorf("OpenRouter API error: %s", apiResp.Error.Message)
	}

	if apiResp.Text == "" {
		return nil, fmt.Errorf("transcription produced empty result")
	}

	language := opts.Language
	if apiResp.Language != "" {
		language = apiResp.Language
	}

	duration := apiResp.Duration
	if duration == 0 {
		duration = apiResp.Usage.Seconds
	}

	return &Result{
		Text:     apiResp.Text,
		Language: language,
		Duration: duration,
	}, nil
}

// apiError builds a readable error for a non-200 response, preferring the
// structured error message over the raw body.
func (t *OpenRouterTranscriber) apiError(status int, respBody []byte, parseErr error, apiResp *openRouterResponse) error {
	detail := strings.TrimSpace(string(respBody))
	if parseErr == nil && apiResp.Error != nil && apiResp.Error.Message != "" {
		detail = apiResp.Error.Message
	}

	if status == http.StatusUnauthorized {
		return fmt.Errorf("OpenRouter API error (HTTP %d): %s\n"+
			"Check that your key is an OpenRouter key (they start with 'sk-or-'), not an OpenAI key.\n"+
			"Set one with: openscribe config --set-openrouter-api-key <key>", status, detail)
	}

	return fmt.Errorf("OpenRouter API error (HTTP %d): %s", status, detail)
}

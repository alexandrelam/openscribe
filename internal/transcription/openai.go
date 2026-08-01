package transcription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DefaultOpenAIModel is the default OpenAI transcription model.
const DefaultOpenAIModel = "gpt-transcribe"

// OpenAITranscriber handles speech-to-text transcription using the OpenAI API.
type OpenAITranscriber struct {
	apiKey   string
	model    string
	prompt   string
	keywords []string
}

// openAIResponse represents the JSON response from the OpenAI transcription API.
type openAIResponse struct {
	Text      string `json:"text"`
	Languages []struct {
		Code string `json:"code"`
	} `json:"languages"`
}

// NewOpenAITranscriber creates a new OpenAI-based transcriber.
func NewOpenAITranscriber(apiKey, model, prompt string, keywords []string) (*OpenAITranscriber, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		model = DefaultOpenAIModel
	}
	if err := validateKeywords(keywords); err != nil {
		return nil, err
	}
	return &OpenAITranscriber{
		apiKey:   apiKey,
		model:    model,
		prompt:   prompt,
		keywords: keywords,
	}, nil
}

func usesNewModelFields(model string) bool {
	return strings.HasPrefix(model, "gpt-transcribe") || strings.HasPrefix(model, "gpt-live-transcribe")
}

func validateKeywords(keywords []string) error {
	for _, kw := range keywords {
		if strings.ContainsAny(kw, "<>\r\n") {
			return fmt.Errorf("invalid keyword %q: must be a single line and must not contain '<', '>', or line breaks", kw)
		}
	}
	return nil
}

// TranscribeFile transcribes an audio file using the OpenAI API and returns the text.
func (t *OpenAITranscriber) TranscribeFile(audioPath string, opts Options) (*Result, error) {
	// Open the audio file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Build multipart form request
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add the audio file
	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy audio data: %w", err)
	}

	// Add model field
	if err := writer.WriteField("model", t.model); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	newModel := usesNewModelFields(t.model)

	if opts.Language != "" {
		if newModel {
			if err := writer.WriteField("languages[]", opts.Language); err != nil {
				return nil, fmt.Errorf("failed to write languages field: %w", err)
			}
		} else {
			if err := writer.WriteField("language", opts.Language); err != nil {
				return nil, fmt.Errorf("failed to write language field: %w", err)
			}
		}
	}

	if t.prompt != "" {
		if err := writer.WriteField("prompt", t.prompt); err != nil {
			return nil, fmt.Errorf("failed to write prompt field: %w", err)
		}
	}

	if newModel {
		for _, kw := range t.keywords {
			if err := writer.WriteField("keywords[]", kw); err != nil {
				return nil, fmt.Errorf("failed to write keywords field: %w", err)
			}
		}
	}

	// Add response format
	if err := writer.WriteField("response_format", "json"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResp openAIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if apiResp.Text == "" {
		return nil, fmt.Errorf("transcription produced empty result")
	}

	language := opts.Language
	if len(apiResp.Languages) > 0 && apiResp.Languages[0].Code != "" {
		language = apiResp.Languages[0].Code
	}

	return &Result{
		Text:     apiResp.Text,
		Language: language,
	}, nil
}

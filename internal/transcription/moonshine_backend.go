//go:build moonshine

package transcription

import (
	"fmt"
	"os"

	"github.com/alexandrelam/openscribe/internal/config"
	"github.com/alexandrelam/openscribe/internal/models"
	"github.com/alexandrelam/openscribe/internal/transcription/moonshine"
)

// MoonshineTranscriber implements the Transcriber interface using Moonshine.
type MoonshineTranscriber struct {
	engine   *moonshine.Transcriber
	modelDir string
}

func newMoonshineTranscriber(cfg *config.Config) (Transcriber, error) {
	modelSize := models.MoonshineModelSize(cfg.MoonshineModel)
	if modelSize == "" {
		modelSize = models.MoonshineTiny
	}

	// Validate model is downloaded
	ok, err := models.IsMoonshineModelDownloaded(modelSize)
	if err != nil {
		return nil, fmt.Errorf("failed to check moonshine model: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("moonshine model '%s' is not downloaded. Run 'openscribe models download --backend moonshine %s' first", modelSize, modelSize)
	}

	modelDir, err := models.GetMoonshineModelDir(modelSize)
	if err != nil {
		return nil, err
	}

	engine, err := moonshine.New(modelDir, string(modelSize))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize moonshine: %w", err)
	}

	return &MoonshineTranscriber{engine: engine, modelDir: modelDir}, nil
}

// TranscribeFile reads a WAV file and transcribes it using Moonshine.
func (t *MoonshineTranscriber) TranscribeFile(audioPath string, opts Options) (*Result, error) {
	// Read WAV file and convert to float32 samples
	samples, sampleRate, err := readWAVAsFloat32(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	text, err := t.engine.Transcribe(samples, sampleRate)
	if err != nil {
		return nil, fmt.Errorf("moonshine transcription failed: %w", err)
	}

	if text == "" {
		return nil, fmt.Errorf("transcription produced empty result")
	}

	return &Result{
		Text:     text,
		Language: opts.Language, // Moonshine doesn't do language detection
	}, nil
}

// readWAVAsFloat32 reads a 16-bit PCM WAV file and returns float32 samples normalized to [-1, 1]
// along with the sample rate.
func readWAVAsFloat32(path string) ([]float32, int32, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	if len(data) < 44 {
		return nil, 0, fmt.Errorf("file too small to be a valid WAV")
	}

	// Verify RIFF header
	if string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, 0, fmt.Errorf("not a valid WAV file")
	}

	// Extract sample rate from fmt chunk (bytes 24-27)
	sampleRate := int32(data[24]) | int32(data[25])<<8 | int32(data[26])<<16 | int32(data[27])<<24

	// Find data chunk
	offset := 12
	for offset < len(data)-8 {
		chunkID := string(data[offset : offset+4])
		chunkSize := int(data[offset+4]) | int(data[offset+5])<<8 | int(data[offset+6])<<16 | int(data[offset+7])<<24
		if chunkID == "data" {
			pcmData := data[offset+8 : offset+8+chunkSize]
			numSamples := len(pcmData) / 2
			samples := make([]float32, numSamples)
			for i := 0; i < numSamples; i++ {
				sample := int16(pcmData[i*2]) | int16(pcmData[i*2+1])<<8
				samples[i] = float32(sample) / 32768.0
			}
			return samples, sampleRate, nil
		}
		offset += 8 + chunkSize
		if chunkSize%2 != 0 {
			offset++
		}
	}

	return nil, 0, fmt.Errorf("no data chunk found in WAV file")
}

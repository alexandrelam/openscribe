package audio

import (
	"encoding/binary"
	"fmt"
	"os"
)

// WAVHeader represents the header of a WAV file
type WAVHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32
}

// SaveWAV saves audio data as a WAV file
func SaveWAV(filename string, audioData []byte, sampleRate, channels uint32) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create WAV file: %w", err)
	}
	defer file.Close()

	bitsPerSample := uint16(16) // 16-bit audio
	byteRate := sampleRate * uint32(channels) * uint32(bitsPerSample) / 8
	blockAlign := uint16(channels) * bitsPerSample / 8
	dataSize := uint32(len(audioData))

	// Create WAV header
	header := WAVHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     36 + dataSize,
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1, // PCM
		NumChannels:   uint16(channels),
		SampleRate:    sampleRate,
		ByteRate:      byteRate,
		BlockAlign:    blockAlign,
		BitsPerSample: bitsPerSample,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: dataSize,
	}

	// Write header
	if err := binary.Write(file, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("failed to write WAV header: %w", err)
	}

	// Write audio data
	if _, err := file.Write(audioData); err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}

	return nil
}

// LoadWAV loads audio data from a WAV file
func LoadWAV(filename string) ([]byte, uint32, uint32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to open WAV file: %w", err)
	}
	defer file.Close()

	// Read header
	var header WAVHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read WAV header: %w", err)
	}

	// Validate WAV file
	if string(header.ChunkID[:]) != "RIFF" || string(header.Format[:]) != "WAVE" {
		return nil, 0, 0, fmt.Errorf("not a valid WAV file")
	}

	// Read audio data
	audioData := make([]byte, header.Subchunk2Size)
	if _, err := file.Read(audioData); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read audio data: %w", err)
	}

	return audioData, header.SampleRate, uint32(header.NumChannels), nil
}

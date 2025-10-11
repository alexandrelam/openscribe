package audio

import (
	"fmt"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
)

// Recorder handles audio recording from a microphone
type Recorder struct {
	deviceName     string
	sampleRate     uint32
	channels       uint32
	isRecording    bool
	audioData      []byte
	audioDataMutex sync.Mutex
	device         *malgo.Device
	context        *malgo.AllocatedContext
}

// NewRecorder creates a new audio recorder
func NewRecorder(deviceName string) *Recorder {
	return &Recorder{
		deviceName:  deviceName,
		sampleRate:  16000, // Whisper-compatible sample rate
		channels:    1,     // Mono
		isRecording: false,
		audioData:   make([]byte, 0),
	}
}

// Start begins recording audio
func (r *Recorder) Start() error {
	if r.isRecording {
		return fmt.Errorf("already recording")
	}

	// Initialize audio context
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize audio context: %w", err)
	}
	r.context = ctx

	// Find the device to use
	var deviceInfo *malgo.DeviceInfo
	if r.deviceName != "" {
		// Find specific device by name
		infos, devicesErr := ctx.Devices(malgo.Capture)
		if devicesErr != nil {
			_ = ctx.Uninit()
			ctx.Free()
			return fmt.Errorf("failed to enumerate devices: %w", devicesErr)
		}

		found := false
		for _, info := range infos {
			if info.Name() == r.deviceName {
				deviceInfo = &info
				found = true
				break
			}
		}

		if !found {
			_ = ctx.Uninit()
			ctx.Free()
			return fmt.Errorf("device not found: %s", r.deviceName)
		}
	}

	// Configure device
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = r.channels
	deviceConfig.SampleRate = r.sampleRate
	deviceConfig.Alsa.NoMMap = 1

	if deviceInfo != nil {
		deviceConfig.Capture.DeviceID = deviceInfo.ID.Pointer()
	}

	// Reset audio data buffer
	r.audioDataMutex.Lock()
	r.audioData = make([]byte, 0)
	r.audioDataMutex.Unlock()

	// Callback to capture audio data
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		r.audioDataMutex.Lock()
		r.audioData = append(r.audioData, pSample...)
		r.audioDataMutex.Unlock()
	}

	// Initialize and start device
	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: onRecvFrames,
	})
	if err != nil {
		_ = ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("failed to initialize device: %w", err)
	}

	err = device.Start()
	if err != nil {
		device.Uninit()
		_ = ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("failed to start device: %w", err)
	}

	r.device = device
	r.isRecording = true

	return nil
}

// Stop ends the recording and returns the captured audio data
func (r *Recorder) Stop() ([]byte, error) {
	if !r.isRecording {
		return nil, fmt.Errorf("not currently recording")
	}

	// Stop the device
	if r.device != nil {
		r.device.Uninit()
	}

	// Cleanup context
	if r.context != nil {
		_ = r.context.Uninit()
		r.context.Free()
	}

	r.isRecording = false

	// Return the captured audio data
	r.audioDataMutex.Lock()
	data := make([]byte, len(r.audioData))
	copy(data, r.audioData)
	r.audioDataMutex.Unlock()

	return data, nil
}

// IsRecording returns whether the recorder is currently recording
func (r *Recorder) IsRecording() bool {
	return r.isRecording
}

// GetSampleRate returns the recorder's sample rate
func (r *Recorder) GetSampleRate() uint32 {
	return r.sampleRate
}

// GetChannels returns the number of audio channels
func (r *Recorder) GetChannels() uint32 {
	return r.channels
}

// RecordDuration records audio for a specific duration
func (r *Recorder) RecordDuration(duration time.Duration) ([]byte, error) {
	if err := r.Start(); err != nil {
		return nil, err
	}

	// Wait for the specified duration
	time.Sleep(duration)

	return r.Stop()
}

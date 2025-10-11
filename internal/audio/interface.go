package audio

import "github.com/gen2brain/malgo"

// DeviceInfo is an interface that abstracts device information
// This allows for mocking in tests
type DeviceInfo interface {
	Name() string
	IsDefault() uint32
}

// MalgoDeviceInfo wraps malgo.DeviceInfo to implement our DeviceInfo interface
type MalgoDeviceInfo struct {
	info *malgo.DeviceInfo
}

// Name implements DeviceInfo
func (m MalgoDeviceInfo) Name() string {
	return m.info.Name()
}

// IsDefault implements DeviceInfo
func (m MalgoDeviceInfo) IsDefault() uint32 {
	return m.info.IsDefault
}

// DeviceEnumerator is an interface for enumerating audio devices
// This allows for mocking in tests
type DeviceEnumerator interface {
	Devices(deviceType malgo.DeviceType) ([]DeviceInfo, error)
}

// MalgoContext wraps malgo.AllocatedContext to implement DeviceEnumerator
type MalgoContext struct {
	ctx *malgo.AllocatedContext
}

// Devices implements DeviceEnumerator
func (m *MalgoContext) Devices(deviceType malgo.DeviceType) ([]DeviceInfo, error) {
	infos, err := m.ctx.Devices(deviceType)
	if err != nil {
		return nil, err
	}

	// Wrap malgo.DeviceInfo in our interface
	result := make([]DeviceInfo, len(infos))
	for i := range infos {
		result[i] = MalgoDeviceInfo{info: &infos[i]}
	}

	return result, nil
}

// Uninit uninitializes the context
func (m *MalgoContext) Uninit() error {
	return m.ctx.Uninit()
}

// Free frees the context
func (m *MalgoContext) Free() {
	m.ctx.Free()
}

// NewMalgoContext creates a new MalgoContext
func NewMalgoContext() (*MalgoContext, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	return &MalgoContext{ctx: ctx}, nil
}

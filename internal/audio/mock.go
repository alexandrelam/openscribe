package audio

import "github.com/gen2brain/malgo"

// MockDeviceEnumerator is a mock implementation of DeviceEnumerator for testing
type MockDeviceEnumerator struct {
	DevicesFunc func(deviceType malgo.DeviceType) ([]DeviceInfo, error)
}

// Devices implements DeviceEnumerator
func (m *MockDeviceEnumerator) Devices(deviceType malgo.DeviceType) ([]DeviceInfo, error) {
	if m.DevicesFunc != nil {
		return m.DevicesFunc(deviceType)
	}
	return nil, nil
}

// MockDeviceInfo is a mock implementation of DeviceInfo for testing
type MockDeviceInfo struct {
	name      string
	isDefault uint32
}

// Name returns the device name
func (m MockDeviceInfo) Name() string {
	return m.name
}

// IsDefault returns whether the device is default
func (m MockDeviceInfo) IsDefault() uint32 {
	return m.isDefault
}

// NewMockDeviceInfo creates a new MockDeviceInfo
func NewMockDeviceInfo(name string, isDefault bool) DeviceInfo {
	defaultValue := uint32(0)
	if isDefault {
		defaultValue = 1
	}
	return MockDeviceInfo{
		name:      name,
		isDefault: defaultValue,
	}
}

// CreateMockEnumerator creates a mock enumerator with predefined devices
func CreateMockEnumerator(devices []DeviceInfo, err error) *MockDeviceEnumerator {
	return &MockDeviceEnumerator{
		DevicesFunc: func(deviceType malgo.DeviceType) ([]DeviceInfo, error) {
			return devices, err
		},
	}
}

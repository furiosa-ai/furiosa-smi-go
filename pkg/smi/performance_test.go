package smi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testDeviceTemperature(t *testing.T, arch Arch, expected FuriosaSmiDeviceTemperature) {
	mockDevice := GetStaticMockDevice(arch, 0)

	temperature, err := mockDevice.DeviceTemperature()
	assert.NoError(t, err)

	assert.Equal(t, expected.SocPeak(), temperature.SocPeak())
	assert.Equal(t, expected.Ambient(), temperature.Ambient())
}

func TestDeviceTemperature(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    FuriosaSmiDeviceTemperature
	}{
		{
			description: "Test RNGD Device Temperature",
			arch:        FuriosaSmiArchRngd,
			expected:    FuriosaSmiDeviceTemperature{socPeak: 20.0, ambient: 10.0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceTemperature(t, tc.arch, tc.expected)
		})
	}
}

func testPowerConsumption(t *testing.T, arch Arch, expected float64) {
	mockDevice := GetStaticMockDevice(arch, 0)

	power, err := mockDevice.PowerConsumption()
	assert.NoError(t, err)

	assert.Equal(t, expected, power)
}

func TestPowerConsumption(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    float64
	}{
		{
			description: "Test RNGD Device Power Consumption",
			arch:        FuriosaSmiArchRngd,
			expected:    100.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testPowerConsumption(t, tc.arch, tc.expected)
		})
	}
}

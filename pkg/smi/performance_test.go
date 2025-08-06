package smi

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
	"github.com/stretchr/testify/assert"
)

func testDeviceTemperature(t *testing.T, arch Arch, expected deviceTemperature) {
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
		expected    deviceTemperature
	}{
		{
			description: "Test RNGD Device Temperature",
			arch:        ArchRngd,
			expected:    deviceTemperature{binding.FuriosaSmiDeviceTemperature{SocPeak: 20.0, Ambient: 10.0}},
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
			arch:        ArchRngd,
			expected:    100.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testPowerConsumption(t, tc.arch, tc.expected)
		})
	}
}

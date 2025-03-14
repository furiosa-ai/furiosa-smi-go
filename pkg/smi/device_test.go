package smi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCoreStatus(t *testing.T, arch Arch) {
	mockDevice := GetStaticMockDevice(arch, 0)

	coreStat, err := mockDevice.CoreStatus()
	assert.NoError(t, err)

	for i, peStatus := range coreStat.PeStatus() {
		assert.Equal(t, uint32(i), peStatus.Core())
		assert.Equal(t, CoreStatusAvailable, peStatus.Status())
	}
}

func TestCoreStatus(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    map[uint32]CoreStatus
	}{
		{
			description: "Test Warboy Core Status",
			arch:        ArchWarboy,
		},
		{
			description: "Test RNGD Core Status",
			arch:        ArchRngd,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			testCoreStatus(t, test.arch)
		})
	}
}

func testLiveness(t *testing.T, arch Arch, expected bool) {
	mockDevice := GetStaticMockDevice(arch, 0)

	liveness, err := mockDevice.Liveness()
	assert.NoError(t, err)

	assert.Equal(t, expected, liveness)
}

func TestLiveness(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    bool
	}{
		{
			description: "Test Warboy Liveness",
			arch:        ArchWarboy,
			expected:    true,
		},
		{
			description: "Test RNGD Liveness",
			arch:        ArchRngd,
			expected:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testLiveness(t, tc.arch, tc.expected)
		})
	}
}

func testDeviceToDeviceLinkType(t *testing.T, devices []Device, expectedMap map[int]map[int]LinkType) {
	for i, device0 := range devices {
		for j, device1 := range devices {
			linkType, err := device0.DeviceToDeviceLinkType(device1)
			assert.NoError(t, err)

			idx0, idx1 := i, j
			if i > j {
				idx0, idx1 = j, i
			}

			expected := expectedMap[idx0][idx1]
			assert.Equalf(t, expected, linkType, "expected linktype between npu%d, npu%d is %v but got %v", i, j, expected, linkType)
		}
	}
}

func TestDeviceToDeviceLinkType(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
	}{
		{
			description: "Test Warboy DeviceToDeviceLinkType",
			arch:        ArchWarboy,
		},
		{
			description: "Test RNGD DeviceToDeviceLinkType",
			arch:        ArchRngd,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceToDeviceLinkType(t, GetStaticMockDevices(tc.arch), linkTypeHintMap)
		})
	}
}

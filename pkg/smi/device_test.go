package smi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCoreStatus(t *testing.T, arch Arch) {
	mockDevice := GetStaticMockDevice(arch, 0)

	coreStat, err := mockDevice.CoreStatus()
	assert.NoError(t, err)

	for i, peStat := range coreStat.PeStatus() {
		assert.Equal(t, uint32(i), peStat.Core())
		assert.Equal(t, FuriosaSmiCoreStatusAvailable, peStat.Status())
	}
}

func TestCoreStatus(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    map[uint32]FuriosaSmiCoreStatus
	}{
		{
			description: "Test RNGD Core Status",
			arch:        FuriosaSmiArchRngd,
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
			description: "Test RNGD Liveness",
			arch:        FuriosaSmiArchRngd,
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
			description: "Test RNGD DeviceToDeviceLinkType",
			arch:        FuriosaSmiArchRngd,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceToDeviceLinkType(t, GetStaticMockDevices(tc.arch), linkTypeHintMap)
		})
	}
}

func testCoreFrequency(t *testing.T, arch Arch, expected FuriosaSmiCoreFrequency) {
	mockDevice := GetStaticMockDevice(arch, 0)

	freq, err := mockDevice.CoreFrequency()
	assert.NoError(t, err)

	assert.Equal(t, len(expected.pe), len(freq.PeFrequency()))
	for i := 0; i < len(expected.pe); i++ {
		assert.Equal(t, expected.pe[i].Core, freq.PeFrequency()[i].Core())
		assert.Equal(t, expected.pe[i].Frequency, freq.PeFrequency()[i].Frequency())
	}
}

func TestCoreFrequency(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    FuriosaSmiCoreFrequency
	}{
		{
			description: "Test RNGD Core Frequency",
			arch:        FuriosaSmiArchRngd,
			expected: func() FuriosaSmiCoreFrequency {
				exp := FuriosaSmiCoreFrequency{pe: []FuriosaSmiPeFrequency{}}
				for i := 0; i < 8; i++ {
					exp.pe[i] = FuriosaSmiPeFrequency{core: uint32(i), frequency: 500}
				}

				return exp
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testCoreFrequency(t, tc.arch, tc.expected)
		})
	}
}

func testMemoryFrequency(t *testing.T, arch Arch, expected uint32) {
	mockDevice := GetStaticMockDevice(arch, 0)

	freq, err := mockDevice.MemoryFrequency()
	assert.NoError(t, err)

	assert.Equal(t, expected, freq.Frequency())
}

func TestMemoryFrequency(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    uint32
	}{
		{
			description: "Test RNGD Memory Frequency",
			arch:        FuriosaSmiArchRngd,
			expected:    6000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testMemoryFrequency(t, tc.arch, tc.expected)
		})
	}
}

package smi

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
	"github.com/stretchr/testify/assert"
)

func testCoreStatus(t *testing.T, arch Arch, expected map[uint32]CoreStatus) {
	mockDevice := GetStaticMockDevice(arch, 0)

	coreStat, err := mockDevice.CoreStatus()
	assert.NoError(t, err)

	assert.Equal(t, expected, coreStat)
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
			expected: func() map[uint32]CoreStatus {
				exp := make(map[uint32]CoreStatus)
				for i := 0; i < 2; i++ {
					exp[uint32(i)] = CoreStatusAvailable
				}

				return exp
			}(),
		},
		{
			description: "Test RNGD Core Status",
			arch:        ArchRngd,
			expected: func() map[uint32]CoreStatus {
				exp := make(map[uint32]CoreStatus)
				for i := 0; i < 8; i++ {
					exp[uint32(i)] = CoreStatusAvailable
				}

				return exp
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			testCoreStatus(t, test.arch, test.expected)
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

func testCoreFrequency(t *testing.T, arch Arch, expected coreFrequency) {
	mockDevice := GetStaticMockDevice(arch, 0)

	freq, err := mockDevice.CoreFrequency()
	assert.NoError(t, err)

	for i := 0; i < int(expected.raw.PeCount); i++ {
		assert.Equal(t, expected.raw.Pe[i].Core, freq.PeFrequency()[i].Core())
		assert.Equal(t, expected.raw.Pe[i].Frequency, freq.PeFrequency()[i].Frequency())
	}
}

func TestCoreFrequency(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    coreFrequency
	}{
		{
			description: "Test Warboy Core Frequency",
			arch:        ArchWarboy,
			expected: func() coreFrequency {
				exp := coreFrequency{binding.FuriosaSmiCoreFrequency{PeCount: 2, Pe: [64]binding.FuriosaSmiPeFrequency{}}}
				for i := 0; i < 2; i++ {
					exp.raw.Pe[i] = binding.FuriosaSmiPeFrequency{Core: uint32(i), Frequency: 2000}
				}

				return exp
			}(),
		},
		{
			description: "Test RNGD Core Frequency",
			arch:        ArchRngd,
			expected: func() coreFrequency {
				exp := coreFrequency{binding.FuriosaSmiCoreFrequency{PeCount: 8, Pe: [64]binding.FuriosaSmiPeFrequency{}}}
				for i := 0; i < 8; i++ {
					exp.raw.Pe[i] = binding.FuriosaSmiPeFrequency{Core: uint32(i), Frequency: 500}
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
			description: "Test Warboy Memory Frequency",
			arch:        ArchWarboy,
			expected:    4266,
		},
		{
			description: "Test RNGD Memory Frequency",
			arch:        ArchRngd,
			expected:    6000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testMemoryFrequency(t, tc.arch, tc.expected)
		})
	}
}

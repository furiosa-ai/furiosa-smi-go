package smi

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
	"github.com/stretchr/testify/assert"
)

func testDeviceErrorInfo(t *testing.T, arch Arch, expected DeviceErrorInfo) {
	mockDevice := GetStaticMockDevice(arch, 0)

	deviceErrInfo, err := mockDevice.DeviceErrorInfo()
	assert.NoErrorf(t, err, "Failed to get Device Error Info")

	assert.Equal(t, expected.AxiPostErrorCount(), deviceErrInfo.AxiPostErrorCount())
	assert.Equal(t, expected.AxiFetchErrorCount(), deviceErrInfo.AxiFetchErrorCount())
	assert.Equal(t, expected.AxiDiscardErrorCount(), deviceErrInfo.AxiDiscardErrorCount())
	assert.Equal(t, expected.AxiDoorbellErrorCount(), deviceErrInfo.AxiDoorbellErrorCount())
	assert.Equal(t, expected.PciePostErrorCount(), deviceErrInfo.PciePostErrorCount())
	assert.Equal(t, expected.PcieFetchErrorCount(), deviceErrInfo.PcieFetchErrorCount())
	assert.Equal(t, expected.PcieDiscardErrorCount(), deviceErrInfo.PcieDiscardErrorCount())
	assert.Equal(t, expected.PcieDoorbellErrorCount(), deviceErrInfo.PcieDoorbellErrorCount())
	assert.Equal(t, expected.DeviceErrorCount(), deviceErrInfo.DeviceErrorCount())
}

func TestDeviceErrorInfo(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    DeviceErrorInfo
	}{
		{
			description: "Test Warboy Device Error Info",
			arch:        ArchWarboy,
			expected: newDeviceErrorInfo(
				binding.FuriosaSmiDeviceErrorInfo{
					AxiPostErrorCount:      0,
					AxiFetchErrorCount:     0,
					AxiDiscardErrorCount:   0,
					AxiDoorbellErrorCount:  0,
					PciePostErrorCount:     0,
					PcieFetchErrorCount:    0,
					PcieDiscardErrorCount:  0,
					PcieDoorbellErrorCount: 0,
					DeviceErrorCount:       0,
				},
			),
		},
		{
			description: "Test RNGD Device Error Info",
			arch:        ArchRngd,
			expected: newDeviceErrorInfo(
				binding.FuriosaSmiDeviceErrorInfo{
					AxiPostErrorCount:      0,
					AxiFetchErrorCount:     0,
					AxiDiscardErrorCount:   0,
					AxiDoorbellErrorCount:  0,
					PciePostErrorCount:     0,
					PcieFetchErrorCount:    0,
					PcieDiscardErrorCount:  0,
					PcieDoorbellErrorCount: 0,
					DeviceErrorCount:       0,
				},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceErrorInfo(t, tc.arch, tc.expected)
		})
	}
}

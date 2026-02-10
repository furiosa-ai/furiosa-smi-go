package smi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testDeviceInfo(t *testing.T, arch Arch, expected DeviceInfo) {
	mockDevice := GetStaticMockDevice(arch, 0)

	devInfo, err := mockDevice.DeviceInfo()
	assert.NoError(t, err)

	assert.Equal(t, expected.Arch(), devInfo.Arch())
	assert.Equal(t, expected.CoreNum(), devInfo.CoreNum())
	assert.Equal(t, expected.NumaNode(), devInfo.NumaNode())
	assert.Equal(t, expected.Name(), devInfo.Name())
	assert.Equal(t, expected.Serial(), devInfo.Serial())
	assert.Equal(t, expected.UUID(), devInfo.UUID())
	assert.Equal(t, expected.BDF(), devInfo.BDF())
	assert.Equal(t, expected.Major(), devInfo.Major())
	assert.Equal(t, expected.Minor(), devInfo.Minor())
	assert.Equal(t, expected.FirmwareVersion().Major(), devInfo.FirmwareVersion().Major())
	assert.Equal(t, expected.FirmwareVersion().Minor(), devInfo.FirmwareVersion().Minor())
	assert.Equal(t, expected.FirmwareVersion().Patch(), devInfo.FirmwareVersion().Patch())
	assert.Equal(t, expected.FirmwareVersion().Metadata(), devInfo.FirmwareVersion().Metadata())
	assert.Equal(t, expected.FirmwareVersion().Prerelease(), devInfo.FirmwareVersion().Prerelease())
}

func stringTo96ByteArray(str string) [96]byte {
	var arr [96]byte
	copy(arr[:], str)
	return arr
}

func TestDeviceInfo(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    DeviceInfo
	}{
		{
			description: "Test RNGD Device Info",
			arch:        FuriosaSmiArchRngd,
			expected: &FuriosaSmiDeviceInfo{
				arch:     FuriosaSmiArchRngd,
				coreNum:  8,
				numaNode: 0,
				name:     "npu0",
				serial:   "TEST0236FH505KRE0",
				uuid:     "A76AAD68-6855-40B1-9E86-D080852D1C80",
				bdf:      "0000:27:00.0",
				major:    234,
				minor:    0,
				firmwareVersion: &FuriosaSmiVersion{
					major:      1,
					minor:      6,
					patch:      0,
					metadata:   "c1bebfd",
					prerelease: "dev0",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceInfo(t, tc.arch, tc.expected)
		})
	}
}

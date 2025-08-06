package smi

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
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
	assert.Equal(t, expected.PertVersion().Major(), devInfo.PertVersion().Major())
	assert.Equal(t, expected.PertVersion().Minor(), devInfo.PertVersion().Minor())
	assert.Equal(t, expected.PertVersion().Patch(), devInfo.PertVersion().Patch())
	assert.Equal(t, expected.PertVersion().Metadata(), devInfo.PertVersion().Metadata())
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
			arch:        ArchRngd,
			expected: newDeviceInfo(
				binding.FuriosaSmiDeviceInfo{
					Arch:     binding.FuriosaSmiArchRngd,
					CoreNum:  8,
					NumaNode: 0,
					Name:     stringTo96ByteArray("npu0"),
					Serial:   stringTo96ByteArray("TEST0236FH505KRE0"),
					Uuid:     stringTo96ByteArray("A76AAD68-6855-40B1-9E86-D080852D1C80"),
					Bdf:      stringTo96ByteArray("0000:27:00.0"),
					Major:    234,
					Minor:    0,
					FirmwareVersion: binding.FuriosaSmiVersion{
						Major:    1,
						Minor:    6,
						Patch:    0,
						Metadata: stringTo96ByteArray("c1bebfd"),
					},
					PertVersion: binding.FuriosaSmiVersion{
						Major:    0,
						Minor:    0,
						Patch:    0,
						Metadata: stringTo96ByteArray(""),
					},
				},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceInfo(t, tc.arch, tc.expected)
		})
	}
}

package smi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testDeviceFiles(t *testing.T, arch Arch, expected []DeviceFile) {
	mockDevice := GetStaticMockDevice(arch, 0)

	devFiles, err := mockDevice.DeviceFiles()
	assert.NoErrorf(t, err, "Failed to get Device Files")

	assert.Len(t, devFiles, len(expected))

	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i].Cores(), devFiles[i].Cores())
		assert.Equal(t, expected[i].Path(), devFiles[i].Path())
	}

}
func TestDeviceFiles(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    []DeviceFile
	}{
		{
			description: "Test RNGD Device Files",
			arch:        FuriosaSmiArchRngd,
			expected: []DeviceFile{
				newDeviceFile(0, 0, "/dev/rngd/npu0pe0"),
				newDeviceFile(1, 1, "/dev/rngd/npu0pe1"),
				newDeviceFile(0, 1, "/dev/rngd/npu0pe0-1"),
				newDeviceFile(2, 2, "/dev/rngd/npu0pe2"),
				newDeviceFile(3, 3, "/dev/rngd/npu0pe3"),
				newDeviceFile(2, 3, "/dev/rngd/npu0pe2-3"),
				newDeviceFile(0, 3, "/dev/rngd/npu0pe0-3"),
				newDeviceFile(4, 4, "/dev/rngd/npu0pe4"),
				newDeviceFile(5, 5, "/dev/rngd/npu0pe5"),
				newDeviceFile(4, 5, "/dev/rngd/npu0pe4-5"),
				newDeviceFile(6, 6, "/dev/rngd/npu0pe6"),
				newDeviceFile(7, 7, "/dev/rngd/npu0pe7"),
				newDeviceFile(6, 7, "/dev/rngd/npu0pe6-7"),
				newDeviceFile(4, 7, "/dev/rngd/npu0pe4-7"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceFiles(t, tc.arch, tc.expected)
		})
	}
}

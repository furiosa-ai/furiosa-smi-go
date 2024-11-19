package smi

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
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

func stringTo256ByteArray(str string) [256]byte {
	var arr [256]byte
	copy(arr[:], str)
	return arr
}

func TestDeviceFiles(t *testing.T) {
	tests := []struct {
		description string
		arch        Arch
		expected    []DeviceFile
	}{
		{
			description: "Test Warboy Device Files",
			arch:        ArchWarboy,
			expected: []DeviceFile{
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 0, CoreEnd: 0, Path: stringTo256ByteArray("/dev/npu0pe0")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 1, CoreEnd: 1, Path: stringTo256ByteArray("/dev/npu0pe1")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 0, CoreEnd: 1, Path: stringTo256ByteArray("/dev/npu0pe0-1")}),
			},
		},
		{
			description: "Test RNGD Device Files",
			arch:        ArchRngd,
			expected: []DeviceFile{
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 0, CoreEnd: 0, Path: stringTo256ByteArray("/dev/rngd/npu0pe0")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 1, CoreEnd: 1, Path: stringTo256ByteArray("/dev/rngd/npu0pe1")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 0, CoreEnd: 1, Path: stringTo256ByteArray("/dev/rngd/npu0pe0-1")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 2, CoreEnd: 2, Path: stringTo256ByteArray("/dev/rngd/npu0pe2")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 3, CoreEnd: 3, Path: stringTo256ByteArray("/dev/rngd/npu0pe3")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 2, CoreEnd: 3, Path: stringTo256ByteArray("/dev/rngd/npu0pe2-3")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 0, CoreEnd: 3, Path: stringTo256ByteArray("/dev/rngd/npu0pe0-3")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 4, CoreEnd: 4, Path: stringTo256ByteArray("/dev/rngd/npu0pe4")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 5, CoreEnd: 5, Path: stringTo256ByteArray("/dev/rngd/npu0pe5")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 4, CoreEnd: 5, Path: stringTo256ByteArray("/dev/rngd/npu0pe4-5")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 6, CoreEnd: 6, Path: stringTo256ByteArray("/dev/rngd/npu0pe6")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 7, CoreEnd: 7, Path: stringTo256ByteArray("/dev/rngd/npu0pe7")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 6, CoreEnd: 7, Path: stringTo256ByteArray("/dev/rngd/npu0pe6-7")}),
				newDeviceFile(binding.FuriosaSmiDeviceFile{
					CoreStart: 4, CoreEnd: 7, Path: stringTo256ByteArray("/dev/rngd/npu0pe4-7")}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			testDeviceFiles(t, tc.arch, tc.expected)
		})
	}
}

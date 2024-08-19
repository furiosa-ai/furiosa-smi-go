package smi

import (
	"reflect"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
)

func testDeviceFiles(arch Arch, t *testing.T, expected []DeviceFile) {
	device := getStaticMockDevice(arch, 0)

	device_files, err := device.DeviceFiles()

	if err != nil {
		t.Errorf("Failed to get Device Files")
	}

	if len(expected) != len(device_files) {
		t.Errorf("expected device files num %d but got %d", len(expected), len(device_files))
	}

	for i := 0; i < len(expected); i++ {
		if !reflect.DeepEqual(expected[i].Cores(), device_files[i].Cores()) {
			t.Errorf("expected cores %v but got %v", expected[i].Cores(), device_files[i].Cores())
		}

		if !reflect.DeepEqual(expected[i].Path(), device_files[i].Path()) {
			t.Errorf("expected path %s but got %s", expected[i].Path(), device_files[i].Path())
		}
	}

}

func stringTo256ByteArray(str string) [256]byte {
	var arr [256]byte
	copy(arr[:], str)
	return arr
}

func TestWarboyDeviceFiles(t *testing.T) {
	expected := []DeviceFile{
		newDeviceFile(binding.FuriosaSmiDeviceFile{
			CoreStart: 0, CoreEnd: 0, Path: stringTo256ByteArray("/dev/npu0pe0")}),
		newDeviceFile(binding.FuriosaSmiDeviceFile{
			CoreStart: 1, CoreEnd: 1, Path: stringTo256ByteArray("/dev/npu0pe1")}),
		newDeviceFile(binding.FuriosaSmiDeviceFile{
			CoreStart: 0, CoreEnd: 1, Path: stringTo256ByteArray("/dev/npu0pe0-1")}),
	}

	testDeviceFiles(ArchWarboy, t, expected)
}

func TestRngdDeviceFiles(t *testing.T) {
	expected := []DeviceFile{
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
	}

	testDeviceFiles(ArchRngd, t, expected)
}

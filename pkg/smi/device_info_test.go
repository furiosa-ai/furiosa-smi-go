package smi

import (
	"reflect"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
)

func testDeviceInfo(arch Arch, t *testing.T, expected DeviceInfo) {
	device := getStaticMockDevice(arch, 0)

	device_info, err := device.DeviceInfo()

	if err != nil {
		t.Errorf("Failed to get Device Information")
	}

	if !reflect.DeepEqual(expected.Arch(), device_info.Arch()) {
		t.Errorf("expected architecture %s but got %s", expected.Arch().ToString(), device_info.Arch().ToString())
	}

	if !reflect.DeepEqual(expected.CoreNum(), device_info.CoreNum()) {
		t.Errorf("expected corenum %d but got %d", expected.CoreNum(), device_info.CoreNum())
	}

	if !reflect.DeepEqual(expected.NumaNode(), device_info.NumaNode()) {
		t.Errorf("expected numanode %d but got %d", expected.NumaNode(), device_info.NumaNode())
	}

	if !reflect.DeepEqual(expected.Name(), device_info.Name()) {
		t.Errorf("expected name %s but got %s", expected.Name(), device_info.Name())
	}

	if !reflect.DeepEqual(expected.Serial(), device_info.Serial()) {
		t.Errorf("expected serial %s but got %s", expected.Serial(), device_info.Serial())
	}

	if !reflect.DeepEqual(expected.UUID(), device_info.UUID()) {
		t.Errorf("expected uuid %s but got %s", expected.UUID(), device_info.UUID())
	}

	if !reflect.DeepEqual(expected.BDF(), device_info.BDF()) {
		t.Errorf("expected bdf %s but got %s", expected.BDF(), device_info.BDF())
	}

	if !reflect.DeepEqual(expected.Major(), device_info.Major()) {
		t.Errorf("expected major %d but got %d", expected.Major(), device_info.Major())
	}

	if !reflect.DeepEqual(expected.Minor(), device_info.Minor()) {
		t.Errorf("expected minor %d but got %d", expected.Minor(), device_info.Minor())
	}

	if !reflect.DeepEqual(expected.FirmwareVersion().Arch(), device_info.FirmwareVersion().Arch()) {
		t.Errorf("expected firmware arch %s but got %s", expected.FirmwareVersion().Arch().ToString(), device_info.FirmwareVersion().Arch().ToString())
	}

	if !reflect.DeepEqual(expected.FirmwareVersion().Major(), device_info.FirmwareVersion().Major()) {
		t.Errorf("expected firmware major %d but got %d", expected.FirmwareVersion().Major(), device_info.FirmwareVersion().Major())
	}

	if !reflect.DeepEqual(expected.FirmwareVersion().Minor(), device_info.FirmwareVersion().Minor()) {
		t.Errorf("expected firmware minor %d but got %d", expected.FirmwareVersion().Minor(), device_info.FirmwareVersion().Minor())
	}

	if !reflect.DeepEqual(expected.FirmwareVersion().Patch(), device_info.FirmwareVersion().Patch()) {
		t.Errorf("expected firmware patch %d but got %d", expected.FirmwareVersion().Patch(), device_info.FirmwareVersion().Patch())
	}

	if !reflect.DeepEqual(expected.FirmwareVersion().Metadata(), device_info.FirmwareVersion().Metadata()) {
		t.Errorf("expected firmware metadata %s but got %s", expected.FirmwareVersion().Metadata(), device_info.FirmwareVersion().Metadata())
	}

	if !reflect.DeepEqual(expected.DriverVersion().Arch(), device_info.DriverVersion().Arch()) {
		t.Errorf("expected driver arch %s but got %s", expected.DriverVersion().Arch().ToString(), device_info.DriverVersion().Arch().ToString())
	}

	if !reflect.DeepEqual(expected.DriverVersion().Major(), device_info.DriverVersion().Major()) {
		t.Errorf("expected driver major %d but got %d", expected.DriverVersion().Major(), device_info.DriverVersion().Major())
	}

	if !reflect.DeepEqual(expected.DriverVersion().Minor(), device_info.DriverVersion().Minor()) {
		t.Errorf("expected driver minor %d but got %d", expected.DriverVersion().Minor(), device_info.DriverVersion().Minor())
	}

	if !reflect.DeepEqual(expected.DriverVersion().Patch(), device_info.DriverVersion().Patch()) {
		t.Errorf("expected driver patch %d but got %d", expected.DriverVersion().Patch(), device_info.DriverVersion().Patch())
	}

	if !reflect.DeepEqual(expected.DriverVersion().Metadata(), device_info.DriverVersion().Metadata()) {
		t.Errorf("expected driver metadata %s but got %s", expected.DriverVersion().Metadata(), device_info.DriverVersion().Metadata())
	}
}

func stringTo96ByteArray(str string) [96]byte {
	var arr [96]byte
	copy(arr[:], str)
	return arr
}

func TestWarboyDeviceInfo(t *testing.T) {
	expected := newDeviceInfo(binding.FuriosaSmiDeviceInfo{
		Arch:     binding.FuriosaSmiArchWarboy,
		CoreNum:  2,
		NumaNode: 0,
		Name:     stringTo96ByteArray("/dev/npu0"),
		Serial:   stringTo96ByteArray("TEST0236FH505KRE0"),
		Uuid:     stringTo96ByteArray("A76AAD68-6855-40B1-9E86-D080852D1C80"),
		Bdf:      stringTo96ByteArray("0000:27:00.0"),
		Major:    234,
		Minor:    0,
		FirmwareVersion: binding.FuriosaSmiVersion{
			Arch:     binding.FuriosaSmiArchWarboy,
			Major:    1,
			Minor:    6,
			Patch:    0,
			Metadata: stringTo96ByteArray("c1bebfd")},
		DriverVersion: binding.FuriosaSmiVersion{
			Arch:     binding.FuriosaSmiArchWarboy,
			Major:    1,
			Minor:    9,
			Patch:    2,
			Metadata: stringTo96ByteArray("3def9c2")}})
	testDeviceInfo(ArchWarboy, t, expected)
}

func TestRngdDeviceInfo(t *testing.T) {
	expected := newDeviceInfo(binding.FuriosaSmiDeviceInfo{
		Arch:     binding.FuriosaSmiArchRngd,
		CoreNum:  8,
		NumaNode: 0,
		Name:     stringTo96ByteArray("/dev/rngd/npu0"),
		Serial:   stringTo96ByteArray("TEST0236FH505KRE0"),
		Uuid:     stringTo96ByteArray("A76AAD68-6855-40B1-9E86-D080852D1C80"),
		Bdf:      stringTo96ByteArray("0000:27:00.0"),
		Major:    234,
		Minor:    0,
		FirmwareVersion: binding.FuriosaSmiVersion{
			Arch:     binding.FuriosaSmiArchRngd,
			Major:    1,
			Minor:    6,
			Patch:    0,
			Metadata: stringTo96ByteArray("c1bebfd")},
		DriverVersion: binding.FuriosaSmiVersion{
			Arch:     binding.FuriosaSmiArchRngd,
			Major:    1,
			Minor:    9,
			Patch:    2,
			Metadata: stringTo96ByteArray("3def9c2")}})
	testDeviceInfo(ArchRngd, t, expected)
}

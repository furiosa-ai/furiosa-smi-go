package smi

import (
	"reflect"
	"testing"
)

func testCoreStatus(arch Arch, t *testing.T, expected map[uint32]CoreStatus) {
	device := GetStaticMockDevice(arch, 0)

	core_status, err := device.CoreStatus()

	if err != nil {
		t.Errorf("Failed to get core status")
	}

	if !reflect.DeepEqual(expected, core_status) {
		t.Errorf("expected core status %v but got %v", expected, core_status)
	}
}

func TestWarboyCoreStatus(t *testing.T) {
	expected := make(map[uint32]CoreStatus)

	for i := 0; i < 2; i++ {
		expected[uint32(i)] = CoreStatusAvailable
	}

	testCoreStatus(ArchWarboy, t, expected)
}

func TestRngdCoreStatus(t *testing.T) {
	expected := make(map[uint32]CoreStatus)

	for i := 0; i < 8; i++ {
		expected[uint32(i)] = CoreStatusAvailable
	}

	testCoreStatus(ArchRngd, t, expected)
}

func testLiveness(arch Arch, t *testing.T, expected bool) {
	device := GetStaticMockDevice(arch, 0)

	liveness, err := device.Liveness()

	if err != nil {
		t.Errorf("Failed to get liveness")
	}

	if !reflect.DeepEqual(expected, liveness) {
		t.Errorf("expected liveness %v but got %v", expected, liveness)
	}
}

func TestWarboyLiveness(t *testing.T) {
	expected := true

	testLiveness(ArchWarboy, t, expected)
}

func TestRngdLiveness(t *testing.T) {
	expected := true

	testLiveness(ArchRngd, t, expected)
}

func testGetDeviceToDeviceLinkType(device0 Device, device1 Device, t *testing.T, expected LinkType) {

	linktype, err := device0.GetDeviceToDeviceLinkType(device1)

	if err != nil {
		t.Errorf("Failed to get linktype")
	}

	if !reflect.DeepEqual(expected, linktype) {
		t.Errorf("expected linktype %v but got %v", expected, linktype)
	}
}

func TestWarboyGetDeviceToDeviceLinkTypeNoc(t *testing.T) {
	expected := LinkTypeNoc

	device0 := GetStaticMockDevice(ArchWarboy, 0)

	testGetDeviceToDeviceLinkType(device0, device0, t, expected)
}

func TestRngdGetDeviceToDeviceLinkTypeNoc(t *testing.T) {
	expected := LinkTypeNoc

	device0 := GetStaticMockDevice(ArchRngd, 0)

	testGetDeviceToDeviceLinkType(device0, device0, t, expected)
}

func TestWarboyGetDeviceToDeviceLinkTypeHostBridge(t *testing.T) {
	expected := LinkTypeHostBridge

	device0 := GetStaticMockDevice(ArchWarboy, 0)
	device1 := GetStaticMockDevice(ArchWarboy, 1)

	testGetDeviceToDeviceLinkType(device0, device1, t, expected)
}

func TestRngdGetDeviceToDeviceLinkTypeHostBridge(t *testing.T) {
	expected := LinkTypeHostBridge

	device0 := GetStaticMockDevice(ArchRngd, 0)
	device1 := GetStaticMockDevice(ArchRngd, 1)

	testGetDeviceToDeviceLinkType(device0, device1, t, expected)
}

func TestWarboyGetDeviceToDeviceLinkTypeCpu(t *testing.T) {
	expected := LinkTypeCpu

	device0 := GetStaticMockDevice(ArchWarboy, 0)
	device1 := GetStaticMockDevice(ArchWarboy, 2)

	testGetDeviceToDeviceLinkType(device0, device1, t, expected)
}

func TestRngdGetDeviceToDeviceLinkTypeCpu(t *testing.T) {
	expected := LinkTypeCpu

	device0 := GetStaticMockDevice(ArchRngd, 0)
	device1 := GetStaticMockDevice(ArchRngd, 2)

	testGetDeviceToDeviceLinkType(device0, device1, t, expected)
}

func TestWarboyGetDeviceToDeviceLinkTypeInterconnect(t *testing.T) {
	expected := LinkTypeInterconnect

	device0 := GetStaticMockDevice(ArchWarboy, 0)
	device1 := GetStaticMockDevice(ArchWarboy, 4)

	testGetDeviceToDeviceLinkType(device0, device1, t, expected)
}

func TestRngdGetDeviceToDeviceLinkTypeInterconnect(t *testing.T) {
	expected := LinkTypeInterconnect

	device0 := GetStaticMockDevice(ArchRngd, 0)
	device1 := GetStaticMockDevice(ArchRngd, 4)

	testGetDeviceToDeviceLinkType(device0, device1, t, expected)
}

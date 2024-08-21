package smi

import (
	"reflect"
	"testing"
)

func testCoreStatus(arch Arch, t *testing.T, expected map[uint32]CoreStatus) {
	mockdevice := GetStaticMockDevice(arch, 0)

	core_status, err := mockdevice.CoreStatus()

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
	mockdevice := GetStaticMockDevice(arch, 0)

	liveness, err := mockdevice.Liveness()

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

func testGetDeviceToDeviceLinkType(devices []Device, t *testing.T, expectedMap map[int]map[int]LinkType) {

	for i, device0 := range devices {
		for j, device1 := range devices {
			linktype, err := device0.GetDeviceToDeviceLinkType(device1)

			if err != nil {
				t.Errorf("Failed to get linktype")
			}

			idx0, idx1 := i, j
			if i > j {
				idx0, idx1 = j, i
			}

			expected := expectedMap[idx0][idx1]
			if !reflect.DeepEqual(expected, linktype) {
				t.Errorf("expected linktype between npu%d, npu%d is %v but got %v", i, j, expected, linktype)
			}
		}
	}
}

func TestWarboyGetDeviceToDeviceLinkType(t *testing.T) {
	devices := GetStaticMockDevices(ArchRngd)

	testGetDeviceToDeviceLinkType(devices, t, linkTypeHintMap)
}

func TestRngdGetDeviceToDeviceLinkType(t *testing.T) {
	devices := GetStaticMockDevices(ArchRngd)

	testGetDeviceToDeviceLinkType(devices, t, linkTypeHintMap)
}

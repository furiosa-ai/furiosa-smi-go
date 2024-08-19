package smi

import (
	"reflect"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
)

func testDeviceErrorInfo(arch Arch, t *testing.T, expected DeviceErrorInfo) {
	device := getStaticMockDevice(arch, 0)

	device_error_info, err := device.DeviceErrorInfo()

	if err != nil {
		t.Errorf("Failed to get Device Error Info")
	}

	if !reflect.DeepEqual(expected.AxiPostErrorCount(), device_error_info.AxiPostErrorCount()) {
		t.Errorf("expected AxiPostErrorCount %v but got %v", expected.AxiPostErrorCount(), device_error_info.AxiPostErrorCount())
	}

	if !reflect.DeepEqual(expected.AxiFetchErrorCount(), device_error_info.AxiFetchErrorCount()) {
		t.Errorf("expected AxiFetchErrorCount %v but got %v", expected.AxiFetchErrorCount(), device_error_info.AxiFetchErrorCount())
	}

	if !reflect.DeepEqual(expected.AxiDiscardErrorCount(), device_error_info.AxiDiscardErrorCount()) {
		t.Errorf("expected AxiDiscardErrorCount %v but got %v", expected.AxiDiscardErrorCount(), device_error_info.AxiDiscardErrorCount())
	}

	if !reflect.DeepEqual(expected.AxiDoorbellErrorCount(), device_error_info.AxiDoorbellErrorCount()) {
		t.Errorf("expected AxiDoorbellErrorCount %v but got %v", expected.AxiDoorbellErrorCount(), device_error_info.AxiDoorbellErrorCount())
	}

	if !reflect.DeepEqual(expected.PciePostErrorCount(), device_error_info.PciePostErrorCount()) {
		t.Errorf("expected PciePostErrorCount %v but got %v", expected.PciePostErrorCount(), device_error_info.PciePostErrorCount())
	}

	if !reflect.DeepEqual(expected.PcieFetchErrorCount(), device_error_info.PcieFetchErrorCount()) {
		t.Errorf("expected PcieFetchErrorCount %v but got %v", expected.PcieFetchErrorCount(), device_error_info.PcieFetchErrorCount())
	}

	if !reflect.DeepEqual(expected.PcieDiscardErrorCount(), device_error_info.PcieDiscardErrorCount()) {
		t.Errorf("expected PcieDiscardErrorCount %v but got %v", expected.PcieDiscardErrorCount(), device_error_info.PcieDiscardErrorCount())
	}

	if !reflect.DeepEqual(expected.PcieDoorbellErrorCount(), device_error_info.PcieDoorbellErrorCount()) {
		t.Errorf("expected PcieDoorbellErrorCount %v but got %v", expected.PcieDoorbellErrorCount(), device_error_info.PcieDoorbellErrorCount())
	}
}

func TestWarboyDeviceErrorInfo(t *testing.T) {
	expected := newDeviceErrorInfo(binding.FuriosaSmiDeviceErrorInfo{
		AxiPostErrorCount:      1,
		AxiFetchErrorCount:     2,
		AxiDiscardErrorCount:   3,
		AxiDoorbellErrorCount:  4,
		PciePostErrorCount:     5,
		PcieFetchErrorCount:    6,
		PcieDiscardErrorCount:  7,
		PcieDoorbellErrorCount: 8,
		DeviceErrorCount:       9})

	testDeviceErrorInfo(ArchWarboy, t, expected)
}

func TestRngdDeviceErrorInfo(t *testing.T) {
	expected := newDeviceErrorInfo(binding.FuriosaSmiDeviceErrorInfo{
		AxiPostErrorCount:      1,
		AxiFetchErrorCount:     2,
		AxiDiscardErrorCount:   3,
		AxiDoorbellErrorCount:  4,
		PciePostErrorCount:     5,
		PcieFetchErrorCount:    6,
		PcieDiscardErrorCount:  7,
		PcieDoorbellErrorCount: 8,
		DeviceErrorCount:       9})

	testDeviceErrorInfo(ArchRngd, t, expected)
}

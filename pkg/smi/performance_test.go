package smi

import (
	"reflect"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"
)

func testDeviceTemperature(arch Arch, t *testing.T, expected deviceTemperature) {
	device := getStaticMockDevice(arch, 0)

	temperature, err := device.DeviceTemperature()
	if err != nil {
		t.Errorf("Failed to get Device Temperature")
	}

	if !reflect.DeepEqual(expected.SocPeak(), temperature.SocPeak()) {
		t.Errorf("expected SocPeak %f but got %f", expected.SocPeak(), temperature.SocPeak())
	}
	if !reflect.DeepEqual(expected.Ambient(), temperature.Ambient()) {
		t.Errorf("expected Ambient %f but got %f", expected.Ambient(), temperature.Ambient())
	}
}

func TestWarboyDeviceTemperature(t *testing.T) {
	expected := deviceTemperature{binding.FuriosaSmiDeviceTemperature{SocPeak: 20.0, Ambient: 10.0}}

	testDeviceTemperature(ArchWarboy, t, expected)
}

func TestRngdDeviceTemperature(t *testing.T) {
	expected := deviceTemperature{binding.FuriosaSmiDeviceTemperature{SocPeak: 20.0, Ambient: 10.0}}

	testDeviceTemperature(ArchRngd, t, expected)
}

func testPowerConsumption(arch Arch, t *testing.T, expected float64) {
	device := getStaticMockDevice(arch, 0)

	power, err := device.PowerConsumption()
	if err != nil {
		t.Errorf("Failed to get Power Consumption")
	}

	if !reflect.DeepEqual(expected, power) {
		t.Errorf("expected power consumption %f but got %f", expected, power)
	}
}

func TestWarboyPowerConsumption(t *testing.T) {
	expected := 100.0

	testPowerConsumption(ArchWarboy, t, expected)
}

func TestRngdPowerConsumption(t *testing.T) {
	expected := 100.0

	testPowerConsumption(ArchRngd, t, expected)
}

func testDeviceUtilization(arch Arch, t *testing.T) {
	device := getStaticMockDevice(arch, 0)

	_, err := device.DeviceUtilization() // Currenlty, not to check the value.

	if err != nil {
		t.Errorf("Failed to get Device Utilization")
	}
}

func TestWarboyDeviceUtilization(t *testing.T) {
	testDeviceUtilization(ArchWarboy, t)
}

func TestRngdDeviceUtilization(t *testing.T) {
	testDeviceUtilization(ArchRngd, t)
}

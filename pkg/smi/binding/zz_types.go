// WARNING: This file has automatically been generated
// Code generated by https://git.io/c-for-go. DO NOT EDIT.

package binding

/*
#include "furiosa/furiosa_smi.h"
#include <stdlib.h>
#include "cgo_helpers.h"
*/
import "C"

type FuriosaSmiObserver C.FuriosaSmiObserver

type FuriosaSmiDeviceHandle uint32

type FuriosaSmiDeviceHandles struct {
	Count         uint32
	DeviceHandles [64]FuriosaSmiDeviceHandle
}

type FuriosaSmiVersion struct {
	Arch     FuriosaSmiArch
	Major    uint32
	Minor    uint32
	Patch    uint32
	Metadata [96]byte
}

type FuriosaSmiDeviceInfo struct {
	Arch            FuriosaSmiArch
	CoreNum         uint32
	NumaNode        uint32
	Name            [96]byte
	Serial          [96]byte
	Uuid            [96]byte
	Bdf             [96]byte
	Major           uint16
	Minor           uint16
	FirmwareVersion FuriosaSmiVersion
	DriverVersion   FuriosaSmiVersion
}

type FuriosaSmiDeviceFile struct {
	CoreStart uint32
	CoreEnd   uint32
	Path      [256]byte
}

type FuriosaSmiDeviceFiles struct {
	Count       uint32
	DeviceFiles [64]FuriosaSmiDeviceFile
}

type FuriosaSmiCoreStatuses struct {
	Count      uint32
	CoreStatus [128]FuriosaSmiCoreStatus
}

type FuriosaSmiDeviceErrorInfo struct {
	AxiPostErrorCount      uint32
	AxiFetchErrorCount     uint32
	AxiDiscardErrorCount   uint32
	AxiDoorbellErrorCount  uint32
	PciePostErrorCount     uint32
	PcieFetchErrorCount    uint32
	PcieDiscardErrorCount  uint32
	PcieDoorbellErrorCount uint32
	DeviceErrorCount       uint32
}

type FuriosaSmiDriverInfo struct {
	Count      uint32
	DriverInfo [24]FuriosaSmiVersion
}

type FuriosaSmiPeUtilization struct {
	Core              uint32
	TimeWindowMil     uint32
	PeUsagePercentage float64
}

type FuriosaSmiMemoryUtilization struct {
	TotalBytes uint64
	InUseBytes uint64
}

type FuriosaSmiDeviceUtilization struct {
	PeCount uint32
	Pe      [64]FuriosaSmiPeUtilization
	Memory  FuriosaSmiMemoryUtilization
}

type FuriosaSmiDevicePowerConsumption struct {
	RmsTotal float64
}

type FuriosaSmiDeviceTemperature struct {
	SocPeak float64
	Ambient float64
}

package rngd

// RNGDMgmtConfig, RNGDDevfsOps, RNGDMgmtOps, and RNGDHwmonOps satisfy
// internal/arch/config.MgmtConfig and internal/device's
// MgmtOps/DevfsOps/HwmonOps interfaces for RNGD/RNGD+.

import (
	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/version"
)

// Management-node sysfs attribute names. mgmt.go reads these, mock.go
// writes these same names when building a fake device, so a single shared
// constant per attribute keeps production reads and mock writes from
// silently drifting apart. attrBusname/attrDeviceUUID/attrDeviceSN also
// back the config.MgmtConfig interface's AttrBusname/AttrDeviceUUID/
// AttrDeviceSN getters, since discovery.go reads those two generically
// across architectures; the rest are read directly by RNGD-specific
// functions here and never cross that interface boundary, but get the same
// treatment for consistency -- the drift risk is identical either way.
const (
	attrBusname        = "busname"
	attrDeviceUUID     = "device_uuid"
	attrDeviceSN       = "device_sn"
	attrFwVersion      = "fw_version"
	attrDriverVersion  = "version"
	attrDev            = "dev"
	attrPEOccupancy    = "pe_occupancy"
	attrNPUClocks      = "npu_clocks"
	attrGovernor       = "npu_governor"
	attrDeviceState    = "device_state"
	attrThrottleReason = "throttle_reason"
)

// RNGDMgmtConfig implements internal/arch/config.MgmtConfig.
type RNGDMgmtConfig struct{}

func NewRNGDMgmtConfig() *RNGDMgmtConfig { return &RNGDMgmtConfig{} }

func (*RNGDMgmtConfig) ClassRoot() string      { return RNGDManagementRoot() }
func (*RNGDMgmtConfig) AttrBusname() string    { return attrBusname }
func (*RNGDMgmtConfig) AttrDeviceUUID() string { return attrDeviceUUID }
func (*RNGDMgmtConfig) AttrDeviceSN() string   { return attrDeviceSN }

func (*RNGDMgmtConfig) ParseNodeIndex(entryName string) *uint8 {
	return ParseMgmtNodeIndex(entryName)
}

func (*RNGDMgmtConfig) FormatDeviceName(devNodeIndex uint8) string {
	return FormatDeviceName(devNodeIndex)
}

func (*RNGDMgmtConfig) FormatDebugfsPath(devNodeIndex uint8) string {
	return FormatDebugfsMgmtPath(devNodeIndex)
}

// RNGDDevfsOps implements pkg/device.DevfsOps.
type RNGDDevfsOps struct{}

func NewRNGDDevfsOps() *RNGDDevfsOps { return &RNGDDevfsOps{} }

func (*RNGDDevfsOps) GetDevNodeIndices(root string) map[uint8]struct{} {
	return GetDevNodeIndices(root)
}

func (*RNGDDevfsOps) GetCoreCount(devRoot string, devNodeIndex uint8) uint32 {
	return GetCoreCount(devRoot, devNodeIndex)
}

func (*RNGDDevfsOps) GetAccessibleCoreIndices(devRoot string, devNodeIndex uint8) map[uint32]struct{} {
	return GetAccessibleCoreIndices(devRoot, devNodeIndex)
}

func (*RNGDDevfsOps) GetDeviceFiles(devRoot string, devNodeIndex uint8) []device.DeviceFile {
	return GetDeviceFiles(devRoot, devNodeIndex)
}

func (*RNGDDevfsOps) HasP2PInterfaces(devRoot string, devNodeIndex uint8) bool {
	return HasP2PInterfaces(devRoot, devNodeIndex)
}

// RNGDMgmtOps implements pkg/device.MgmtOps.
type RNGDMgmtOps struct{}

func NewRNGDMgmtOps() *RNGDMgmtOps { return &RNGDMgmtOps{} }

func (*RNGDMgmtOps) ReadFirmwareVersion(mgmtPath string) (version.VersionInfo, error) {
	return ReadFirmwareVersion(mgmtPath)
}

func (*RNGDMgmtOps) ReadDriverVersion(mgmtPath string) (version.VersionInfo, error) {
	return ReadDriverVersion(mgmtPath)
}

func (*RNGDMgmtOps) ReadMajorMinor(mgmtPath string) (uint16, uint16, error) {
	return ReadMajorMinor(mgmtPath)
}

func (*RNGDMgmtOps) ReadCoreStatus(mgmtPath string) (map[uint32]device.CoreStatus, error) {
	return ReadCoreStatus(mgmtPath)
}

func (*RNGDMgmtOps) ReadCoreFrequency(mgmtPath string) (map[uint32]uint32, error) {
	return ReadCoreFrequency(mgmtPath)
}

func (*RNGDMgmtOps) ReadGovernorProfile(mgmtPath string) (device.GovernorProfile, error) {
	return ReadGovernorProfile(mgmtPath)
}

func (*RNGDMgmtOps) WriteGovernorProfile(mgmtPath string, profile device.GovernorProfile) error {
	return WriteGovernorProfile(mgmtPath, profile)
}

func (*RNGDMgmtOps) ReadLiveness(mgmtPath string) bool {
	return ReadLiveness(mgmtPath)
}

func (*RNGDMgmtOps) ReadPerformanceCounter(mgmtPath string, accessible map[uint32]struct{}, coreStatus map[uint32]device.CoreStatus) (map[uint32]device.PePerformanceCounter, error) {
	return ReadPerformanceCounter(mgmtPath, accessible, coreStatus)
}

func (*RNGDMgmtOps) ReadMemoryUsage(mgmtPath string, accessible map[uint32]struct{}, coreStatus map[uint32]device.CoreStatus, memoryConfig sysfs.MemoryConfig) (map[device.DeviceCoreRange]device.MemoryInfo, error) {
	return ReadMemoryUsage(mgmtPath, accessible, coreStatus, memoryConfig)
}

func (*RNGDMgmtOps) ReadThrottleReason(mgmtPath string) ([]device.ThrottleReason, error) {
	return ReadThrottleReason(mgmtPath)
}

func (*RNGDMgmtOps) WritePowerLimit(debugfsMgmtPath string, powerLimitMW uint32) error {
	return WritePowerLimit(debugfsMgmtPath, powerLimitMW)
}

// RNGDHwmonOps implements pkg/device.HwmonOps.
type RNGDHwmonOps struct{}

func NewRNGDHwmonOps() *RNGDHwmonOps { return &RNGDHwmonOps{} }

func (*RNGDHwmonOps) ReadDeviceTemperature(hwmonDir string) (device.DeviceTemperature, error) {
	return ReadDeviceTemperature(hwmonDir)
}

func (*RNGDHwmonOps) ReadPowerConsumption(hwmonDir string) (float64, error) {
	return ReadPowerConsumption(hwmonDir)
}

func (*RNGDHwmonOps) ReadPowerLimit(hwmonDir string) (uint32, error) {
	return ReadPowerLimit(hwmonDir)
}

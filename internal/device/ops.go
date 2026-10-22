package device

// furiosa-smi/src/device.rs

import (
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/version"
)

// MgmtOps is the set of architecture-specific operations for reading and
// writing an already-located management node -- Go's equivalent of the
// Rust source's SysfsMgmtOps. Unlike device.rs, which re-resolves these via
// retrieve_arch_config(self.device_id()) on every call, Go stores an
// implementation as a Device field, injected once at construction by
// internal/discovery: Go's package-import-cycle rules (unlike Rust's
// single-crate compilation) don't allow Device itself to import
// internal/arch/config, since that package needs this very type. One
// concrete implementation exists per architecture family (e.g.
// rngd.RNGDMgmtOps).
type MgmtOps interface {
	ReadFirmwareVersion(mgmtPath string) (version.VersionInfo, error)
	ReadDriverVersion(mgmtPath string) (version.VersionInfo, error)
	ReadMajorMinor(mgmtPath string) (uint16, uint16, error)
	ReadCoreStatus(mgmtPath string) (map[uint32]CoreStatus, error)
	ReadCoreFrequency(mgmtPath string) (map[uint32]uint32, error)
	ReadGovernorProfile(mgmtPath string) (GovernorProfile, error)
	WriteGovernorProfile(mgmtPath string, profile GovernorProfile) error
	ReadLiveness(mgmtPath string) bool
	ReadPerformanceCounter(mgmtPath string, accessible map[uint32]struct{}, coreStatus map[uint32]CoreStatus) (map[uint32]PePerformanceCounter, error)
	ReadMemoryUsage(mgmtPath string, accessible map[uint32]struct{}, coreStatus map[uint32]CoreStatus, memoryConfig sysfs.MemoryConfig) (map[DeviceCoreRange]MemoryInfo, error)
	ReadThrottleReason(mgmtPath string) ([]ThrottleReason, error)
	WritePowerLimit(debugfsMgmtPath string, powerLimitMW uint32) error
}

// HwmonOps is the set of architecture-specific operations for reading a
// device's hwmon sensors -- Go's equivalent of SysfsHwmonOps, injected the
// same way as MgmtOps.
type HwmonOps interface {
	ReadDeviceTemperature(hwmonDir string) (DeviceTemperature, error)
	ReadPowerConsumption(hwmonDir string) (float64, error)
	ReadPowerLimit(hwmonDir string) (uint32, error)
}

// DevfsOps is the set of architecture-specific operations for enumerating a
// device's devfs nodes. internal/arch/config also uses GetDevNodeIndices
// and GetCoreCount at discovery time, before any Device exists -- matching
// the Rust source, where ArchConfig.devfs_ops is one shared table used by
// both discovery.rs and device.rs.
type DevfsOps interface {
	GetDevNodeIndices(root string) map[uint8]struct{}
	GetCoreCount(devRoot string, devNodeIndex uint8) uint32
	GetAccessibleCoreIndices(devRoot string, devNodeIndex uint8) map[uint32]struct{}
	GetDeviceFiles(devRoot string, devNodeIndex uint8) []DeviceFile
	HasP2PInterfaces(devRoot string, devNodeIndex uint8) bool
}

// Ops bundles everything architecture-specific a Device needs at runtime,
// resolved once by internal/discovery via internal/arch/config.RetrieveConfig
// and injected through NewDevice.
type Ops struct {
	MgmtPath        string
	DebugfsMgmtPath string
	DevRoot         string
	DevfsOps        DevfsOps
	MgmtOps         MgmtOps
	HwmonOps        HwmonOps
	MemoryConfig    sysfs.MemoryConfig
}

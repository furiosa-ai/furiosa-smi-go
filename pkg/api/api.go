package api

// furiosa-smi/src/lib.rs

import (
	"github.com/furiosa-ai/furiosa-smi-go/internal/arch"
	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/discovery"
	"github.com/furiosa-ai/furiosa-smi-go/internal/pcie"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/topology"
	"github.com/furiosa-ai/furiosa-smi-go/internal/version"
)

type (
	// Arch represents the architecture type of a device.
	Arch = arch.Arch
	// BDF is a PCI Bus/Device/Function address (e.g., `0000:27:00.0`).
	BDF = bdf.BDF
	// CoreStatus represents a core status.
	CoreStatus = device.CoreStatus
	// Device is an available Furiosa NPU device with fully resolved device information.
	Device = device.Device
	// DeviceCoreRange is a range of NPU cores designated by a device file name (e.g., `npu0pe0-3`).
	DeviceCoreRange = device.DeviceCoreRange
	// DeviceFile is a device file under a Furiosa NPU device.
	DeviceFile = device.DeviceFile
	// DeviceTemperature is the temperature information of a device.
	DeviceTemperature = device.DeviceTemperature
	ErrorKind         = smierrors.ErrorKind
	// GovernorProfile represents a governor profile of the device.
	GovernorProfile = device.GovernorProfile
	// LinkType represents a device-to-device link type.
	LinkType = topology.LinkType
	// MemoryInfo is the DRAM usage of a PE fusion group.
	MemoryInfo = device.MemoryInfo
	// PCIDevice is a PCI device recognized by the system with Furiosa vendor ID.
	PCIDevice = device.PCIDevice
	// PcieDeviceInfo is PCIe device information.
	PcieDeviceInfo = pcie.PcieDeviceInfo
	// PcieInfo is PCIe information.
	PcieInfo = pcie.PcieInfo
	// PcieLinkInfo is PCIe link information.
	PcieLinkInfo = pcie.PcieLinkInfo
	// PcieRootComplexInfo is PCIe root complex information.
	PcieRootComplexInfo = pcie.PcieRootComplexInfo
	// PePerformanceCounter is a performance counter snapshot.
	PePerformanceCounter = device.PePerformanceCounter
	// SMIError represents an error from the Furiosa SMI library.
	SMIError = smierrors.SMIError
	// SriovInfo is SR-IOV information.
	SriovInfo = pcie.SriovInfo
	// ThrottleReason represents a throttle reason.
	ThrottleReason = device.ThrottleReason
	// VersionInfo is version information.
	VersionInfo = version.VersionInfo
)

const (
	// ArchRNGD is the RNGD architecture.
	ArchRNGD = arch.RNGD
)

const (
	// CoreStatusAvailable means a core is available.
	CoreStatusAvailable = device.CoreStatusAvailable
	// CoreStatusOccupied means a core is occupied.
	CoreStatusOccupied = device.CoreStatusOccupied
)

const (
	// GovernorProfilePerformance is the Performance governor profile.
	GovernorProfilePerformance = device.GovernorProfilePerformance
	// GovernorProfilePowerSave is the PowerSave governor profile.
	GovernorProfilePowerSave = device.GovernorProfilePowerSave
)

const (
	// LinkTypeUnknown means the connection type is unidentified.
	LinkTypeUnknown = topology.LinkTypeUnknown
	// LinkTypeInterconnect means devices communicate via inter-socket links (e.g., QPI, GMI).
	LinkTypeInterconnect = topology.LinkTypeInterconnect
	// LinkTypeCPU means devices communicate exclusively within a single CPU socket.
	LinkTypeCPU = topology.LinkTypeCPU
	// LinkTypeBridge means devices communicate via one or more PCIe switches.
	LinkTypeBridge = topology.LinkTypeBridge
	// LinkTypeNoC means the connection is within the same NPU chip.
	LinkTypeNoC = topology.LinkTypeNoC
)

const (
	// ErrorKindUnknownArch means a device architecture is unknown.
	ErrorKindUnknownArch = smierrors.ErrorKindUnknownArch
	// ErrorKindParse means a parsing operation fails.
	ErrorKindParse = smierrors.ErrorKindParse
	// ErrorKindIO means a filesystem operation fails.
	ErrorKindIO = smierrors.ErrorKindIO
	// ErrorKindDeviceNotAvailable means a device is not available (driver unbound, no mgmt node, etc.).
	ErrorKindDeviceNotAvailable = smierrors.ErrorKindDeviceNotAvailable
	// ErrorKindDeviceNotFound means a device is not found with the given identifier.
	ErrorKindDeviceNotFound = smierrors.ErrorKindDeviceNotFound
	// ErrorKindUnexpectedValue means a device reports a value inconsistent with the expected state.
	ErrorKindUnexpectedValue = smierrors.ErrorKindUnexpectedValue
)

const (
	// ThrottleReasonIdle means throttling in idle or unused state.
	ThrottleReasonIdle = device.ThrottleReasonIdle
	// ThrottleReasonThermalSlowdown means throttling triggered by high temperature.
	ThrottleReasonThermalSlowdown = device.ThrottleReasonThermalSlowdown
	// ThrottleReasonAppPowerCap means throttling due to host-defined power limit.
	ThrottleReasonAppPowerCap = device.ThrottleReasonAppPowerCap
	// ThrottleReasonAppClockCap means throttling due to host-defined clock limit.
	ThrottleReasonAppClockCap = device.ThrottleReasonAppClockCap
	// ThrottleReasonHwClockCap means throttling from device-internal clock limit.
	ThrottleReasonHwClockCap = device.ThrottleReasonHwClockCap
	// ThrottleReasonHwBusLimit means throttling from internal bus/NoC bandwidth limit.
	ThrottleReasonHwBusLimit = device.ThrottleReasonHwBusLimit
	// ThrottleReasonHwPowerCap means throttling from device-enforced power limit.
	ThrottleReasonHwPowerCap = device.ThrottleReasonHwPowerCap
	// ThrottleReasonGovClockCap means throttling due to governor-defined clock limit.
	ThrottleReasonGovClockCap = device.ThrottleReasonGovClockCap
	// ThrottleReasonOtherReason means throttling due to other undefined reasons.
	ThrottleReasonOtherReason = device.ThrottleReasonOtherReason
)

// ParseBDF parses either "dddd:bb:dd.f" or "bb:dd.f" notation.
func ParseBDF(value string) (BDF, error) {
	return bdf.ParseBDF(value)
}

// NewDeviceCoreRange creates a new device core range.
func NewDeviceCoreRange(devNodeIndex uint8, startCore, endCore uint32) DeviceCoreRange {
	return device.NewDeviceCoreRange(devNodeIndex, startCore, endCore)
}

// ParseDeviceCoreRange parses a device file name such as "npu0pe0-3".
func ParseDeviceCoreRange(name string) (DeviceCoreRange, error) {
	return device.ParseDeviceCoreRange(name)
}

// ListDevices lists all available Furiosa NPU devices in the system.
func ListDevices() ([]Device, error) {
	return discovery.ListDevices()
}

// ListPCIDevices lists all Furiosa PCI devices recognized by the system.
func ListPCIDevices() ([]PCIDevice, error) {
	return discovery.ListPCIDevices()
}

// GetPCIDevice gets a Furiosa PCI device by its BDF address.
func GetPCIDevice(target BDF) (PCIDevice, error) {
	return discovery.GetPCIDevice(target)
}

// ListDisabledDevices lists all disabled Furiosa NPU devices in the system.
func ListDisabledDevices() ([]PCIDevice, error) {
	return discovery.ListDisabledDevices()
}

// DriverInfo returns the driver version information.
func DriverInfo() (VersionInfo, error) {
	return discovery.DriverInfo()
}

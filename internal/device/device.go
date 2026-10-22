package device

// furiosa-smi/src/device.rs

import (
	"github.com/furiosa-ai/furiosa-smi-go/internal/arch"
	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/pcie"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/topology"
	"github.com/furiosa-ai/furiosa-smi-go/internal/version"
)

// DeviceInfo holds a Device's identity fields resolved at discovery time
// (UUID, serial, core count, ...) -- Go's equivalent of Rust's DeviceInfo,
// kept separate so Device itself stays a thin (pciDevice, info, ops) triple
// instead of a flat bag of fields.
type DeviceInfo struct {
	devNodeIndex uint8
	arch         arch.Arch
	name         string
	uuid         string
	serial       string
	coreNum      uint32
	numaNode     int32
}

func NewDeviceInfo(devNodeIndex uint8, deviceArch arch.Arch, name, uuid, serial string, coreNum uint32, numaNode int32) DeviceInfo {
	return DeviceInfo{
		devNodeIndex: devNodeIndex,
		arch:         deviceArch,
		name:         name,
		uuid:         uuid,
		serial:       serial,
		coreNum:      coreNum,
		numaNode:     numaNode,
	}
}

// Device is a Furiosa NPU device with fully resolved identity (UUID,
// serial, core count, ...); PCIDevice alone doesn't carry this.
type Device struct {
	pciDevice PCIDevice
	info      DeviceInfo
	ops       Ops
}

func NewDevice(pciDevice PCIDevice, info DeviceInfo, ops Ops) Device {
	return Device{
		pciDevice: pciDevice,
		info:      info,
		ops:       ops,
	}
}

// BDF returns the PCI BDF address of the device.
func (d *Device) BDF() bdf.BDF {
	return d.pciDevice.BDF()
}

// DeviceID returns the PCI device ID.
func (d *Device) DeviceID() uint16 {
	return d.pciDevice.DeviceID()
}

// VendorID returns the PCI vendor ID.
func (d *Device) VendorID() uint16 {
	return d.pciDevice.VendorID()
}

// UUID returns the UUID of the device.
func (d *Device) UUID() string {
	return d.info.uuid
}

// Serial returns the serial number of the device.
func (d *Device) Serial() string {
	return d.info.serial
}

// CoreNum returns the number of cores of the device.
func (d *Device) CoreNum() uint32 {
	return d.info.coreNum
}

// NUMANode returns the NUMA node of the device.
func (d *Device) NUMANode() int32 {
	return d.info.numaNode
}

// DevNodeIndex returns the device node index.
func (d *Device) DevNodeIndex() uint8 {
	return d.info.devNodeIndex
}

// FirmwareVersion returns the firmware version of the device.
func (d *Device) FirmwareVersion() (version.VersionInfo, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return version.VersionInfo{}, err
	}

	return d.ops.MgmtOps.ReadFirmwareVersion(mgmtPath)
}

func (d *Device) driverVersion() (version.VersionInfo, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return version.VersionInfo{}, err
	}

	return d.ops.MgmtOps.ReadDriverVersion(mgmtPath)
}

// DriverVersion returns d's driver version. This is a package-level
// function, not a Device method, because Rust's driver_version is
// pub(crate) -- reachable only from the crate-internal driver_info() free
// function, never directly by an external caller -- and Go has no
// method-level visibility narrower than the whole package to mirror that
// with. Keeping driverVersion unexported and only reachable through this
// function is the closest Go equivalent; the find-first-available-device
// policy stays in the caller (pkg/api.DriverInfo), same as before.
func DriverVersion(d Device) (version.VersionInfo, error) {
	return d.driverVersion()
}

// Arch returns the architecture of the device.
func (d *Device) Arch() arch.Arch {
	return d.info.arch
}

// ProductName returns the marketing name for this device's device ID.
func (d *Device) ProductName() string {
	switch d.DeviceID() {
	case arch.RNGDDeviceID:
		return "RNGD"

	case arch.RNGDPlusDeviceID:
		return "RNGD+"

	default:
		return "Unknown"
	}
}

// DeviceName returns the device node name (e.g., "npu0").
func (d *Device) DeviceName() string {
	return d.info.name
}

// Major returns the major part of the device number.
func (d *Device) Major() (uint16, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return 0, err
	}

	major, _, err := d.ops.MgmtOps.ReadMajorMinor(mgmtPath)

	return major, err
}

// Minor returns the minor part of the device number.
func (d *Device) Minor() (uint16, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return 0, err
	}

	_, minor, err := d.ops.MgmtOps.ReadMajorMinor(mgmtPath)

	return minor, err
}

// DeviceTemperature returns the temperature of the device.
func (d *Device) DeviceTemperature() (DeviceTemperature, error) {
	hwmonPath, err := d.resolveHwmon()
	if err != nil {
		return DeviceTemperature{}, err
	}

	return d.ops.HwmonOps.ReadDeviceTemperature(hwmonPath)
}

// PowerConsumption returns the power consumption of the device in Watts.
func (d *Device) PowerConsumption() (float64, error) {
	hwmonPath, err := d.resolveHwmon()
	if err != nil {
		return 0, err
	}

	return d.ops.HwmonOps.ReadPowerConsumption(hwmonPath)
}

// PowerLimit returns the device's current power limit in milliwatts.
func (d *Device) PowerLimit() (uint32, error) {
	hwmonPath, err := d.resolveHwmon()
	if err != nil {
		return 0, err
	}

	return d.ops.HwmonOps.ReadPowerLimit(hwmonPath)
}

// SetPowerLimit sets the device's power limit in milliwatts. This requires
// root privileges.
func (d *Device) SetPowerLimit(powerLimitMW uint32) error {
	debugfsMgmtPath, err := d.resolveDebugfsMgmt()
	if err != nil {
		return err
	}

	return d.ops.MgmtOps.WritePowerLimit(debugfsMgmtPath, powerLimitMW)
}

// CoreStatus examines each core of the device, whether it is available or not.
func (d *Device) CoreStatus() (map[uint32]CoreStatus, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return nil, err
	}

	allStatuses, err := d.ops.MgmtOps.ReadCoreStatus(mgmtPath)
	if err != nil {
		return nil, err
	}

	// Filter to assigned PEs only; not all PEs may be visible in
	// containerized environments.
	accessible := d.ops.DevfsOps.GetAccessibleCoreIndices(d.ops.DevRoot, d.info.devNodeIndex)
	result := make(map[uint32]CoreStatus, len(accessible))
	for core, status := range allStatuses {
		if _, ok := accessible[core]; ok {
			result[core] = status
		}
	}

	return result, nil
}

// CoreFrequency returns the per-core frequency in MHz for each accessible PE.
func (d *Device) CoreFrequency() (map[uint32]uint32, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return nil, err
	}

	allFrequencies, err := d.ops.MgmtOps.ReadCoreFrequency(mgmtPath)
	if err != nil {
		return nil, err
	}

	accessible := d.ops.DevfsOps.GetAccessibleCoreIndices(d.ops.DevRoot, d.info.devNodeIndex)
	result := make(map[uint32]uint32, len(accessible))
	for core, freq := range allFrequencies {
		if _, ok := accessible[core]; ok {
			result[core] = freq
		}
	}

	return result, nil
}

// TCPFrequency returns the per-core TCP frequency in MHz for each
// accessible PE -- half of CoreFrequency, per the hardware's TCP/NPU clock
// relationship.
func (d *Device) TCPFrequency() (map[uint32]uint32, error) {
	coreFrequencies, err := d.CoreFrequency()
	if err != nil {
		return nil, err
	}

	result := make(map[uint32]uint32, len(coreFrequencies))
	for core, freq := range coreFrequencies {
		result[core] = freq / 2
	}

	return result, nil
}

// GovernorProfile returns the current governor profile of the device.
func (d *Device) GovernorProfile() (GovernorProfile, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return 0, err
	}

	return d.ops.MgmtOps.ReadGovernorProfile(mgmtPath)
}

// SetGovernorProfile sets the governor profile of the device. This requires root privileges.
func (d *Device) SetGovernorProfile(profile GovernorProfile) error {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return err
	}

	return d.ops.MgmtOps.WriteGovernorProfile(mgmtPath, profile)
}

// Liveness swallows any resolve error into false rather than propagating
// it, matching the Rust source's unwrap_or(false).
func (d *Device) Liveness() bool {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return false
	}

	return d.ops.MgmtOps.ReadLiveness(mgmtPath)
}

// PerformanceCounter returns a performance counter snapshot for each accessible PE.
func (d *Device) PerformanceCounter() (map[uint32]PePerformanceCounter, error) {
	coreStatus, err := d.CoreStatus()
	if err != nil {
		return nil, err
	}

	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return nil, err
	}

	accessible := d.ops.DevfsOps.GetAccessibleCoreIndices(d.ops.DevRoot, d.info.devNodeIndex)

	return d.ops.MgmtOps.ReadPerformanceCounter(mgmtPath, accessible, coreStatus)
}

// MemoryUsage returns the DRAM usage (including shared DRAM) of the device, keyed by fusion PE range.
func (d *Device) MemoryUsage() (map[DeviceCoreRange]MemoryInfo, error) {
	coreStatus, err := d.CoreStatus()
	if err != nil {
		return nil, err
	}

	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return nil, err
	}

	accessible := d.ops.DevfsOps.GetAccessibleCoreIndices(d.ops.DevRoot, d.info.devNodeIndex)

	return d.ops.MgmtOps.ReadMemoryUsage(mgmtPath, accessible, coreStatus, d.ops.MemoryConfig)
}

// DeviceToDeviceLinkType returns the link type between this device and another device.
func (d *Device) DeviceToDeviceLinkType(other *Device) (topology.LinkType, error) {
	if other == nil {
		return topology.LinkTypeUnknown, smierrors.UnexpectedValue("other device is nil")
	}

	return topology.GetLinkType(d.BDF(), other.BDF())
}

// ThrottleReason returns every currently-active throttle reason of the
// device.
func (d *Device) ThrottleReason() ([]ThrottleReason, error) {
	mgmtPath, err := d.resolveMgmt()
	if err != nil {
		return nil, err
	}

	return d.ops.MgmtOps.ReadThrottleReason(mgmtPath)
}

// DeviceFiles lists this device's devfs entries -- the real host paths
// (e.g. "/dev/rngd/npu0pe0-3") a caller like a device plugin or CDI spec
// generator injects into a container.
func (d *Device) DeviceFiles() ([]DeviceFile, error) {
	devRoot, err := d.resolveDevfs()
	if err != nil {
		return nil, err
	}

	return d.ops.DevfsOps.GetDeviceFiles(devRoot, d.info.devNodeIndex), nil
}

// P2PAccessible reports whether both this device and other have their P2P
// DMA interfaces available in devfs.
func (d *Device) P2PAccessible(other *Device) bool {
	if other == nil {
		return false
	}

	return d.hasP2PInterfaces() && other.hasP2PInterfaces()
}

// hasP2PInterfaces checks the devfs P2P interface files directly, not
// gated on driver-bound availability like most other accessors -- matching
// the Rust source, where p2p_accessible never calls ensure_available.
//
// A zero-value Device has no DevfsOps and reports false rather than
// panicking.
func (d *Device) hasP2PInterfaces() bool {
	if d.ops.DevfsOps == nil {
		return false
	}

	return d.ops.DevfsOps.HasP2PInterfaces(d.ops.DevRoot, d.info.devNodeIndex)
}

// PcieInfo returns the device's PCIe topology and link information. Unlike
// most other accessors, this doesn't gate on device availability -- PCIe
// sysfs attributes are readable even for a disabled device.
func (d *Device) PcieInfo() (pcie.PcieInfo, error) {
	return pcie.ReadPcieInfo(d.BDF())
}

// resolveMgmt is the availability gate every mgmt-node-reading method goes
// through first. The mgmt path itself is precomputed once at construction
// (it's a pure function of devNodeIndex), so all that's left to check here
// is whether the device is still available.
func (d *Device) resolveMgmt() (string, error) {
	if err := d.ensureAvailable(); err != nil {
		return "", err
	}

	return d.ops.MgmtPath, nil
}

// resolveHwmon is the availability gate every hwmon-reading method goes
// through first.
func (d *Device) resolveHwmon() (string, error) {
	if err := d.ensureAvailable(); err != nil {
		return "", err
	}

	return sysfs.FindHwmonDir(d.BDF())
}

// resolveDebugfsMgmt is the availability gate every debugfs-mgmt-writing
// method goes through first.
func (d *Device) resolveDebugfsMgmt() (string, error) {
	if err := d.ensureAvailable(); err != nil {
		return "", err
	}

	return d.ops.DebugfsMgmtPath, nil
}

// resolveDevfs is the availability gate every devfs-reading method goes
// through first.
func (d *Device) resolveDevfs() (string, error) {
	if err := d.ensureAvailable(); err != nil {
		return "", err
	}

	return d.ops.DevRoot, nil
}

func (d *Device) ensureAvailable() error {
	if !sysfs.IsDriverBound(d.BDF()) {
		return smierrors.DeviceNotAvailable(d.BDF().String())
	}

	return nil
}

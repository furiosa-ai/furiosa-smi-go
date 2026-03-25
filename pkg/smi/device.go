package smi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ListDevices lists all Furiosa NPU devices in the system.
//
// This mirrors the discovery flow in furiosa-smi/src/initialize/mod.rs:
//  1. Scan /dev/rngd for character device files matching npu{N}pe{S}[-{E}]
//     to discover node indices and their accessible PE cores.
//  2. For each node index (sorted ascending), read all device attributes from
//     /sys/class/rngd_mgmt/rngd!npu{N}mgmt/ and build a fully-populated Device.
//  3. Devices whose sysfs attributes are unreadable (e.g. unbound) are skipped,
//     matching the Rust behaviour of filtering out invalid device contexts.
func ListDevices() ([]Device, error) {
	nodeMap, err := scanRngdDevFiles()
	if err != nil {
		return nil, err
	}

	nodeIndices := make([]uint32, 0, len(nodeMap))
	for idx := range nodeMap {
		nodeIndices = append(nodeIndices, idx)
	}
	sort.Slice(nodeIndices, func(i, j int) bool { return nodeIndices[i] < nodeIndices[j] })

	devices := make([]Device, 0, len(nodeIndices))
	for i, nodeIdx := range nodeIndices {
		info, err := buildDeviceInfo(uint32(i), nodeIdx, nodeMap[nodeIdx])
		if err != nil {
			continue
		}
		devices = append(devices, &FuriosaSmiDevice{bdf: info.bdf, info: *info})
	}

	return devices, nil
}

// scanRngdDevFiles scans rngdDevRootPath for character device files matching
// npu{N}pe{S}[-{E}] and returns a map of nodeIdx → sorted unique core list.
// This mirrors search_furiosa_device + filter_dev_files in discovery.rs.
// A missing /dev/rngd directory (ENOENT) is treated as "no devices present"
// and returns an empty map — matching legacy ListDevices behaviour on systems
// without Furiosa hardware.
func scanRngdDevFiles() (map[uint32][]uint32, error) {
	entries, err := os.ReadDir(rngdDevRootPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[uint32][]uint32{}, nil
		}
		return nil, err
	}

	nodeCores := make(map[uint32]map[uint32]struct{})

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if !isRngdDeviceFile(info) {
			continue
		}

		parsed, ok := parseDeviceFilename(e.Name())
		if !ok {
			continue
		}

		if _, exists := nodeCores[parsed.nodeIdx]; !exists {
			nodeCores[parsed.nodeIdx] = make(map[uint32]struct{})
		}
		for c := parsed.coreStart; c <= parsed.coreEnd; c++ {
			nodeCores[parsed.nodeIdx][c] = struct{}{}
		}
	}

	result := make(map[uint32][]uint32, len(nodeCores))
	for nodeIdx, coreSet := range nodeCores {
		cores := make([]uint32, 0, len(coreSet))
		for c := range coreSet {
			cores = append(cores, c)
		}
		sort.Slice(cores, func(i, j int) bool { return cores[i] < cores[j] })
		result[nodeIdx] = cores
	}

	return result, nil
}

// buildDeviceInfo reads all sysfs attributes for device node nodeIdx and
// constructs a FuriosaSmiDeviceInfo. index is the sequential position in the
// sorted device list (matches DeviceContext.index in the Rust implementation).
//
// Attribute → sysfs file mapping (under rngd!npu{N}mgmt/):
//
//	bdf             ← busname
//	uuid            ← device_uuid
//	serial          ← device_sn
//	major, minor    ← dev        ("235:0")
//	firmwareVersion ← fw_version ("1.6.0, c1bebfd")
//	numaNode        ← /sys/bus/pci/devices/{bdf}/numa_node
func buildDeviceInfo(index, nodeIdx uint32, cores []uint32) (*FuriosaSmiDeviceInfo, error) {
	mgmtDir := filepath.Join(rngdMgmtRoot(), fmt.Sprintf("rngd!npu%dmgmt", nodeIdx))

	bdf, err := readMgmtAttr(mgmtDir, "busname")
	if err != nil {
		return nil, err
	}
	uuid, err := readMgmtAttr(mgmtDir, "device_uuid")
	if err != nil {
		return nil, err
	}
	serial, err := readMgmtAttr(mgmtDir, "device_sn")
	if err != nil {
		return nil, err
	}
	devStr, err := readMgmtAttr(mgmtDir, "dev")
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(devStr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("%w: %q", ErrParse, devStr)
	}
	maj, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}
	min, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}
	fwStr, err := readMgmtAttr(mgmtDir, "fw_version")
	if err != nil {
		return nil, err
	}
	fwVersion, err := parseVersionInfo(fwStr)
	if err != nil {
		return nil, err
	}
	numaData, err := os.ReadFile(filepath.Join(pciDevicesRoot(), bdf, "numa_node"))
	if err != nil {
		return nil, err
	}
	numaNode, err := strconv.ParseInt(strings.TrimSpace(string(numaData)), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}

	return &FuriosaSmiDeviceInfo{
		index:           index,
		arch:            ArchRngd,
		coreNum:         uint32(len(cores)),
		numaNode:        int32(numaNode),
		name:            fmt.Sprintf("npu%d", nodeIdx),
		serial:          serial,
		uuid:            uuid,
		bdf:             bdf,
		major:           uint16(maj),
		minor:           uint16(min),
		firmwareVersion: fwVersion,
	}, nil
}

// ListDisabledDevices lists all disabled Furiosa NPU devices in the system. It returns a list of BDF strings representing the disabled devices.
func ListDisabledDevices() ([]string, error) {
	//logic here

	return nil, nil
}

// EnableDevice enables a Furiosa NPU device by bdf. This requires root privileges.
func EnableDevice(bdf string) error {
	//logic here

	return nil
}

// DisableDevice disables a Furiosa NPU device by bdf. This requires root privileges.
func DisableDevice(bdf string) error {
	//logic here

	return nil
}

// DriverInfo returns the driver version by reading the "version" sysfs
// attribute from the first reachable rngd management interface.
//
// This mirrors system/mod.rs:furiosa_smi_get_driver_info in furiosa-smi:
// iterate over valid device contexts, read version from sysfs, parse and return
// the first success. Returns ErrDeviceNotFound when no device is readable.
func DriverInfo() (VersionInfo, error) {
	entries, err := os.ReadDir(rngdMgmtRoot())
	if err != nil || len(entries) == 0 {
		return nil, ErrDeviceNotFound
	}

	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "rngd!npu") || !strings.HasSuffix(name, "mgmt") {
			continue
		}
		raw, err := readMgmtAttr(filepath.Join(rngdMgmtRoot(), name), "version")
		if err != nil {
			continue
		}
		if v, err := parseVersionInfo(raw); err == nil {
			return v, nil
		}
	}

	return nil, ErrDeviceNotFound
}

func CreateObserverWithOpt(opt ObserverOpt) (Observer, error) {
	return newObserverWithOpt(opt)
}

func CreateDefaultObserver() (Observer, error) {
	opt, err := NewOptForObserver()
	if err != nil {
		return nil, err
	}

	return newObserverWithOpt(opt)
}

// Device represents the abstraction for a single Furiosa NPU device.
type Device interface {
	// DeviceInfo returns `DeviceInfo` which contains information about NPU device. (e.g. arch, serial, ...)
	DeviceInfo() (DeviceInfo, error)
	// DeviceFiles list device files under this device.
	DeviceFiles() ([]DeviceFile, error)
	// CoreStatus examine each core of the device, whether it is occupied or available.
	CoreStatus() (CoreStatuses, error)
	// Liveness returns a liveness state of the device.
	Liveness() (bool, error)
	// CoreFrequency returns a core frequency (MHz) of the device.
	CoreFrequency() (CoreFrequency, error)
	// MemoryFrequency returns a memory frequency (MHz) of the device.
	MemoryFrequency() (MemoryFrequency, error)
	// PowerConsumption returns a power consumption of the device.
	PowerConsumption() (float64, error)
	// DeviceTemperature returns a temperature of the device.
	DeviceTemperature() (DeviceTemperature, error)
	// DeviceToDeviceLinkType returns a device link type between two devices.
	DeviceToDeviceLinkType(target Device) (LinkType, error)
	// P2PAccessible returns whether two devices are p2p accessible each other or not.
	P2PAccessible(target Device) (bool, error)
	// DevicePerformanceCounter returns a performance counter of the device.
	DevicePerformanceCounter() (DevicePerformanceCounter, error)
	// GovernorProfile returns a governor profile of the device.
	GovernorProfile() (GovernorProfile, error)
	// SetGovernorProfile set a governor profile of the device.
	SetGovernorProfile(governorProfile GovernorProfile) error
	// PcieInfo returns a PCIe information of the device.
	PcieInfo() (PcieInfo, error)
	// ThrottleReason returns a throttle reason of the device.
	ThrottleReason() (ThrottleReason, error)
	// MemoryUtilization returns a memory utilization of the device.
	MemoryUtilization() (MemoryUtilization, error)
}

var _ Device = new(FuriosaSmiDevice)

type FuriosaSmiDevice struct {
	bdf  string
	info FuriosaSmiDeviceInfo
}

func newDevice(bdf string, info FuriosaSmiDeviceInfo) Device {
	return &FuriosaSmiDevice{bdf: bdf, info: info}
}

func (d *FuriosaSmiDevice) DeviceInfo() (DeviceInfo, error) {
	return &d.info, nil
}

func (d *FuriosaSmiDevice) DeviceFiles() ([]DeviceFile, error) {
	out, err := FuriosaSmiGetDeviceFiles(d)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return out.DeviceFiles(), nil
}

func (d *FuriosaSmiDevice) CoreStatus() (CoreStatuses, error) {
	// logic here
	return nil, nil
}

func (d *FuriosaSmiDevice) Liveness() (bool, error) {
	// logic here
	return false, nil
}

func (d *FuriosaSmiDevice) CoreFrequency() (CoreFrequency, error) {
	// logic here
	return nil, nil
}

func (d *FuriosaSmiDevice) MemoryFrequency() (MemoryFrequency, error) {
	// logic here
	return nil, nil
}

func (d *FuriosaSmiDevice) PowerConsumption() (float64, error) {
	// logic here
	return 0, nil
}

func (d *FuriosaSmiDevice) DeviceTemperature() (DeviceTemperature, error) {
	// logic here
	return nil, nil
}

func (d *FuriosaSmiDevice) DeviceToDeviceLinkType(target Device) (LinkType, error) {
	// logic here
	return LinkTypeUnknown, nil
}

func (d *FuriosaSmiDevice) P2PAccessible(target Device) (bool, error) {
	// logic here
	return false, nil
}

func (d *FuriosaSmiDevice) DevicePerformanceCounter() (DevicePerformanceCounter, error) {
	// logic here
	return nil, nil
}

func (d *FuriosaSmiDevice) GovernorProfile() (GovernorProfile, error) {
	// logic here
	return 0, nil
}

func (d *FuriosaSmiDevice) SetGovernorProfile(profile GovernorProfile) error {
	// logic here
	return nil
}

func (d *FuriosaSmiDevice) PcieInfo() (PcieInfo, error) {
	var outPcieDeviceInfo FuriosaSmiPcieDeviceInfo
	var outPcieLinkInfo FuriosaSmiPcieLinkInfo
	var outSriovInfo FuriosaSmiSriovInfo
	var outPcieRootComplexInfo FuriosaSmiPcieRootComplexInfo
	var outPcieSwitchInfo FuriosaSmiPcieSwitchInfo

	if err := FuriosaSmiGetPcieDeviceInfo(d, &outPcieDeviceInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetPcieLinkInfo(d, &outPcieLinkInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetSriovInfo(d, &outSriovInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetPcieRootComplexInfo(d, &outPcieRootComplexInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetPcieSwitchInfo(d, &outPcieSwitchInfo); err != nil {
		return nil, err
	}

	return newPcieInfo(&outPcieDeviceInfo, &outPcieLinkInfo, &outSriovInfo, &outPcieRootComplexInfo, &outPcieSwitchInfo), nil
}

func (d *FuriosaSmiDevice) ThrottleReason() (ThrottleReason, error) {
	// logic here
	return 0, nil
}

func (d *FuriosaSmiDevice) MemoryUtilization() (MemoryUtilization, error) {
	// logic here
	return nil, nil
}

package discovery

// furiosa-smi/src/discovery.rs

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch"
	"github.com/furiosa-ai/furiosa-smi-go/internal/arch/config"
	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
	"github.com/furiosa-ai/furiosa-smi-go/internal/version"
)

type mgmtInfo struct {
	devNodeIndex uint8
	mgmtPath     string
}

// ListDevices lists all available Furiosa NPU devices in the system.
func ListDevices() ([]device.Device, error) {
	pciDevices, err := enumerateFuriosaPCIDevices()
	if err != nil {
		return nil, err
	}

	deviceIDs := make(map[uint16]struct{})
	for _, d := range pciDevices {
		deviceIDs[d.DeviceID()] = struct{}{}
	}

	mgmtMap := getMgmtNodes(deviceIDs)
	assignedNodes := getDevNodes(deviceIDs)

	devices := make([]device.Device, 0)
	for _, pci := range pciDevices {
		if d := resolveDevice(pci, mgmtMap, assignedNodes); d != nil {
			devices = append(devices, *d)
		}
	}

	slices.SortFunc(devices, func(a, b device.Device) int {
		return cmp.Or(
			cmp.Compare(a.VendorID(), b.VendorID()),
			cmp.Compare(a.DeviceID(), b.DeviceID()),
			cmp.Compare(a.DevNodeIndex(), b.DevNodeIndex()),
		)
	})

	return devices, nil
}

// ListPCIDevices lists all Furiosa PCI devices recognized by the system.
func ListPCIDevices() ([]device.PCIDevice, error) {
	devices, err := enumerateFuriosaPCIDevices()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(devices, device.PCIDevice.Compare)

	return devices, nil
}

// GetPCIDevice gets a Furiosa PCI device by its BDF address.
func GetPCIDevice(target bdf.BDF) (device.PCIDevice, error) {
	path := filepath.Join(sysroot.PCIDevicesRoot(), target.String())
	pciDevice := parseFuriosaPCIDevice(path, target)
	if pciDevice == nil {
		return device.PCIDevice{}, smierrors.DeviceNotFound(fmt.Sprintf("PCI device %s not found", target.String()))
	}

	return *pciDevice, nil
}

// ListDisabledDevices lists all disabled Furiosa NPU devices in the
// system. Rust's equivalent (drain::list_disabled_devices) lives in its
// own module; the Go port has no such module, so this stays alongside its
// sibling PCI-device listers here instead.
func ListDisabledDevices() ([]device.PCIDevice, error) {
	pciDevices, err := ListPCIDevices()
	if err != nil {
		return nil, err
	}

	disabled := make([]device.PCIDevice, 0, len(pciDevices))
	for _, pciDevice := range pciDevices {
		if !sysfs.IsDriverBound(pciDevice.BDF()) {
			disabled = append(disabled, pciDevice)
		}
	}

	return disabled, nil
}

func enumerateFuriosaPCIDevices() ([]device.PCIDevice, error) {
	root := sysroot.PCIDevicesRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, smierrors.IO(root, err)
	}

	devices := make([]device.PCIDevice, 0)
	for _, entry := range entries {
		parsedBDF, err := bdf.ParseBDF(entry.Name())
		if err != nil {
			continue
		}

		if pciDevice := parseFuriosaPCIDevice(filepath.Join(root, entry.Name()), parsedBDF); pciDevice != nil {
			devices = append(devices, *pciDevice)
		}
	}

	return devices, nil
}

func parseFuriosaPCIDevice(path string, target bdf.BDF) *device.PCIDevice {
	vendorID := sysfs.ReadSysfsHexU16(filepath.Join(path, "vendor"))
	if vendorID == nil || *vendorID != arch.FuriosaVendorID {
		return nil
	}

	deviceID := sysfs.ReadSysfsHexU16(filepath.Join(path, "device"))
	if deviceID == nil {
		return nil
	}

	// Filter unsupported devices.
	if _, err := config.RetrieveArchConfig(*deviceID); err != nil {
		return nil
	}

	pciDevice := device.NewPCIDevice(target, *deviceID, *vendorID)

	return &pciDevice
}

func getMgmtNodes(deviceIDs map[uint16]struct{}) map[bdf.BDF]mgmtInfo {
	result := make(map[bdf.BDF]mgmtInfo)
	for deviceID := range deviceIDs {
		config, err := config.RetrieveArchConfig(deviceID)
		if err != nil {
			continue
		}

		maps.Copy(result, getMgmtRoot(config))
	}

	return result
}

func getMgmtRoot(config config.ArchConfig) map[bdf.BDF]mgmtInfo {
	entries, err := os.ReadDir(config.Mgmt.ClassRoot())
	if err != nil {
		return nil
	}

	result := make(map[bdf.BDF]mgmtInfo)
	for _, entry := range entries {
		if e := parseMgmtEntry(config, entry.Name()); e != nil {
			result[e.target] = e.info
		}
	}

	return result
}

// mgmtEntry is parseMgmtEntry's result: the BDF a management node belongs
// to, plus the mgmtInfo parsed from that node.
type mgmtEntry struct {
	target bdf.BDF
	info   mgmtInfo
}

func parseMgmtEntry(config config.ArchConfig, name string) *mgmtEntry {
	devNodeIndex := config.Mgmt.ParseNodeIndex(name)
	if devNodeIndex == nil {
		return nil
	}

	entryPath := filepath.Join(config.Mgmt.ClassRoot(), name)
	busname, err := sysfs.ReadSysfsAttr(filepath.Join(entryPath, config.Mgmt.AttrBusname()))
	if err != nil {
		return nil
	}

	target, err := bdf.ParseBDF(busname)
	if err != nil {
		return nil
	}

	return &mgmtEntry{target: target, info: mgmtInfo{devNodeIndex: *devNodeIndex, mgmtPath: entryPath}}
}

func getDevNodes(deviceIDs map[uint16]struct{}) map[uint16]map[uint8]struct{} {
	result := make(map[uint16]map[uint8]struct{})
	for deviceID := range deviceIDs {
		config, err := config.RetrieveArchConfig(deviceID)
		if err != nil {
			continue
		}

		result[deviceID] = config.DevfsOps.GetDevNodeIndices(config.DevRoot)
	}

	return result
}

func resolveDevice(pciDevice device.PCIDevice, mgmtMap map[bdf.BDF]mgmtInfo, assignedNodes map[uint16]map[uint8]struct{}) *device.Device {
	mgmt, ok := mgmtMap[pciDevice.BDF()]
	if !ok {
		return nil
	}

	if !isAssigned(assignedNodes, pciDevice.DeviceID(), mgmt.devNodeIndex) {
		return nil
	}

	return readDeviceInfo(mgmt, pciDevice)
}

func isAssigned(assignedNodes map[uint16]map[uint8]struct{}, deviceID uint16, nodeIndex uint8) bool {
	nodes, ok := assignedNodes[deviceID]
	if !ok {
		return false
	}

	_, exists := nodes[nodeIndex]

	return exists
}

func readDeviceInfo(mgmt mgmtInfo, pciDevice device.PCIDevice) *device.Device {
	deviceID := pciDevice.DeviceID()
	config, err := config.RetrieveArchConfig(deviceID)
	if err != nil {
		return nil
	}

	uuid, err := sysfs.ReadSysfsAttr(filepath.Join(mgmt.mgmtPath, config.Mgmt.AttrDeviceUUID()))
	if err != nil {
		return nil
	}

	serial, err := sysfs.ReadSysfsAttr(filepath.Join(mgmt.mgmtPath, config.Mgmt.AttrDeviceSN()))
	if err != nil {
		return nil
	}

	coreNum := config.DevfsOps.GetCoreCount(config.DevRoot, mgmt.devNodeIndex)
	numaNode, err := readNUMANode(pciDevice.BDF())
	if err != nil {
		return nil
	}

	name := config.Mgmt.FormatDeviceName(mgmt.devNodeIndex)
	ops := device.Ops{
		MgmtPath:        mgmt.mgmtPath,
		DebugfsMgmtPath: config.Mgmt.FormatDebugfsPath(mgmt.devNodeIndex),
		DevRoot:         config.DevRoot,
		DevfsOps:        config.DevfsOps,
		MgmtOps:         config.MgmtOps,
		HwmonOps:        config.HwmonOps,
		MemoryConfig:    config.MemoryConfig,
	}
	info := device.NewDeviceInfo(mgmt.devNodeIndex, config.Arch, name, uuid, serial, coreNum, numaNode)
	d := device.NewDevice(pciDevice, info, ops)

	return &d
}

// DriverInfo returns the driver version information, read from the first
// available device. Rust's equivalent (driver_info) lives directly in
// lib.rs rather than any module; this is the closest Go analog now that
// pkg/api is a pure re-export layer with no logic of its own.
func DriverInfo() (version.VersionInfo, error) {
	devices, err := ListDevices()
	if err != nil {
		return version.VersionInfo{}, err
	}

	for _, d := range devices {
		if info, err := device.DriverVersion(d); err == nil {
			return info, nil
		}
	}

	return version.VersionInfo{}, smierrors.DeviceNotFound("no available device to read driver version")
}

func readNUMANode(target bdf.BDF) (int32, error) {
	path := filepath.Join(sysroot.PCIDevicesRoot(), target.String(), "numa_node")
	value, err := sysfs.ReadSysfsAttr(path)
	if err != nil {
		return 0, err
	}

	parsed, parseErr := strconv.ParseInt(value, 10, 32)
	if parseErr != nil {
		return 0, smierrors.Parse(fmt.Sprintf("invalid numa_node: '%s'", value))
	}

	return int32(parsed), nil
}

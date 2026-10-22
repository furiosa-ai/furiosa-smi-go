package rngd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/consts"
	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
	"github.com/furiosa-ai/furiosa-smi-go/internal/utils"
)

const (
	managementRootPath = "sys/class/rngd_mgmt"
	devicesRootPath    = "dev/rngd"
	debugfsRootPath    = "sys/kernel/debug/rngd"
)

var memory = sysfs.MemoryConfig{DRAMTotalSize: 48 * consts.GiB, FirmwareReservedSize: 512 * consts.MiB, PECount: 8}
var plusMemory = sysfs.MemoryConfig{DRAMTotalSize: 72 * consts.GiB, FirmwareReservedSize: 512 * consts.MiB, PECount: 8}

// Memory returns the memory config for a plain RNGD device.
func Memory() sysfs.MemoryConfig { return memory }

// PlusMemory returns the memory config for an RNGD+ device.
func PlusMemory() sysfs.MemoryConfig { return plusMemory }

var mgmtPattern = regexp.MustCompile(`^rngd!npu(\d+)mgmt$`)

// Static RNGD specific method.
// called by ListDevices() -> getMgmtNodes(DeviceIds) -> getMgmtRoot(config) ->
// config.Mgmt.ClassRoot() -> RNGDManagementRoot() in the future.
func RNGDManagementRoot() string {
	return sysroot.Join(managementRootPath)
}

// later used as ArchConfig{DevRoot: RNGDDeviceRoot()}
func RNGDDeviceRoot() string {
	return sysroot.Join(devicesRootPath)
}

// RNGDDebugfsRoot returns the debugfs directory exposing RNGD's
// root-only management interface (used for writes like power limit). See
// RNGDManagementRoot for why this lives here instead of internal/sysroot.
func RNGDDebugfsRoot() string {
	return sysroot.Join(debugfsRootPath)
}

// ParseMgmtNodeIndex extracts the device node index from a management node
// entry name (e.g. "rngd!npu3mgmt" -> 3).
func ParseMgmtNodeIndex(entryName string) *uint8 {
	matches := mgmtPattern.FindStringSubmatch(entryName)
	if matches == nil {
		return nil
	}

	value, err := strconv.ParseUint(matches[1], 10, 8)
	if err != nil {
		return nil
	}

	result := uint8(value)

	return &result
}

// FormatDeviceName returns the device name (e.g. "npu3") for devNodeIndex.
func FormatDeviceName(devNodeIndex uint8) string {
	return fmt.Sprintf("npu%d", devNodeIndex)
}

// FormatMgmtPath returns the management node path for devNodeIndex.
func FormatMgmtPath(devNodeIndex uint8) string {
	return filepath.Join(RNGDManagementRoot(), fmt.Sprintf("rngd!npu%dmgmt", devNodeIndex))
}

// FormatDebugfsMgmtPath returns the debugfs management path for
// devNodeIndex.
func FormatDebugfsMgmtPath(devNodeIndex uint8) string {
	return filepath.Join(RNGDDebugfsRoot(), fmt.Sprintf("mgmt%d", devNodeIndex))
}

// GetDevNodeIndices returns every RNGD device node index visible under
// root.
func GetDevNodeIndices(root string) map[uint8]struct{} {
	entries, err := sysroot.ReadDir(root)
	if err != nil {
		return map[uint8]struct{}{}
	}

	result := make(map[uint8]struct{})
	for _, entry := range entries {
		if !utils.IsCharDevice(entry) {
			continue
		}

		if nodeIndex := parseDevNodeIndex(entry.Name()); nodeIndex != nil {
			result[*nodeIndex] = struct{}{}
		}
	}

	return result
}

// parseDevNodeIndex reuses the public device-file parser (like the Rust
// source's parse_dev_node_index), so "npuN" bare-node names -- valid there
// -- are recognized here too, instead of enforcing a stricter, discovery-only
// pattern that requires "peS" and silently drops unpartitioned devices.
func parseDevNodeIndex(name string) *uint8 {
	r, err := device.ParseDeviceCoreRange(name)
	if err != nil {
		return nil
	}

	result := r.DevNodeIndex()

	return &result
}

func GetCoreCount(devRoot string, devNodeIndex uint8) uint32 {
	entries, err := sysroot.ReadDir(devRoot)
	if err != nil {
		return 0
	}

	prefix := fmt.Sprintf("npu%dpe", devNodeIndex)
	var count uint32
	for _, entry := range entries {
		if isSinglePE(prefix, entry.Name()) {
			count++
		}
	}

	return count
}

func isSinglePE(prefix, name string) bool {
	suffix, ok := strings.CutPrefix(name, prefix)
	if !ok {
		return false
	}

	_, err := strconv.ParseUint(suffix, 10, 32)

	return err == nil
}

func GetAccessibleCoreIndices(devRoot string, devNodeIndex uint8) map[uint32]struct{} {
	entries, err := sysroot.ReadDir(devRoot)
	if err != nil {
		return map[uint32]struct{}{}
	}

	result := make(map[uint32]struct{})
	for _, entry := range entries {
		if !utils.IsCharDevice(entry) {
			continue
		}

		coreRange := parseCoreRange(devNodeIndex, entry.Name())
		if coreRange == nil {
			continue
		}

		for _, core := range coreRange.Cores() {
			result[core] = struct{}{}
		}
	}

	return result
}

// parseCoreRange reuses the public device-file parser, matching Rust's
// parse_core_range, so it accepts every name ParseDeviceCoreRange does
// (including bare "npuN") rather than a stricter local pattern.
func parseCoreRange(devNodeIndex uint8, name string) *device.DeviceCoreRange {
	r, err := device.ParseDeviceCoreRange(name)
	if err != nil || r.DevNodeIndex() != devNodeIndex {
		return nil
	}

	return &r
}

// GetDeviceFiles lists every RNGD devfs entry for devNodeIndex under
// devRoot, sorted by core range. Paths are reported as host paths (see
// sysroot.HostPath), so under a mock root they still read "/dev/rngd/...".
func GetDeviceFiles(devRoot string, devNodeIndex uint8) []device.DeviceFile {
	entries, err := sysroot.ReadDir(devRoot)
	if err != nil {
		return nil
	}

	var files []device.DeviceFile
	for _, entry := range entries {
		if !utils.IsCharDevice(entry) {
			continue
		}

		name := entry.Name()
		coreRange := parseCoreRange(devNodeIndex, name)
		if coreRange == nil {
			continue
		}

		files = append(files, device.NewDeviceFile(*coreRange, sysroot.HostPath(filepath.Join(devRoot, name))))
	}

	slices.SortFunc(files, func(a, b device.DeviceFile) int {
		return a.CoreRange().Compare(b.CoreRange())
	})

	return files
}

// HasP2PInterfaces reports whether every mandatory P2P DMA interface
// (remote channel files + the DMA remapping file) exists under devRoot for
// devNodeIndex.
func HasP2PInterfaces(devRoot string, devNodeIndex uint8) bool {
	for _, name := range p2pMandatoryInterfaces(devNodeIndex) {
		if _, err := os.Stat(filepath.Join(devRoot, name)); err != nil {
			return false
		}
	}

	return true
}

func p2pMandatoryInterfaces(devNodeIndex uint8) []string {
	interfaces := make([]string, 0, 9)
	interfaces = append(interfaces, fmt.Sprintf("npu%ddmar", devNodeIndex))
	for channel := 0; channel <= 7; channel++ {
		interfaces = append(interfaces, fmt.Sprintf("npu%dch%dr", devNodeIndex, channel))
	}

	return interfaces
}

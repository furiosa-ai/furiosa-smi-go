package topology

// furiosa-smi/src/topology.rs

import (
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

// LinkType represents a device-to-device link type.
type LinkType int

const (
	// LinkTypeUnknown means the connection type is unidentified.
	LinkTypeUnknown LinkType = iota
	// LinkTypeInterconnect means devices communicate via inter-socket links
	// (e.g. QPI, GMI).
	LinkTypeInterconnect
	// LinkTypeCPU means devices communicate exclusively within a single
	// CPU socket.
	LinkTypeCPU
	// LinkTypeBridge means devices communicate via one or more PCIe
	// switches.
	LinkTypeBridge
	// LinkTypeNoC means the connection is within the same NPU chip.
	LinkTypeNoC
)

func (l LinkType) String() string {
	switch l {
	case LinkTypeUnknown:
		return "Unknown"

	case LinkTypeInterconnect:
		return "Interconnect"

	case LinkTypeCPU:
		return "Cpu"

	case LinkTypeBridge:
		return "Bridge"

	case LinkTypeNoC:
		return "Noc"

	default:
		return fmt.Sprintf("LinkType(%d)", int(l))
	}
}

// GetLinkType mirrors hwloc classification: the PCIe ancestor chain
// determines Bridge, then physical_package_id (CPU socket) distinguishes
// Cpu vs Interconnect. Unlike ancestor-lookup errors, a CPU-socket lookup
// failure on either side falls back to Unknown instead of propagating,
// matching the Rust source.
func GetLinkType(first, second bdf.BDF) (LinkType, error) {
	if first == second {
		return LinkTypeNoC, nil
	}

	ancestors1, err := pciAncestors(first)
	if err != nil {
		return LinkTypeUnknown, err
	}

	ancestors2, err := pciAncestors(second)
	if err != nil {
		return LinkTypeUnknown, err
	}

	for _, a := range ancestors1 {
		if slices.Contains(ancestors2, a) {
			return LinkTypeBridge, nil
		}
	}

	socket1, err1 := pciDeviceSocket(first)
	socket2, err2 := pciDeviceSocket(second)
	if err1 == nil && socket1 != nil && err2 == nil && socket2 != nil {
		if *socket1 == *socket2 {
			return LinkTypeCPU, nil
		}

		return LinkTypeInterconnect, nil
	}

	return LinkTypeUnknown, nil
}

// pciAncestors returns the PCIe device ancestor chain for the given BDF,
// self excluded. Resolves the sysfs device symlink to its real path (e.g.
// /sys/devices/pci0000:16/0000:16:00.0/0000:17:00.0/0000:18:00.0) and
// collects all path components that parse as a BDF.
func pciAncestors(target bdf.BDF) ([]bdf.BDF, error) {
	devicePath := filepath.Join(sysroot.PCIDevicesRoot(), target.String())
	real, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return nil, smierrors.IO(devicePath, err)
	}

	var ancestors []bdf.BDF
	for component := range strings.SplitSeq(real, string(filepath.Separator)) {
		value := componentAsBDF(component)
		if value == nil || *value == target {
			continue
		}

		ancestors = append(ancestors, *value)
	}

	return ancestors, nil
}

// pciDeviceSocket returns the CPU socket (physical_package_id) for a PCI
// device, if one can be determined. Reads local_cpus to find the first
// local CPU, then reads that CPU's physical_package_id.
func pciDeviceSocket(target bdf.BDF) (*uint32, error) {
	localCPUsPath := filepath.Join(sysroot.PCIDevicesRoot(), target.String(), "local_cpus")
	data, err := os.ReadFile(localCPUsPath)
	if err != nil {
		return nil, smierrors.IO(localCPUsPath, err)
	}

	firstCPU := firstCPUFromCPUMask(strings.TrimSpace(string(data)))
	if firstCPU == nil {
		return nil, nil
	}

	pkgPath := filepath.Join(sysroot.CPUDevicesRoot(), fmt.Sprintf("cpu%d", *firstCPU), "topology", "physical_package_id")
	pkgData, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, smierrors.IO(pkgPath, err)
	}

	trimmed := strings.TrimSpace(string(pkgData))
	value, parseErr := strconv.ParseUint(trimmed, 10, 32)
	if parseErr != nil {
		return nil, smierrors.Parse(fmt.Sprintf("invalid physical_package_id: '%s'", trimmed))
	}

	socket := uint32(value)

	return &socket, nil
}

// firstCPUFromCPUMask extracts the first set CPU number from a kernel
// cpumask hex string. Kernel cpumask format: comma-separated 32-bit hex
// groups, MSB first (e.g. "00000000,0000ffff" means CPUs 0-15, returns 0).
func firstCPUFromCPUMask(mask string) *uint32 {
	groups := strings.Split(mask, ",")
	for i, group := range slices.Backward(groups) {
		value, err := strconv.ParseUint(strings.TrimSpace(group), 16, 32)
		if err != nil {
			return nil
		}

		if value != 0 {
			bit := bits.TrailingZeros32(uint32(value))
			groupIndex := uint32(len(groups) - 1 - i)
			cpu := groupIndex*32 + uint32(bit)

			return &cpu
		}
	}

	return nil
}

func componentAsBDF(component string) *bdf.BDF {
	parsed, err := bdf.ParseBDF(component)
	if err != nil {
		return nil
	}

	return &parsed
}

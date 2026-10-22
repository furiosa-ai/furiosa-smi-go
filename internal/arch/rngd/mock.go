package rngd

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch"
	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

// PEMock describes one occupied PE. A core listed in RngdMock.OccupiedPEs
// reports as occupied in pe_occupancy and gets a "rngd!npuNpeC" node
// carrying the "metrics" and "alloc_status" attributes pe_mgmt.go reads.
// Cores not listed stay unoccupied and get no PE node at all, matching a
// real device: the driver only populates those attributes for a PE that
// has actually been allocated.
type PEMock struct {
	Core           uint32
	CycleCount     uint64
	TaskCycleCount uint64
	// DRAMCapacity defaults to one PE's share of DRAM (Memory().DRAMPerPE()),
	// which is what a singly-allocated, unfused PE reports. Set it to a
	// multiple of that share to model a fusion group spanning several PEs --
	// every PE in a fusion reports the whole fusion's capacity.
	DRAMCapacity   uint64
	DRAMUsage      uint64
	SharedCapacity uint64
	SharedUsage    uint64
}

// HWMonMock describes the hwmon sensors of a mock device. hwmon nodes are
// created by the driver, so a RngdMock only gets them when HWMon is set
// and the device is driver-bound. Values are given in the units callers
// see, not the units sysfs reports: Build converts to millidegrees and
// microwatts on the way out.
type HWMonMock struct {
	SoCPeakCelsius   float64 // defaults to 45.5
	AmbientCelsius   float64 // defaults to 30.2
	PowerConsumption float64 // watts, defaults to 12.5
	PowerLimitMW     uint32  // milliwatts, defaults to 150000
}

// PCIeMock describes the PCIe link attributes and upstream topology of a
// mock device. These live in PCI config space rather than being published
// by the driver, so unlike hwmon they're built even for a Disabled device
// -- matching internal/pcie, which reads them without gating on
// availability.
type PCIeMock struct {
	LinkWidth    uint32  // defaults to 16
	LinkSpeed    float64 // GT/s, defaults to 32.0 (Gen5)
	MaxLinkWidth uint32  // defaults to LinkWidth
	MaxLinkSpeed float64 // defaults to LinkSpeed
	// SriovTotalVFs and SriovNumVFs leave their attributes absent when nil,
	// which is what a device with SR-IOV disabled looks like -- the case
	// internal/pcie defaults to "0" for.
	SriovTotalVFs *uint32
	SriovNumVFs   *uint32
	// Upstream is the BDF of the bridge this device sits under. When set,
	// the /sys/bus/pci/devices entry becomes a symlink into a
	// /sys/devices/pciDDDD:BB/<upstream>/<device> tree, which is the shape
	// real sysfs has and the one internal/pcie resolves to find the root
	// complex and any upstream switch. When empty, the device entry is a
	// plain directory with no ancestry, so neither is discoverable.
	Upstream string
	// UpstreamSecondaryBus and UpstreamSubordinateBus are the bridge's bus
	// range. A subordinate above the secondary marks a real switch; equal
	// values make it a pass-through bridge that internal/pcie skips.
	UpstreamSecondaryBus   uint32
	UpstreamSubordinateBus uint32
}

// SocketMock describes the CPU socket a mock device is local to. Build
// writes the device's local_cpus mask with only FirstCPU set, and that CPU's
// topology/physical_package_id as PackageID -- the two files
// internal/topology reads to tell a same-socket (Cpu) link from a
// cross-socket (Interconnect) one.
type SocketMock struct {
	FirstCPU  uint32
	PackageID uint32
}

// RngdMock describes one mock RNGD device for tests (see
// internal/sysroot.MockDevice).
type RngdMock struct {
	BDF             string
	DevNodeIndex    uint8
	NUMANode        int32
	UUID            string
	Serial          string
	VendorHex       string     // defaults to FuriosaVendorID, as a hex string
	DeviceHex       string     // defaults to RNGDDeviceID, as a hex string
	FirmwareVersion string     // defaults to "1.0.0"
	DriverVersion   string     // defaults to "2025.1.0"
	Governor        string     // defaults to "performance"
	CoreNum         int        // defaults to 8; every core gets 1500MHz
	OccupiedPEs     []PEMock   // defaults to none, leaving every core unoccupied
	HWMon           *HWMonMock // defaults to nil, leaving the device with no hwmon node
	PCIe            *PCIeMock  // defaults to nil, leaving PCIe link attributes unwritten
	Disabled        bool       // if true, no driver/mgmt/devfs nodes are created
	// FusedPEs adds a "npuNpeS-E" devfs entry per {S, E} pair, alongside the
	// per-core ones. Each pair must satisfy S < E < CoreNum.
	FusedPEs  [][2]uint32
	DevNumber string      // mgmt "dev" attribute, defaults to "240:<DevNodeIndex>"
	P2P       bool        // if true, the P2P DMA devfs entries (npuNdmar, npuNch0r..ch7r) are created
	Socket    *SocketMock // defaults to nil, leaving the device with no CPU-socket affinity
}

// buildPCIDir creates the device's /sys/bus/pci/devices entry and returns
// its path. With a PCIe upstream configured the entry is a symlink into a
// /sys/devices tree, which is the shape real sysfs has; without one it's a
// plain directory, leaving the device with no discoverable ancestry.
func (m *RngdMock) buildPCIDir(t *testing.T) string {
	t.Helper()

	entry := filepath.Join(sysroot.PCIDevicesRoot(), m.BDF)
	if m.PCIe == nil || m.PCIe.Upstream == "" {
		sysroot.MustMkdirAll(t, entry)

		return entry
	}

	upstream, err := bdf.ParseBDF(m.PCIe.Upstream)
	if err != nil {
		t.Fatalf("upstream BDF %q: %v", m.PCIe.Upstream, err)
	}

	hostBridge := fmt.Sprintf("pci%04x:%02x", upstream.Domain(), upstream.Bus())
	bridgeDir := sysroot.Join(filepath.Join("sys/devices", hostBridge, m.PCIe.Upstream))
	deviceDir := filepath.Join(bridgeDir, m.BDF)
	sysroot.MustMkdirAll(t, deviceDir)

	sysroot.MustWriteFile(t, filepath.Join(bridgeDir, "secondary_bus_number"),
		fmt.Sprintf("%d\n", m.PCIe.UpstreamSecondaryBus))
	sysroot.MustWriteFile(t, filepath.Join(bridgeDir, "subordinate_bus_number"),
		fmt.Sprintf("%d\n", m.PCIe.UpstreamSubordinateBus))

	// Bridges are listed under /sys/bus/pci/devices alongside endpoints, and
	// that's the path internal/pcie reads their bus range through.
	sysroot.MustMkdirAll(t, sysroot.PCIDevicesRoot())
	mustSymlink(t, bridgeDir, filepath.Join(sysroot.PCIDevicesRoot(), m.PCIe.Upstream))
	mustSymlink(t, deviceDir, entry)

	return entry
}

// build writes the PCI config-space attributes internal/pcie reads.
func (p *PCIeMock) build(t *testing.T, pciDir, deviceHex string) {
	t.Helper()

	linkWidth := orDefaultUint32(p.LinkWidth, 16)
	linkSpeed := orDefaultFloat64(p.LinkSpeed, 32.0)
	maxLinkWidth := orDefaultUint32(p.MaxLinkWidth, linkWidth)
	maxLinkSpeed := orDefaultFloat64(p.MaxLinkSpeed, linkSpeed)

	sysroot.MustWriteFile(t, filepath.Join(pciDir, "subsystem_vendor"), fmt.Sprintf("0x%04x\n", arch.FuriosaVendorID))
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "subsystem_device"), deviceHex+"\n")
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "revision"), "0x01\n")
	// Class 0x12 subclass 0x00 is "processing accelerator", what the kernel
	// reports for an NPU.
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "class"), "0x120000\n")

	sysroot.MustWriteFile(t, filepath.Join(pciDir, "current_link_width"), fmt.Sprintf("%d\n", linkWidth))
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "current_link_speed"), fmt.Sprintf("%.1f GT/s PCIe\n", linkSpeed))
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "max_link_width"), fmt.Sprintf("%d\n", maxLinkWidth))
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "max_link_speed"), fmt.Sprintf("%.1f GT/s PCIe\n", maxLinkSpeed))

	if p.SriovTotalVFs != nil {
		sysroot.MustWriteFile(t, filepath.Join(pciDir, "sriov_totalvfs"), fmt.Sprintf("%d\n", *p.SriovTotalVFs))
	}

	if p.SriovNumVFs != nil {
		sysroot.MustWriteFile(t, filepath.Join(pciDir, "sriov_numvfs"), fmt.Sprintf("%d\n", *p.SriovNumVFs))
	}
}

// build writes pciDir's local_cpus mask and the matching CPU's
// physical_package_id. The mask uses the kernel's format: comma-separated
// 32-bit hex groups, most significant first, so CPU 0 is "00000001" and CPU
// 32 is "00000001,00000000".
func (c *SocketMock) build(t *testing.T, pciDir string) {
	t.Helper()

	groups := make([]string, c.FirstCPU/32+1)
	for i := range groups {
		groups[i] = "00000000"
	}
	groups[0] = fmt.Sprintf("%08x", uint32(1)<<(c.FirstCPU%32))

	sysroot.MustWriteFile(t, filepath.Join(pciDir, "local_cpus"), strings.Join(groups, ",")+"\n")
	sysroot.MustWriteFile(t, filepath.Join(sysroot.CPUDevicesRoot(), fmt.Sprintf("cpu%d", c.FirstCPU), "topology", "physical_package_id"),
		fmt.Sprintf("%d\n", c.PackageID))
}

// build writes one hwmon node's sensors. The kernel reports temperatures in
// millidegrees and power in microwatts, and pairs each sensor with a
// "<prefix>N_label" naming it -- internal/utils looks sensors up by that
// label rather than by index, so the labels are what make this fixture work.
func (h *HWMonMock) build(t *testing.T, hwmonDir string) {
	t.Helper()

	socPeak := orDefaultFloat64(h.SoCPeakCelsius, 45.5)
	ambient := orDefaultFloat64(h.AmbientCelsius, 30.2)
	power := orDefaultFloat64(h.PowerConsumption, 12.5)
	powerLimitMW := orDefaultUint32(h.PowerLimitMW, 150_000)

	sysroot.MustWriteFile(t, filepath.Join(hwmonDir, "temp1_label"), "PEAK\n")
	sysroot.MustWriteFile(t, filepath.Join(hwmonDir, "temp1_input"), fmt.Sprintf("%d\n", int64(math.Round(socPeak*1000))))
	sysroot.MustWriteFile(t, filepath.Join(hwmonDir, "temp2_label"), "AMBIENT\n")
	sysroot.MustWriteFile(t, filepath.Join(hwmonDir, "temp2_input"), fmt.Sprintf("%d\n", int64(math.Round(ambient*1000))))
	sysroot.MustWriteFile(t, filepath.Join(hwmonDir, "power1_label"), "RMS_TOTAL\n")
	sysroot.MustWriteFile(t, filepath.Join(hwmonDir, "power1_average"), fmt.Sprintf("%d\n", int64(math.Round(power*1_000_000))))
	sysroot.MustWriteFile(t, filepath.Join(hwmonDir, "power1_max"), fmt.Sprintf("%d\n", uint64(powerLimitMW)*1000))
}

// mustSymlink tolerates an existing link so several mocks in one MockSpec
// can share an upstream bridge.
func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("symlink %s -> %s: %v", link, target, err)
	}
}

func orDefaultUint32(value, fallback uint32) uint32 {
	if value == 0 {
		return fallback
	}

	return value
}

func orDefaultFloat64(value, fallback float64) float64 {
	if value == 0 {
		return fallback
	}

	return value
}

// validateFusedPE checks that fused names a PE range start < end < coreNum,
// which is what Build requires of every FusedPEs entry.
func validateFusedPE(fused [2]uint32, coreNum int) error {
	start, end := fused[0], fused[1]
	if start >= end || end >= uint32(coreNum) {
		return fmt.Errorf("fused PE range %d-%d: want start < end < CoreNum (%d)", start, end, coreNum)
	}

	return nil
}

// Build implements sysroot.MockDevice.
func (m *RngdMock) Build(t *testing.T) {
	t.Helper()

	vendorHex := sysroot.OrDefault(m.VendorHex, fmt.Sprintf("0x%04x", arch.FuriosaVendorID))
	deviceHex := sysroot.OrDefault(m.DeviceHex, fmt.Sprintf("0x%04x", arch.RNGDDeviceID))
	firmwareVersion := sysroot.OrDefault(m.FirmwareVersion, "1.0.0")
	driverVersion := sysroot.OrDefault(m.DriverVersion, "2025.1.0")
	governor := sysroot.OrDefault(m.Governor, "performance")
	coreNum := m.CoreNum
	if coreNum == 0 {
		coreNum = 8
	}

	pciDir := m.buildPCIDir(t)
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "vendor"), vendorHex+"\n")
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "device"), deviceHex+"\n")
	sysroot.MustWriteFile(t, filepath.Join(pciDir, "numa_node"), fmt.Sprintf("%d\n", m.NUMANode))
	if m.PCIe != nil {
		m.PCIe.build(t, pciDir, deviceHex)
	}

	if m.Socket != nil {
		m.Socket.build(t, pciDir)
	}

	if m.Disabled {
		return
	}

	sysroot.MustMkdirAll(t, filepath.Join(pciDir, "driver"))
	if m.HWMon != nil {
		m.HWMon.build(t, filepath.Join(pciDir, "hwmon", "hwmon0"))
	}

	mgmtDir := filepath.Join(RNGDManagementRoot(), fmt.Sprintf("rngd!npu%dmgmt", m.DevNodeIndex))
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrBusname), m.BDF+"\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrDeviceUUID), m.UUID+"\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrDeviceSN), m.Serial+"\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrFwVersion), firmwareVersion+"\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrDriverVersion), driverVersion+"\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrDev),
		sysroot.OrDefault(m.DevNumber, fmt.Sprintf("240:%d", m.DevNodeIndex))+"\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrGovernor), governor+"\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrDeviceState), "0\n")
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrThrottleReason), "Throttle reason = 0x00000000\n")

	occupied := make(map[uint32]struct{}, len(m.OccupiedPEs))
	for _, pe := range m.OccupiedPEs {
		occupied[pe.Core] = struct{}{}
	}

	occupancy, clocks := "", ""
	for core := range coreNum {
		if _, ok := occupied[uint32(core)]; ok {
			occupancy += "1\n"
		} else {
			occupancy += "0\n"
		}

		clocks += "1500000000\n"
	}

	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrPEOccupancy), occupancy)
	sysroot.MustWriteFile(t, filepath.Join(mgmtDir, attrNPUClocks), clocks)

	// The PE nodes are siblings of the mgmt node, which is how pe_mgmt.go
	// locates them (it walks up from the mgmt path to the class root). The
	// "rngd!npuNpeC" name is spelled out here rather than reused from
	// pePath, so a change to that convention has to be made in both places
	// and can't silently agree with itself.
	for _, pe := range m.OccupiedPEs {
		peDir := filepath.Join(RNGDManagementRoot(), fmt.Sprintf("rngd!npu%dpe%d", m.DevNodeIndex, pe.Core))
		dramCapacity := pe.DRAMCapacity
		if dramCapacity == 0 {
			dramCapacity = Memory().DRAMPerPE()
		}

		sysroot.MustWriteFile(t, filepath.Join(peDir, "metrics"),
			fmt.Sprintf("CycleCount: 0x%x\nTaskCycleCount: 0x%x\n", pe.CycleCount, pe.TaskCycleCount))
		sysroot.MustWriteFile(t, filepath.Join(peDir, "alloc_status"),
			fmt.Sprintf("dram_capacity: 0x%x\ndram_usage: 0x%x\nshared_capacity: 0x%x\nshared_usage: 0x%x\n",
				dramCapacity, pe.DRAMUsage, pe.SharedCapacity, pe.SharedUsage))
	}

	devRoot := RNGDDeviceRoot()
	sysroot.MustMkdirAll(t, devRoot)
	target := filepath.Join(t.TempDir(), "null-target")
	sysroot.MustWriteFile(t, target, "")
	names := make([]string, 0, coreNum+len(m.FusedPEs))
	for core := range coreNum {
		names = append(names, fmt.Sprintf("npu%dpe%d", m.DevNodeIndex, core))
	}

	for _, fused := range m.FusedPEs {
		if err := validateFusedPE(fused, coreNum); err != nil {
			t.Fatal(err)
		}

		names = append(names, fmt.Sprintf("npu%dpe%d-%d", m.DevNodeIndex, fused[0], fused[1]))
	}

	// Spelled out rather than reused from p2pMandatoryInterfaces, for the
	// same reason as the PE node names above.
	if m.P2P {
		names = append(names, fmt.Sprintf("npu%ddmar", m.DevNodeIndex))
		for channel := range 8 {
			names = append(names, fmt.Sprintf("npu%dch%dr", m.DevNodeIndex, channel))
		}
	}

	for _, name := range names {
		link := filepath.Join(devRoot, name)
		if err := os.Symlink(target, link); err != nil {
			t.Fatalf("symlink %s: %v", link, err)
		}
	}
}

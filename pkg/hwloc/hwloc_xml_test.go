package hwloc

import (
	"fmt"
	"testing"
)

const testXMLPath = "testdata/rngd_8cards.xml"

const furiosAIVendorID = 0x1ed2

func loadTestTopology(t *testing.T) *Topology {
	t.Helper()
	topo, err := NewTopology()
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	topo.SetAllTypesFilter(TypeFilterKeepNone)
	topo.SetTypeFilter(ObjPCIDevice, TypeFilterKeepImportant)
	topo.SetTypeFilter(ObjBridge, TypeFilterKeepImportant)
	topo.SetTypeFilter(ObjPackage, TypeFilterKeepImportant)
	topo.SetXML(testXMLPath)
	if err := topo.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	return topo
}

func bdf(domain uint32, bus, dev, fn uint8) string {
	return fmt.Sprintf("%04x:%02x:%02x.%x", domain, bus, dev, fn)
}

func objBDF(obj *Object) string {
	if obj == nil || obj.PCIDev == nil {
		return "<nil>"
	}
	return bdf(obj.PCIDev.Domain, obj.PCIDev.Bus, obj.PCIDev.Dev, obj.PCIDev.Func)
}

// getFuriosDevices mirrors furiosa-smi's get_furiosa_devices: iterate PCI
// devices, filter by vendor ID, return sorted by BDF.
func getFuriosDevices(topo *Topology) []*Object {
	var found []*Object
	obj := topo.GetNextPCIDev(nil)
	for obj != nil {
		if obj.PCIDev != nil && obj.PCIDev.VendorID == furiosAIVendorID {
			found = append(found, obj)
		}
		obj = topo.GetNextPCIDev(obj)
	}
	return found
}

// TestFuriosaDeviceCount verifies that exactly 8 Furiosa NPUs are discovered.
func TestFuriosaDeviceCount(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	devices := getFuriosDevices(topo)
	if got := len(devices); got != 8 {
		t.Fatalf("expected 8 Furiosa devices, got %d", got)
	}
}

// TestFuriosaDeviceBDFs verifies the exact BDF addresses and ordering.
func TestFuriosaDeviceBDFs(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	devices := getFuriosDevices(topo)

	expectedBDFs := []string{
		"0000:27:00.0",
		"0000:2a:00.0",
		"0000:51:00.0",
		"0000:57:00.0",
		"0000:9e:00.0",
		"0000:a4:00.0",
		"0000:c7:00.0",
		"0000:ca:00.0",
	}

	if len(devices) != len(expectedBDFs) {
		t.Fatalf("expected %d devices, got %d", len(expectedBDFs), len(devices))
	}

	for i, dev := range devices {
		got := objBDF(dev)
		if got != expectedBDFs[i] {
			t.Errorf("device[%d]: expected BDF %s, got %s", i, expectedBDFs[i], got)
		}
	}
}

// TestFuriosaDeviceAttributes verifies PCI attributes parsed from XML.
func TestFuriosaDeviceAttributes(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	devices := getFuriosDevices(topo)
	for i, dev := range devices {
		if dev.PCIDev.VendorID != 0x1ed2 {
			t.Errorf("device[%d] vendor: expected 0x1ed2, got 0x%04x", i, dev.PCIDev.VendorID)
		}
		if dev.PCIDev.DeviceID != 0x0001 {
			t.Errorf("device[%d] device: expected 0x0001, got 0x%04x", i, dev.PCIDev.DeviceID)
		}
		if dev.PCIDev.ClassID != 0x1200 {
			t.Errorf("device[%d] class: expected 0x1200, got 0x%04x", i, dev.PCIDev.ClassID)
		}
		if dev.PCIDev.Revision != 0x01 {
			t.Errorf("device[%d] revision: expected 0x01, got 0x%02x", i, dev.PCIDev.Revision)
		}
		if dev.PCIDev.LinkSpeed != 31.507692 {
			t.Errorf("device[%d] link speed: expected 31.507692, got %f", i, dev.PCIDev.LinkSpeed)
		}
		if dev.Type != ObjPCIDevice {
			t.Errorf("device[%d] type: expected ObjPCIDevice, got %d", i, dev.Type)
		}
	}
}

// TestGetPCIDevByBusID verifies device lookup by domain:bus:dev.func.
func TestGetPCIDevByBusID(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	tests := []struct {
		domain uint32
		bus    uint8
		dev    uint8
		fn     uint8
	}{
		{0, 0x27, 0, 0},
		{0, 0x2a, 0, 0},
		{0, 0x51, 0, 0},
		{0, 0x57, 0, 0},
		{0, 0x9e, 0, 0},
		{0, 0xa4, 0, 0},
		{0, 0xc7, 0, 0},
		{0, 0xca, 0, 0},
	}

	for _, tc := range tests {
		name := bdf(tc.domain, tc.bus, tc.dev, tc.fn)
		t.Run(name, func(t *testing.T) {
			obj := topo.GetPCIDevByBusID(tc.domain, tc.bus, tc.dev, tc.fn)
			if obj == nil {
				t.Fatalf("GetPCIDevByBusID returned nil for %s", name)
			}
			if obj.PCIDev.VendorID != furiosAIVendorID {
				t.Errorf("wrong vendor: expected 0x%04x, got 0x%04x", furiosAIVendorID, obj.PCIDev.VendorID)
			}
		})
	}

	// Non-existent device should return nil.
	if obj := topo.GetPCIDevByBusID(0, 0xFF, 0, 0); obj != nil {
		t.Errorf("expected nil for non-existent device, got %v", objBDF(obj))
	}
}

// TestCommonAncestor verifies the lowest common ancestor algorithm matches
// hwloc behavior for the 8-NPU topology described in furiosa-smi tests:
//
//	Machine
//	├── Package 0
//	│   ├── HostBridge [20-3a]  → NPU0(27:00.0), NPU1(2a:00.0)
//	│   └── HostBridge [4a-64]  → NPU2(51:00.0), NPU3(57:00.0)
//	└── Package 1
//	    ├── HostBridge [97-b1]  → NPU4(9e:00.0), NPU5(a4:00.0)
//	    └── HostBridge [c0-e3]  → NPU6(c7:00.0), NPU7(ca:00.0)
func TestCommonAncestor(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	lookup := func(bus uint8) *Object {
		obj := topo.GetPCIDevByBusID(0, bus, 0, 0)
		if obj == nil {
			t.Fatalf("device 0000:%02x:00.0 not found", bus)
		}
		return obj
	}

	npu0 := lookup(0x27)
	npu1 := lookup(0x2a)
	npu2 := lookup(0x51)
	npu3 := lookup(0x57)
	npu4 := lookup(0x9e)
	npu5 := lookup(0xa4)
	npu6 := lookup(0xc7)
	npu7 := lookup(0xca)

	tests := []struct {
		name         string
		obj1, obj2   *Object
		expectedType ObjType
	}{
		// Same host bridge, same PCIe switch tree → Bridge
		{"NPU0-NPU1 (same switch)", npu0, npu1, ObjBridge},
		{"NPU2-NPU3 (same switch)", npu2, npu3, ObjBridge},
		{"NPU4-NPU5 (same switch)", npu4, npu5, ObjBridge},
		{"NPU6-NPU7 (same switch)", npu6, npu7, ObjBridge},

		// Same package, different host bridges → Package
		{"NPU0-NPU2 (same pkg, diff HB)", npu0, npu2, ObjPackage},
		{"NPU0-NPU3 (same pkg, diff HB)", npu0, npu3, ObjPackage},
		{"NPU1-NPU2 (same pkg, diff HB)", npu1, npu2, ObjPackage},
		{"NPU1-NPU3 (same pkg, diff HB)", npu1, npu3, ObjPackage},
		{"NPU4-NPU6 (same pkg, diff HB)", npu4, npu6, ObjPackage},
		{"NPU4-NPU7 (same pkg, diff HB)", npu4, npu7, ObjPackage},
		{"NPU5-NPU6 (same pkg, diff HB)", npu5, npu6, ObjPackage},
		{"NPU5-NPU7 (same pkg, diff HB)", npu5, npu7, ObjPackage},

		// Different packages → Machine
		{"NPU0-NPU4 (diff pkg)", npu0, npu4, ObjMachine},
		{"NPU0-NPU6 (diff pkg)", npu0, npu6, ObjMachine},
		{"NPU1-NPU5 (diff pkg)", npu1, npu5, ObjMachine},
		{"NPU2-NPU6 (diff pkg)", npu2, npu6, ObjMachine},
		{"NPU3-NPU7 (diff pkg)", npu3, npu7, ObjMachine},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ancestor := GetCommonAncestorObj(tc.obj1, tc.obj2)
			if ancestor == nil {
				t.Fatal("GetCommonAncestorObj returned nil")
			}
			if ancestor.Type != tc.expectedType {
				t.Errorf("expected ancestor type %d, got %d", tc.expectedType, ancestor.Type)
			}
		})
	}
}

// TestCommonAncestorSelf verifies that a device is its own ancestor.
func TestCommonAncestorSelf(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	dev := topo.GetPCIDevByBusID(0, 0x27, 0, 0)
	if dev == nil {
		t.Fatal("device not found")
	}

	ancestor := GetCommonAncestorObj(dev, dev)
	if ancestor != dev {
		t.Error("expected same device as ancestor of itself")
	}
}

// getRootComplex mirrors furiosa-smi's get_root_complex: walk parents until
// a host bridge (UpstreamType == BridgeHost) is found.
func getRootComplex(topo *Topology, domain uint32, bus, dev, fn uint8) (uint16, uint8, bool) {
	obj := topo.GetPCIDevByBusID(domain, bus, dev, fn)
	if obj == nil {
		return 0, 0, false
	}
	for p := obj.Parent; p != nil; p = p.Parent {
		if p.Type == ObjBridge && p.Bridge != nil &&
			p.Bridge.UpstreamType == BridgeHost {
			return uint16(p.Bridge.Downstream.Domain), p.Bridge.Downstream.SecondaryBus, true
		}
	}
	return 0, 0, false
}

// TestRootComplex verifies root complex discovery for each NPU.
func TestRootComplex(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	tests := []struct {
		bus            uint8
		expectedDomain uint16
		expectedSecBus uint8
	}{
		{0x27, 0x0000, 0x20},
		{0x2a, 0x0000, 0x20},
		{0x51, 0x0000, 0x4a},
		{0x57, 0x0000, 0x4a},
		{0x9e, 0x0000, 0x97},
		{0xa4, 0x0000, 0x97},
		{0xc7, 0x0000, 0xc0},
		{0xca, 0x0000, 0xc0},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("NPU@%02x", tc.bus)
		t.Run(name, func(t *testing.T) {
			domain, secBus, ok := getRootComplex(topo, 0, tc.bus, 0, 0)
			if !ok {
				t.Fatal("root complex not found")
			}
			if domain != tc.expectedDomain {
				t.Errorf("root complex domain: expected 0x%04x, got 0x%04x", tc.expectedDomain, domain)
			}
			if secBus != tc.expectedSecBus {
				t.Errorf("root complex secondary bus: expected 0x%02x, got 0x%02x", tc.expectedSecBus, secBus)
			}
		})
	}
}

// getPCIeSwitch mirrors furiosa-smi's get_pcie_switch: walk parents until
// a PCI-to-PCI bridge with subordinate_bus > secondary_bus is found.
func getPCIeSwitch(topo *Topology, domain uint32, bus, dev, fn uint8) (uint16, uint8, uint8, uint8, bool) {
	obj := topo.GetPCIDevByBusID(domain, bus, dev, fn)
	if obj == nil {
		return 0, 0, 0, 0, false
	}
	for p := obj.Parent; p != nil; p = p.Parent {
		if p.Type == ObjBridge && p.Bridge != nil &&
			p.Bridge.UpstreamType == BridgePCI &&
			p.Bridge.DownstreamType == BridgePCI &&
			p.Bridge.Downstream.SubordinateBus > p.Bridge.Downstream.SecondaryBus {
			return uint16(p.Bridge.Upstream.Domain),
				p.Bridge.Upstream.Bus,
				p.Bridge.Upstream.Dev,
				p.Bridge.Upstream.Func,
				true
		}
	}
	return 0, 0, 0, 0, false
}

// TestPCIeSwitch verifies PCIe switch detection for each NPU.
func TestPCIeSwitch(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	tests := []struct {
		bus         uint8
		expectedBDF string
	}{
		{0x27, "0000:25:00.0"},
		{0x2a, "0000:28:00.0"},
		{0x51, "0000:4f:00.0"},
		{0x57, "0000:55:00.0"},
		{0x9e, "0000:9c:00.0"},
		{0xa4, "0000:a2:00.0"},
		{0xc7, "0000:c5:00.0"},
		{0xca, "0000:c8:00.0"},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("NPU@%02x", tc.bus)
		t.Run(name, func(t *testing.T) {
			domain, bus, dev, fn, ok := getPCIeSwitch(topo, 0, tc.bus, 0, 0)
			if !ok {
				t.Fatal("PCIe switch not found")
			}
			got := bdf(uint32(domain), bus, dev, fn)
			if got != tc.expectedBDF {
				t.Errorf("PCIe switch: expected %s, got %s", tc.expectedBDF, got)
			}
		})
	}
}

// TestLinkType mirrors furiosa-smi's get_link_type classification.
func TestLinkType(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	type linkType int
	const (
		linkBridge       linkType = iota // ancestor is ObjBridge
		linkCPU                          // ancestor is ObjPackage
		linkInterconnect                 // ancestor is ObjMachine
	)

	classifyLink := func(obj1, obj2 *Object) linkType {
		ancestor := GetCommonAncestorObj(obj1, obj2)
		switch ancestor.Type {
		case ObjBridge:
			return linkBridge
		case ObjPackage:
			return linkCPU
		case ObjMachine:
			return linkInterconnect
		default:
			return -1
		}
	}

	lookup := func(bus uint8) *Object {
		return topo.GetPCIDevByBusID(0, bus, 0, 0)
	}

	tests := []struct {
		name     string
		bus1     uint8
		bus2     uint8
		expected linkType
	}{
		{"same switch", 0x27, 0x2a, linkBridge},
		{"same pkg diff HB", 0x27, 0x51, linkCPU},
		{"diff pkg", 0x27, 0x9e, linkInterconnect},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyLink(lookup(tc.bus1), lookup(tc.bus2))
			if got != tc.expected {
				t.Errorf("expected link type %d, got %d", tc.expected, got)
			}
		})
	}
}

// TestTopologyStructure verifies the overall tree structure.
func TestTopologyStructure(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	root := topo.Root()
	if root == nil {
		t.Fatal("root is nil")
	}
	if root.Type != ObjMachine {
		t.Fatalf("root type: expected ObjMachine, got %d", root.Type)
	}
	if root.Depth != 0 {
		t.Fatalf("root depth: expected 0, got %d", root.Depth)
	}

	// Machine should have 2 Packages as children.
	packages := 0
	for _, child := range root.Children {
		if child.Type == ObjPackage {
			packages++
			if child.Depth != 1 {
				t.Errorf("package depth: expected 1, got %d", child.Depth)
			}
			if child.Parent != root {
				t.Error("package parent is not root")
			}
		}
	}
	if packages != 2 {
		t.Errorf("expected 2 packages under Machine, got %d", packages)
	}
}

// TestGetTypeDepth verifies virtual depth constants.
func TestGetTypeDepth(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	if d := topo.GetTypeDepth(ObjBridge); d != TypeDepthBridge {
		t.Errorf("bridge depth: expected %d, got %d", TypeDepthBridge, d)
	}
	if d := topo.GetTypeDepth(ObjPCIDevice); d != TypeDepthPCIDevice {
		t.Errorf("PCI device depth: expected %d, got %d", TypeDepthPCIDevice, d)
	}
	if d := topo.GetTypeDepth(ObjMachine); d != 0 {
		t.Errorf("machine depth: expected 0, got %d", d)
	}
	if d := topo.GetTypeDepth(ObjPackage); d != 1 {
		t.Errorf("package depth: expected 1, got %d", d)
	}
}

// TestNextCousin verifies that NextCousin links form a complete chain.
func TestNextCousin(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	count := 0
	obj := topo.GetObjByDepth(TypeDepthPCIDevice, 0)
	for obj != nil {
		count++
		obj = obj.NextCousin
	}

	expectedFromIteration := 0
	iter := topo.GetNextPCIDev(nil)
	for iter != nil {
		expectedFromIteration++
		iter = topo.GetNextPCIDev(iter)
	}

	if count != expectedFromIteration {
		t.Errorf("NextCousin chain length %d != iteration count %d", count, expectedFromIteration)
	}
}

// TestHostBridgeCount verifies the number of host bridges.
func TestHostBridgeCount(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	hostBridges := 0
	obj := topo.GetObjByDepth(TypeDepthBridge, 0)
	for i := uint(0); obj != nil; i++ {
		obj = topo.GetObjByDepth(TypeDepthBridge, i)
		if obj != nil && obj.Bridge != nil && obj.Bridge.UpstreamType == BridgeHost {
			hostBridges++
		}
	}

	// The XML has host bridges for domains: [00-03], [16-17], [20-3a],
	// [4a-64], [74-75], [97-b1], [c0-e3], plus more under Package 1.
	if hostBridges < 4 {
		t.Errorf("expected at least 4 host bridges (one per NPU pair), got %d", hostBridges)
	}
}

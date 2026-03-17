package hwloc

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func mkPCIDev(domain uint32, bus, dev, fn uint8) *Object {
	return &Object{
		Type: ObjPCIDevice,
		PCIDev: &PCIDevAttr{
			Domain: domain,
			Bus:    bus,
			Dev:    dev,
			Func:   fn,
		},
	}
}

func mkPCIBridge(domain uint32, bus, dev, fn, sec, sub uint8) *Object {
	pci := &PCIDevAttr{
		Domain:  domain,
		Bus:     bus,
		Dev:     dev,
		Func:    fn,
		ClassID: pciClassBridgePCI,
	}
	return &Object{
		Type:   ObjBridge,
		PCIDev: pci,
		Bridge: &BridgeAttr{
			Upstream:       *pci,
			UpstreamType:   BridgePCI,
			DownstreamType: BridgePCI,
			Downstream: BridgeDownstreamPCI{
				Domain:         domain,
				SecondaryBus:   sec,
				SubordinateBus: sub,
			},
		},
	}
}

func TestDefaultIOFilters(t *testing.T) {
	topo, err := NewTopology()
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}

	if got := topo.getFilter(ObjBridge); got != TypeFilterKeepNone {
		t.Fatalf("ObjBridge default filter: expected %d, got %d", TypeFilterKeepNone, got)
	}
	if got := topo.getFilter(ObjPCIDevice); got != TypeFilterKeepNone {
		t.Fatalf("ObjPCIDevice default filter: expected %d, got %d", TypeFilterKeepNone, got)
	}
	if got := topo.getFilter(ObjOSDevice); got != TypeFilterKeepNone {
		t.Fatalf("ObjOSDevice default filter: expected %d, got %d", TypeFilterKeepNone, got)
	}

	topo.SetAllTypesFilter(TypeFilterKeepAll)
	if got := topo.getFilter(ObjBridge); got != TypeFilterKeepAll {
		t.Fatalf("ObjBridge filter after SetAllTypesFilter: expected %d, got %d", TypeFilterKeepAll, got)
	}
}

func TestSpecialTypeDepthWithoutObjects(t *testing.T) {
	topo, err := NewTopology()
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}

	if d := topo.GetTypeDepth(ObjBridge); d != TypeDepthBridge {
		t.Fatalf("ObjBridge depth: expected %d, got %d", TypeDepthBridge, d)
	}
	if d := topo.GetTypeDepth(ObjPCIDevice); d != TypeDepthPCIDevice {
		t.Fatalf("ObjPCIDevice depth: expected %d, got %d", TypeDepthPCIDevice, d)
	}
	if d := topo.GetTypeDepth(ObjOSDevice); d != TypeDepthOSDevice {
		t.Fatalf("ObjOSDevice depth: expected %d, got %d", TypeDepthOSDevice, d)
	}

	if obj := topo.GetObjByDepth(TypeDepthBridge, 0); obj != nil {
		t.Fatalf("expected nil bridge object before load, got %#v", obj)
	}
	if obj := topo.GetObjByDepth(TypeDepthPCIDevice, 0); obj != nil {
		t.Fatalf("expected nil PCI device object before load, got %#v", obj)
	}
}

func TestGetNextObjByDepthRejectsMismatchedPrev(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	dev := topo.GetNextPCIDev(nil)
	if dev == nil {
		t.Fatal("no PCI device in test topology")
	}
	if got := topo.GetNextObjByDepth(TypeDepthBridge, dev); got != nil {
		t.Fatalf("expected nil when prev depth mismatches bridge depth, got %v", objBDF(got))
	}

	root := topo.Root()
	if root == nil {
		t.Fatal("nil root")
	}
	if got := topo.GetNextObjByDepth(1, root); got != nil {
		t.Fatalf("expected nil when prev depth mismatches normal depth, got type=%d", got.Type)
	}
}

func TestSpecialObjectsHaveSpecialDepth(t *testing.T) {
	topo := loadTestTopology(t)
	defer topo.Destroy()

	for i := uint(0); ; i++ {
		obj := topo.GetObjByDepth(TypeDepthBridge, i)
		if obj == nil {
			break
		}
		if obj.Depth != TypeDepthBridge {
			t.Fatalf("bridge depth: expected %d, got %d", TypeDepthBridge, obj.Depth)
		}
	}

	for i := uint(0); ; i++ {
		obj := topo.GetObjByDepth(TypeDepthPCIDevice, i)
		if obj == nil {
			break
		}
		if obj.Depth != TypeDepthPCIDevice {
			t.Fatalf("pcidev depth: expected %d, got %d", TypeDepthPCIDevice, obj.Depth)
		}
	}
}

func TestPCITreeInsertDuplicateBusIDIgnored(t *testing.T) {
	var tree []*Object
	d1 := mkPCIDev(0, 0x20, 0x00, 0)
	d2 := mkPCIDev(0, 0x20, 0x00, 0)

	pciTreeInsertByBusID(&tree, d1)
	pciTreeInsertByBusID(&tree, d2)

	if len(tree) != 1 {
		t.Fatalf("expected duplicate busid to be ignored, tree len=%d", len(tree))
	}
	if tree[0] != d1 {
		t.Fatal("first object should remain after duplicate insert")
	}
}

func TestPCITreeInsertSupersetReparentsSiblings(t *testing.T) {
	var tree []*Object
	dev := mkPCIDev(0, 0x22, 0x00, 0)
	bridge := mkPCIBridge(0, 0x20, 0x00, 0, 0x21, 0x2f)

	pciTreeInsertByBusID(&tree, dev)
	pciTreeInsertByBusID(&tree, bridge)

	if len(tree) != 1 {
		t.Fatalf("expected one top-level object, got %d", len(tree))
	}
	if tree[0] != bridge {
		t.Fatal("bridge should become top-level superset")
	}
	if len(bridge.Children) != 1 || bridge.Children[0] != dev {
		t.Fatal("device should be moved under bridge")
	}
	if dev.Parent != bridge {
		t.Fatal("device parent should be bridge")
	}
}

func TestCreateHostBridgesGroupsByDomainBus(t *testing.T) {
	bridge20 := mkPCIBridge(0, 0x20, 0x00, 0, 0x21, 0x24)
	dev20 := mkPCIDev(0, 0x20, 0x01, 0)
	dev30 := mkPCIDev(0, 0x30, 0x00, 0)

	tree := []*Object{bridge20, dev20, dev30}
	hostBridges := createHostBridges(tree)

	if len(hostBridges) != 2 {
		t.Fatalf("expected 2 host bridges, got %d", len(hostBridges))
	}

	hb0 := hostBridges[0]
	if hb0.Bridge == nil || hb0.Bridge.UpstreamType != BridgeHost {
		t.Fatal("first host bridge has invalid bridge attributes")
	}
	if hb0.Bridge.Downstream.SecondaryBus != 0x20 {
		t.Fatalf("first host bridge secondary bus: expected 0x20, got 0x%02x", hb0.Bridge.Downstream.SecondaryBus)
	}
	if hb0.Bridge.Downstream.SubordinateBus != 0x24 {
		t.Fatalf("first host bridge subordinate bus: expected 0x24, got 0x%02x", hb0.Bridge.Downstream.SubordinateBus)
	}
	if len(hb0.Children) != 2 {
		t.Fatalf("first host bridge should have 2 children, got %d", len(hb0.Children))
	}

	hb1 := hostBridges[1]
	if hb1.Bridge.Downstream.SecondaryBus != 0x30 || hb1.Bridge.Downstream.SubordinateBus != 0x30 {
		t.Fatalf("second host bridge bus range mismatch: got [%02x-%02x]",
			hb1.Bridge.Downstream.SecondaryBus, hb1.Bridge.Downstream.SubordinateBus)
	}
}

func TestXMLBridgeFilteredChildPromotion(t *testing.T) {
	const xmlData = `<?xml version="1.0"?>
<topology version="2.0">
  <object type="Machine">
    <object type="Package">
      <object type="Bridge" bridge_type="1-1" bridge_pci="0000:[20-20]" pci_busid="0000:20:00.0" pci_type="0604 [0000:0000] [0000:0000] 00 00">
        <object type="PCIDev" pci_busid="0000:20:01.0" pci_type="1200 [1ed2:0001] [0000:0000] 01 00" />
      </object>
    </object>
  </object>
</topology>`

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "topology.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlData), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	topo, err := NewTopology()
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	defer topo.Destroy()

	topo.SetAllTypesFilter(TypeFilterKeepNone)
	topo.SetTypeFilter(ObjPackage, TypeFilterKeepAll)
	topo.SetTypeFilter(ObjPCIDevice, TypeFilterKeepAll)
	topo.SetTypeFilter(ObjBridge, TypeFilterKeepNone)
	topo.SetXML(xmlPath)
	if err := topo.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := topo.GetObjByDepth(TypeDepthBridge, 0); got != nil {
		t.Fatalf("expected no bridges after filtering, got %+v", got)
	}

	dev := topo.GetPCIDevByBusID(0, 0x20, 0x01, 0)
	if dev == nil {
		t.Fatal("expected promoted PCI device to exist")
	}
	if dev.Parent == nil || dev.Parent.Type != ObjPackage {
		t.Fatalf("promoted PCI device parent: expected ObjPackage, got %+v", dev.Parent)
	}
}

func TestXMLKeepImportantPCIDeviceClassFiltering(t *testing.T) {
	const xmlData = `<?xml version="1.0"?>
<topology version="2.0">
  <object type="Machine">
    <object type="Package">
      <object type="PCIDev" pci_busid="0000:20:00.0" pci_type="0108 [1111:0001] [0000:0000] 01 00" />
      <object type="PCIDev" pci_busid="0000:21:00.0" pci_type="0c03 [2222:0002] [0000:0000] 01 00" />
      <object type="PCIDev" pci_busid="0000:22:00.0" pci_type="0c04 [3333:0003] [0000:0000] 01 00" />
    </object>
  </object>
</topology>`

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "important.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlData), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	topo, err := NewTopology()
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	defer topo.Destroy()

	topo.SetAllTypesFilter(TypeFilterKeepNone)
	topo.SetTypeFilter(ObjPackage, TypeFilterKeepImportant)
	topo.SetTypeFilter(ObjPCIDevice, TypeFilterKeepImportant)
	topo.SetXML(xmlPath)
	if err := topo.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	count := 0
	for obj := topo.GetNextPCIDev(nil); obj != nil; obj = topo.GetNextPCIDev(obj) {
		count++
	}
	if count != 2 {
		t.Fatalf("expected 2 important PCI devices, got %d", count)
	}

	if obj := topo.GetPCIDevByBusID(0, 0x20, 0, 0); obj == nil {
		t.Fatal("expected storage class device to be kept")
	}
	if obj := topo.GetPCIDevByBusID(0, 0x22, 0, 0); obj == nil {
		t.Fatal("expected 0c04 class device to be kept")
	}
	if obj := topo.GetPCIDevByBusID(0, 0x21, 0, 0); obj != nil {
		t.Fatal("expected non-important 0c03 class device to be filtered out")
	}
}

func TestXMLKeepImportantBridgePruning(t *testing.T) {
	const xmlData = `<?xml version="1.0"?>
<topology version="2.0">
  <object type="Machine">
    <object type="Package">
      <object type="Bridge" bridge_type="1-1" bridge_pci="0000:[20-20]" pci_busid="0000:20:00.0" pci_type="0604 [0000:0000] [0000:0000] 00 00" />
      <object type="Bridge" bridge_type="1-1" bridge_pci="0000:[30-30]" pci_busid="0000:30:00.0" pci_type="0604 [0000:0000] [0000:0000] 00 00">
        <object type="PCIDev" pci_busid="0000:30:01.0" pci_type="1200 [1ed2:0001] [0000:0000] 01 00" />
      </object>
      <object type="PCIDev" pci_busid="0000:40:00.0" pci_type="0601 [1234:5678] [0000:0000] 01 00" />
      <object type="PCIDev" subtype="NVSwitch" pci_busid="0000:41:00.0" pci_type="0601 [1234:5679] [0000:0000] 01 00" />
    </object>
  </object>
</topology>`

	dir := t.TempDir()
	xmlPath := filepath.Join(dir, "bridge-important.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlData), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	topo, err := NewTopology()
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	defer topo.Destroy()

	topo.SetAllTypesFilter(TypeFilterKeepNone)
	topo.SetTypeFilter(ObjPackage, TypeFilterKeepImportant)
	topo.SetTypeFilter(ObjBridge, TypeFilterKeepImportant)
	topo.SetTypeFilter(ObjPCIDevice, TypeFilterKeepImportant)
	topo.SetXML(xmlPath)
	if err := topo.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	bridges := 0
	for i := uint(0); ; i++ {
		obj := topo.GetObjByDepth(TypeDepthBridge, i)
		if obj == nil {
			break
		}
		bridges++
	}
	if bridges != 1 {
		t.Fatalf("expected only one non-empty bridge with KEEP_IMPORTANT, got %d", bridges)
	}

	if obj := topo.GetPCIDevByBusID(0, 0x40, 0, 0); obj != nil {
		t.Fatal("expected class-0x06 PCI leaf without children to be dropped")
	}
	if obj := topo.GetPCIDevByBusID(0, 0x41, 0, 0); obj == nil {
		t.Fatal("expected NVSwitch subtype leaf to be kept")
	}
}

func TestParseLinkSpeedMatchesHwlocLinuxParsing(t *testing.T) {
	if got := parseLinkSpeed("2.5 GT/s PCIe"); math.Abs(got-2.0) > 1e-9 {
		t.Fatalf("Gen1 parse mismatch: expected 2.0, got %.12f", got)
	}
	if got := parseLinkSpeed("5 GT/s PCIe"); math.Abs(got-4.0) > 1e-9 {
		t.Fatalf("Gen2 parse mismatch: expected 4.0, got %.12f", got)
	}

	wantGen3 := 8.0 * 128.0 / 130.0
	if got := parseLinkSpeed("8 GT/s PCIe"); math.Abs(got-wantGen3) > 1e-9 {
		t.Fatalf("Gen3 parse mismatch: expected %.12f, got %.12f", wantGen3, got)
	}
}

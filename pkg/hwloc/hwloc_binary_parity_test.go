package hwloc

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

type parityXMLTopology struct {
	XMLName xml.Name       `xml:"topology"`
	Objects []parityXMLObj `xml:"object"`
}

type parityXMLObj struct {
	Type       string         `xml:"type,attr"`
	Subtype    string         `xml:"subtype,attr"`
	BusID      string         `xml:"pci_busid,attr"`
	PCIType    string         `xml:"pci_type,attr"`
	BridgeType string         `xml:"bridge_type,attr"`
	BridgePCI  string         `xml:"bridge_pci,attr"`
	Children   []parityXMLObj `xml:"object"`
}

type parityRefNode struct {
	Type     ObjType
	Depth    int
	Subtype  string
	PCIDev   *PCIDevAttr
	Bridge   *BridgeAttr
	Parent   *parityRefNode
	Children []*parityRefNode
}

type parityRefModel struct {
	root       *parityRefNode
	bridges    []*parityRefNode
	pciDevices []*parityRefNode
	typeDepths map[ObjType]int
	byBDF      map[string]*parityRefNode
}

func runFilteredLstopoXML(t *testing.T) []byte {
	t.Helper()
	lstopoPath := requireLstopo(t)

	cmd := exec.Command(
		lstopoPath,
		"--input", testXMLPath,
		"--of", "xml",
		"--filter", "all:none",
		"--filter", "pci_device:important",
		"--filter", "bridge:important",
		"--filter", "package:important",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lstopo filtered XML failed: %v\n%s", err, string(out))
	}
	return out
}

func buildParityRefModel(t *testing.T, xmlBytes []byte) *parityRefModel {
	t.Helper()

	var topo parityXMLTopology
	if err := xml.Unmarshal(xmlBytes, &topo); err != nil {
		t.Fatalf("parsing lstopo XML output: %v", err)
	}
	if len(topo.Objects) == 0 {
		t.Fatal("lstopo XML has no root object")
	}

	root := convertParityXMLNode(&topo.Objects[0], nil)
	assignParityDepths(root, 0)

	m := &parityRefModel{
		root:       root,
		typeDepths: make(map[ObjType]int),
		byBDF:      make(map[string]*parityRefNode),
	}
	collectParityModel(m, root)
	return m
}

func convertParityXMLNode(x *parityXMLObj, parent *parityRefNode) *parityRefNode {
	n := &parityRefNode{
		Type:    parseObjType(x.Type),
		Subtype: x.Subtype,
		Parent:  parent,
	}

	if x.BusID != "" {
		pci := &PCIDevAttr{}
		var domain, bus, dev, fn uint
		if count, _ := fmt.Sscanf(x.BusID, "%x:%x:%x.%x", &domain, &bus, &dev, &fn); count == 4 {
			pci.Domain = uint32(domain)
			pci.Bus = uint8(bus)
			pci.Dev = uint8(dev)
			pci.Func = uint8(fn)
		}
		if x.PCIType != "" {
			parsePCIType(x.PCIType, pci)
		}
		n.PCIDev = pci
	}

	if x.BridgeType != "" {
		bridge := &BridgeAttr{}
		var upType, downType int
		fmt.Sscanf(x.BridgeType, "%d-%d", &upType, &downType)
		bridge.UpstreamType = BridgeType(upType)
		bridge.DownstreamType = BridgeType(downType)

		if x.BridgePCI != "" {
			var domain, secBus, subBus uint
			if count, _ := fmt.Sscanf(x.BridgePCI, "%x:[%x-%x]", &domain, &secBus, &subBus); count == 3 {
				bridge.Downstream.Domain = uint32(domain)
				bridge.Downstream.SecondaryBus = uint8(secBus)
				bridge.Downstream.SubordinateBus = uint8(subBus)
			}
		}

		if bridge.UpstreamType == BridgePCI && n.PCIDev != nil {
			bridge.Upstream = *n.PCIDev
		}

		n.Bridge = bridge
	}

	for i := range x.Children {
		child := convertParityXMLNode(&x.Children[i], n)
		n.Children = append(n.Children, child)
	}

	return n
}

func assignParityDepths(n *parityRefNode, depth int) {
	if n == nil {
		return
	}
	switch n.Type {
	case ObjBridge:
		n.Depth = TypeDepthBridge
	case ObjPCIDevice:
		n.Depth = TypeDepthPCIDevice
	case ObjOSDevice:
		n.Depth = TypeDepthOSDevice
	default:
		n.Depth = depth
	}
	for _, c := range n.Children {
		assignParityDepths(c, depth+1)
	}
}

func collectParityModel(m *parityRefModel, n *parityRefNode) {
	if n == nil {
		return
	}

	switch n.Type {
	case ObjBridge:
		m.bridges = append(m.bridges, n)
	case ObjPCIDevice:
		m.pciDevices = append(m.pciDevices, n)
		m.byBDF[refNodeBDF(n)] = n
	}

	if n.Type >= ObjMachine && n.Type <= ObjPackage {
		if _, ok := m.typeDepths[n.Type]; !ok {
			m.typeDepths[n.Type] = n.Depth
		}
	}

	for _, c := range n.Children {
		collectParityModel(m, c)
	}
}

func (m *parityRefModel) getTypeDepth(typ ObjType) int {
	switch typ {
	case ObjBridge:
		return TypeDepthBridge
	case ObjPCIDevice:
		return TypeDepthPCIDevice
	case ObjOSDevice:
		return TypeDepthOSDevice
	default:
		if d, ok := m.typeDepths[typ]; ok {
			return d
		}
		return TypeDepthUnknown
	}
}

func parityCommonAncestor(a, b *parityRefNode) *parityRefNode {
	for a != b {
		for a.Depth > b.Depth {
			a = a.Parent
		}
		for b.Depth > a.Depth {
			b = b.Parent
		}
		if a != b && a.Depth == b.Depth {
			a = a.Parent
			b = b.Parent
		}
	}
	return a
}

func refNodeBDF(n *parityRefNode) string {
	if n == nil || n.PCIDev == nil {
		return ""
	}
	return bdf(n.PCIDev.Domain, n.PCIDev.Bus, n.PCIDev.Dev, n.PCIDev.Func)
}

func refBridgeKey(n *parityRefNode) string {
	if n == nil || n.Bridge == nil {
		return ""
	}
	busID := "-"
	if n.PCIDev != nil {
		busID = refNodeBDF(n)
	}
	return fmt.Sprintf("%d-%d:%04x:[%02x-%02x]:%s",
		n.Bridge.UpstreamType,
		n.Bridge.DownstreamType,
		n.Bridge.Downstream.Domain,
		n.Bridge.Downstream.SecondaryBus,
		n.Bridge.Downstream.SubordinateBus,
		busID,
	)
}

func goBridgeKey(obj *Object) string {
	if obj == nil || obj.Bridge == nil {
		return ""
	}
	busID := "-"
	if obj.PCIDev != nil {
		busID = objBDF(obj)
	}
	return fmt.Sprintf("%d-%d:%04x:[%02x-%02x]:%s",
		obj.Bridge.UpstreamType,
		obj.Bridge.DownstreamType,
		obj.Bridge.Downstream.Domain,
		obj.Bridge.Downstream.SecondaryBus,
		obj.Bridge.Downstream.SubordinateBus,
		busID,
	)
}

func parseBDF(t *testing.T, s string) (uint32, uint8, uint8, uint8) {
	t.Helper()
	var domain uint32
	var bus, dev, fn uint8
	if n, err := fmt.Sscanf(s, "%x:%x:%x.%x", &domain, &bus, &dev, &fn); err != nil || n != 4 {
		t.Fatalf("invalid BDF %q", s)
	}
	return domain, bus, dev, fn
}

func refAncestorKey(node *parityRefNode) string {
	if node == nil {
		return "nil"
	}
	switch node.Type {
	case ObjMachine:
		return "Machine"
	case ObjPackage:
		return "Package:" + strings.Join(refDescendantBDFs(node), ",")
	case ObjBridge:
		return "Bridge:" + refBridgeKey(node)
	case ObjPCIDevice:
		return "PCIDev:" + refNodeBDF(node)
	default:
		return fmt.Sprintf("Type%d", node.Type)
	}
}

func goAncestorKey(node *Object) string {
	if node == nil {
		return "nil"
	}
	switch node.Type {
	case ObjMachine:
		return "Machine"
	case ObjPackage:
		return "Package:" + strings.Join(goDescendantBDFs(node), ",")
	case ObjBridge:
		return "Bridge:" + goBridgeKey(node)
	case ObjPCIDevice:
		return "PCIDev:" + objBDF(node)
	default:
		return fmt.Sprintf("Type%d", node.Type)
	}
}

func refDescendantBDFs(n *parityRefNode) []string {
	var out []string
	var walk func(*parityRefNode)
	walk = func(cur *parityRefNode) {
		if cur == nil {
			return
		}
		if cur.Type == ObjPCIDevice && cur.PCIDev != nil {
			out = append(out, refNodeBDF(cur))
		}
		for _, c := range cur.Children {
			walk(c)
		}
	}
	walk(n)
	sort.Strings(out)
	return out
}

func goDescendantBDFs(n *Object) []string {
	var out []string
	var walk func(*Object)
	walk = func(cur *Object) {
		if cur == nil {
			return
		}
		if cur.Type == ObjPCIDevice && cur.PCIDev != nil {
			out = append(out, objBDF(cur))
		}
		for _, c := range cur.Children {
			walk(c)
		}
	}
	walk(n)
	sort.Strings(out)
	return out
}

func collectGoPCIBDFByNextPCIDev(topo *Topology) []string {
	var out []string
	for obj := topo.GetNextPCIDev(nil); obj != nil; obj = topo.GetNextPCIDev(obj) {
		out = append(out, objBDF(obj))
	}
	return out
}

func collectGoPCIBDFByNextObjByType(topo *Topology) []string {
	var out []string
	for obj := topo.GetNextObjByType(ObjPCIDevice, nil); obj != nil; obj = topo.GetNextObjByType(ObjPCIDevice, obj) {
		out = append(out, objBDF(obj))
	}
	return out
}

func collectGoPCIBDFByDepthIteration(topo *Topology) []string {
	var out []string
	for obj := topo.GetNextObjByDepth(TypeDepthPCIDevice, nil); obj != nil; obj = topo.GetNextObjByDepth(TypeDepthPCIDevice, obj) {
		out = append(out, objBDF(obj))
	}
	return out
}

func TestParityAgainstLocalHwlocBinary(t *testing.T) {
	refXML := runFilteredLstopoXML(t)
	ref := buildParityRefModel(t, refXML)

	topo := loadTestTopology(t)
	defer topo.Destroy()

	for _, typ := range []ObjType{ObjMachine, ObjPackage, ObjBridge, ObjPCIDevice} {
		want := ref.getTypeDepth(typ)
		got := topo.GetTypeDepth(typ)
		if got != want {
			t.Fatalf("GetTypeDepth(%d): hwloc=%d, go=%d", typ, want, got)
		}
	}

	var refPCIByOrder []string
	for _, node := range ref.pciDevices {
		refPCIByOrder = append(refPCIByOrder, refNodeBDF(node))
	}

	if got := collectGoPCIBDFByNextPCIDev(topo); strings.Join(got, ",") != strings.Join(refPCIByOrder, ",") {
		t.Fatalf("GetNextPCIDev order mismatch\nref=%v\ngo =%v", refPCIByOrder, got)
	}
	if got := collectGoPCIBDFByNextObjByType(topo); strings.Join(got, ",") != strings.Join(refPCIByOrder, ",") {
		t.Fatalf("GetNextObjByType(PCIDevice) order mismatch\nref=%v\ngo =%v", refPCIByOrder, got)
	}
	if got := collectGoPCIBDFByDepthIteration(topo); strings.Join(got, ",") != strings.Join(refPCIByOrder, ",") {
		t.Fatalf("GetNextObjByDepth(PCIDevice) order mismatch\nref=%v\ngo =%v", refPCIByOrder, got)
	}

	for i, bdfStr := range refPCIByOrder {
		obj := topo.GetObjByDepth(TypeDepthPCIDevice, uint(i))
		if obj == nil {
			t.Fatalf("GetObjByDepth(PCIDevice,%d) returned nil; expected %s", i, bdfStr)
		}
		if got := objBDF(obj); got != bdfStr {
			t.Fatalf("GetObjByDepth(PCIDevice,%d): hwloc=%s, go=%s", i, bdfStr, got)
		}
	}

	for i, refBridge := range ref.bridges {
		obj := topo.GetObjByDepth(TypeDepthBridge, uint(i))
		if obj == nil {
			t.Fatalf("GetObjByDepth(Bridge,%d) returned nil", i)
		}
		want := refBridgeKey(refBridge)
		got := goBridgeKey(obj)
		if got != want {
			t.Fatalf("GetObjByDepth(Bridge,%d) mismatch\nref=%s\ngo =%s", i, want, got)
		}
	}
	if obj := topo.GetObjByDepth(TypeDepthBridge, uint(len(ref.bridges))); obj != nil {
		t.Fatalf("expected nil after last bridge depth index, got %s", goBridgeKey(obj))
	}

	goByBDF := make(map[string]*Object)
	for _, bdfStr := range refPCIByOrder {
		domain, bus, dev, fn := parseBDF(t, bdfStr)
		obj := topo.GetPCIDevByBusID(domain, bus, dev, fn)
		if obj == nil {
			t.Fatalf("GetPCIDevByBusID(%s) returned nil", bdfStr)
		}
		if got := objBDF(obj); got != bdfStr {
			t.Fatalf("GetPCIDevByBusID(%s) returned %s", bdfStr, got)
		}
		goByBDF[bdfStr] = obj
	}

	var furiosaBDFs []string
	for _, bdfStr := range refPCIByOrder {
		n := ref.byBDF[bdfStr]
		if n != nil && n.PCIDev != nil && n.PCIDev.VendorID == furiosAIVendorID {
			furiosaBDFs = append(furiosaBDFs, bdfStr)
		}
	}
	if len(furiosaBDFs) < 2 {
		t.Fatalf("need at least two Furiosa devices for ancestor parity check, got %d", len(furiosaBDFs))
	}

	for i := range furiosaBDFs {
		for j := range furiosaBDFs {
			bdf1 := furiosaBDFs[i]
			bdf2 := furiosaBDFs[j]

			refAnc := parityCommonAncestor(ref.byBDF[bdf1], ref.byBDF[bdf2])
			goAnc := GetCommonAncestorObj(goByBDF[bdf1], goByBDF[bdf2])
			if refAnc == nil || goAnc == nil {
				t.Fatalf("nil ancestor for pair (%s,%s)", bdf1, bdf2)
			}
			if refAnc.Type != goAnc.Type {
				t.Fatalf("ancestor type mismatch for (%s,%s): hwloc=%d, go=%d", bdf1, bdf2, refAnc.Type, goAnc.Type)
			}

			wantKey := refAncestorKey(refAnc)
			gotKey := goAncestorKey(goAnc)
			if wantKey != gotKey {
				t.Fatalf("ancestor mismatch for (%s,%s)\nref=%s\ngo =%s", bdf1, bdf2, wantKey, gotKey)
			}
		}
	}
}

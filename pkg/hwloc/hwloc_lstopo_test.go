package hwloc

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"sort"
	"testing"
)

// lstopoXMLTopology maps the root of lstopo's XML output.
type lstopoXMLTopology struct {
	XMLName xml.Name       `xml:"topology"`
	Objects []lstopoXMLObj `xml:"object"`
}

type lstopoXMLObj struct {
	Type       string         `xml:"type,attr"`
	BusID      string         `xml:"pci_busid,attr"`
	PCIType    string         `xml:"pci_type,attr"`
	BridgeType string         `xml:"bridge_type,attr"`
	BridgePCI  string         `xml:"bridge_pci,attr"`
	Children   []lstopoXMLObj `xml:"object"`
}

type pciDevEntry struct {
	BDF      string
	VendorID uint
	DeviceID uint
}

type bridgeEntry struct {
	BDF          string
	UpstreamType int
	SecBus       uint
	SubBus       uint
	Domain       uint
}

func requireLstopo(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("lstopo")
	if err != nil {
		path, err = exec.LookPath("lstopo-no-graphics")
		if err != nil {
			t.Skip("lstopo not found in PATH; install hwloc to run this test")
		}
	}
	return path
}

// parseLstopoPCIDevices extracts PCI device entries from lstopo XML.
func parseLstopoPCIDevices(obj *lstopoXMLObj) []pciDevEntry {
	var result []pciDevEntry
	if obj.Type == "PCIDev" && obj.BusID != "" {
		entry := pciDevEntry{BDF: obj.BusID}
		if obj.PCIType != "" {
			var classID, vendorID, deviceID uint
			fmt.Sscanf(obj.PCIType, "%x [%x:%x]", &classID, &vendorID, &deviceID)
			entry.VendorID = vendorID
			entry.DeviceID = deviceID
		}
		result = append(result, entry)
	}
	for i := range obj.Children {
		result = append(result, parseLstopoPCIDevices(&obj.Children[i])...)
	}
	return result
}

// parseLstopoBridges extracts bridge entries from lstopo XML.
func parseLstopoBridges(obj *lstopoXMLObj) []bridgeEntry {
	var result []bridgeEntry
	if obj.Type == "Bridge" {
		entry := bridgeEntry{BDF: obj.BusID}
		if obj.BridgeType != "" {
			var up, down int
			fmt.Sscanf(obj.BridgeType, "%d-%d", &up, &down)
			entry.UpstreamType = up
		}
		if obj.BridgePCI != "" {
			fmt.Sscanf(obj.BridgePCI, "%x:[%x-%x]", &entry.Domain, &entry.SecBus, &entry.SubBus)
		}
		result = append(result, entry)
	}
	for i := range obj.Children {
		result = append(result, parseLstopoBridges(&obj.Children[i])...)
	}
	return result
}

// countObjTypes counts the total number of each type in the lstopo XML tree.
func countObjTypes(obj *lstopoXMLObj) map[string]int {
	counts := make(map[string]int)
	counts[obj.Type]++
	for i := range obj.Children {
		for k, v := range countObjTypes(&obj.Children[i]) {
			counts[k] += v
		}
	}
	return counts
}

// TestLstopoPCIDeviceList compares PCI device lists between lstopo and Go.
func TestLstopoPCIDeviceList(t *testing.T) {
	lstopoPath := requireLstopo(t)

	cmd := exec.Command(lstopoPath, "--input", testXMLPath, "--of", "xml")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("lstopo failed: %v", err)
	}

	var lstopoTopo lstopoXMLTopology
	if err := xml.Unmarshal(out, &lstopoTopo); err != nil {
		t.Fatalf("parsing lstopo XML: %v", err)
	}

	if len(lstopoTopo.Objects) == 0 {
		t.Fatal("lstopo returned empty topology")
	}

	lstopoDevices := parseLstopoPCIDevices(&lstopoTopo.Objects[0])

	topo := loadTestTopology(t)
	defer topo.Destroy()

	var goDevices []pciDevEntry
	obj := topo.GetNextPCIDev(nil)
	for obj != nil {
		if obj.PCIDev != nil {
			goDevices = append(goDevices, pciDevEntry{
				BDF:      bdf(obj.PCIDev.Domain, obj.PCIDev.Bus, obj.PCIDev.Dev, obj.PCIDev.Func),
				VendorID: uint(obj.PCIDev.VendorID),
				DeviceID: uint(obj.PCIDev.DeviceID),
			})
		}
		obj = topo.GetNextPCIDev(obj)
	}

	sort.Slice(lstopoDevices, func(i, j int) bool { return lstopoDevices[i].BDF < lstopoDevices[j].BDF })
	sort.Slice(goDevices, func(i, j int) bool { return goDevices[i].BDF < goDevices[j].BDF })

	if len(lstopoDevices) != len(goDevices) {
		t.Errorf("PCI device count mismatch: lstopo=%d, go=%d", len(lstopoDevices), len(goDevices))

		lstopoBDFs := make(map[string]bool)
		for _, d := range lstopoDevices {
			lstopoBDFs[d.BDF] = true
		}
		goBDFs := make(map[string]bool)
		for _, d := range goDevices {
			goBDFs[d.BDF] = true
		}

		for b := range lstopoBDFs {
			if !goBDFs[b] {
				t.Errorf("  missing in Go:    %s", b)
			}
		}
		for b := range goBDFs {
			if !lstopoBDFs[b] {
				t.Errorf("  extra in Go:      %s", b)
			}
		}
		return
	}

	for i := range lstopoDevices {
		if lstopoDevices[i].BDF != goDevices[i].BDF {
			t.Errorf("device[%d] BDF: lstopo=%s, go=%s", i, lstopoDevices[i].BDF, goDevices[i].BDF)
		}
		if lstopoDevices[i].VendorID != goDevices[i].VendorID {
			t.Errorf("device[%d] (%s) vendor: lstopo=0x%04x, go=0x%04x",
				i, lstopoDevices[i].BDF, lstopoDevices[i].VendorID, goDevices[i].VendorID)
		}
		if lstopoDevices[i].DeviceID != goDevices[i].DeviceID {
			t.Errorf("device[%d] (%s) device: lstopo=0x%04x, go=0x%04x",
				i, lstopoDevices[i].BDF, lstopoDevices[i].DeviceID, goDevices[i].DeviceID)
		}
	}
}

// TestLstopoBridgeList compares bridge objects between lstopo and Go.
func TestLstopoBridgeList(t *testing.T) {
	lstopoPath := requireLstopo(t)

	cmd := exec.Command(lstopoPath, "--input", testXMLPath, "--of", "xml")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("lstopo failed: %v", err)
	}

	var lstopoTopo lstopoXMLTopology
	if err := xml.Unmarshal(out, &lstopoTopo); err != nil {
		t.Fatalf("parsing lstopo XML: %v", err)
	}

	lstopoBridges := parseLstopoBridges(&lstopoTopo.Objects[0])

	topo := loadTestTopology(t)
	defer topo.Destroy()

	var goBridges []bridgeEntry
	for i := uint(0); ; i++ {
		obj := topo.GetObjByDepth(TypeDepthBridge, i)
		if obj == nil {
			break
		}
		if obj.Bridge == nil {
			continue
		}
		entry := bridgeEntry{
			UpstreamType: int(obj.Bridge.UpstreamType),
			Domain:       uint(obj.Bridge.Downstream.Domain),
			SecBus:       uint(obj.Bridge.Downstream.SecondaryBus),
			SubBus:       uint(obj.Bridge.Downstream.SubordinateBus),
		}
		if obj.PCIDev != nil {
			entry.BDF = bdf(obj.PCIDev.Domain, obj.PCIDev.Bus, obj.PCIDev.Dev, obj.PCIDev.Func)
		}
		goBridges = append(goBridges, entry)
	}

	// Compare host bridge counts.
	lstopoHostBridges := 0
	goHostBridges := 0
	for _, b := range lstopoBridges {
		if b.UpstreamType == 0 {
			lstopoHostBridges++
		}
	}
	for _, b := range goBridges {
		if b.UpstreamType == 0 {
			goHostBridges++
		}
	}

	if lstopoHostBridges != goHostBridges {
		t.Errorf("host bridge count mismatch: lstopo=%d, go=%d", lstopoHostBridges, goHostBridges)
	}

	if len(lstopoBridges) != len(goBridges) {
		t.Errorf("total bridge count mismatch: lstopo=%d, go=%d", len(lstopoBridges), len(goBridges))
	}

	// Compare PCI-to-PCI bridges by their BDF (host bridges have no BDF).
	lstopoPCIBridges := make(map[string]bridgeEntry)
	for _, b := range lstopoBridges {
		if b.BDF != "" {
			lstopoPCIBridges[b.BDF] = b
		}
	}
	goPCIBridges := make(map[string]bridgeEntry)
	for _, b := range goBridges {
		if b.BDF != "" {
			goPCIBridges[b.BDF] = b
		}
	}

	for bdfKey, lb := range lstopoPCIBridges {
		gb, ok := goPCIBridges[bdfKey]
		if !ok {
			t.Errorf("bridge %s present in lstopo but missing in Go", bdfKey)
			continue
		}
		if lb.SecBus != gb.SecBus {
			t.Errorf("bridge %s secondary bus: lstopo=0x%02x, go=0x%02x", bdfKey, lb.SecBus, gb.SecBus)
		}
		if lb.SubBus != gb.SubBus {
			t.Errorf("bridge %s subordinate bus: lstopo=0x%02x, go=0x%02x", bdfKey, lb.SubBus, gb.SubBus)
		}
	}
}

// TestLstopoFuriosaDevices uses lstopo's XML output to verify that
// Furiosa devices (vendor 0x1ed2) match between lstopo and Go.
func TestLstopoFuriosaDevices(t *testing.T) {
	lstopoPath := requireLstopo(t)

	cmd := exec.Command(lstopoPath, "--input", testXMLPath, "--of", "xml")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("lstopo failed: %v", err)
	}

	var lstopoTopo lstopoXMLTopology
	if err := xml.Unmarshal(out, &lstopoTopo); err != nil {
		t.Fatalf("parsing lstopo XML: %v", err)
	}

	allDevices := parseLstopoPCIDevices(&lstopoTopo.Objects[0])
	var lstopoFuriosaBDFs []string
	for _, d := range allDevices {
		if d.VendorID == 0x1ed2 {
			lstopoFuriosaBDFs = append(lstopoFuriosaBDFs, d.BDF)
		}
	}

	topo := loadTestTopology(t)
	defer topo.Destroy()

	goFuriosa := getFuriosDevices(topo)
	var goFuriosaBDFs []string
	for _, dev := range goFuriosa {
		goFuriosaBDFs = append(goFuriosaBDFs, objBDF(dev))
	}

	sort.Strings(lstopoFuriosaBDFs)
	sort.Strings(goFuriosaBDFs)

	if len(lstopoFuriosaBDFs) != len(goFuriosaBDFs) {
		t.Fatalf("Furiosa device count: lstopo=%d, go=%d\n  lstopo: %v\n  go:     %v",
			len(lstopoFuriosaBDFs), len(goFuriosaBDFs), lstopoFuriosaBDFs, goFuriosaBDFs)
	}

	for i := range lstopoFuriosaBDFs {
		if lstopoFuriosaBDFs[i] != goFuriosaBDFs[i] {
			t.Errorf("Furiosa device[%d]: lstopo=%s, go=%s",
				i, lstopoFuriosaBDFs[i], goFuriosaBDFs[i])
		}
	}
	t.Logf("All %d Furiosa devices match: %v", len(goFuriosaBDFs), goFuriosaBDFs)
}

// TestLstopoTopologyDepth compares the depth of the topology tree reported
// by lstopo with the Go implementation.
func TestLstopoTopologyDepth(t *testing.T) {
	lstopoPath := requireLstopo(t)

	cmd := exec.Command(lstopoPath, "--input", testXMLPath, "--of", "xml")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("lstopo failed: %v", err)
	}

	var lstopoTopo lstopoXMLTopology
	if err := xml.Unmarshal(out, &lstopoTopo); err != nil {
		t.Fatalf("parsing lstopo XML: %v", err)
	}

	lstopoCounts := countObjTypes(&lstopoTopo.Objects[0])

	topo := loadTestTopology(t)
	defer topo.Destroy()

	goPCIDevCount := 0
	dev := topo.GetNextPCIDev(nil)
	for dev != nil {
		goPCIDevCount++
		dev = topo.GetNextPCIDev(dev)
	}

	goBridgeCount := 0
	for i := uint(0); ; i++ {
		if topo.GetObjByDepth(TypeDepthBridge, i) == nil {
			break
		}
		goBridgeCount++
	}

	if lstopoCounts["PCIDev"] != goPCIDevCount {
		t.Errorf("PCIDev count: lstopo=%d, go=%d", lstopoCounts["PCIDev"], goPCIDevCount)
	}
	if lstopoCounts["Bridge"] != goBridgeCount {
		t.Errorf("Bridge count: lstopo=%d, go=%d", lstopoCounts["Bridge"], goBridgeCount)
	}

	t.Logf("Object counts match: PCIDev=%d, Bridge=%d, Package=%d",
		goPCIDevCount, goBridgeCount, lstopoCounts["Package"])
}

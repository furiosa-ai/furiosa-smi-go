package hwloc

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	sysfsDevicesPath = "/sys/bus/pci/devices"
	sysfsCPUPath     = "/sys/devices/system/cpu"
	sysfsNodePath    = "/sys/devices/system/node"

	pciClassBridgePCI   = 0x0604
	pciHeaderTypeBridge = 0x01

	pciRevisionID     = 0x08
	pciHeaderType     = 0x0E
	pciSecondaryBus   = 0x19
	pciSubordinateBus = 0x1A
)

// loadFromSysfs discovers topology from Linux sysfs.
func (t *Topology) loadFromSysfs() error {
	t.root = &Object{Type: ObjMachine}

	var packagesByID map[int]*Object
	if t.getFilter(ObjPackage) != TypeFilterKeepNone {
		var packages []*Object
		packages, packagesByID = discoverPackages()
		for _, pkg := range packages {
			t.root.Children = append(t.root.Children, pkg)
			pkg.Parent = t.root
		}
	}

	if t.getFilter(ObjPCIDevice) != TypeFilterKeepNone || t.getFilter(ObjBridge) != TypeFilterKeepNone {
		if err := t.discoverPCI(packagesByID); err != nil {
			return err
		}
	}

	t.connect()
	return nil
}

// discoverPackages reads physical_package_id from sysfs to find CPU packages.
func discoverPackages() ([]*Object, map[int]*Object) {
	packageMap := make(map[int]*Object)

	cpuDir, err := os.ReadDir(sysfsCPUPath)
	if err != nil {
		return nil, packageMap
	}

	for _, entry := range cpuDir {
		if !strings.HasPrefix(entry.Name(), "cpu") {
			continue
		}
		cpuNum := strings.TrimPrefix(entry.Name(), "cpu")
		if _, err := strconv.Atoi(cpuNum); err != nil {
			continue
		}

		pkgIDPath := filepath.Join(sysfsCPUPath, entry.Name(), "topology", "physical_package_id")
		data, err := os.ReadFile(pkgIDPath)
		if err != nil {
			continue
		}
		pkgID, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}
		if _, exists := packageMap[pkgID]; !exists {
			packageMap[pkgID] = &Object{Type: ObjPackage}
		}
	}

	ids := make([]int, 0, len(packageMap))
	for id := range packageMap {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	packages := make([]*Object, len(ids))
	for i, id := range ids {
		packages[i] = packageMap[id]
	}
	return packages, packageMap
}

// discoverPCI scans /sys/bus/pci/devices/, builds the PCI bridge hierarchy,
// creates host bridges, and attaches them under the appropriate packages.
func (t *Topology) discoverPCI(packagesByID map[int]*Object) error {
	entries, err := os.ReadDir(sysfsDevicesPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", sysfsDevicesPath, err)
	}

	var pciObjs []*Object

	for _, entry := range entries {
		name := entry.Name()
		var domain uint32
		var bus, dev, fn uint8
		n, scanErr := fmt.Sscanf(name, "%x:%x:%x.%x", &domain, &bus, &dev, &fn)
		if scanErr != nil || n != 4 {
			continue
		}

		devPath := filepath.Join(sysfsDevicesPath, name)

		configData, _ := os.ReadFile(filepath.Join(devPath, "config"))

		var classID uint16
		var progIF uint8
		var headerType uint8

		classStr := readSysfsString(filepath.Join(devPath, "class"))
		if classStr != "" {
			val, _ := strconv.ParseUint(strings.TrimPrefix(classStr, "0x"), 16, 32)
			classID = uint16(val >> 8)
			progIF = uint8(val & 0xFF)
		}

		if len(configData) > pciHeaderType {
			headerType = configData[pciHeaderType] & 0x7F
		}

		isBridge := classID == pciClassBridgePCI && headerType == pciHeaderTypeBridge

		var secondaryBus, subordinateBus uint8
		if isBridge && len(configData) >= pciSubordinateBus+1 {
			secondaryBus = configData[pciSecondaryBus]
			subordinateBus = configData[pciSubordinateBus]
			if secondaryBus <= bus || subordinateBus <= bus || secondaryBus > subordinateBus {
				continue
			}
		}

		var obj *Object
		if isBridge {
			if t.getFilter(ObjBridge) == TypeFilterKeepNone {
				continue
			}
			pciAttr := &PCIDevAttr{
				Domain:  domain,
				Bus:     bus,
				Dev:     dev,
				Func:    fn,
				ClassID: classID,
				ProgIF:  progIF,
			}
			obj = &Object{
				Type:   ObjBridge,
				PCIDev: pciAttr,
				Bridge: &BridgeAttr{
					UpstreamType:   BridgePCI,
					DownstreamType: BridgePCI,
					Downstream: BridgeDownstreamPCI{
						Domain:         domain,
						SecondaryBus:   secondaryBus,
						SubordinateBus: subordinateBus,
					},
				},
			}
		} else {
			filter := t.getFilter(ObjPCIDevice)
			if filter == TypeFilterKeepNone {
				continue
			}
			if filter == TypeFilterKeepImportant && !isImportantPCIDeviceClass(classID) {
				continue
			}
			obj = &Object{
				Type: ObjPCIDevice,
				PCIDev: &PCIDevAttr{
					Domain:  domain,
					Bus:     bus,
					Dev:     dev,
					Func:    fn,
					ClassID: classID,
					ProgIF:  progIF,
				},
			}
		}

		readSysfsAttrs(obj.PCIDev, devPath)

		obj.PCIDev.Revision = 0xFF
		if len(configData) > pciRevisionID {
			obj.PCIDev.Revision = configData[pciRevisionID]
		}

		readLinkSpeed(obj.PCIDev, devPath)

		if obj.Bridge != nil {
			obj.Bridge.Upstream = *obj.PCIDev
		}

		pciObjs = append(pciObjs, obj)
	}

	var tree []*Object
	for _, obj := range pciObjs {
		pciTreeInsertByBusID(&tree, obj)
	}

	hostBridges := createHostBridges(tree)

	for _, hb := range hostBridges {
		parent := findParentForHostBridge(hb, t.root, packagesByID)
		parent.Children = append(parent.Children, hb)
		hb.Parent = parent
	}

	return nil
}

func readSysfsString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readSysfsAttrs(pci *PCIDevAttr, devPath string) {
	if s := readSysfsString(filepath.Join(devPath, "vendor")); s != "" {
		val, _ := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16)
		pci.VendorID = uint16(val)
	}
	if s := readSysfsString(filepath.Join(devPath, "device")); s != "" {
		val, _ := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16)
		pci.DeviceID = uint16(val)
	}
	if s := readSysfsString(filepath.Join(devPath, "subsystem_vendor")); s != "" {
		val, _ := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16)
		pci.SubvendorID = uint16(val)
	}
	if s := readSysfsString(filepath.Join(devPath, "subsystem_device")); s != "" {
		val, _ := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16)
		pci.SubdeviceID = uint16(val)
	}
}

func readLinkSpeed(pci *PCIDevAttr, devPath string) {
	speedStr := readSysfsString(filepath.Join(devPath, "current_link_speed"))
	widthStr := readSysfsString(filepath.Join(devPath, "current_link_width"))
	if speedStr == "" || widthStr == "" {
		return
	}
	speed := parseLinkSpeed(speedStr)
	width, _ := strconv.ParseFloat(widthStr, 64)
	if speed > 0 && width > 0 {
		pci.LinkSpeed = float32(speed * width / 8.0)
	}
}

// parseLinkSpeed extracts the GT/s value from sysfs current_link_speed strings
// like "8.0 GT/s PCIe", "16.0 GT/s PCIe", etc.
func parseLinkSpeed(s string) float64 {
	if strings.HasPrefix(s, "2.5 ") {
		return 2.5 * 0.8
	}
	if strings.HasPrefix(s, "5 ") {
		return 5 * 0.8
	}

	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0
	}
	v, _ := strconv.ParseFloat(fields[0], 64)
	return v * 128.0 / 130.0
}

// findParentForHostBridge determines which Package (or Machine root) a host
// bridge should be attached to, based on PCI NUMA locality in sysfs.
func findParentForHostBridge(hb *Object, root *Object, packagesByID map[int]*Object) *Object {
	if len(packagesByID) == 0 {
		return root
	}

	dev := findFirstPCIDevice(hb)
	if dev == nil {
		return firstPackageOrRoot(root, packagesByID)
	}

	bdf := fmt.Sprintf("%04x:%02x:%02x.%x",
		dev.PCIDev.Domain, dev.PCIDev.Bus, dev.PCIDev.Dev, dev.PCIDev.Func)

	numaStr := readSysfsString(filepath.Join(sysfsDevicesPath, bdf, "numa_node"))
	if numaStr == "" || numaStr == "-1" {
		return firstPackageOrRoot(root, packagesByID)
	}

	numaNode, err := strconv.Atoi(numaStr)
	if err != nil || numaNode < 0 {
		return firstPackageOrRoot(root, packagesByID)
	}

	cpuList := readSysfsString(filepath.Join(sysfsNodePath, fmt.Sprintf("node%d", numaNode), "cpulist"))
	if cpuList == "" {
		return firstPackageOrRoot(root, packagesByID)
	}

	firstCPU := parseFirstCPU(cpuList)
	if firstCPU < 0 {
		return firstPackageOrRoot(root, packagesByID)
	}

	pkgStr := readSysfsString(filepath.Join(sysfsCPUPath,
		fmt.Sprintf("cpu%d", firstCPU), "topology", "physical_package_id"))
	if pkgStr == "" {
		return firstPackageOrRoot(root, packagesByID)
	}

	pkgID, err := strconv.Atoi(pkgStr)
	if err != nil {
		return firstPackageOrRoot(root, packagesByID)
	}

	if pkg, ok := packagesByID[pkgID]; ok {
		return pkg
	}
	return firstPackageOrRoot(root, packagesByID)
}

func firstPackageOrRoot(root *Object, packagesByID map[int]*Object) *Object {
	for _, pkg := range packagesByID {
		return pkg
	}
	return root
}

func findFirstPCIDevice(obj *Object) *Object {
	if obj.Type == ObjPCIDevice {
		return obj
	}
	for _, child := range obj.Children {
		if dev := findFirstPCIDevice(child); dev != nil {
			return dev
		}
	}
	return nil
}

func parseFirstCPU(cpuList string) int {
	parts := strings.Split(cpuList, ",")
	if len(parts) == 0 {
		return -1
	}
	rangeParts := strings.Split(strings.TrimSpace(parts[0]), "-")
	v, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
	if err != nil {
		return -1
	}
	return v
}

// --- PCI tree building (equivalent to pci-common.c) ---

type busIDComparison int

const (
	busIDLower    busIDComparison = -2
	busIDIncluded busIDComparison = -1
	busIDEqual    busIDComparison = 0
	busIDSuperset busIDComparison = 1
	busIDHigher   busIDComparison = 2
)

// compareBusIDs compares two PCI objects by their bus addresses,
// also detecting bridge containment (superset/included).
// Equivalent to hwloc_pci_compare_busids in pci-common.c.
func compareBusIDs(a, b *Object) busIDComparison {
	if a.PCIDev.Domain < b.PCIDev.Domain {
		return busIDLower
	}
	if a.PCIDev.Domain > b.PCIDev.Domain {
		return busIDHigher
	}

	if a.Type == ObjBridge && a.Bridge != nil && a.Bridge.DownstreamType == BridgePCI &&
		b.PCIDev.Bus >= a.Bridge.Downstream.SecondaryBus &&
		b.PCIDev.Bus <= a.Bridge.Downstream.SubordinateBus {
		return busIDSuperset
	}
	if b.Type == ObjBridge && b.Bridge != nil && b.Bridge.DownstreamType == BridgePCI &&
		a.PCIDev.Bus >= b.Bridge.Downstream.SecondaryBus &&
		a.PCIDev.Bus <= b.Bridge.Downstream.SubordinateBus {
		return busIDIncluded
	}

	if a.PCIDev.Bus < b.PCIDev.Bus {
		return busIDLower
	}
	if a.PCIDev.Bus > b.PCIDev.Bus {
		return busIDHigher
	}
	if a.PCIDev.Dev < b.PCIDev.Dev {
		return busIDLower
	}
	if a.PCIDev.Dev > b.PCIDev.Dev {
		return busIDHigher
	}
	if a.PCIDev.Func < b.PCIDev.Func {
		return busIDLower
	}
	if a.PCIDev.Func > b.PCIDev.Func {
		return busIDHigher
	}
	return busIDEqual
}

// pciTreeInsertByBusID inserts an object into the PCI tree sorted by bus ID.
// Equivalent to hwloc_pcidisc_tree_insert_by_busid.
func pciTreeInsertByBusID(tree *[]*Object, obj *Object) {
	pciAddObject(nil, tree, obj)
}

// pciAddObject inserts newObj into the sorted sibling list, handling bridge
// containment: devices whose bus falls within a bridge's secondary-subordinate
// range are placed as children of that bridge.
// Equivalent to hwloc_pci_add_object in pci-common.c.
func pciAddObject(parent *Object, siblings *[]*Object, newObj *Object) {
	for i := 0; i < len(*siblings); i++ {
		cur := (*siblings)[i]
		comp := compareBusIDs(newObj, cur)

		switch comp {
		case busIDHigher:
			continue

		case busIDIncluded:
			pciAddObject(cur, &cur.Children, newObj)
			return

		case busIDLower, busIDSuperset:
			newObj.Parent = parent

			*siblings = append(*siblings, nil)
			copy((*siblings)[i+1:], (*siblings)[i:])
			(*siblings)[i] = newObj

			if newObj.Type == ObjBridge && newObj.Bridge != nil &&
				newObj.Bridge.DownstreamType == BridgePCI {
				j := i + 1
				for j < len(*siblings) {
					sibling := (*siblings)[j]
					if compareBusIDs(newObj, sibling) == busIDLower {
						if sibling.PCIDev.Domain > newObj.PCIDev.Domain ||
							sibling.PCIDev.Bus > newObj.Bridge.Downstream.SubordinateBus {
							break
						}
						j++
					} else {
						newObj.Children = append(newObj.Children, sibling)
						sibling.Parent = newObj
						*siblings = append((*siblings)[:j], (*siblings)[j+1:]...)
					}
				}
			}
			return

		case busIDEqual:
			return
		}
	}

	newObj.Parent = parent
	*siblings = append(*siblings, newObj)
}

// createHostBridges groups top-level PCI objects by upstream domain:bus and
// wraps each group in a host bridge (upstream_type = BridgeHost).
// Equivalent to hwloc_pcidisc_add_hostbridges in pci-common.c.
func createHostBridges(tree []*Object) []*Object {
	var hostBridges []*Object

	i := 0
	for i < len(tree) {
		child := tree[i]
		currentDomain := child.PCIDev.Domain
		currentBus := child.PCIDev.Bus
		currentSubordinate := currentBus

		hb := &Object{
			Type: ObjBridge,
			Bridge: &BridgeAttr{
				UpstreamType:   BridgeHost,
				DownstreamType: BridgePCI,
			},
		}

		hb.Children = append(hb.Children, child)
		child.Parent = hb

		if child.Type == ObjBridge && child.Bridge != nil &&
			child.Bridge.DownstreamType == BridgePCI &&
			child.Bridge.Downstream.SubordinateBus > currentSubordinate {
			currentSubordinate = child.Bridge.Downstream.SubordinateBus
		}

		i++
		for i < len(tree) {
			next := tree[i]
			if next.PCIDev.Domain != currentDomain || next.PCIDev.Bus != currentBus {
				break
			}
			hb.Children = append(hb.Children, next)
			next.Parent = hb
			if next.Type == ObjBridge && next.Bridge != nil &&
				next.Bridge.DownstreamType == BridgePCI &&
				next.Bridge.Downstream.SubordinateBus > currentSubordinate {
				currentSubordinate = next.Bridge.Downstream.SubordinateBus
			}
			i++
		}

		hb.Bridge.Downstream = BridgeDownstreamPCI{
			Domain:         currentDomain,
			SecondaryBus:   currentBus,
			SubordinateBus: currentSubordinate,
		}

		hostBridges = append(hostBridges, hb)
	}

	return hostBridges
}

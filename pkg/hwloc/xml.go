package hwloc

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
)

// xmlTopology maps the root <topology> element of hwloc2 XML.
type xmlTopology struct {
	XMLName xml.Name    `xml:"topology"`
	Version string      `xml:"version,attr"`
	Objects []xmlObject `xml:"object"`
}

// xmlObject maps a single <object> element in hwloc2 XML.
// Unknown attributes and child elements (info, distances, etc.) are silently ignored.
type xmlObject struct {
	Type         string      `xml:"type,attr"`
	Subtype      string      `xml:"subtype,attr"`
	PCIBusID     string      `xml:"pci_busid,attr"`
	PCIType      string      `xml:"pci_type,attr"`
	PCILinkSpeed string      `xml:"pci_link_speed,attr"`
	BridgeType   string      `xml:"bridge_type,attr"`
	BridgePCI    string      `xml:"bridge_pci,attr"`
	Children     []xmlObject `xml:"object"`
}

// loadFromXML parses an hwloc2 XML topology file.
func (t *Topology) loadFromXML(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading XML file: %w", err)
	}

	var xtopo xmlTopology
	if err := xml.Unmarshal(data, &xtopo); err != nil {
		return fmt.Errorf("parsing XML: %w", err)
	}

	if len(xtopo.Objects) == 0 {
		return fmt.Errorf("no root object in topology XML")
	}

	t.root = t.importXMLObject(&xtopo.Objects[0], nil)
	if t.root == nil {
		t.root = &Object{Type: ObjMachine}
	}

	t.connect()
	return nil
}

// importXMLObject recursively converts an xmlObject into an Object tree.
// Filtered types are skipped and their children are promoted to the parent.
func (t *Topology) importXMLObject(xobj *xmlObject, parent *Object) *Object {
	objType := parseObjType(xobj.Type)
	filter := t.getFilter(objType)

	if objType != ObjMachine && filter == TypeFilterKeepNone {
		for i := range xobj.Children {
			child := t.importXMLObject(&xobj.Children[i], parent)
			if child != nil && parent != nil {
				parent.Children = append(parent.Children, child)
			}
		}
		return nil
	}

	if objType == ObjPCIDevice && filter == TypeFilterKeepImportant {
		if !isImportantPCIDeviceClass(parsePCIClassID(xobj.PCIType)) {
			for i := range xobj.Children {
				child := t.importXMLObject(&xobj.Children[i], parent)
				if child != nil && parent != nil {
					parent.Children = append(parent.Children, child)
				}
			}
			return nil
		}
	}

	obj := &Object{
		Type:    objType,
		Parent:  parent,
		Subtype: xobj.Subtype,
	}

	if xobj.PCIBusID != "" {
		pci := &PCIDevAttr{}
		var domain, bus, dev, fn uint
		if n, _ := fmt.Sscanf(xobj.PCIBusID, "%x:%x:%x.%x", &domain, &bus, &dev, &fn); n == 4 {
			pci.Domain = uint32(domain)
			pci.Bus = uint8(bus)
			pci.Dev = uint8(dev)
			pci.Func = uint8(fn)
		}

		if xobj.PCIType != "" {
			parsePCIType(xobj.PCIType, pci)
		}

		if xobj.PCILinkSpeed != "" {
			speed, _ := strconv.ParseFloat(xobj.PCILinkSpeed, 32)
			pci.LinkSpeed = float32(speed)
		}

		obj.PCIDev = pci
	}

	if xobj.BridgeType != "" {
		bridge := &BridgeAttr{}

		var upType, downType int
		fmt.Sscanf(xobj.BridgeType, "%d-%d", &upType, &downType)
		bridge.UpstreamType = BridgeType(upType)
		bridge.DownstreamType = BridgeType(downType)

		if xobj.BridgePCI != "" {
			var domain, secBus, subBus uint
			if n, _ := fmt.Sscanf(xobj.BridgePCI, "%x:[%x-%x]", &domain, &secBus, &subBus); n == 3 {
				bridge.Downstream.Domain = uint32(domain)
				bridge.Downstream.SecondaryBus = uint8(secBus)
				bridge.Downstream.SubordinateBus = uint8(subBus)
			}
		}

		if bridge.UpstreamType == BridgePCI && obj.PCIDev != nil {
			bridge.Upstream = *obj.PCIDev
		}

		obj.Bridge = bridge
	}

	for i := range xobj.Children {
		child := t.importXMLObject(&xobj.Children[i], obj)
		if child != nil {
			obj.Children = append(obj.Children, child)
		}
	}

	return obj
}

func parsePCIClassID(s string) uint16 {
	var classID uint
	if n, _ := fmt.Sscanf(s, "%x", &classID); n == 1 {
		return uint16(classID)
	}
	return 0
}

func parseObjType(s string) ObjType {
	switch s {
	case "Machine":
		return ObjMachine
	case "Package":
		return ObjPackage
	case "Bridge":
		return ObjBridge
	case "PCIDev":
		return ObjPCIDevice
	case "OSDev":
		return ObjOSDevice
	default:
		return objTypeUnknown
	}
}

// parsePCIType parses the pci_type XML attribute.
// Format: "CCCC [VVVV:DDDD] [VVVV:DDDD] RR PP"
// (class_id, vendor:device, subvendor:subdevice, revision, prog_if)
func parsePCIType(s string, pci *PCIDevAttr) {
	var classID, vendorID, deviceID, subvendorID, subdeviceID uint
	var revision, progIF uint
	n, _ := fmt.Sscanf(s, "%x [%x:%x] [%x:%x] %x %x",
		&classID, &vendorID, &deviceID, &subvendorID, &subdeviceID, &revision, &progIF)
	if n >= 1 {
		pci.ClassID = uint16(classID)
	}
	if n >= 3 {
		pci.VendorID = uint16(vendorID)
		pci.DeviceID = uint16(deviceID)
	}
	if n >= 5 {
		pci.SubvendorID = uint16(subvendorID)
		pci.SubdeviceID = uint16(subdeviceID)
	}
	if n >= 6 {
		pci.Revision = uint8(revision)
	}
	if n >= 7 {
		pci.ProgIF = uint8(progIF)
	}
}

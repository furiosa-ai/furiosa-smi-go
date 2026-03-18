// Package hwloc provides a pure-Go reimplementation of the hwloc topology
// discovery library, focused on PCI device enumeration and bridge hierarchy.
package hwloc

// ObjType represents the type of a topology object.
type ObjType int

const (
	ObjMachine   ObjType = 0
	ObjPackage   ObjType = 1
	ObjBridge    ObjType = 6
	ObjPCIDevice ObjType = 7
	ObjOSDevice  ObjType = 8

	objTypeUnknown ObjType = -1
)

// BridgeType represents the type of a bridge endpoint (upstream or downstream).
type BridgeType int

const (
	BridgeHost BridgeType = 0
	BridgePCI  BridgeType = 1
)

// TypeFilter controls which object types are included in the topology.
type TypeFilter int

const (
	TypeFilterKeepNone      TypeFilter = 0
	TypeFilterKeepAll       TypeFilter = 1
	TypeFilterKeepStructure TypeFilter = 2
	TypeFilterKeepImportant TypeFilter = 3
)

// Virtual depth constants for I/O object types.
// These are returned by GetTypeDepth for types that don't occupy a fixed level.
const (
	TypeDepthUnknown   = -1
	TypeDepthMultiple  = -2
	TypeDepthBridge    = -4
	TypeDepthPCIDevice = -5
	TypeDepthOSDevice  = -6
)

// PCIDevAttr holds PCI device attributes (domain:bus:dev.func, IDs, link speed).
type PCIDevAttr struct {
	Domain      uint32
	Bus         uint8
	Dev         uint8
	Func        uint8
	ClassID     uint16
	VendorID    uint16
	DeviceID    uint16
	SubvendorID uint16
	SubdeviceID uint16
	Revision    uint8
	ProgIF      uint8
	LinkSpeed   float32
}

// BridgeDownstreamPCI holds the downstream PCI bus range of a bridge.
type BridgeDownstreamPCI struct {
	Domain         uint32
	SecondaryBus   uint8
	SubordinateBus uint8
}

// BridgeAttr holds bridge-specific attributes.
type BridgeAttr struct {
	Upstream       PCIDevAttr
	UpstreamType   BridgeType
	Downstream     BridgeDownstreamPCI
	DownstreamType BridgeType
	Depth          uint
}

// Object represents a single topology object (Machine, Package, Bridge, PCI device, etc.).
type Object struct {
	Type       ObjType
	Depth      int
	Parent     *Object
	Children   []*Object
	NextCousin *Object
	Subtype    string

	// PCIDev holds PCI device attributes. Non-nil for ObjPCIDevice objects
	// and for ObjBridge objects with UpstreamType == BridgePCI.
	PCIDev *PCIDevAttr

	// Bridge holds bridge-specific attributes. Non-nil for ObjBridge objects.
	Bridge *BridgeAttr
}

// Topology represents a hardware topology.
type Topology struct {
	root *Object

	bridges    []*Object
	pciDevices []*Object

	levels     [][]*Object
	typeDepths map[ObjType]int

	defaultFilter   TypeFilter
	filterOverrides map[ObjType]TypeFilter
	xmlPath         string
}

// NewTopology creates and returns a new Topology handle.
// Equivalent to hwloc_topology_init.
func NewTopology() (*Topology, error) {
	return &Topology{
		defaultFilter: TypeFilterKeepAll,
		filterOverrides: map[ObjType]TypeFilter{
			ObjBridge:    TypeFilterKeepNone,
			ObjPCIDevice: TypeFilterKeepNone,
			ObjOSDevice:  TypeFilterKeepNone,
		},
		typeDepths: make(map[ObjType]int),
	}, nil
}

// SetAllTypesFilter sets the default type filter applied to all object types.
// Machine type is always kept regardless of the filter value.
// Equivalent to hwloc_topology_set_all_types_filter.
func (t *Topology) SetAllTypesFilter(filter TypeFilter) {
	t.defaultFilter = filter
	t.filterOverrides = make(map[ObjType]TypeFilter)
}

// SetTypeFilter overrides the filter for a specific object type.
// Equivalent to hwloc_topology_set_type_filter.
func (t *Topology) SetTypeFilter(typ ObjType, filter TypeFilter) {
	t.filterOverrides[typ] = filter
}

func (t *Topology) getFilter(typ ObjType) TypeFilter {
	if typ == ObjMachine {
		return TypeFilterKeepAll
	}
	if f, ok := t.filterOverrides[typ]; ok {
		return f
	}
	return t.defaultFilter
}

func isImportantPCIDeviceClass(classID uint16) bool {
	baseClass := classID >> 8
	return baseClass == 0x03 ||
		baseClass == 0x02 ||
		baseClass == 0x01 ||
		baseClass == 0x00 ||
		baseClass == 0x0b ||
		classID == 0x0c04 ||
		classID == 0x0c06 ||
		classID == 0x0502 ||
		baseClass == 0x06 ||
		baseClass == 0x12
}

// SetXML configures the topology to load from an XML file instead of sysfs.
// Must be called before Load.
// Equivalent to hwloc_topology_set_xml.
func (t *Topology) SetXML(path string) {
	t.xmlPath = path
}

// Load discovers the topology, either from sysfs or from a previously
// configured XML file.
// Equivalent to hwloc_topology_load.
func (t *Topology) Load() error {
	if t.xmlPath != "" {
		return t.loadFromXML(t.xmlPath)
	}
	return t.loadFromSysfs()
}

// Destroy releases all topology data.
// Equivalent to hwloc_topology_destroy.
func (t *Topology) Destroy() {
	t.root = nil
	t.bridges = nil
	t.pciDevices = nil
	t.levels = nil
	t.typeDepths = nil
}

// Root returns the root object of the topology (always ObjMachine).
func (t *Topology) Root() *Object {
	return t.root
}

// GetTypeDepth returns the depth of objects of the given type.
// For I/O object types (Bridge, PCIDevice, OSDevice), returns a virtual
// depth constant (TypeDepthBridge, TypeDepthPCIDevice, etc.).
// Returns TypeDepthUnknown if no objects of that type exist.
// Equivalent to hwloc_get_type_depth.
func (t *Topology) GetTypeDepth(typ ObjType) int {
	switch typ {
	case ObjBridge:
		return TypeDepthBridge
	case ObjPCIDevice:
		return TypeDepthPCIDevice
	case ObjOSDevice:
		return TypeDepthOSDevice
	default:
		if d, ok := t.typeDepths[typ]; ok {
			return d
		}
		return TypeDepthUnknown
	}
}

// GetObjByDepth returns the idx-th object at the given depth.
// For virtual depths (TypeDepthBridge, TypeDepthPCIDevice), returns from
// the corresponding special level list.
// Returns nil if idx is out of range.
// Equivalent to hwloc_get_obj_by_depth.
func (t *Topology) GetObjByDepth(depth int, idx uint) *Object {
	switch depth {
	case TypeDepthBridge:
		if int(idx) < len(t.bridges) {
			return t.bridges[idx]
		}
		return nil
	case TypeDepthPCIDevice:
		if int(idx) < len(t.pciDevices) {
			return t.pciDevices[idx]
		}
		return nil
	default:
		if depth >= 0 && depth < len(t.levels) && int(idx) < len(t.levels[depth]) {
			return t.levels[depth][idx]
		}
		return nil
	}
}

// GetNextPCIDev returns the next PCI device after prev, or the first if prev is nil.
// Equivalent to hwloc_get_next_pcidev.
func (t *Topology) GetNextPCIDev(prev *Object) *Object {
	return t.GetNextObjByType(ObjPCIDevice, prev)
}

// GetNextObjByType returns the next object of the given type after prev.
// Returns the first object of that type if prev is nil.
// Equivalent to hwloc_get_next_obj_by_type.
func (t *Topology) GetNextObjByType(typ ObjType, prev *Object) *Object {
	depth := t.GetTypeDepth(typ)
	if depth == TypeDepthUnknown || depth == TypeDepthMultiple {
		return nil
	}
	return t.GetNextObjByDepth(depth, prev)
}

// GetNextObjByDepth returns the next object at the given depth after prev.
// Returns the first object at that depth if prev is nil.
// Equivalent to hwloc_get_next_obj_by_depth.
func (t *Topology) GetNextObjByDepth(depth int, prev *Object) *Object {
	var list []*Object
	switch depth {
	case TypeDepthBridge:
		list = t.bridges
	case TypeDepthPCIDevice:
		list = t.pciDevices
	default:
		if depth >= 0 && depth < len(t.levels) {
			list = t.levels[depth]
		}
	}

	if prev == nil {
		if len(list) > 0 {
			return list[0]
		}
		return nil
	}

	if prev.Depth != depth {
		return nil
	}

	for i, obj := range list {
		if obj == prev && i+1 < len(list) {
			return list[i+1]
		}
	}
	return nil
}

// GetPCIDevByBusID finds a PCI device by its domain:bus:dev.func address.
// Returns nil if no matching device is found.
// Equivalent to hwloc_get_pcidev_by_busid.
func (t *Topology) GetPCIDevByBusID(domain uint32, bus, dev, fn uint8) *Object {
	for _, obj := range t.pciDevices {
		if obj.PCIDev != nil &&
			obj.PCIDev.Domain == domain &&
			obj.PCIDev.Bus == bus &&
			obj.PCIDev.Dev == dev &&
			obj.PCIDev.Func == fn {
			return obj
		}
	}
	return nil
}

// GetCommonAncestorObj finds the lowest common ancestor of two objects
// by walking parent chains using depth to guide direction.
// Equivalent to hwloc_get_common_ancestor_obj.
func GetCommonAncestorObj(obj1, obj2 *Object) *Object {
	if obj1 == nil || obj2 == nil {
		return nil
	}
	for obj1 != obj2 {
		for obj1.Depth > obj2.Depth {
			if obj1.Parent == nil {
				return nil
			}
			obj1 = obj1.Parent
		}
		for obj2.Depth > obj1.Depth {
			if obj2.Parent == nil {
				return nil
			}
			obj2 = obj2.Parent
		}
		if obj1 != obj2 && obj1.Depth == obj2.Depth {
			if obj1.Parent == nil || obj2.Parent == nil {
				return nil
			}
			obj1 = obj1.Parent
			obj2 = obj2.Parent
		}
	}
	return obj1
}

// connect rebuilds internal indexes after the object tree is constructed.
// Assigns depths, builds special-level lists, and sets up cousin links.
func (t *Topology) connect() {
	t.filterImportantBridges(t.root, 0)
	t.assignDepths(t.root, 0)

	t.bridges = nil
	t.pciDevices = nil
	t.collectSpecialObjects(t.root)

	for i := 0; i < len(t.pciDevices)-1; i++ {
		t.pciDevices[i].NextCousin = t.pciDevices[i+1]
	}
	for i := 0; i < len(t.bridges)-1; i++ {
		t.bridges[i].NextCousin = t.bridges[i+1]
	}

	t.buildLevels()
}

func (t *Topology) filterImportantBridges(root *Object, depth uint) {
	if root == nil {
		return
	}

	kept := root.Children[:0]
	for _, child := range root.Children {
		t.filterImportantBridges(child, depth+1)

		if child.Bridge != nil {
			child.Bridge.Depth = depth
		}

		if t.shouldDropAfterImportantFilter(child) {
			child.Parent = nil
			continue
		}

		kept = append(kept, child)
	}
	root.Children = kept
}

func (t *Topology) shouldDropAfterImportantFilter(obj *Object) bool {
	if t.getFilter(obj.Type) != TypeFilterKeepImportant {
		return false
	}

	if len(obj.Children) != 0 {
		return false
	}

	if obj.Type == ObjBridge {
		return true
	}

	if obj.Type == ObjPCIDevice && obj.PCIDev != nil && (obj.PCIDev.ClassID>>8) == 0x06 {
		return obj.Subtype != "NVSwitch"
	}

	return false
}

func (t *Topology) assignDepths(obj *Object, depth int) {
	if obj == nil {
		return
	}
	switch obj.Type {
	case ObjBridge:
		obj.Depth = TypeDepthBridge
	case ObjPCIDevice:
		obj.Depth = TypeDepthPCIDevice
	case ObjOSDevice:
		obj.Depth = TypeDepthOSDevice
	default:
		obj.Depth = depth
	}
	for _, child := range obj.Children {
		t.assignDepths(child, depth+1)
	}
}

func (t *Topology) collectSpecialObjects(obj *Object) {
	if obj == nil {
		return
	}
	switch obj.Type {
	case ObjBridge:
		t.bridges = append(t.bridges, obj)
	case ObjPCIDevice:
		t.pciDevices = append(t.pciDevices, obj)
	}
	for _, child := range obj.Children {
		t.collectSpecialObjects(child)
	}
}

func (t *Topology) buildLevels() {
	t.levels = nil
	t.typeDepths = make(map[ObjType]int)
	t.buildLevelsRecurse(t.root)

	for _, level := range t.levels {
		for i := 0; i < len(level)-1; i++ {
			level[i].NextCousin = level[i+1]
		}
	}
}

func (t *Topology) buildLevelsRecurse(obj *Object) {
	if obj == nil {
		return
	}
	switch obj.Type {
	case ObjBridge, ObjPCIDevice, ObjOSDevice:
		return
	}

	depth := obj.Depth
	for len(t.levels) <= depth {
		t.levels = append(t.levels, nil)
	}
	t.levels[depth] = append(t.levels[depth], obj)

	if _, exists := t.typeDepths[obj.Type]; !exists {
		t.typeDepths[obj.Type] = depth
	}

	for _, child := range obj.Children {
		t.buildLevelsRecurse(child)
	}
}

package smi

// DeviceInfo represents a device information.
type DeviceInfo interface {
	// Index returns an index number of the device based on hardware topology.
	Index() uint32
	// Arch returns an architecture of device.
	Arch() Arch
	// CoreNum returns the number of PE cores.
	CoreNum() uint32
	// NumaNode returns a numa node of device.
	NumaNode() int32
	// Name returns a name of device.
	Name() string
	// Serial returns a serial of device.
	Serial() string
	// UUID returns an uuid of device.
	UUID() string
	// BDF returns a bdf of device.
	BDF() string
	// Major returns a major part of pci device.
	Major() uint16
	// Minor returns a minor part of pci device.
	Minor() uint16
	// FirmwareVersion returns a firmware version of device.
	FirmwareVersion() VersionInfo
}

var _ DeviceInfo = new(FuriosaSmiDeviceInfo)

type FuriosaSmiDeviceInfo struct {
	index           uint32
	arch            Arch
	coreNum         uint32
	numaNode        int32
	name            string
	serial          string
	uuid            string
	bdf             string
	major           uint16
	minor           uint16
	firmwareVersion VersionInfo
}

func (d *FuriosaSmiDeviceInfo) Index() uint32 {
	return d.index
}

func (d *FuriosaSmiDeviceInfo) Arch() Arch {
	return Arch(d.arch)
}

func (d *FuriosaSmiDeviceInfo) CoreNum() uint32 {
	return d.coreNum
}

func (d *FuriosaSmiDeviceInfo) NumaNode() int32 {
	return d.numaNode
}

func (d *FuriosaSmiDeviceInfo) Name() string {
	return d.name
}

func (d *FuriosaSmiDeviceInfo) Serial() string {
	return d.serial
}

func (d *FuriosaSmiDeviceInfo) UUID() string {
	return d.uuid
}

func (d *FuriosaSmiDeviceInfo) BDF() string {
	return d.bdf
}

func (d *FuriosaSmiDeviceInfo) Major() uint16 {
	return d.major
}

func (d *FuriosaSmiDeviceInfo) Minor() uint16 {
	return d.minor
}

func (d *FuriosaSmiDeviceInfo) FirmwareVersion() VersionInfo {
	return d.firmwareVersion
}

package smi

import (
	"fmt"
)

type PcieInfo interface {
	// DeviceInfo returns PCIe device information.
	DeviceInfo() PcieDeviceInfo
	// LinkInfo returns PCIe link information.
	LinkInfo() PcieLinkInfo
	// SriovInfo returns SR-IOV information.
	SriovInfo() SriovInfo
	// RootComplexInfo returns PCIe Root Complex information.
	RootComplexInfo() PcieRootComplexInfo
	// SwitchInfo returns PCIe switch information if available.
	SwitchInfo() PcieSwitchInfo
}
type pcieInfo struct {
	pcieDeviceInfo      PcieDeviceInfo
	pcieLinkInfo        PcieLinkInfo
	sriovInfo           SriovInfo
	pcieRootComplexInfo PcieRootComplexInfo
	pcieSwitchInfo      PcieSwitchInfo
}

func newPcieInfo(pcieDeviceInfo PcieDeviceInfo,
	pcieLinkInfo PcieLinkInfo,
	sriovInfo SriovInfo,
	pcieRootComplexInfo PcieRootComplexInfo,
	pcieSwitchInfo PcieSwitchInfo) PcieInfo {
	return &pcieInfo{
		pcieDeviceInfo:      pcieDeviceInfo,
		pcieLinkInfo:        pcieLinkInfo,
		sriovInfo:           sriovInfo,
		pcieRootComplexInfo: pcieRootComplexInfo,
		pcieSwitchInfo:      pcieSwitchInfo,
	}
}

func (p *pcieInfo) DeviceInfo() PcieDeviceInfo {
	return p.pcieDeviceInfo
}

func (p *pcieInfo) LinkInfo() PcieLinkInfo {
	return p.pcieLinkInfo
}

func (p *pcieInfo) SriovInfo() SriovInfo {
	return p.sriovInfo
}

func (p *pcieInfo) RootComplexInfo() PcieRootComplexInfo {
	return p.pcieRootComplexInfo
}

func (p *pcieInfo) SwitchInfo() PcieSwitchInfo {
	if p.pcieSwitchInfo == nil {
		return nil
	}
	return p.pcieSwitchInfo
}

type PcieDeviceInfo interface {
	// DeviceId returns device id.
	DeviceId() uint16
	// VendorId returns vendor id.
	VendorId() uint16
	// SubsystemId returns subsystem device id.
	SubsystemId() uint16
	// RevisionId returns revision id.
	RevisionId() uint8
	// ClassId returns class id.
	ClassId() uint8
	// SubClassId returns subclass id.
	SubClassId() uint8
}

var _ PcieDeviceInfo = new(FuriosaSmiPcieDeviceInfo)

type FuriosaSmiPcieDeviceInfo struct {
	deviceId          uint16
	subsystemVendorId uint16
	subsystemDeviceId uint16
	revisionId        byte
	classId           byte
	subClassId        byte
}

func (p *FuriosaSmiPcieDeviceInfo) DeviceId() uint16 {
	return p.deviceId
}

func (p *FuriosaSmiPcieDeviceInfo) VendorId() uint16 {
	return p.subsystemVendorId
}

func (p *FuriosaSmiPcieDeviceInfo) SubsystemId() uint16 {
	return p.subsystemDeviceId
}

func (p *FuriosaSmiPcieDeviceInfo) RevisionId() uint8 {
	return p.revisionId
}

func (p *FuriosaSmiPcieDeviceInfo) ClassId() uint8 {
	return p.classId
}

func (p *FuriosaSmiPcieDeviceInfo) SubClassId() uint8 {
	return p.subClassId
}

type PcieLinkInfo interface {
	// PcieGenStatus returns PCIe generation status.
	PcieGenStatus() uint8
	// PcieWidthStatus returns link width status.
	LinkWidthStatus() uint32
	// PcieSpeedStatus returns link speed status in GT/s.
	LinkSpeedStatus() float64
	// MaxLinkGenCapability returns maximum link generation capability.
	MaxLinkWidthCapability() uint32
	// MaxLinkSpeedCapability returns maximum link speed capability in GT/s.
	MaxLinkSpeedCapability() float64
}

type FuriosaSmiPcieLinkInfo struct {
	pcieGenStatus          byte
	linkWidthStatus        uint32
	linkSpeedStatus        float64
	maxLinkWidthCapability uint32
	maxLinkSpeedCapability float64
}

var _ PcieLinkInfo = new(FuriosaSmiPcieLinkInfo)

func (p *FuriosaSmiPcieLinkInfo) PcieGenStatus() uint8 {
	return p.pcieGenStatus
}

func (p *FuriosaSmiPcieLinkInfo) LinkWidthStatus() uint32 {
	return p.linkWidthStatus
}

func (p *FuriosaSmiPcieLinkInfo) LinkSpeedStatus() float64 {
	return p.linkSpeedStatus
}

func (p *FuriosaSmiPcieLinkInfo) MaxLinkWidthCapability() uint32 {
	return p.maxLinkWidthCapability
}

func (p *FuriosaSmiPcieLinkInfo) MaxLinkSpeedCapability() float64 {
	return p.maxLinkSpeedCapability
}

type SriovInfo interface {
	// SriovTotalVfs returns the total number of VFs
	SriovTotalVfs() uint32
	// SriovEnabledVfs returns the number of enabled VFs
	SriovEnabledVfs() uint32
}

type FuriosaSmiSriovInfo struct {
	sriovTotalVfs   uint32
	sriovEnabledVfs uint32
}

var _ SriovInfo = new(FuriosaSmiSriovInfo)

func (s *FuriosaSmiSriovInfo) SriovTotalVfs() uint32 {
	return s.sriovTotalVfs
}

func (s *FuriosaSmiSriovInfo) SriovEnabledVfs() uint32 {
	return s.sriovEnabledVfs
}

type PcieRootComplexInfo interface {
	// Domain returns domain information.
	Domain() uint16
	// Bus returns bus information.
	Bus() uint8
	// String returns a string representation of the PCIe root complex information in BDF format.
	String() string
}
type FuriosaSmiPcieRootComplexInfo struct {
	domain uint16
	bus    byte
}

var _ PcieRootComplexInfo = new(FuriosaSmiPcieRootComplexInfo)

func (p *FuriosaSmiPcieRootComplexInfo) Domain() uint16 {
	return p.domain
}

func (p *FuriosaSmiPcieRootComplexInfo) Bus() uint8 {
	return p.bus
}

func (p *FuriosaSmiPcieRootComplexInfo) String() string {
	return fmt.Sprintf("%04x:%02x", p.Domain(), p.Bus())
}

type PcieSwitchInfo interface {
	// Domain returns domain information.
	Domain() uint16
	// Bus returns bus information.
	Bus() uint8
	// Device returns device information.
	Device() uint8
	// Function returns function information.
	Function() uint8
	// String returns a string representation of the PCIe switch information in BDF format.
	String() string
}
type FuriosaSmiPcieSwitchInfo struct {
	domain   uint16
	bus      byte
	device   byte
	function byte
}

var _ PcieSwitchInfo = new(FuriosaSmiPcieSwitchInfo)

func (p *FuriosaSmiPcieSwitchInfo) Domain() uint16 {
	return p.domain
}

func (p *FuriosaSmiPcieSwitchInfo) Bus() uint8 {
	return p.bus
}

func (p *FuriosaSmiPcieSwitchInfo) Device() uint8 {
	return p.device
}

func (p *FuriosaSmiPcieSwitchInfo) Function() uint8 {
	return p.function
}

func (p *FuriosaSmiPcieSwitchInfo) String() string {
	return fmt.Sprintf("%04x:%02x:%02x.%d",
		p.Domain(),
		p.Bus(),
		p.Device(),
		p.Function(),
	)
}

func FuriosaSmiGetPcieDeviceInfo(device Device, outPcieDeviceInfo *FuriosaSmiPcieDeviceInfo) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetPcieLinkInfo(device Device, outPcieLinkInfo *FuriosaSmiPcieLinkInfo) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetPcieRootComplexInfo(device Device, outRootComplexInfo *FuriosaSmiPcieRootComplexInfo) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetPcieSwitchInfo(device Device, outPcieSwitchInfo *FuriosaSmiPcieSwitchInfo) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetSriovInfo(device Device, outSriovInfo *FuriosaSmiSriovInfo) error {
	// TODO: Implement this function
	return nil
}

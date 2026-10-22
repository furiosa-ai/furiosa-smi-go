package pcie

// furiosa-smi/src/pcie.rs

import "github.com/furiosa-ai/furiosa-smi-go/internal/bdf"

// PcieInfo is the PCIe topology and link information for a device.
type PcieInfo struct {
	deviceInfo      PcieDeviceInfo
	linkInfo        PcieLinkInfo
	sriovInfo       SriovInfo
	rootComplexInfo PcieRootComplexInfo
	switchInfo      *bdf.BDF
}

func NewPcieInfo(deviceInfo PcieDeviceInfo, linkInfo PcieLinkInfo, sriovInfo SriovInfo, rootComplexInfo PcieRootComplexInfo, switchInfo *bdf.BDF) PcieInfo {
	return PcieInfo{
		deviceInfo:      deviceInfo,
		linkInfo:        linkInfo,
		sriovInfo:       sriovInfo,
		rootComplexInfo: rootComplexInfo,
		switchInfo:      switchInfo,
	}
}

// DeviceInfo returns PCIe device information.
func (i PcieInfo) DeviceInfo() PcieDeviceInfo { return i.deviceInfo }

// LinkInfo returns PCIe link information.
func (i PcieInfo) LinkInfo() PcieLinkInfo { return i.linkInfo }

// SriovInfo returns SR-IOV information.
func (i PcieInfo) SriovInfo() SriovInfo { return i.sriovInfo }

// RootComplexInfo returns PCIe root complex information.
func (i PcieInfo) RootComplexInfo() PcieRootComplexInfo { return i.rootComplexInfo }

// SwitchInfo returns the BDF of the PCIe switch this device is attached
// under, if any.
func (i PcieInfo) SwitchInfo() *bdf.BDF { return i.switchInfo }

// PcieDeviceInfo is a device's PCIe identity fields.
type PcieDeviceInfo struct {
	deviceID    uint16
	vendorID    uint16
	subsystemID uint16
	revisionID  uint8
	classID     uint8
	subClassID  uint8
}

func NewPcieDeviceInfo(deviceID, vendorID, subsystemID uint16, revisionID, classID, subClassID uint8) PcieDeviceInfo {
	return PcieDeviceInfo{
		deviceID:    deviceID,
		vendorID:    vendorID,
		subsystemID: subsystemID,
		revisionID:  revisionID,
		classID:     classID,
		subClassID:  subClassID,
	}
}

// DeviceID returns device id.
func (i PcieDeviceInfo) DeviceID() uint16 { return i.deviceID }

// VendorID returns subsystem vendor id.
func (i PcieDeviceInfo) VendorID() uint16 { return i.vendorID }

// SubsystemID returns subsystem device id.
func (i PcieDeviceInfo) SubsystemID() uint16 { return i.subsystemID }

// RevisionID returns revision id.
func (i PcieDeviceInfo) RevisionID() uint8 { return i.revisionID }

// ClassID returns class id.
func (i PcieDeviceInfo) ClassID() uint8 { return i.classID }

// SubClassID returns sub class id.
func (i PcieDeviceInfo) SubClassID() uint8 { return i.subClassID }

// PcieLinkInfo is a device's current and maximum PCIe link parameters.
type PcieLinkInfo struct {
	pcieGenStatus          uint8
	linkWidthStatus        uint32
	linkSpeedStatus        float64
	maxLinkWidthCapability uint32
	maxLinkSpeedCapability float64
}

func NewPcieLinkInfo(pcieGenStatus uint8, linkWidthStatus uint32, linkSpeedStatus float64, maxLinkWidthCapability uint32, maxLinkSpeedCapability float64) PcieLinkInfo {
	return PcieLinkInfo{
		pcieGenStatus:          pcieGenStatus,
		linkWidthStatus:        linkWidthStatus,
		linkSpeedStatus:        linkSpeedStatus,
		maxLinkWidthCapability: maxLinkWidthCapability,
		maxLinkSpeedCapability: maxLinkSpeedCapability,
	}
}

// PCIeGenStatus returns PCIe generation status.
func (i PcieLinkInfo) PCIeGenStatus() uint8 { return i.pcieGenStatus }

// LinkWidthStatus returns link width status.
func (i PcieLinkInfo) LinkWidthStatus() uint32 { return i.linkWidthStatus }

// LinkSpeedStatus returns link speed status in GT/s.
func (i PcieLinkInfo) LinkSpeedStatus() float64 { return i.linkSpeedStatus }

// MaxLinkWidthCapability returns maximum link width capability.
func (i PcieLinkInfo) MaxLinkWidthCapability() uint32 { return i.maxLinkWidthCapability }

// MaxLinkSpeedCapability returns maximum link speed capability in GT/s.
func (i PcieLinkInfo) MaxLinkSpeedCapability() float64 { return i.maxLinkSpeedCapability }

// SriovInfo is a device's SR-IOV virtual function counts.
type SriovInfo struct {
	sriovTotalVFs   uint32
	sriovEnabledVFs uint32
}

func NewSriovInfo(sriovTotalVFs, sriovEnabledVFs uint32) SriovInfo {
	return SriovInfo{sriovTotalVFs: sriovTotalVFs, sriovEnabledVFs: sriovEnabledVFs}
}

// SriovTotalVFs returns total number of VFs.
func (i SriovInfo) SriovTotalVFs() uint32 { return i.sriovTotalVFs }

// SriovEnabledVFs returns enabled number of VFs.
func (i SriovInfo) SriovEnabledVFs() uint32 { return i.sriovEnabledVFs }

// PcieRootComplexInfo identifies the PCIe root complex (host bridge) a
// device is attached under.
type PcieRootComplexInfo struct {
	domain uint16
	bus    uint8
}

func NewPcieRootComplexInfo(domain uint16, bus uint8) PcieRootComplexInfo {
	return PcieRootComplexInfo{domain: domain, bus: bus}
}

// Domain returns domain.
func (i PcieRootComplexInfo) Domain() uint16 { return i.domain }

// Bus returns bus number.
func (i PcieRootComplexInfo) Bus() uint8 { return i.bus }

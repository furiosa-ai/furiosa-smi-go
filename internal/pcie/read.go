package pcie

// furiosa-smi/src/pcie.rs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

const (
	attrCurrentLinkWidth     = "current_link_width"
	attrCurrentLinkSpeed     = "current_link_speed"
	attrMaxLinkWidth         = "max_link_width"
	attrMaxLinkSpeed         = "max_link_speed"
	attrClass                = "class"
	attrDevice               = "device"
	attrSubsystemVendor      = "subsystem_vendor"
	attrSubsystemDevice      = "subsystem_device"
	attrRevision             = "revision"
	attrSriovTotalVFs        = "sriov_totalvfs"
	attrSriovNumVFs          = "sriov_numvfs"
	attrSecondaryBusNumber   = "secondary_bus_number"
	attrSubordinateBusNumber = "subordinate_bus_number"
)

// ReadPcieInfo reads the PCIe topology and link information for target.
func ReadPcieInfo(target bdf.BDF) (PcieInfo, error) {
	devicePath := filepath.Join(sysroot.PCIDevicesRoot(), target.String())
	realPath, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return PcieInfo{}, smierrors.IO(devicePath, err)
	}

	deviceInfo, err := readPcieDeviceInfo(target)
	if err != nil {
		return PcieInfo{}, err
	}

	linkInfo, err := readLinkInfo(target)
	if err != nil {
		return PcieInfo{}, err
	}

	sriovInfo, err := readSriovInfo(target)
	if err != nil {
		return PcieInfo{}, err
	}

	rootComplexInfo, err := readRootComplexInfo(realPath)
	if err != nil {
		return PcieInfo{}, err
	}

	switchInfo, err := readSwitchInfo(realPath, target)
	if err != nil {
		return PcieInfo{}, err
	}

	return NewPcieInfo(deviceInfo, linkInfo, sriovInfo, rootComplexInfo, switchInfo), nil
}

func readPcieDeviceInfo(target bdf.BDF) (PcieDeviceInfo, error) {
	deviceID, err := readHexU16Attr(target, attrDevice, "couldn't parse device ID to u16")
	if err != nil {
		return PcieDeviceInfo{}, err
	}

	vendorID, err := readHexU16Attr(target, attrSubsystemVendor, "couldn't parse subsystem vendor ID to u16")
	if err != nil {
		return PcieDeviceInfo{}, err
	}

	subsystemID, err := readHexU16Attr(target, attrSubsystemDevice, "couldn't parse subsystem device ID to u16")
	if err != nil {
		return PcieDeviceInfo{}, err
	}

	revisionID, err := readHexU8Attr(target, attrRevision, "couldn't parse revision ID to u8")
	if err != nil {
		return PcieDeviceInfo{}, err
	}

	classRaw, err := readAttr(target, attrClass)
	if err != nil {
		return PcieDeviceInfo{}, err
	}

	classID, subClassID, err := parseClassID(classRaw)
	if err != nil {
		return PcieDeviceInfo{}, err
	}

	return NewPcieDeviceInfo(deviceID, vendorID, subsystemID, revisionID, classID, subClassID), nil
}

func readLinkInfo(target bdf.BDF) (PcieLinkInfo, error) {
	widthRaw, err := readAttr(target, attrCurrentLinkWidth)
	if err != nil {
		return PcieLinkInfo{}, err
	}

	linkWidthStatus, err := parseFirstTokenUint32(widthRaw, "couldn't parse current link width", "couldn't parse current link width to number")
	if err != nil {
		return PcieLinkInfo{}, err
	}

	speedRaw, err := readAttr(target, attrCurrentLinkSpeed)
	if err != nil {
		return PcieLinkInfo{}, err
	}

	linkSpeedStatus, err := parseFirstTokenFloat64(speedRaw, "couldn't parse current link speed", "couldn't parse current link speed to number")
	if err != nil {
		return PcieLinkInfo{}, err
	}

	maxWidthRaw, err := readAttr(target, attrMaxLinkWidth)
	if err != nil {
		return PcieLinkInfo{}, err
	}

	maxLinkWidthCapability, err := parseFirstTokenUint32(maxWidthRaw, "couldn't parse max link width", "couldn't parse max link width to number")
	if err != nil {
		return PcieLinkInfo{}, err
	}

	maxSpeedRaw, err := readAttr(target, attrMaxLinkSpeed)
	if err != nil {
		return PcieLinkInfo{}, err
	}

	maxLinkSpeedCapability, err := parseFirstTokenFloat64(maxSpeedRaw, "couldn't parse max link speed", "couldn't parse max link speed to number")
	if err != nil {
		return PcieLinkInfo{}, err
	}

	pcieGenStatus, err := parsePCIeGenStatus(maxLinkSpeedCapability)
	if err != nil {
		return PcieLinkInfo{}, err
	}

	return NewPcieLinkInfo(pcieGenStatus, linkWidthStatus, linkSpeedStatus, maxLinkWidthCapability, maxLinkSpeedCapability), nil
}

func readSriovInfo(target bdf.BDF) (SriovInfo, error) {
	// SR-IOV is off by default after 2025.3.0 release, so both attributes
	// may legitimately be absent.
	totalRaw := readAttrOrDefault(target, attrSriovTotalVFs, "0")
	numRaw := readAttrOrDefault(target, attrSriovNumVFs, "0")
	total, err := parseUint32(totalRaw, "couldn't parse SR-IOV total VFs to u32")
	if err != nil {
		return SriovInfo{}, err
	}

	enabled, err := parseUint32(numRaw, "couldn't parse SR-IOV number of VFs to u32")
	if err != nil {
		return SriovInfo{}, err
	}

	return NewSriovInfo(total, enabled), nil
}

func readRootComplexInfo(realPath string) (PcieRootComplexInfo, error) {
	for component := range strings.SplitSeq(realPath, string(filepath.Separator)) {
		if hb := parseHostBridgeComponent(component); hb != nil {
			return NewPcieRootComplexInfo(hb.domain, hb.bus), nil
		}
	}

	return PcieRootComplexInfo{}, smierrors.Parse("couldn't find root complex")
}

// readSwitchInfo walks realPath's BDF-shaped ancestors from the device
// outward, returning the nearest one whose bridge bus range actually spans
// more than its own secondary bus (i.e. a real switch, not a pass-through
// bridge).
func readSwitchInfo(realPath string, target bdf.BDF) (*bdf.BDF, error) {
	var ancestors []bdf.BDF
	for component := range strings.SplitSeq(realPath, string(filepath.Separator)) {
		candidate, err := bdf.ParseBDF(component)
		if err != nil {
			continue
		}

		if candidate == target {
			continue
		}

		ancestors = append(ancestors, candidate)
	}

	for _, ancestor := range slices.Backward(ancestors) {

		secondary, err := readBridgeBusNumber(ancestor, attrSecondaryBusNumber)
		if err != nil {
			if isIOError(err) {
				continue
			}

			return nil, err
		}

		subordinate, err := readBridgeBusNumber(ancestor, attrSubordinateBusNumber)
		if err != nil {
			if isIOError(err) {
				continue
			}

			return nil, err
		}

		if subordinate > secondary {
			result := ancestor

			return &result, nil
		}
	}

	return nil, nil
}

func isIOError(err error) bool {
	var ioErr *smierrors.IOError

	return errors.As(err, &ioErr)
}

// readBridgeBusNumber reads attr under target's PCI sysfs directory. Not
// every path ancestor is actually a PCI bridge exposing secondary_bus_number
// /subordinate_bus_number, so the caller treats an IOError here (attribute
// missing) as "skip this ancestor" rather than a hard failure.
func readBridgeBusNumber(target bdf.BDF, attr string) (uint32, error) {
	path := filepath.Join(sysroot.PCIDevicesRoot(), target.String(), attr)
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, smierrors.IO(path, err)
	}

	value, parseErr := strconv.ParseUint(strings.TrimSpace(string(content)), 10, 32)
	if parseErr != nil {
		return 0, smierrors.Parse(fmt.Sprintf("couldn't parse %s to u32", attr))
	}

	return uint32(value), nil
}

// hostBridgeComponent is parseHostBridgeComponent's result: the domain and
// secondary bus number extracted from a "pciDDDD:BB"-shaped sysfs path
// component.
type hostBridgeComponent struct {
	domain uint16
	bus    uint8
}

func parseHostBridgeComponent(component string) *hostBridgeComponent {
	raw, ok := strings.CutPrefix(component, "pci")
	if !ok {
		return nil
	}

	domainStr, busStr, ok := strings.Cut(raw, ":")
	if !ok {
		return nil
	}

	domain, err := strconv.ParseUint(domainStr, 16, 16)
	if err != nil {
		return nil
	}

	bus, err := strconv.ParseUint(busStr, 16, 8)
	if err != nil {
		return nil
	}

	return &hostBridgeComponent{domain: uint16(domain), bus: uint8(bus)}
}

func readHexU16Attr(target bdf.BDF, attr, message string) (uint16, error) {
	content, err := readAttr(target, attr)
	if err != nil {
		return 0, err
	}

	return parseHexU16(content, message)
}

func readHexU8Attr(target bdf.BDF, attr, message string) (uint8, error) {
	content, err := readAttr(target, attr)
	if err != nil {
		return 0, err
	}

	return parseHexU8(content, message)
}

func parseClassID(content string) (uint8, uint8, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(content), "0x")
	value, err := strconv.ParseUint(trimmed, 16, 32)
	if err != nil {
		return 0, 0, smierrors.Parse("couldn't parse class ID to u32")
	}

	return uint8(value >> 16), uint8(value >> 8), nil
}

func parseFirstTokenUint32(content, missingMessage, invalidMessage string) (uint32, error) {
	fields := strings.Fields(content)
	if len(fields) == 0 {
		return 0, smierrors.Parse(missingMessage)
	}

	value, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return 0, smierrors.Parse(invalidMessage)
	}

	return uint32(value), nil
}

func parseFirstTokenFloat64(content, missingMessage, invalidMessage string) (float64, error) {
	fields := strings.Fields(content)
	if len(fields) == 0 {
		return 0, smierrors.Parse(missingMessage)
	}

	value, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, smierrors.Parse(invalidMessage)
	}

	return value, nil
}

func parsePCIeGenStatus(speed float64) (uint8, error) {
	switch speed {
	case 2.5:
		return 1, nil

	case 5.0:
		return 2, nil

	case 8.0:
		return 3, nil

	case 16.0:
		return 4, nil

	case 32.0:
		return 5, nil

	default:
		return 0, smierrors.Parse(fmt.Sprintf("Invalid PCIe link speed: %v", speed))
	}
}

func parseUint32(content, message string) (uint32, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(content), 10, 32)
	if err != nil {
		return 0, smierrors.Parse(message)
	}

	return uint32(value), nil
}

func parseHexU16(content, message string) (uint16, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(content), "0x")
	value, err := strconv.ParseUint(trimmed, 16, 16)
	if err != nil {
		return 0, smierrors.Parse(message)
	}

	return uint16(value), nil
}

func parseHexU8(content, message string) (uint8, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(content), "0x")
	value, err := strconv.ParseUint(trimmed, 16, 8)
	if err != nil {
		return 0, smierrors.Parse(message)
	}

	return uint8(value), nil
}

func readAttrOrDefault(target bdf.BDF, attr, fallback string) string {
	content, err := readAttr(target, attr)
	if err != nil {
		return fallback
	}

	return content
}

func readAttr(target bdf.BDF, attr string) (string, error) {
	path := filepath.Join(sysroot.PCIDevicesRoot(), target.String(), attr)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", smierrors.IO(path, err)
	}

	return string(content), nil
}

package rngd

// furiosa-smi/src/sysfs/rngd/mgmt.rs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/version"
)

var majorMinorPattern = regexp.MustCompile(`^(\d+):(\d+)$`)

func ReadFirmwareVersion(mgmtPath string) (version.VersionInfo, error) {
	raw, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrFwVersion))
	if err != nil {
		return version.VersionInfo{}, err
	}

	return version.ParseVersionInfo(raw)
}

func ReadDriverVersion(mgmtPath string) (version.VersionInfo, error) {
	raw, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrDriverVersion))
	if err != nil {
		return version.VersionInfo{}, err
	}

	return version.ParseVersionInfo(raw)
}

func ReadMajorMinor(mgmtPath string) (uint16, uint16, error) {
	raw, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrDev))
	if err != nil {
		return 0, 0, err
	}

	matches := majorMinorPattern.FindStringSubmatch(raw)
	if matches == nil {
		return 0, 0, smierrors.Parse(fmt.Sprintf("unexpected dev format: '%s'", raw))
	}

	major, majorErr := strconv.ParseUint(matches[1], 10, 16)
	minor, minorErr := strconv.ParseUint(matches[2], 10, 16)
	if majorErr != nil || minorErr != nil {
		return 0, 0, smierrors.Parse(fmt.Sprintf("invalid major/minor in dev: '%s'", raw))
	}

	return uint16(major), uint16(minor), nil
}

func ReadCoreStatus(mgmtPath string) (map[uint32]device.CoreStatus, error) {
	text, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrPEOccupancy))
	if err != nil {
		return nil, err
	}

	return parsePEOccupancy(text)
}

func parsePEOccupancy(text string) (map[uint32]device.CoreStatus, error) {
	lines := strings.Split(strings.TrimRight(text, " \t\r\n"), "\n")
	result := make(map[uint32]device.CoreStatus, len(lines))
	for core, line := range lines {
		status, err := parseOccupancyLine(core, line)
		if err != nil {
			return nil, err
		}

		result[uint32(core)] = status
	}

	return result, nil
}

func parseOccupancyLine(core int, line string) (device.CoreStatus, error) {
	switch strings.TrimSpace(line) {
	case "0":
		return device.CoreStatusAvailable, nil

	case "1":
		return device.CoreStatusOccupied, nil

	default:
		return 0, smierrors.Parse(fmt.Sprintf("invalid pe_occupancy status '%s' for core %d", strings.TrimSpace(line), core))
	}
}

func ReadCoreFrequency(mgmtPath string) (map[uint32]uint32, error) {
	text, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrNPUClocks))
	if err != nil {
		return nil, err
	}

	return parseCoreFrequency(text)
}

func parseCoreFrequency(text string) (map[uint32]uint32, error) {
	lines := strings.Split(strings.TrimRight(text, " \t\r\n"), "\n")
	result := make(map[uint32]uint32, len(lines))
	for core, line := range lines {
		freq, err := parseFrequencyLine(core, line)
		if err != nil {
			return nil, err
		}

		result[uint32(core)] = freq
	}

	return result, nil
}

func parseFrequencyLine(core int, line string) (uint32, error) {
	raw, err := strconv.ParseUint(strings.TrimSpace(line), 10, 32)
	if err != nil {
		return 0, smierrors.Parse(fmt.Sprintf("invalid frequency '%s' for core %d", line, core))
	}

	return uint32(raw) / 1_000_000, nil
}

func ReadGovernorProfile(mgmtPath string) (device.GovernorProfile, error) {
	raw, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrGovernor))
	if err != nil {
		return 0, err
	}

	return parseGovernorProfile(raw)
}

func parseGovernorProfile(raw string) (device.GovernorProfile, error) {
	firstLine := strings.TrimSpace(strings.SplitN(raw, "\n", 2)[0])
	switch {
	case strings.HasSuffix(firstLine, "performance"):
		return device.GovernorProfilePerformance, nil

	case strings.HasSuffix(firstLine, "powersave"):
		return device.GovernorProfilePowerSave, nil

	default:
		return 0, smierrors.Parse(fmt.Sprintf("unknown governor profile: '%s'", firstLine))
	}
}

func WriteGovernorProfile(mgmtPath string, profile device.GovernorProfile) error {
	attrPath := filepath.Join(mgmtPath, attrGovernor)
	var value string
	switch profile {
	case device.GovernorProfilePerformance:
		value = "performance"

	case device.GovernorProfilePowerSave:
		value = "powersave"

	default:
		return smierrors.UnexpectedValue(fmt.Sprintf("invalid governor profile: %s", profile))
	}
	if err := os.WriteFile(attrPath, []byte(value+"\n"), 0o644); err != nil {
		return smierrors.IO(attrPath, err)
	}

	return nil
}

func ReadLiveness(mgmtPath string) bool {
	raw, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrDeviceState))
	if err != nil {
		return false
	}

	return parseLiveness(raw)
}

func parseLiveness(raw string) bool {
	return strings.TrimSpace(raw) == "0"
}

func ReadThrottleReason(mgmtPath string) ([]device.ThrottleReason, error) {
	raw, err := sysfs.ReadSysfsAttr(filepath.Join(mgmtPath, attrThrottleReason))
	if err != nil {
		return nil, err
	}

	bits, err := parseThrottleReasonBits(raw)
	if err != nil {
		return nil, err
	}

	return device.ThrottleReasonsFromBits(bits), nil
}

func parseThrottleReasonBits(raw string) (uint32, error) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return 0, smierrors.Parse(fmt.Sprintf("throttle_reason: %s", raw))
	}

	hex := fields[len(fields)-1]
	value, err := strconv.ParseUint(strings.TrimPrefix(hex, "0x"), 16, 32)
	if err != nil {
		return 0, smierrors.Parse(fmt.Sprintf("throttle_reason hex value: %s", hex))
	}

	return uint32(value), nil
}

func WritePowerLimit(debugfsMgmtPath string, powerLimitMW uint32) error {
	attrPath := filepath.Join(debugfsMgmtPath, "power_limit_mw")
	if err := os.WriteFile(attrPath, []byte(fmt.Sprintf("%d\n", powerLimitMW)), 0o644); err != nil {
		return smierrors.IO(attrPath, err)
	}

	return nil
}

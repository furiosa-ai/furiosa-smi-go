package sysfs

// furiosa-smi/src/sysfs.rs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

type MemoryConfig struct {
	DRAMTotalSize        uint64
	FirmwareReservedSize uint64
	PECount              uint64
}

func (m MemoryConfig) DRAMPerPE() uint64 {
	return (m.DRAMTotalSize - m.FirmwareReservedSize) / m.PECount
}

func FindHwmonDir(target bdf.BDF) (string, error) {
	hwmonRoot := filepath.Join(sysroot.PCIDevicesRoot(), target.String(), "hwmon")
	entries, err := os.ReadDir(hwmonRoot)
	if err != nil {
		return "", smierrors.IO(hwmonRoot, err)
	}

	if len(entries) == 0 {
		return "", smierrors.IO(hwmonRoot, fmt.Errorf("no hwmon directory found"))
	}

	return filepath.Join(hwmonRoot, entries[0].Name()), nil
}

func IsDriverBound(target bdf.BDF) bool {
	driverPath := filepath.Join(sysroot.PCIDevicesRoot(), target.String(), "driver")
	_, err := os.Stat(driverPath)

	return err == nil
}

// ReadSysfsAttr rejects empty content as invalid rather than returning "".
func ReadSysfsAttr(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", smierrors.IO(path, err)
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return "", smierrors.Parse(fmt.Sprintf("empty attribute: %s", path))
	}

	return trimmed, nil
}

func ReadSysfsHexU16(path string) *uint16 {
	raw, err := ReadSysfsAttr(path)
	if err != nil {
		return nil
	}

	value, err := parseHexU16(raw)
	if err != nil {
		return nil
	}

	return &value
}

func parseHexU16(s string) (uint16, error) {
	trimmed := strings.ToLower(strings.TrimSpace(s))
	digits := strings.TrimPrefix(trimmed, "0x")
	value, err := strconv.ParseUint(digits, 16, 16)
	if err != nil {
		return 0, smierrors.Parse(fmt.Sprintf("invalid hex u16: '%s'", trimmed))
	}

	return uint16(value), nil
}

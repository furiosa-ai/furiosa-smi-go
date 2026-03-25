package smi

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	rngdSysClassRoot = "/sys/class/rngd_mgmt"
	rngdDevRoot      = "/dev/rngd"
	pciDevRoot       = "/sys/bus/pci/devices"
)

// Test-mode overrides — set these in tests to redirect sysfs/devfs reads.
var rngdSysClassRootForTest string
var rngdDevRootForTest string
var pciDevRootForTest string

func rngdMgmtRoot() string {
	if rngdSysClassRootForTest != "" {
		return rngdSysClassRootForTest
	}
	return rngdSysClassRoot
}

func rngdDevRootPath() string {
	if rngdDevRootForTest != "" {
		return rngdDevRootForTest
	}
	return rngdDevRoot
}

func pciDevicesRoot() string {
	if pciDevRootForTest != "" {
		return pciDevRootForTest
	}
	return pciDevRoot
}

// isRngdDeviceFile reports whether the file is a candidate rngd device node.
// In test mode (rngdDevRootForTest set) regular files are accepted so tests
// can work without actual character devices present.
func isRngdDeviceFile(info os.FileInfo) bool {
	if rngdDevRootForTest != "" {
		return info.Mode().IsRegular()
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// readMgmtAttr reads a sysfs attribute file from a rngd management directory
// and returns its content with leading/trailing whitespace stripped.
func readMgmtAttr(mgmtDir, attr string) (string, error) {
	data, err := os.ReadFile(filepath.Join(mgmtDir, attr))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

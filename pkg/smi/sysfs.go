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

// findRngdMgmtDirs returns all rngd!npu{N}mgmt directories under the rngd
// sysfs class root, in lexicographic order so npu0 comes before npu1.
func findRngdMgmtDirs() ([]string, error) {
	entries, err := os.ReadDir(rngdMgmtRoot())
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() && strings.HasPrefix(name, "rngd!npu") && strings.HasSuffix(name, "mgmt") {
			dirs = append(dirs, filepath.Join(rngdMgmtRoot(), name))
		}
	}

	return dirs, nil
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

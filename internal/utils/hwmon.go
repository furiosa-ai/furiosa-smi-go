package utils

// The Linux kernel's hwmon sysfs interface (hwmonN/<prefix><i>_label,
// <prefix><i>_<suffix>) is a kernel-standard protocol, not specific to any
// one architecture's driver -- any driver exposing hwmon sensors uses this
// same lookup-by-label mechanism. Rust keeps the equivalent of this code
// inside sysfs/rngd/hwmon.rs too (not the generic sysfs.rs), presumably for
// the same reason we do: there's only one architecture consuming it so
// far, so it hasn't been worth factoring out until a second one needs it.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
)

// FindSensorPath finds the hwmon sensor file at hwmonDir whose "<prefix>N_label"
// entry equals label, returning "<prefix>N_<suffix>"'s path.
func FindSensorPath(hwmonDir, prefix, label, suffix string) (string, error) {
	entries, err := os.ReadDir(hwmonDir)
	if err != nil {
		return "", smierrors.IO(hwmonDir, err)
	}

	for _, entry := range entries {
		if base := matchLabelEntry(hwmonDir, entry.Name(), prefix, label); base != nil {
			return filepath.Join(hwmonDir, fmt.Sprintf("%s_%s", *base, suffix)), nil
		}
	}

	return "", smierrors.IO(hwmonDir, fmt.Errorf("%s sensor '%s' not found", prefix, label))
}

func matchLabelEntry(hwmonDir, name, prefix, label string) *string {
	rest, ok := strings.CutPrefix(name, prefix)
	if !ok {
		return nil
	}

	base, ok := strings.CutSuffix(rest, "_label")
	if !ok {
		return nil
	}

	value, err := sysfs.ReadSysfsAttr(filepath.Join(hwmonDir, name))
	if err != nil || value != label {
		return nil
	}

	result := prefix + base

	return &result
}

// ReadSensorValue reads and parses a hwmon sensor value file.
func ReadSensorValue(path string) (float64, error) {
	raw, err := sysfs.ReadSysfsAttr(path)
	if err != nil {
		return 0, err
	}

	value, parseErr := strconv.ParseFloat(raw, 64)
	if parseErr != nil {
		return 0, smierrors.Parse(fmt.Sprintf("invalid sensor value: %s", path))
	}

	return value, nil
}

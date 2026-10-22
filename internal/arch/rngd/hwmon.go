package rngd

// furiosa-smi/src/sysfs/rngd/hwmon.rs

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/utils"
)

func ReadDeviceTemperature(hwmonDir string) (device.DeviceTemperature, error) {
	socPeak, err := readTemperature(hwmonDir, "PEAK")
	if err != nil {
		return device.DeviceTemperature{}, err
	}

	ambient, err := readTemperature(hwmonDir, "AMBIENT")
	if err != nil {
		return device.DeviceTemperature{}, err
	}

	return device.NewDeviceTemperature(socPeak, ambient), nil
}

func ReadPowerConsumption(hwmonDir string) (float64, error) {
	path, err := utils.FindSensorPath(hwmonDir, "power", "RMS_TOTAL", "average")
	if err != nil {
		return 0, err
	}

	value, err := utils.ReadSensorValue(path)
	if err != nil {
		return 0, err
	}

	return value / 1_000_000.0, nil
}

func ReadPowerLimit(hwmonDir string) (uint32, error) {
	raw, err := sysfs.ReadSysfsAttr(filepath.Join(hwmonDir, "power1_max"))
	if err != nil {
		return 0, err
	}

	return parsePowerLimit(raw)
}

func parsePowerLimit(raw string) (uint32, error) {
	microwatts, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, smierrors.Parse(fmt.Sprintf("power1_max value: '%s'", raw))
	}

	milliwatts := microwatts / 1000
	if milliwatts > math.MaxUint32 {
		return 0, smierrors.Parse(fmt.Sprintf("power1_max value: '%s'", raw))
	}

	return uint32(milliwatts), nil
}

func readTemperature(hwmonDir, label string) (float64, error) {
	path, err := utils.FindSensorPath(hwmonDir, "temp", label, "input")
	if err != nil {
		return 0, err
	}

	value, err := utils.ReadSensorValue(path)
	if err != nil {
		return 0, err
	}

	return value / 1000.0, nil
}

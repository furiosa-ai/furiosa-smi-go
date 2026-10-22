package config

// furiosa-smi/src/arch.rs

import (
	"github.com/furiosa-ai/furiosa-smi-go/internal/arch"
	"github.com/furiosa-ai/furiosa-smi-go/internal/arch/rngd"
	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
)

// MgmtConfig describes how to locate and parse an architecture's
// management nodes. One concrete implementation exists per architecture
// family (e.g. rngd.RNGDMgmtConfig); this package only holds the interface,
// not the values.
type MgmtConfig interface {
	ClassRoot() string
	AttrBusname() string
	AttrDeviceUUID() string
	AttrDeviceSN() string
	ParseNodeIndex(entryName string) *uint8
	FormatDeviceName(devNodeIndex uint8) string
	FormatDebugfsPath(devNodeIndex uint8) string
}

type ArchConfig struct {
	Arch         arch.Arch
	DevRoot      string
	Mgmt         MgmtConfig
	DevfsOps     device.DevfsOps
	MgmtOps      device.MgmtOps
	HwmonOps     device.HwmonOps
	MemoryConfig sysfs.MemoryConfig
}

func RetrieveArchConfig(deviceID uint16) (ArchConfig, error) {
	switch deviceID {
	case arch.RNGDDeviceID:
		return rngdArchConfig(rngd.Memory()), nil

	case arch.RNGDPlusDeviceID:
		return rngdArchConfig(rngd.PlusMemory()), nil

	default:
		return ArchConfig{}, smierrors.UnknownArch(deviceID)
	}
}

func rngdArchConfig(memoryConfig sysfs.MemoryConfig) ArchConfig {
	return ArchConfig{
		Arch:         arch.RNGD,
		DevRoot:      rngd.RNGDDeviceRoot(),
		Mgmt:         rngd.NewRNGDMgmtConfig(),
		DevfsOps:     rngd.NewRNGDDevfsOps(),
		MgmtOps:      rngd.NewRNGDMgmtOps(),
		HwmonOps:     rngd.NewRNGDHwmonOps(),
		MemoryConfig: memoryConfig,
	}
}

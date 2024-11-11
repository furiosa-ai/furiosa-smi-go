package smi

import "github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"

// Arch represents type for NPU architecture.
type Arch uint32

const (
	// ArchWarboy represents Gen 1 - Vision NPU.
	ArchWarboy = Arch(binding.FuriosaSmiArchWarboy)
	// ArchRngd represents Gen 2 - RNGD.
	ArchRngd = Arch(binding.FuriosaSmiArchRngd)
	// ArchRngdMax represents RNGD-Max architecture (not released yet)
	ArchRngdMax = Arch(binding.FuriosaSmiArchRngdMax)
	// ArchRngdS represents RNGD-S architecture (not released yet)
	ArchRngdS = Arch(binding.FuriosaSmiArchRngdS)
)

func (a Arch) ToString() string {
	switch a {
	case ArchWarboy:
		return "warboy"
	case ArchRngd:
		return "rngd"
	case ArchRngdMax:
		return "rngd-max"
	case ArchRngdS:
		return "rngd-s"
	default:
		return "unknown"
	}
}

// CoreStatus represents a device core status
type CoreStatus uint32

const (
	// CoreStatusAvailable represents core is available.
	CoreStatusAvailable = CoreStatus(binding.FuriosaSmiCoreStatusAvailable)
	// CoreStatusOccupied represents core is occupied.
	CoreStatusOccupied = CoreStatus(binding.FuriosaSmiCoreStatusOccupied)
)

// LinkType represents a topology link type between 2 NPU devices.
type LinkType uint32

const (
	// LinkTypeUnknown means unknown link type.
	LinkTypeUnknown = LinkType(binding.FuriosaSmiDeviceToDeviceLinkTypeUnknown)
	// LinkTypeInterconnect represents link type under same machine.
	LinkTypeInterconnect = LinkType(binding.FuriosaSmiDeviceToDeviceLinkTypeInterconnect)
	// LinkTypeCpu represents link type under same cpu.
	LinkTypeCpu = LinkType(binding.FuriosaSmiDeviceToDeviceLinkTypeCpu)
	// LinkTypeHostBridge represents link type under same switch.
	LinkTypeHostBridge = LinkType(binding.FuriosaSmiDeviceToDeviceLinkTypeBridge)
	// LinkTypeNoc represents link type under same socket.
	LinkTypeNoc = LinkType(binding.FuriosaSmiDeviceToDeviceLinkTypeNoc)
)

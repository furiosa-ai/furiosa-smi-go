package smi

// Arch represents NPU architecture.
type Arch uint32

const (
	ArchRngd Arch = iota + 1
	ArchRngdMax
	ArchRngdS

	FuriosaSmiArchWarboy  Arch = 0
	FuriosaSmiArchRngd         = ArchRngd
	FuriosaSmiArchRngdMax      = ArchRngdMax
	FuriosaSmiArchRngdS        = ArchRngdS
)

// ToString converts given arch into the string representation.
func (a Arch) ToString() string {
	switch a {
	case FuriosaSmiArchRngd:
		return "rngd"
	case FuriosaSmiArchRngdMax:
		return "rngd-max"
	case FuriosaSmiArchRngdS:
		return "rngd-s"
	default:
		return "unknown"
	}
}

type CoreStatus uint32

type FuriosaSmiCoreStatus = CoreStatus

const (
	CoreStatusAvailable CoreStatus = iota
	CoreStatusOccupied

	FuriosaSmiCoreStatusAvailable = CoreStatusAvailable
	FuriosaSmiCoreStatusOccupied  = CoreStatusOccupied
)

type CoreStatuses interface {
	// PeStatus returns a core status of the device.
	PeStatus() []PeStatus
}

type FuriosaSmiPeStatus struct {
	core   uint32
	status CoreStatus
}

type FuriosaSmiCoreStatuses struct {
	coreStatuses []FuriosaSmiPeStatus
}

var _ CoreStatuses = new(FuriosaSmiCoreStatuses)

func (c *FuriosaSmiCoreStatuses) PeStatus() (ret []PeStatus) {
	ret = make([]PeStatus, len(c.coreStatuses))
	for i := range c.coreStatuses {
		ret[i] = &c.coreStatuses[i]
	}
	return
}

// PeStatus represents a device core status.
type PeStatus interface {
	// Core returns a core index.
	Core() uint32
	// Status returns a core status.
	Status() CoreStatus
}

var _ PeStatus = new(FuriosaSmiPeStatus)

func (p *FuriosaSmiPeStatus) Core() uint32 {
	return p.core
}

func (p *FuriosaSmiPeStatus) Status() CoreStatus {
	return p.status
}

// LinkType represents a topology link type between 2 NPU devices.
type LinkType uint32

const (
	LinkTypeUnknown      LinkType = iota
	LinkTypeInterconnect LinkType = 10
	LinkTypeCpu          LinkType = 20
	LinkTypeHostBridge   LinkType = 30
	LinkTypeNoc          LinkType = 70

	FuriosaSmiLinkTypeUnknown      = LinkTypeUnknown
	FuriosaSmiLinkTypeInterconnect = LinkTypeInterconnect
	FuriosaSmiLinkTypeCpu          = LinkTypeCpu
	FuriosaSmiLinkTypeHostBridge   = LinkTypeHostBridge
	FuriosaSmiLinkTypeNoc          = LinkTypeNoc
)

type FuriosaSmiPeFrequency struct {
	core      uint32
	frequency uint32
}

type FuriosaSmiCoreFrequency struct {
	pe []FuriosaSmiPeFrequency
}

type FuriosaSmiMemoryFrequency struct {
	frequency uint32
}

type PeFrequency interface {
	Core() uint32
	Frequency() uint32
}

var _ PeFrequency = new(FuriosaSmiPeFrequency)

func (p *FuriosaSmiPeFrequency) Core() uint32 {
	return p.core
}

func (p *FuriosaSmiPeFrequency) Frequency() uint32 {
	return p.frequency
}

type CoreFrequency interface {
	PeFrequency() []PeFrequency
}

var _ CoreFrequency = new(FuriosaSmiCoreFrequency)

func (c *FuriosaSmiCoreFrequency) PeFrequency() (ret []PeFrequency) {
	for i := 0; i < len(c.pe); i++ {
		ret = append(ret, &c.pe[i])
	}
	return
}

type MemoryFrequency interface {
	Frequency() uint32
}

var _ MemoryFrequency = new(FuriosaSmiMemoryFrequency)

func (m *FuriosaSmiMemoryFrequency) Frequency() uint32 {
	return m.frequency
}

// GovernorProfile Represents a governor profile
type GovernorProfile uint32

const (
	GovernorProfilePerformance GovernorProfile = iota
	GovernorProfilePowerSave

	FuriosaSmiGovernorProfilePerformance = GovernorProfilePerformance
	FuriosaSmiGovernorProfilePowerSave   = GovernorProfilePowerSave
)

func (p GovernorProfile) String() string {
	switch p {
	case FuriosaSmiGovernorProfilePerformance:
		return "Performance"

	case FuriosaSmiGovernorProfilePowerSave:
		return "PowerSave"

	default: // should not reach here!
		return "Unknown"
	}
}

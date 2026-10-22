package arch

// furiosa-smi/src/arch.rs

import "fmt"

// Arch represents the architecture type of a device.
type Arch int

const (
	// RNGD is the RNGD architecture.
	RNGD Arch = iota
)

func (a Arch) String() string {
	switch a {
	case RNGD:
		return "RNGD"

	default:
		return fmt.Sprintf("Arch(%d)", int(a))
	}
}

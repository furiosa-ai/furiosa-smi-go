package device

// furiosa-smi/src/device/core_status.rs

import "fmt"

// CoreStatus represents a core status.
type CoreStatus int

const (
	// CoreStatusAvailable means a core is available.
	CoreStatusAvailable CoreStatus = iota
	// CoreStatusOccupied means a core is occupied.
	CoreStatusOccupied
)

func (s CoreStatus) String() string {
	switch s {
	case CoreStatusAvailable:
		return "Available"

	case CoreStatusOccupied:
		return "Occupied"

	default:
		return fmt.Sprintf("Unknown CoreStatus(%d)", int(s))
	}
}

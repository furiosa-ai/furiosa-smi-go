package device

// furiosa-smi/src/device/governor.rs

import "fmt"

// GovernorProfile represents a governor profile of the device.
type GovernorProfile int

const (
	// GovernorProfilePerformance is the Performance governor profile.
	GovernorProfilePerformance GovernorProfile = iota
	// GovernorProfilePowerSave is the PowerSave governor profile.
	GovernorProfilePowerSave
)

func (g GovernorProfile) String() string {
	switch g {
	case GovernorProfilePerformance:
		return "Performance"

	case GovernorProfilePowerSave:
		return "PowerSave"

	default:
		return fmt.Sprintf("Unknown GovernorProfile(%d)", int(g))
	}
}

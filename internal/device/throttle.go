package device

// furiosa-smi/src/device/throttle.rs

import "fmt"

// ThrottleReason identifies a single active reason a device is throttled,
// decoded from mgmt's throttle_reason bitmask. Rust represents this as an
// enum with an Unknown(u32) catch-all for unrecognized bits; Go instead
// keeps ThrottleReason as its own bit value, matching how bitmask-shaped
// data is idiomatically represented in Go -- an unrecognized bit is simply
// a ThrottleReason value that doesn't match any of the named constants.
type ThrottleReason uint32

const (
	// ThrottleReasonIdle means throttling in idle or unused state.
	ThrottleReasonIdle ThrottleReason = 1 << 0
	// ThrottleReasonThermalSlowdown means throttling triggered by high temperature.
	ThrottleReasonThermalSlowdown ThrottleReason = 1 << 1
	// ThrottleReasonAppPowerCap means throttling due to host-defined power limit.
	ThrottleReasonAppPowerCap ThrottleReason = 1 << 2
	// ThrottleReasonAppClockCap means throttling due to host-defined clock limit.
	ThrottleReasonAppClockCap ThrottleReason = 1 << 3
	// ThrottleReasonHwClockCap means throttling from device-internal clock limit.
	ThrottleReasonHwClockCap ThrottleReason = 1 << 4
	// ThrottleReasonHwBusLimit means throttling from internal bus/NoC bandwidth limit.
	ThrottleReasonHwBusLimit ThrottleReason = 1 << 5
	// ThrottleReasonHwPowerCap means throttling from device-enforced power limit.
	ThrottleReasonHwPowerCap ThrottleReason = 1 << 6
	// ThrottleReasonGovClockCap means throttling due to governor-defined clock limit.
	ThrottleReasonGovClockCap ThrottleReason = 1 << 7
	// ThrottleReasonOtherReason means throttling due to other undefined reasons.
	ThrottleReasonOtherReason ThrottleReason = 1 << 8
)

var knownThrottleReasons = []ThrottleReason{
	ThrottleReasonIdle,
	ThrottleReasonThermalSlowdown,
	ThrottleReasonAppPowerCap,
	ThrottleReasonAppClockCap,
	ThrottleReasonHwClockCap,
	ThrottleReasonHwBusLimit,
	ThrottleReasonHwPowerCap,
	ThrottleReasonGovClockCap,
	ThrottleReasonOtherReason,
}

// ThrottleReasonsFromBits decomposes a raw throttle_reason bitmask into its
// individual set reasons -- known reasons first (in declaration order),
// then any remaining unrecognized bits as their own ThrottleReason values.
func ThrottleReasonsFromBits(bits uint32) []ThrottleReason {
	remaining := bits
	var reasons []ThrottleReason
	for _, reason := range knownThrottleReasons {
		if bits&uint32(reason) != 0 {
			reasons = append(reasons, reason)
			remaining &^= uint32(reason)
		}
	}

	for position := range uint(32) {
		bit := uint32(1) << position
		if remaining&bit != 0 {
			reasons = append(reasons, ThrottleReason(bit))
		}
	}

	return reasons
}

func (r ThrottleReason) String() string {
	switch r {
	case ThrottleReasonIdle:
		return "Idle"

	case ThrottleReasonThermalSlowdown:
		return "Thermal Slowdown"

	case ThrottleReasonAppPowerCap:
		return "App Power Cap"

	case ThrottleReasonAppClockCap:
		return "App Clock Cap"

	case ThrottleReasonHwClockCap:
		return "HW Clock Cap"

	case ThrottleReasonHwBusLimit:
		return "HW Bus Limit"

	case ThrottleReasonHwPowerCap:
		return "HW Power Cap"

	case ThrottleReasonGovClockCap:
		return "Gov Clock Cap"

	case ThrottleReasonOtherReason:
		return "Other Reason"

	default:
		return fmt.Sprintf("Unknown(0x%08x)", uint32(r))
	}
}

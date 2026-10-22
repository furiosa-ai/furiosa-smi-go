package device

// furiosa-smi/src/device/core_range.rs

import (
	"cmp"
	"fmt"
	"regexp"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
)

// Matches "npuN", "npuNpeS" or "npuNpeS-E".
var deviceFilePattern = regexp.MustCompile(`^npu(?P<dev_node_index>\d+)((?:pe)(?P<start_core>\d+)(-(?P<end_core>\d+))?)?$`)

// DeviceCoreRange is a range of NPU cores designated by a device file name
// (e.g. "npu0pe0-3").
type DeviceCoreRange struct {
	devNodeIndex uint8
	startCore    uint32
	endCore      uint32
}

// NewDeviceCoreRange creates a new device core range.
func NewDeviceCoreRange(devNodeIndex uint8, startCore, endCore uint32) DeviceCoreRange {
	return DeviceCoreRange{devNodeIndex: devNodeIndex, startCore: startCore, endCore: endCore}
}

// ParseDeviceCoreRange parses a device file name such as "npu0", "npu0pe1",
// or "npu0pe1-3".
func ParseDeviceCoreRange(name string) (DeviceCoreRange, error) {
	matches := deviceFilePattern.FindStringSubmatch(name)
	if matches == nil {
		return DeviceCoreRange{}, smierrors.Parse(fmt.Sprintf("invalid device file name: '%s'", name))
	}

	group := func(groupName string) string {
		for i, subexpName := range deviceFilePattern.SubexpNames() {
			if subexpName == groupName {
				return matches[i]
			}
		}

		return ""
	}

	devNodeIndex, err := strconv.ParseUint(group("dev_node_index"), 10, 8)
	if err != nil {
		return DeviceCoreRange{}, smierrors.Parse(fmt.Sprintf("invalid device file name: '%s'", name))
	}

	startCore := uint64(0)
	if raw := group("start_core"); raw != "" {
		startCore, err = strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return DeviceCoreRange{}, smierrors.Parse(fmt.Sprintf("invalid device file name: '%s'", name))
		}
	}

	endCore := startCore
	if raw := group("end_core"); raw != "" {
		endCore, err = strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return DeviceCoreRange{}, smierrors.Parse(fmt.Sprintf("invalid device file name: '%s'", name))
		}
	}

	return DeviceCoreRange{
		devNodeIndex: uint8(devNodeIndex),
		startCore:    uint32(startCore),
		endCore:      uint32(endCore),
	}, nil
}

// DevNodeIndex returns the device node index (the `N` in `npuN`).
func (r DeviceCoreRange) DevNodeIndex() uint8 {
	return r.devNodeIndex
}

// StartCore returns the first core of the range.
func (r DeviceCoreRange) StartCore() uint32 {
	return r.startCore
}

// EndCore returns the last core of the range.
func (r DeviceCoreRange) EndCore() uint32 {
	return r.endCore
}

// Cores returns the core indices designated by this range.
func (r DeviceCoreRange) Cores() []uint32 {
	if r.startCore > r.endCore {
		return []uint32{}
	}

	cores := make([]uint32, 0, uint64(r.endCore)-uint64(r.startCore)+1)
	for core := r.startCore; ; core++ {
		cores = append(cores, core)
		if core == r.endCore {
			break
		}
	}

	return cores
}

// Compare orders DeviceCoreRanges by device node index, then start core,
// then end core -- matching the Rust source's derived Ord on
// DeviceCoreRange (field declaration order).
func (r DeviceCoreRange) Compare(other DeviceCoreRange) int {
	return cmp.Or(
		cmp.Compare(r.devNodeIndex, other.devNodeIndex),
		cmp.Compare(r.startCore, other.startCore),
		cmp.Compare(r.endCore, other.endCore),
	)
}

func (r DeviceCoreRange) String() string {
	s := fmt.Sprintf("npu%dpe%d", r.devNodeIndex, r.startCore)
	if r.startCore != r.endCore {
		s += fmt.Sprintf("-%d", r.endCore)
	}

	return s
}

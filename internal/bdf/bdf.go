package bdf

// furiosa-smi/src/device/bdf.rs

import (
	"cmp"
	"fmt"
	"regexp"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
)

// Matches "dddd:bb:dd.f" or "bb:dd.f" (domain optional).
var bdfPattern = regexp.MustCompile(`^(?:([0-9a-fA-F]{4}):)?([0-9a-fA-F]{2}):([0-1][0-9a-fA-F])\.([0-7])$`)

// BDF is a PCI domain, bus, device, and function address (e.g.
// "0000:27:00.0").
type BDF struct {
	domain   uint16
	bus      uint8
	device   uint8
	function uint8
}

// ParseBDF parses either "dddd:bb:dd.f" or "bb:dd.f" notation.
func ParseBDF(value string) (BDF, error) {
	matches := bdfPattern.FindStringSubmatch(value)
	if matches == nil {
		return BDF{}, smierrors.Parse(fmt.Sprintf("invalid BDF: '%s'", value))
	}

	var domain uint64
	if matches[1] != "" {
		domain, _ = strconv.ParseUint(matches[1], 16, 16)
	}

	bus, _ := strconv.ParseUint(matches[2], 16, 8)
	device, _ := strconv.ParseUint(matches[3], 16, 8)
	function, _ := strconv.ParseUint(matches[4], 16, 8)

	return BDF{
		domain:   uint16(domain),
		bus:      uint8(bus),
		device:   uint8(device),
		function: uint8(function),
	}, nil
}

func (b BDF) String() string {
	return fmt.Sprintf("%04x:%02x:%02x.%x", b.domain, b.bus, b.device, b.function)
}

func (b BDF) Domain() uint16 {
	return b.domain
}

func (b BDF) Bus() uint8 {
	return b.bus
}

func (b BDF) Device() uint8 {
	return b.device
}

func (b BDF) Function() uint8 {
	return b.function
}

// Compare orders BDFs by domain, then bus, then device, then function --
// matching the Rust source's derived Ord on Bdf (field declaration order).
func (b BDF) Compare(other BDF) int {
	return cmp.Or(
		cmp.Compare(b.domain, other.domain),
		cmp.Compare(b.bus, other.bus),
		cmp.Compare(b.device, other.device),
		cmp.Compare(b.function, other.function),
	)
}

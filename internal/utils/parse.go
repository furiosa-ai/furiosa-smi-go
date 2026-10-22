package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
)

// ParseColonFields parses lines of the form "Key: value" into a map. Not
// specific to any one sysfs attribute's format -- several RNGD attribute
// files (performance "metrics", PE "alloc_status") happen to share this
// same colon-separated-fields shape.
func ParseColonFields(text string) map[string]string {
	fields := make(map[string]string)
	for line := range strings.SplitSeq(text, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}

		fields[strings.TrimSpace(key)] = value
	}

	return fields
}

// ParseHexU64 parses a hex string, with or without a "0x"/"0X" prefix.
func ParseHexU64(s string) (uint64, error) {
	trimmed := strings.TrimSpace(s)
	digits := trimmed
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		digits = trimmed[2:]
	}

	value, err := strconv.ParseUint(digits, 16, 64)
	if err != nil {
		return 0, smierrors.Parse(fmt.Sprintf("invalid hex u64: '%s'", trimmed))
	}

	return value, nil
}

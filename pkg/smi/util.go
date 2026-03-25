package smi

import (
	"fmt"
	"regexp"
	"strconv"
)

// versionRe mirrors the Rust VERSION_PATTERN used in furiosa-smi:
//
//	^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)(?:~(?P<internal>[^,]*))?(,( (?P<hash>.+)))?
//
// Capture groups: [full, major, minor, patch, prerelease(~internal), metadata(, hash)]
var versionRe = regexp.MustCompile(
	`^(\d+)\.(\d+)\.(\d+)(?:~([^,]*))?(?:,\s+(.+))?`,
)

// deviceFileRe mirrors Rust's DEVICE_FILE_PATTERN in furiosa-smi/src/discovery.rs.
// It matches filenames of the form npu{N}pe{S} or npu{N}pe{S}-{E}.
// Bare "npu{N}" without a pe-suffix is intentionally rejected (same as Rust).
var deviceFileRe = regexp.MustCompile(`^npu(\d+)pe(\d+)(?:-(\d+))?$`)

type parsedDevFile struct {
	nodeIdx   uint32
	coreStart uint32
	coreEnd   uint32
}

// parseDeviceFilename parses an rngd device filename into node index and core
// range. Returns false when the name does not match the expected pattern.
func parseDeviceFilename(name string) (parsedDevFile, bool) {
	m := deviceFileRe.FindStringSubmatch(name)
	if m == nil {
		return parsedDevFile{}, false
	}
	nodeIdx, _ := strconv.ParseUint(m[1], 10, 32)
	coreStart, _ := strconv.ParseUint(m[2], 10, 32)
	coreEnd := coreStart
	if m[3] != "" {
		coreEnd, _ = strconv.ParseUint(m[3], 10, 32)
	}
	return parsedDevFile{
		nodeIdx:   uint32(nodeIdx),
		coreStart: uint32(coreStart),
		coreEnd:   uint32(coreEnd),
	}, true
}

// parseVersionInfo parses a semver string of the form produced by the
// furiosa kernel driver, e.g. "2025.1.0, 696efad" or "2025.4.0~dev0, 7e25130".
//
// The mapping onto VersionInfo follows system/mod.rs in furiosa-smi:
//   - prerelease ← ~internal  (e.g. "dev0")
//   - metadata   ← , hash     (e.g. "696efad")
func parseVersionInfo(s string) (VersionInfo, error) {
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return nil, fmt.Errorf("%w: %q", ErrParse, s)
	}

	major, _ := strconv.ParseUint(m[1], 10, 32)
	minor, _ := strconv.ParseUint(m[2], 10, 32)
	patch, _ := strconv.ParseUint(m[3], 10, 32)
	prerelease := m[4] // from ~internal
	metadata := m[5]   // from , hash

	return &FuriosaSmiVersion{
		major:      uint32(major),
		minor:      uint32(minor),
		patch:      uint32(patch),
		metadata:   metadata,
		prerelease: prerelease,
	}, nil
}

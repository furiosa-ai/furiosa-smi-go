package version

// furiosa-smi/src/version.rs

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
)

// Matches "major.minor.patch[~prerelease][, metadata]". Deliberately not
// end-anchored, mirroring the Rust source: trailing content after a valid
// major.minor.patch (that doesn't start with "~" or ", ") is ignored rather
// than rejected.
var versionPattern = regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)(?:~(?P<prerelease>[^,]*))?(?:, (?P<metadata>.+))?`)

// VersionInfo is version information.
type VersionInfo struct {
	major      uint32
	minor      uint32
	patch      uint32
	prerelease *string
	metadata   *string
}

// ParseVersionInfo parses a version string of the form
// "major.minor.patch[~prerelease][, metadata]".
func ParseVersionInfo(input string) (VersionInfo, error) {
	match := versionPattern.FindStringSubmatch(input)
	if match == nil {
		return VersionInfo{}, smierrors.Parse(fmt.Sprintf("couldn't parse version: %s", input))
	}

	groups := make(map[string]string, len(match))
	for i, name := range versionPattern.SubexpNames() {
		if name != "" {
			groups[name] = match[i]
		}
	}

	number := func(name string) (uint32, error) {
		raw, ok := groups[name]
		if !ok || raw == "" {
			return 0, smierrors.Parse(fmt.Sprintf("missing '%s' field in version: %s", name, input))
		}

		value, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return 0, smierrors.Parse(fmt.Sprintf("invalid '%s' field in version: %s", name, input))
		}

		return uint32(value), nil
	}

	major, err := number("major")
	if err != nil {
		return VersionInfo{}, err
	}

	minor, err := number("minor")
	if err != nil {
		return VersionInfo{}, err
	}

	patch, err := number("patch")
	if err != nil {
		return VersionInfo{}, err
	}

	info := VersionInfo{major: major, minor: minor, patch: patch}
	if value := groups["prerelease"]; value != "" {
		info.prerelease = &value
	}

	if value := groups["metadata"]; value != "" {
		info.metadata = &value
	}

	return info, nil
}

func (v VersionInfo) Major() uint32 {
	return v.major
}

func (v VersionInfo) Minor() uint32 {
	return v.minor
}

func (v VersionInfo) Patch() uint32 {
	return v.patch
}

// Prerelease returns the version's prerelease tag, if any.
func (v VersionInfo) Prerelease() *string {
	return v.prerelease
}

// Metadata returns the version's build metadata, if any.
func (v VersionInfo) Metadata() *string {
	return v.metadata
}

func (v VersionInfo) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	if v.prerelease != nil {
		result += fmt.Sprintf("(%s)", *v.prerelease)
	}

	if v.metadata != nil {
		result += fmt.Sprintf(", %s", *v.metadata)
	}

	return result
}

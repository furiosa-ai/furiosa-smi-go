package smi

import (
	"fmt"
)

// VersionInfo represents a version information.
type VersionInfo interface {
	fmt.Stringer // added for `String() string` method

	// Major returns a major part of version.
	Major() uint32
	// Minor returns a minor part of version.
	Minor() uint32
	// Patch returns a patch part of version.
	Patch() uint32
	// Metadata returns a metadata of version.
	Metadata() string
	// Prerelease returns a prerelease of version.
	Prerelease() string
}

var _ VersionInfo = new(FuriosaSmiVersion)

type FuriosaSmiVersion struct {
	major      uint32
	minor      uint32
	patch      uint32
	metadata   string
	prerelease string
}

func (v *FuriosaSmiVersion) Major() uint32 {
	return v.major
}

func (v *FuriosaSmiVersion) Minor() uint32 {
	return v.minor
}

func (v *FuriosaSmiVersion) Patch() uint32 {
	return v.patch
}

func (v *FuriosaSmiVersion) Metadata() string {
	return v.metadata
}

func (v *FuriosaSmiVersion) Prerelease() string {
	return v.prerelease
}

func (v *FuriosaSmiVersion) String() string {
	prerelease := v.Prerelease()

	if prerelease == "" {
		return fmt.Sprintf("%d.%d.%d, %s", v.Major(), v.Minor(), v.Patch(), v.Metadata())
	}

	return fmt.Sprintf("%d.%d.%d(%s), %s", v.Major(), v.Minor(), v.Patch(), prerelease, v.Metadata())
}

// Package apitest builds mock sysfs/devfs fixtures that pkg/api reads in
// place of the real host filesystem, so consumers can test against
// api.ListDevices without NPU hardware.
//
// The mock root is process-global state: no test that calls SetMockRoot
// (directly or via StaticRNGDDevices/StaticRNGDDevice), nor any of its
// subtests, may call t.Parallel().
package apitest

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch/rngd"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

type (
	// MockDevice is one mock device that Build writes a fixture for.
	MockDevice = sysroot.MockDevice
	// MockSpec is a set of mock devices to build together via Build.
	MockSpec = sysroot.MockSpec
	// RNGDMock describes one mock RNGD device. It implements MockDevice.
	RNGDMock = rngd.RngdMock
	// PEMock describes one occupied PE of an RNGDMock.
	PEMock = rngd.PEMock
	// HWMonMock describes the hwmon sensors of an RNGDMock.
	HWMonMock = rngd.HWMonMock
	// PCIeMock describes the PCIe link attributes and upstream topology of
	// an RNGDMock.
	PCIeMock = rngd.PCIeMock
	// SocketMock describes the CPU socket an RNGDMock is local to.
	SocketMock = rngd.SocketMock
)

// SetMockRoot redirects every filesystem root pkg/api reads to
// subdirectories of dir for the lifetime of t. It rewrites process-global
// state, so tests using it must not call t.Parallel().
func SetMockRoot(t *testing.T, dir string) {
	t.Helper()
	sysroot.SetMock(t, dir)
}

// Build writes the fixture for every device in spec under the mock root
// installed by SetMockRoot.
func Build(t *testing.T, spec MockSpec) {
	t.Helper()
	sysroot.BuildMock(t, spec)
}

package sysroot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

// SysrootTestSuite covers how paths resolve under the host root and under a
// mock root. It never leaves a mock root installed between tests, since
// the default-root tests depend on starting from the host root. Per this
// package's rule, no test in this suite may call t.Parallel().
type SysrootTestSuite struct {
	suite.Suite
}

func (*SysrootTestSuite) SetupSuite() {
}

func (*SysrootTestSuite) TearDownSuite() {
}

func (s *SysrootTestSuite) SetupTest() {
	s.Require().False(sysroot.IsMock(), "a previous test left a mock root installed")
}

func (*SysrootTestSuite) TearDownTest() {
}

func (s *SysrootTestSuite) TestDefaultRootsResolveAgainstHostRoot() {
	s.Equal("/sys/bus/pci/devices", sysroot.PCIDevicesRoot())
	s.Equal("/sys/bus/pci/drivers_probe", sysroot.PCIDriversProbePath())
	s.Equal("/sys/devices/system/cpu", sysroot.CPUDevicesRoot())
	s.Equal("/dev/rngd", sysroot.Join("dev/rngd"))
}

func (s *SysrootTestSuite) TestMockRootRedirectsEveryRoot() {
	dir := s.T().TempDir()
	restore := sysroot.ApplyMockRoot(dir)
	defer restore()

	s.True(sysroot.IsMock())
	s.Equal(filepath.Join(dir, "sys/bus/pci/devices"), sysroot.PCIDevicesRoot())
	s.Equal(filepath.Join(dir, "sys/bus/pci/drivers_probe"), sysroot.PCIDriversProbePath())
	s.Equal(filepath.Join(dir, "sys/devices/system/cpu"), sysroot.CPUDevicesRoot())
	s.Equal(filepath.Join(dir, "dev/rngd"), sysroot.Join("dev/rngd"))
}

// TestApplyMockRootRestoresPreviousRoot pins that restore goes back to
// whatever was installed before, not unconditionally to the host root, so
// nested swaps unwind correctly.
func (s *SysrootTestSuite) TestApplyMockRootRestoresPreviousRoot() {
	outer, inner := s.T().TempDir(), s.T().TempDir()

	restoreOuter := sysroot.ApplyMockRoot(outer)
	restoreInner := sysroot.ApplyMockRoot(inner)
	s.Equal(filepath.Join(inner, "dev/rngd"), sysroot.Join("dev/rngd"))

	restoreInner()
	s.Equal(filepath.Join(outer, "dev/rngd"), sysroot.Join("dev/rngd"))

	restoreOuter()
	s.False(sysroot.IsMock())
	s.Equal("/dev/rngd", sysroot.Join("dev/rngd"))
}

// TestDefaultReadDirLeavesSymlinksAlone contrasts with the mock root, which
// reports symlinks as character devices: on the host root, entries come
// back exactly as the OS reports them.
func (s *SysrootTestSuite) TestDefaultReadDirLeavesSymlinksAlone() {
	dir := s.T().TempDir()
	target := filepath.Join(dir, "target")
	s.Require().NoError(os.WriteFile(target, nil, 0o644))
	s.Require().NoError(os.Symlink(target, filepath.Join(dir, "link")))

	entries, err := sysroot.ReadDir(dir)
	s.Require().NoError(err)

	types := make(map[string]os.FileMode, len(entries))
	for _, entry := range entries {
		types[entry.Name()] = entry.Type()
	}

	s.NotZero(types["link"] & os.ModeSymlink)
	s.Zero(types["link"] & os.ModeCharDevice)
	s.True(types["target"].IsRegular())
}

func (s *SysrootTestSuite) TestDefaultReadDirPropagatesMissingDirectory() {
	_, err := sysroot.ReadDir(filepath.Join(s.T().TempDir(), "absent"))
	s.Error(err)
}

func TestSysrootSuite(t *testing.T) {
	suite.Run(t, new(SysrootTestSuite))
}

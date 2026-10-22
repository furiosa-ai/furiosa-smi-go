package sysroot_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

// recordingMock is a MockDevice that only records that it was built, so
// BuildMock's fan-out can be checked without any real fixture.
type recordingMock struct {
	name  string
	built *[]string
}

func (m *recordingMock) Build(*testing.T) {
	*m.built = append(*m.built, m.name)
}

// SysrootMockTestSuite covers the test-support surface in mock.go. Each
// test installs its own mock root when it needs one, since SetMock's
// install-and-restore behavior is itself under test. Per this package's
// rule, no test in this suite may call t.Parallel().
type SysrootMockTestSuite struct {
	suite.Suite
}

func (*SysrootMockTestSuite) SetupSuite() {
}

func (*SysrootMockTestSuite) TearDownSuite() {
}

func (s *SysrootMockTestSuite) SetupTest() {
	s.Require().False(sysroot.IsMock(), "a previous test left a mock root installed")
}

func (*SysrootMockTestSuite) TearDownTest() {
}

// readDirByName lists dir through sysroot.ReadDir and indexes the entries by
// name.
func (s *SysrootMockTestSuite) readDirByName(dir string) map[string]fs.DirEntry {
	s.T().Helper()
	entries, err := sysroot.ReadDir(dir)
	s.Require().NoError(err)
	byName := make(map[string]fs.DirEntry, len(entries))
	for _, entry := range entries {
		byName[entry.Name()] = entry
	}

	return byName
}

// TestReadDirReportsSymlinksAsCharDevices covers why the mock root has its
// own ReadDir: mocks stand in for /dev/rngd character devices with plain
// symlinks, and callers filtering on character devices must see them as
// such through every way they can inspect an entry.
func (s *SysrootMockTestSuite) TestReadDirReportsSymlinksAsCharDevices() {
	sysroot.SetMock(s.T(), s.T().TempDir())
	dir := sysroot.Join("dev/rngd")
	target := filepath.Join(s.T().TempDir(), "null-target")
	sysroot.MustWriteFile(s.T(), target, "")
	sysroot.MustMkdirAll(s.T(), filepath.Join(dir, "subdir"))
	sysroot.MustWriteFile(s.T(), filepath.Join(dir, "plain"), "")
	s.Require().NoError(os.Symlink(target, filepath.Join(dir, "npu0pe0")))

	entries := s.readDirByName(dir)
	s.Require().Len(entries, 3)

	device := entries["npu0pe0"]
	s.Equal("npu0pe0", device.Name())
	s.Equal(fs.ModeCharDevice, device.Type())
	s.False(device.IsDir())
	info, err := device.Info()
	s.Require().NoError(err)
	s.NotZero(info.Mode() & fs.ModeCharDevice)
	targetInfo, err := os.Stat(target)
	s.Require().NoError(err)
	s.Equal(targetInfo.Mode().Perm(), info.Mode().Perm(), "permission bits come from the link target")

	s.True(entries["plain"].Type().IsRegular())
	s.True(entries["subdir"].IsDir())
}

// TestReadDirLeavesDanglingSymlinkUnwrapped pins that only a symlink whose
// target exists is dressed up as a character device; a broken one is
// reported as the symlink it is.
func (s *SysrootMockTestSuite) TestReadDirLeavesDanglingSymlinkUnwrapped() {
	sysroot.SetMock(s.T(), s.T().TempDir())
	dir := sysroot.Join("dev/rngd")
	sysroot.MustMkdirAll(s.T(), dir)
	s.Require().NoError(os.Symlink(filepath.Join(dir, "absent"), filepath.Join(dir, "npu0pe0")))

	entries := s.readDirByName(dir)
	s.NotZero(entries["npu0pe0"].Type() & fs.ModeSymlink)
	s.Zero(entries["npu0pe0"].Type() & fs.ModeCharDevice)
}

func (s *SysrootMockTestSuite) TestReadDirPropagatesMissingDirectory() {
	sysroot.SetMock(s.T(), s.T().TempDir())

	_, err := sysroot.ReadDir(sysroot.Join("dev/rngd"))
	s.Error(err)
}

// TestSetMockRestoresHostRootWhenTestEnds covers SetMock's contract: the
// swap is scoped to the test that made it, so nothing leaks into the next.
func (s *SysrootMockTestSuite) TestSetMockRestoresHostRootWhenTestEnds() {
	dir := s.T().TempDir()
	s.Run("installs", func() {
		sysroot.SetMock(s.T(), dir)
		s.True(sysroot.IsMock())
		s.Equal(filepath.Join(dir, "dev/rngd"), sysroot.Join("dev/rngd"))
	})

	s.False(sysroot.IsMock())
	s.Equal("/dev/rngd", sysroot.Join("dev/rngd"))
}

// TestHostPath covers HostPath under both roots: the mock root's prefix is
// stripped so presented paths read like host paths, and the default root
// leaves paths alone.
func (s *SysrootMockTestSuite) TestHostPath() {
	s.Equal("/dev/rngd/npu0pe0-3", sysroot.HostPath("/dev/rngd/npu0pe0-3"), "default root")

	dir := s.T().TempDir()
	s.Run("mock", func() {
		sysroot.SetMock(s.T(), dir)
		for _, test := range []struct {
			name, path, want string
		}{
			{name: "joined path", path: sysroot.Join("dev/rngd/npu0pe0-3"), want: "/dev/rngd/npu0pe0-3"},
			{name: "mock root itself", path: dir, want: "/"},
			{name: "outside mock root", path: "/dev/rngd/npu0pe0", want: "/dev/rngd/npu0pe0"},
			{name: "sibling sharing a prefix", path: dir + "-other/dev", want: dir + "-other/dev"},
		} {
			s.Run(test.name, func() {
				s.Equal(test.want, sysroot.HostPath(test.path))
			})
		}
	})
}

func (s *SysrootMockTestSuite) TestBuildMockBuildsEveryDeviceInOrder() {
	var built []string
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{
		&recordingMock{name: "first", built: &built},
		&recordingMock{name: "second", built: &built},
	}})
	s.Equal([]string{"first", "second"}, built)
}

func (s *SysrootMockTestSuite) TestOrDefault() {
	for _, test := range []struct {
		name, value, fallback, want string
	}{
		{name: "empty value takes fallback", value: "", fallback: "1.0.0", want: "1.0.0"},
		{name: "set value wins", value: "2.0.0", fallback: "1.0.0", want: "2.0.0"},
		{name: "whitespace is a value", value: " ", fallback: "1.0.0", want: " "},
	} {
		s.Run(test.name, func() {
			s.Equal(test.want, sysroot.OrDefault(test.value, test.fallback))
		})
	}
}

func (s *SysrootMockTestSuite) TestMustWriteFileCreatesMissingParents() {
	path := filepath.Join(s.T().TempDir(), "sys", "class", "rngd_mgmt", "rngd!npu0mgmt", "busname")

	sysroot.MustWriteFile(s.T(), path, "0000:27:00.0\n")

	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal("0000:27:00.0\n", string(content))
}

func TestSysrootMockSuite(t *testing.T) {
	suite.Run(t, new(SysrootMockTestSuite))
}

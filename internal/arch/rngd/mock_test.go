package rngd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

// RngdMockTestSuite demonstrates sysroot.MockDevice's intended usage: build
// one mock RNGD device via sysroot.BuildMock, then confirm the resulting
// sysfs/devfs fixture is what a real device would look like. Later PRs'
// production code (internal/discovery, pkg/device) will read these same
// mock files through sysroot.PCIDevicesRoot/rngd.RNGDManagementRoot/
// rngd.RNGDDevicesRoot instead of asserting on raw file contents directly.
type RngdMockTestSuite struct {
	suite.Suite
}

func (*RngdMockTestSuite) SetupSuite() {}

func (*RngdMockTestSuite) TearDownSuite() {}

// SetupTest points internal/sysroot at a fresh mock directory. Per that
// package's rule, no test in this suite may call t.Parallel().
func (s *RngdMockTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
}

func (*RngdMockTestSuite) TearDownTest() {}

func (s *RngdMockTestSuite) readFile(path string) string {
	s.T().Helper()
	content, err := os.ReadFile(path)
	s.Require().NoError(err)

	return string(content)
}

func (s *RngdMockTestSuite) TestBuildMockRNGD() {
	sysroot.BuildMock(s.T(), sysroot.MockSpec{
		Devices: []sysroot.MockDevice{&RngdMock{
			BDF:          "0000:27:00.0",
			DevNodeIndex: 0,
			NUMANode:     0,
			UUID:         "test-uuid",
			Serial:       "test-serial",
			CoreNum:      2,
		}},
	})

	pciDir := filepath.Join(sysroot.PCIDevicesRoot(), "0000:27:00.0")
	s.Equal("0x1ed2\n", s.readFile(filepath.Join(pciDir, "vendor")))
	s.Equal("0x0001\n", s.readFile(filepath.Join(pciDir, "device")))
	s.Equal("0\n", s.readFile(filepath.Join(pciDir, "numa_node")))
	_, err := os.Stat(filepath.Join(pciDir, "driver"))
	s.Require().NoError(err, "an enabled mock device must have a driver directory")

	mgmtDir := filepath.Join(RNGDManagementRoot(), "rngd!npu0mgmt")
	s.Equal("0000:27:00.0\n", s.readFile(filepath.Join(mgmtDir, "busname")))
	s.Equal("test-uuid\n", s.readFile(filepath.Join(mgmtDir, "device_uuid")))
	s.Equal("test-serial\n", s.readFile(filepath.Join(mgmtDir, "device_sn")))
	s.Equal("performance\n", s.readFile(filepath.Join(mgmtDir, "npu_governor")))

	for core := range 2 {
		link := filepath.Join(RNGDDeviceRoot(), fmt.Sprintf("npu0pe%d", core))
		info, err := os.Lstat(link)
		s.Require().NoError(err)
		s.NotZero(info.Mode()&os.ModeSymlink, "devfs entry npu0pe%d should be a symlink standing in for a char device", core)
	}
}

// TestBuildMockDisabledSkipsMgmtAndDevfs confirms a Disabled mock only gets
// the bare PCI identity files -- no driver directory, no mgmt node, no
// devfs entries -- matching a device the kernel hasn't bound a driver to.
func (s *RngdMockTestSuite) TestBuildMockDisabledSkipsMgmtAndDevfs() {
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{&RngdMock{
		BDF:      "0000:99:00.0",
		Disabled: true,
	}}})

	pciDir := filepath.Join(sysroot.PCIDevicesRoot(), "0000:99:00.0")
	_, err := os.Stat(filepath.Join(pciDir, "vendor"))
	s.Require().NoError(err)

	_, err = os.Stat(filepath.Join(pciDir, "driver"))
	s.Require().Error(err, "a disabled mock device must not have a driver directory")
}

func (s *RngdMockTestSuite) TestBuildMockFusedPEs() {
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{&RngdMock{
		BDF:          "0000:27:00.0",
		DevNodeIndex: 1,
		CoreNum:      4,
		FusedPEs:     [][2]uint32{{0, 1}, {0, 3}},
	}}})

	var names []string
	for _, file := range GetDeviceFiles(RNGDDeviceRoot(), 1) {
		names = append(names, filepath.Base(file.Path()))
	}

	s.Equal([]string{"npu1pe0", "npu1pe0-1", "npu1pe0-3", "npu1pe1", "npu1pe2", "npu1pe3"}, names)
	s.Equal(uint32(4), GetCoreCount(RNGDDeviceRoot(), 1), "fused entries must not count as cores")
}

func (s *RngdMockTestSuite) TestBuildMockDevNumber() {
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{
		&RngdMock{BDF: "0000:27:00.0", DevNodeIndex: 0, DevNumber: "234:0"},
		&RngdMock{BDF: "0000:2a:00.0", DevNodeIndex: 1},
	}})

	major, minor, err := ReadMajorMinor(FormatMgmtPath(0))
	s.Require().NoError(err)
	s.Equal([2]uint16{234, 0}, [2]uint16{major, minor})

	major, minor, err = ReadMajorMinor(FormatMgmtPath(1))
	s.Require().NoError(err)
	s.Equal([2]uint16{240, 1}, [2]uint16{major, minor}, "default is 240:<DevNodeIndex>")
}

func (s *RngdMockTestSuite) TestBuildMockP2P() {
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{
		&RngdMock{BDF: "0000:27:00.0", DevNodeIndex: 0, CoreNum: 2, P2P: true},
		&RngdMock{BDF: "0000:2a:00.0", DevNodeIndex: 1, CoreNum: 2},
	}})

	s.True(HasP2PInterfaces(RNGDDeviceRoot(), 0))
	s.False(HasP2PInterfaces(RNGDDeviceRoot(), 1))
	s.Len(GetDeviceFiles(RNGDDeviceRoot(), 0), 2, "P2P entries must not be enumerated as device files")
	s.Equal(map[uint8]struct{}{0: {}, 1: {}}, GetDevNodeIndices(RNGDDeviceRoot()))
}

func (s *RngdMockTestSuite) TestBuildMockSocket() {
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{
		&RngdMock{BDF: "0000:27:00.0", DevNodeIndex: 0, Socket: &SocketMock{FirstCPU: 0, PackageID: 0}},
		&RngdMock{BDF: "0000:9e:00.0", DevNodeIndex: 1, Socket: &SocketMock{FirstCPU: 32, PackageID: 1}},
		&RngdMock{BDF: "0000:c7:00.0", DevNodeIndex: 2},
	}})

	pciDevices := sysroot.PCIDevicesRoot()
	s.Equal("00000001\n", s.readFile(filepath.Join(pciDevices, "0000:27:00.0", "local_cpus")))
	s.Equal("00000001,00000000\n", s.readFile(filepath.Join(pciDevices, "0000:9e:00.0", "local_cpus")))
	s.NoFileExists(filepath.Join(pciDevices, "0000:c7:00.0", "local_cpus"))

	cpus := sysroot.CPUDevicesRoot()
	s.Equal("0\n", s.readFile(filepath.Join(cpus, "cpu0", "topology", "physical_package_id")))
	s.Equal("1\n", s.readFile(filepath.Join(cpus, "cpu32", "topology", "physical_package_id")))
}

func TestRngdMockSuite(t *testing.T) {
	suite.Run(t, new(RngdMockTestSuite))
}

// TestValidateFusedPE covers the range check Build applies to each FusedPEs
// entry.
func TestValidateFusedPE(t *testing.T) {
	for _, test := range []struct {
		fused   [2]uint32
		wantErr string
	}{
		{fused: [2]uint32{0, 7}},
		{fused: [2]uint32{1, 1}, wantErr: "fused PE range 1-1: want start < end < CoreNum (8)"},
		{fused: [2]uint32{2, 1}, wantErr: "fused PE range 2-1: want start < end < CoreNum (8)"},
		{fused: [2]uint32{0, 8}, wantErr: "fused PE range 0-8: want start < end < CoreNum (8)"},
	} {
		t.Run(fmt.Sprintf("%d-%d", test.fused[0], test.fused[1]), func(t *testing.T) {
			err := validateFusedPE(test.fused, 8)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("validateFusedPE(%v, 8) = %v, want nil", test.fused, err)
				}

				return
			}

			if err == nil || err.Error() != test.wantErr {
				t.Fatalf("validateFusedPE(%v, 8) = %v, want %q", test.fused, err, test.wantErr)
			}
		})
	}
}

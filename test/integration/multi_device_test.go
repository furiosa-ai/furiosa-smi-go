package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch/rngd"
	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

const multiMockCoreNum = 8

var multiMockDevices = []rngd.RngdMock{
	{BDF: "0000:22:00.0", DevNodeIndex: 3, NUMANode: 1, UUID: "uuid-22", Serial: "serial-22", CoreNum: multiMockCoreNum},
	{BDF: "0000:27:00.0", DevNodeIndex: 1, NUMANode: 0, UUID: "uuid-27", Serial: "serial-27", CoreNum: multiMockCoreNum},
	{BDF: "0000:2a:00.0", DevNodeIndex: 0, NUMANode: 0, UUID: "uuid-2a", Serial: "serial-2a", CoreNum: multiMockCoreNum},
	{BDF: "0000:2c:00.0", DevNodeIndex: 2, NUMANode: 1, UUID: "uuid-2c", Serial: "serial-2c", CoreNum: multiMockCoreNum},
}

type MultiDeviceTestSuite struct {
	suite.Suite
}

func (*MultiDeviceTestSuite) SetupSuite() {}

func (*MultiDeviceTestSuite) TearDownSuite() {}

// SetupTest builds a fresh mock with all 4 devices enabled before every
// test. Per internal/sysroot's rule, no test in this suite may call
// t.Parallel().
func (s *MultiDeviceTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())

	devices := make([]sysroot.MockDevice, len(multiMockDevices))
	for i, d := range multiMockDevices {
		devices[i] = &d
	}

	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: devices})
}

func (*MultiDeviceTestSuite) TearDownTest() {}

// TestListPCIDevicesSortsByBDF confirms ListPCIDevices sorts by
// vendor/device/BDF when every device shares the same vendor+device ID.
func (s *MultiDeviceTestSuite) TestListPCIDevicesSortsByBDF() {
	devices, err := api.ListPCIDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 4)

	want := []string{"0000:22:00.0", "0000:27:00.0", "0000:2a:00.0", "0000:2c:00.0"}
	for i, wantBDF := range want {
		s.Equal(wantBDF, devices[i].BDF().String(), "position %d", i)
	}
}

// TestListDevicesSortsByDevNodeIndex confirms ListDevices sorts by
// vendor/device/devNodeIndex, which is deliberately not the same order as
// the fixture's BDFs.
func (s *MultiDeviceTestSuite) TestListDevicesSortsByDevNodeIndex() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 4)

	want := []string{"0000:2a:00.0", "0000:27:00.0", "0000:2c:00.0", "0000:22:00.0"}
	for i, wantBDF := range want {
		s.Equal(wantBDF, devices[i].BDF().String(), "position %d", i)
		s.Equal(uint8(i), devices[i].DevNodeIndex())
	}
}

// TestEachDeviceResolvesItsOwnData confirms discovery doesn't cross-wire
// per-device fields (UUID, serial, NUMA node, core status) once more than
// one device is present.
func (s *MultiDeviceTestSuite) TestEachDeviceResolvesItsOwnData() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 4)

	byBDF := make(map[string]device.Device, len(devices))
	for _, d := range devices {
		byBDF[d.BDF().String()] = d
	}

	for _, spec := range multiMockDevices {
		got, ok := byBDF[spec.BDF]
		s.Require().True(ok, "device %s missing from ListDevices", spec.BDF)
		s.Equal(spec.UUID, got.UUID())
		s.Equal(spec.Serial, got.Serial())
		s.Equal(spec.NUMANode, got.NUMANode())
		s.Equal(spec.DevNodeIndex, got.DevNodeIndex())
		s.Equal(uint32(multiMockCoreNum), got.CoreNum())

		coreStatus, err := got.CoreStatus()
		s.Require().NoError(err)
		s.Len(coreStatus, multiMockCoreNum)
	}
}

// TestListDisabledDevicesAmongMultiple disables one of the four devices --
// removing its driver binding along with the mgmt/devfs nodes the real
// driver would tear down alongside it -- and confirms it's excluded from
// ListDevices, still present in ListPCIDevices, and reported by
// ListDisabledDevices, while the other three stay enabled.
func (s *MultiDeviceTestSuite) TestListDisabledDevicesAmongMultiple() {
	disabled := multiMockDevices[1] // 0000:27:00.0, devNodeIndex 1

	s.Require().NoError(os.RemoveAll(filepath.Join(sysroot.PCIDevicesRoot(), disabled.BDF, "driver")))
	s.Require().NoError(os.RemoveAll(filepath.Join(rngd.RNGDManagementRoot(), fmt.Sprintf("rngd!npu%dmgmt", disabled.DevNodeIndex))))
	for core := range multiMockCoreNum {
		s.Require().NoError(os.Remove(filepath.Join(rngd.RNGDDeviceRoot(), fmt.Sprintf("npu%dpe%d", disabled.DevNodeIndex, core))))
	}

	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 3)
	for _, d := range devices {
		s.NotEqual(disabled.BDF, d.BDF().String())
	}

	pciDevices, err := api.ListPCIDevices()
	s.Require().NoError(err)
	s.Require().Len(pciDevices, 4, "disabling a device must not remove it from ListPCIDevices")

	disabledList, err := api.ListDisabledDevices()
	s.Require().NoError(err)
	s.Require().Len(disabledList, 1)
	s.Equal(disabled.BDF, disabledList[0].BDF().String())
}

func TestMultiDeviceSuite(t *testing.T) {
	suite.Run(t, new(MultiDeviceTestSuite))
}

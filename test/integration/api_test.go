package integration_test

// Black-box tests against pkg/api: build one mock RNGD device,
// call every exported API function, check the results.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch/rngd"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

const (
	mockBDF      = "0000:27:00.0"
	mockVendorID = "0x1ed2"
	mockDeviceID = "0x0001"
	mockUUID     = "test-uuid"
	mockSerial   = "test-serial"
	mockFWVer    = "1.0.0"
	mockDrvVer   = "2025.1.0"
	mockCoreNum  = 8
)

type APITestSuite struct {
	suite.Suite
}

func (*APITestSuite) SetupSuite() {
}

func (*APITestSuite) TearDownSuite() {
}

// SetupTest builds a fresh, isolated mock for every test: a single
// enabled RNGD device with 8 accessible cores.
func (s *APITestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{&rngd.RngdMock{
		BDF:             mockBDF,
		DevNodeIndex:    0,
		NUMANode:        0,
		UUID:            mockUUID,
		Serial:          mockSerial,
		VendorHex:       mockVendorID,
		DeviceHex:       mockDeviceID,
		FirmwareVersion: mockFWVer,
		DriverVersion:   mockDrvVer,
		CoreNum:         mockCoreNum,
	}}})
}

func (*APITestSuite) TearDownTest() {
}

func (s *APITestSuite) TestParseBDF() {
	got, err := api.ParseBDF(mockBDF)
	s.Require().NoError(err)
	s.Equal(mockBDF, got.String())
}

func (s *APITestSuite) TestDeviceCoreRange() {
	built := api.NewDeviceCoreRange(0, 2, 5)
	s.Equal([]uint32{2, 3, 4, 5}, built.Cores())

	parsed, err := api.ParseDeviceCoreRange("npu0pe2-5")
	s.Require().NoError(err)
	s.Equal(built, parsed)
}

func (s *APITestSuite) TestListPCIDevices() {
	devices, err := api.ListPCIDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 1)
	s.Equal(mockBDF, devices[0].BDF().String())
	s.Equal(uint16(0x0001), devices[0].DeviceID())
	s.Equal(uint16(0x1ed2), devices[0].VendorID())
}

func (s *APITestSuite) TestGetPCIDevice() {
	target, err := api.ParseBDF(mockBDF)
	s.Require().NoError(err)

	got, err := api.GetPCIDevice(target)
	s.Require().NoError(err)
	s.Equal(mockBDF, got.BDF().String())
}

func (s *APITestSuite) TestListDevices() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 1)

	got := devices[0]
	s.Equal(mockBDF, got.BDF().String())
	s.Equal(mockUUID, got.UUID())
	s.Equal(mockSerial, got.Serial())
	s.Equal(uint32(mockCoreNum), got.CoreNum())
	s.Equal(int32(0), got.NUMANode())

	firmwareVersion, err := got.FirmwareVersion()
	s.Require().NoError(err)
	s.Equal(mockFWVer, firmwareVersion.String())

	coreStatus, err := got.CoreStatus()
	s.Require().NoError(err)
	s.Len(coreStatus, mockCoreNum)
}

func (s *APITestSuite) TestListDisabledDevices() {
	// The one mock device is driver-bound (enabled), so it should not
	// appear here.
	disabled, err := api.ListDisabledDevices()
	s.Require().NoError(err)
	s.NotNil(disabled)
	s.Empty(disabled)
}

func (s *APITestSuite) TestDriverInfo() {
	info, err := api.DriverInfo()
	s.Require().NoError(err)
	s.Equal(uint32(2025), info.Major())
	s.Equal(uint32(1), info.Minor())
	s.Equal(uint32(0), info.Patch())
}

func (s *APITestSuite) TestDeviceFiles() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 1)

	files, err := devices[0].DeviceFiles()
	s.Require().NoError(err)
	s.Require().Len(files, mockCoreNum)

	seenCores := make(map[uint32]struct{}, mockCoreNum)
	for _, file := range files {
		s.Require().Len(file.Cores(), 1)
		seenCores[file.Cores()[0]] = struct{}{}
		s.Contains(file.Path(), "npu0pe")
	}

	s.Len(seenCores, mockCoreNum)
}

func (s *APITestSuite) TestP2PAccessible() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 1)

	// The mock doesn't create P2P DMA interface files, so this must be
	// false rather than erroring.
	s.False(devices[0].P2PAccessible(&devices[0]))
}

// TestP2PAccessibleIgnoresAvailability confirms P2PAccessible checks devfs
// P2P interface files directly, without gating on driver-bound availability
// like most other Device methods -- matching Rust's p2p_accessible, which
// never calls ensure_available.
func (s *APITestSuite) TestP2PAccessibleIgnoresAvailability() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 1)

	s.Require().NoError(os.RemoveAll(filepath.Join(sysroot.PCIDevicesRoot(), mockBDF, "driver")))

	devRoot := rngd.RNGDDeviceRoot()
	for _, name := range []string{
		"npu0dmar",
		"npu0ch0r", "npu0ch1r", "npu0ch2r", "npu0ch3r",
		"npu0ch4r", "npu0ch5r", "npu0ch6r", "npu0ch7r",
	} {
		s.Require().NoError(os.WriteFile(filepath.Join(devRoot, name), nil, 0o644))
	}

	s.True(devices[0].P2PAccessible(&devices[0]))
}

func (s *APITestSuite) TestTCPFrequency() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 1)

	coreFrequency, err := devices[0].CoreFrequency()
	s.Require().NoError(err)

	tcpFrequency, err := devices[0].TCPFrequency()
	s.Require().NoError(err)
	s.Len(tcpFrequency, len(coreFrequency))
	for core, freq := range coreFrequency {
		s.Equal(freq/2, tcpFrequency[core])
	}
}

func (s *APITestSuite) TestThrottleReason() {
	devices, err := api.ListDevices()
	s.Require().NoError(err)
	s.Require().Len(devices, 1)

	reasons, err := devices[0].ThrottleReason()
	s.Require().NoError(err)
	s.Empty(reasons)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

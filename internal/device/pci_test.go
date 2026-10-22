package device

// furiosa-smi/src/device.rs
// furiosa-smi/src/drain.rs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	internalArch "github.com/furiosa-ai/furiosa-smi-go/internal/arch"
	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

type PCIDeviceTestSuite struct {
	suite.Suite
}

func (*PCIDeviceTestSuite) SetupSuite() {}

func (*PCIDeviceTestSuite) TearDownSuite() {}

// SetupTest points internal/sysroot at a fresh mock directory. Per that
// package's rule, no test in this suite may call t.Parallel().
func (s *PCIDeviceTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
}

func (*PCIDeviceTestSuite) TearDownTest() {}

func (s *PCIDeviceTestSuite) TestPCIDeviceAccessors() {
	deviceBDF, err := bdf.ParseBDF("0000:27:00.0")
	s.Require().NoError(err)

	d := NewPCIDevice(deviceBDF, 0x0001, internalArch.FuriosaVendorID)
	s.Equal(deviceBDF, d.BDF())
	s.Equal(uint16(0x0001), d.DeviceID())
	s.Equal(uint16(internalArch.FuriosaVendorID), d.VendorID())
}

// newPCIDeviceMock creates the mock PCI device directory for bdfStr,
// optionally with a "driver" subdirectory to simulate it being bound.
func (s *PCIDeviceTestSuite) newPCIDeviceMock(bdfStr string, bound bool) PCIDevice {
	deviceBDF, err := bdf.ParseBDF(bdfStr)
	s.Require().NoError(err)

	devDir := filepath.Join(sysroot.PCIDevicesRoot(), bdfStr)
	s.Require().NoError(os.MkdirAll(devDir, 0o755))
	if bound {
		s.Require().NoError(os.MkdirAll(filepath.Join(devDir, "driver"), 0o755))
	}

	return NewPCIDevice(deviceBDF, 0x0001, internalArch.FuriosaVendorID)
}

func (s *PCIDeviceTestSuite) requireDeviceNotFound(err error) {
	s.T().Helper()
	s.Require().Error(err)
	var smiErr smierrors.SMIError
	s.Require().ErrorAs(err, &smiErr)
	s.Equal(smierrors.ErrorKindDeviceNotFound, smiErr.Kind())
}

func (s *PCIDeviceTestSuite) TestDisableRejectsUnknownDevice() {
	deviceBDF, err := bdf.ParseBDF("0000:99:00.0")
	s.Require().NoError(err)
	d := NewPCIDevice(deviceBDF, 0x0001, internalArch.FuriosaVendorID)

	s.requireDeviceNotFound(d.Disable())
}

func (s *PCIDeviceTestSuite) TestEnableRejectsUnknownDevice() {
	deviceBDF, err := bdf.ParseBDF("0000:99:00.0")
	s.Require().NoError(err)
	d := NewPCIDevice(deviceBDF, 0x0001, internalArch.FuriosaVendorID)

	s.requireDeviceNotFound(d.Enable())
}

// TestDisableIsNoOpWhenAlreadyUnbound confirms Disable never errors when
// the device isn't bound to a driver in the first place -- it must
// short-circuit before ever touching driver/unbind.
func (s *PCIDeviceTestSuite) TestDisableIsNoOpWhenAlreadyUnbound() {
	d := s.newPCIDeviceMock("0000:27:00.0", false)

	s.Require().NoError(d.Disable())
}

// TestDisableWritesBDFToUnbindAttribute confirms Disable writes the
// device's own BDF to driver/unbind when it is currently bound.
func (s *PCIDeviceTestSuite) TestDisableWritesBDFToUnbindAttribute() {
	d := s.newPCIDeviceMock("0000:27:00.0", true)

	s.Require().NoError(d.Disable())

	content, err := os.ReadFile(filepath.Join(sysroot.PCIDevicesRoot(), "0000:27:00.0", "driver", "unbind"))
	s.Require().NoError(err)
	s.Equal("0000:27:00.0\n", string(content))
}

// TestEnableIsNoOpWhenAlreadyBound confirms Enable never writes to the
// shared drivers_probe attribute when the device is already bound --
// proven by a sentinel value in that file surviving the call untouched.
func (s *PCIDeviceTestSuite) TestEnableIsNoOpWhenAlreadyBound() {
	d := s.newPCIDeviceMock("0000:27:00.0", true)
	probePath := sysroot.PCIDriversProbePath()
	s.Require().NoError(os.MkdirAll(filepath.Dir(probePath), 0o755))
	s.Require().NoError(os.WriteFile(probePath, []byte("sentinel\n"), 0o644))

	s.Require().NoError(d.Enable())

	content, err := os.ReadFile(probePath)
	s.Require().NoError(err)
	s.Equal("sentinel\n", string(content))
}

// TestEnableReportsDeviceNotAvailableWhenStillUnboundAfterProbe confirms
// Enable writes the device's BDF to drivers_probe, but -- since nothing in
// a mock filesystem can simulate the kernel actually binding a driver --
// still surfaces DeviceNotAvailable when the device remains unbound
// afterward, rather than silently reporting success.
func (s *PCIDeviceTestSuite) TestEnableReportsDeviceNotAvailableWhenStillUnboundAfterProbe() {
	d := s.newPCIDeviceMock("0000:27:00.0", false)

	err := d.Enable()
	s.Require().Error(err)
	var smiErr smierrors.SMIError
	s.Require().ErrorAs(err, &smiErr)
	s.Equal(smierrors.ErrorKindDeviceNotAvailable, smiErr.Kind())

	content, readErr := os.ReadFile(sysroot.PCIDriversProbePath())
	s.Require().NoError(readErr)
	s.Equal("0000:27:00.0\n", string(content))
}

func TestPCIDeviceSuite(t *testing.T) {
	suite.Run(t, new(PCIDeviceTestSuite))
}

package pcie_test

// furiosa-smi/src/pcie.rs
//
// ReadPcieInfo's tests live in an external test package, unlike read_test.go's
// tests for the pure parsers: exercising the read path needs internal/arch/rngd
// to build the sysfs fixture, and rngd reaches internal/pcie through
// internal/device, so an in-package test would be an import cycle.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch/rngd"
	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/pcie"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

const (
	pcieTestBDF    = "0000:27:00.0"
	pcieTestBridge = "0000:00:01.0"
)

type PcieReadTestSuite struct {
	suite.Suite
}

func (*PcieReadTestSuite) SetupSuite() {
}

func (*PcieReadTestSuite) TearDownSuite() {
}

// SetupTest points internal/sysroot at a fresh mock directory. Per that
// package's rule, no test in this suite may call t.Parallel().
func (s *PcieReadTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
}

func (*PcieReadTestSuite) TearDownTest() {
}

// buildMock builds a mock device carrying pcieMock and returns its BDF.
func (s *PcieReadTestSuite) buildMock(pcieMock *rngd.PCIeMock) bdf.BDF {
	s.T().Helper()
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{&rngd.RngdMock{
		BDF:    pcieTestBDF,
		UUID:   "test-uuid",
		Serial: "test-serial",
		PCIe:   pcieMock,
	}}})

	target, err := bdf.ParseBDF(pcieTestBDF)
	s.Require().NoError(err)

	return target
}

// switchedTopology is an endpoint behind a bridge whose bus range spans
// past its own secondary bus, which is what marks a real switch.
func (s *PcieReadTestSuite) switchedTopology() *rngd.PCIeMock {
	return &rngd.PCIeMock{
		Upstream:               pcieTestBridge,
		UpstreamSecondaryBus:   0x27,
		UpstreamSubordinateBus: 0x2f,
	}
}

func (s *PcieReadTestSuite) TestReadPcieInfoReadsDeviceIdentity() {
	target := s.buildMock(s.switchedTopology())

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)

	deviceInfo := info.DeviceInfo()
	s.Equal(uint16(0x0001), deviceInfo.DeviceID())
	s.Equal(uint16(0x0001), deviceInfo.SubsystemID())
	// VendorID reports the *subsystem* vendor, not the "vendor" attribute --
	// matching the Rust source this was ported from, despite the name.
	s.Equal(uint16(0x1ed2), deviceInfo.VendorID())
	s.Equal(uint8(0x01), deviceInfo.RevisionID())
	// class "0x120000" splits into class 0x12 (processing accelerator) and
	// subclass 0x00.
	s.Equal(uint8(0x12), deviceInfo.ClassID())
	s.Equal(uint8(0x00), deviceInfo.SubClassID())
}

func (s *PcieReadTestSuite) TestReadPcieInfoReadsLinkInfo() {
	target := s.buildMock(&rngd.PCIeMock{
		LinkWidth:              8,
		LinkSpeed:              16.0,
		MaxLinkWidth:           16,
		MaxLinkSpeed:           32.0,
		Upstream:               pcieTestBridge,
		UpstreamSecondaryBus:   0x27,
		UpstreamSubordinateBus: 0x2f,
	})

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)

	linkInfo := info.LinkInfo()
	s.Equal(uint32(8), linkInfo.LinkWidthStatus())
	s.InDelta(16.0, linkInfo.LinkSpeedStatus(), 0.001)
	s.Equal(uint32(16), linkInfo.MaxLinkWidthCapability())
	s.InDelta(32.0, linkInfo.MaxLinkSpeedCapability(), 0.001)
	// The reported generation comes from the max capability, not the speed
	// the link actually trained to -- a downtrained Gen4 link still reports
	// Gen5 here.
	s.Equal(uint8(5), linkInfo.PCIeGenStatus())
}

func (s *PcieReadTestSuite) TestReadPcieInfoResolvesRootComplexFromDevicePath() {
	target := s.buildMock(s.switchedTopology())

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)

	rootComplex := info.RootComplexInfo()
	s.Equal(uint16(0x0000), rootComplex.Domain())
	s.Equal(uint8(0x00), rootComplex.Bus())
}

func (s *PcieReadTestSuite) TestReadPcieInfoReportsUpstreamSwitch() {
	target := s.buildMock(s.switchedTopology())

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)

	switchInfo := info.SwitchInfo()
	s.Require().NotNil(switchInfo)
	s.Equal(pcieTestBridge, switchInfo.String())
}

// TestReadPcieInfoReportsNoSwitchForPassThroughBridge pins the rule that
// separates a switch from a plain bridge: a bridge whose subordinate bus
// equals its secondary spans nothing beyond itself, so it isn't reported.
func (s *PcieReadTestSuite) TestReadPcieInfoReportsNoSwitchForPassThroughBridge() {
	target := s.buildMock(&rngd.PCIeMock{
		Upstream:               pcieTestBridge,
		UpstreamSecondaryBus:   0x27,
		UpstreamSubordinateBus: 0x27,
	})

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)
	s.Nil(info.SwitchInfo())
}

// TestReadPcieInfoDefaultsSriovToZeroWhenAbsent covers the documented case
// of SR-IOV being off, where neither attribute exists at all.
func (s *PcieReadTestSuite) TestReadPcieInfoDefaultsSriovToZeroWhenAbsent() {
	target := s.buildMock(s.switchedTopology())

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)

	sriovInfo := info.SriovInfo()
	s.Zero(sriovInfo.SriovTotalVFs())
	s.Zero(sriovInfo.SriovEnabledVFs())
}

func (s *PcieReadTestSuite) TestReadPcieInfoReadsSriovWhenPresent() {
	total, enabled := uint32(16), uint32(4)
	topology := s.switchedTopology()
	topology.SriovTotalVFs = &total
	topology.SriovNumVFs = &enabled
	target := s.buildMock(topology)

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)

	sriovInfo := info.SriovInfo()
	s.Equal(uint32(16), sriovInfo.SriovTotalVFs())
	s.Equal(uint32(4), sriovInfo.SriovEnabledVFs())
}

// TestReadPcieInfoFailsWithoutRootComplex covers a device entry that isn't a
// symlink into the /sys/devices tree: there's no host bridge component in
// the resolved path to derive a root complex from.
func (s *PcieReadTestSuite) TestReadPcieInfoFailsWithoutRootComplex() {
	target := s.buildMock(&rngd.PCIeMock{})

	_, err := pcie.ReadPcieInfo(target)
	s.Error(err)
}

func (s *PcieReadTestSuite) TestReadPcieInfoFailsOnUnknownDevice() {
	s.buildMock(s.switchedTopology())

	absent, err := bdf.ParseBDF("0000:99:00.0")
	s.Require().NoError(err)

	_, err = pcie.ReadPcieInfo(absent)
	s.Error(err)
}

// TestReadPcieInfoSkipsAncestorsWithoutBusNumbers covers the walk's
// tolerance for path ancestors that aren't PCI bridges: those expose no
// secondary_bus_number, and a missing attribute there means "not a bridge,
// keep walking" rather than a failure of the whole read.
func (s *PcieReadTestSuite) TestReadPcieInfoSkipsAncestorsWithoutBusNumbers() {
	target := s.buildMock(s.switchedTopology())
	s.Require().NoError(os.Remove(filepath.Join(sysroot.PCIDevicesRoot(), pcieTestBridge, "secondary_bus_number")))

	info, err := pcie.ReadPcieInfo(target)
	s.Require().NoError(err)
	s.Nil(info.SwitchInfo())
}

func (s *PcieReadTestSuite) TestReadPcieInfoFailsOnMissingLinkAttribute() {
	target := s.buildMock(s.switchedTopology())
	s.Require().NoError(os.Remove(filepath.Join(sysroot.PCIDevicesRoot(), pcieTestBDF, "current_link_speed")))

	_, err := pcie.ReadPcieInfo(target)
	s.Error(err)
}

func TestPcieReadSuite(t *testing.T) {
	suite.Run(t, new(PcieReadTestSuite))
}

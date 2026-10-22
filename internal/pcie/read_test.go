package pcie

// furiosa-smi/src/pcie.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
)

type PcieTestSuite struct {
	suite.Suite
}

func (*PcieTestSuite) SetupSuite() {
}

func (*PcieTestSuite) TearDownSuite() {
}

func (*PcieTestSuite) SetupTest() {
}

func (*PcieTestSuite) TearDownTest() {
}

func (s *PcieTestSuite) TestParseCurrentLinkWidthFirstToken() {
	got, err := parseFirstTokenUint32("16\n", "missing", "invalid")
	s.Require().NoError(err)
	s.Equal(uint32(16), got)
}

func (s *PcieTestSuite) TestParseCurrentLinkWidthMalformed() {
	_, err := parseFirstTokenUint32("GT/s\n", "missing", "invalid")
	s.Error(err)

	_, err = parseFirstTokenUint32("\n", "missing", "invalid")
	s.Error(err)
}

func (s *PcieTestSuite) TestParseCurrentLinkSpeedFirstToken() {
	got, err := parseFirstTokenFloat64("32.0 GT/s PCIe\n", "missing", "invalid")
	s.Require().NoError(err)
	s.Equal(32.0, got)
}

func (s *PcieTestSuite) TestParseCurrentLinkSpeedMalformed() {
	_, err := parseFirstTokenFloat64("bogus GT/s\n", "missing", "invalid")
	s.Error(err)

	_, err = parseFirstTokenFloat64("\n", "missing", "invalid")
	s.Error(err)
}

func (s *PcieTestSuite) TestParseClassIDExtractsClassAndSubclass() {
	classID, subClassID, err := parseClassID("0x120000")
	s.Require().NoError(err)
	s.Equal(uint8(0x12), classID)
	s.Equal(uint8(0x00), subClassID)

	classID, subClassID, err = parseClassID("0x060400")
	s.Require().NoError(err)
	s.Equal(uint8(0x06), classID)
	s.Equal(uint8(0x04), subClassID)
}

func (s *PcieTestSuite) TestParseHexIdentifiers() {
	got, err := parseHexU16("0x1ed2", "couldn't parse subsystem vendor ID to u16")
	s.Require().NoError(err)
	s.Equal(uint16(0x1ed2), got)

	gotU8, err := parseHexU8("0x01", "couldn't parse revision ID to u8")
	s.Require().NoError(err)
	s.Equal(uint8(0x01), gotU8)
}

func (s *PcieTestSuite) TestParseSriovContent() {
	total, err := parseUint32("0\n", "invalid total")
	s.Require().NoError(err)
	s.Equal(uint32(0), total)

	enabled, err := parseUint32("0\n", "invalid enabled")
	s.Require().NoError(err)
	s.Equal(uint32(0), enabled)

	_, err = parseUint32("bogus\n", "invalid total")
	s.Error(err)

	_, err = parseUint32("bogus\n", "invalid enabled")
	s.Error(err)
}

func (s *PcieTestSuite) TestParsePCIeGenStatusValidSpeeds() {
	got, err := parsePCIeGenStatus(2.5)
	s.Require().NoError(err)
	s.Equal(uint8(1), got)

	got, err = parsePCIeGenStatus(5.0)
	s.Require().NoError(err)
	s.Equal(uint8(2), got)

	got, err = parsePCIeGenStatus(8.0)
	s.Require().NoError(err)
	s.Equal(uint8(3), got)

	got, err = parsePCIeGenStatus(16.0)
	s.Require().NoError(err)
	s.Equal(uint8(4), got)

	got, err = parsePCIeGenStatus(32.0)
	s.Require().NoError(err)
	s.Equal(uint8(5), got)
}

func (s *PcieTestSuite) TestParsePCIeGenStatusInvalidSpeed() {
	_, err := parsePCIeGenStatus(2.7)
	s.Error(err)
}

func (s *PcieTestSuite) TestParseHostBridgeComponentExtractsDomainAndBus() {
	got := parseHostBridgeComponent("pci0000:16")
	s.Require().NotNil(got)
	s.Equal(uint16(0x0000), got.domain)
	s.Equal(uint8(0x16), got.bus)
}

func (s *PcieTestSuite) TestParseHostBridgeComponentRejectsInvalidInput() {
	s.Nil(parseHostBridgeComponent("0000:16:01.0"))
	s.Nil(parseHostBridgeComponent("pci"))
	s.Nil(parseHostBridgeComponent("pcigarbage"))
}

func (s *PcieTestSuite) TestPcieInfoAccessors() {
	switchBDF, err := bdf.ParseBDF("0000:15:00.0")
	s.Require().NoError(err)

	info := NewPcieInfo(
		NewPcieDeviceInfo(0x0001, 0x1ed2, 0x0001, 0x01, 0x12, 0x00),
		NewPcieLinkInfo(5, 16, 32.0, 16, 32.0),
		NewSriovInfo(0, 0),
		NewPcieRootComplexInfo(0x0000, 0x16),
		&switchBDF,
	)

	s.Equal(uint16(0x0001), info.DeviceInfo().DeviceID())
	s.Equal(uint8(5), info.LinkInfo().PCIeGenStatus())
	s.Equal(uint32(0), info.SriovInfo().SriovTotalVFs())
	s.Equal(uint16(0x0000), info.RootComplexInfo().Domain())
	s.Equal(uint8(0x15), info.SwitchInfo().Bus())
}

func TestPcieSuite(t *testing.T) {
	suite.Run(t, new(PcieTestSuite))
}

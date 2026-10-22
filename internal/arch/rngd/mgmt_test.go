package rngd

// furiosa-smi/src/sysfs/rngd/mgmt.rs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
)

type RNGDMgmtTestSuite struct {
	suite.Suite
}

func (*RNGDMgmtTestSuite) SetupSuite() {
}

func (*RNGDMgmtTestSuite) TearDownSuite() {
}

func (*RNGDMgmtTestSuite) SetupTest() {
}

func (*RNGDMgmtTestSuite) TearDownTest() {
}

func (s *RNGDMgmtTestSuite) TestParsePEOccupancyMixed() {
	result, err := parsePEOccupancy("0\n0\n1\n1\n0\n0\n0\n0")
	s.Require().NoError(err)
	s.Len(result, 8)
	s.Equal(device.CoreStatusAvailable, result[0])
	s.Equal(device.CoreStatusOccupied, result[2])
	s.Equal(device.CoreStatusAvailable, result[7])
}

func (s *RNGDMgmtTestSuite) TestParsePEOccupancyAllOccupied() {
	result, err := parsePEOccupancy("1\n1\n1\n1")
	s.Require().NoError(err)
	s.Len(result, 4)
	for _, status := range result {
		s.Equal(device.CoreStatusOccupied, status)
	}
}

func (s *RNGDMgmtTestSuite) TestParsePEOccupancyTrailingNewline() {
	result, err := parsePEOccupancy("0\n1\n")
	s.Require().NoError(err)
	s.Len(result, 2)
	s.Contains(result, uint32(0))
	s.Contains(result, uint32(1))
}

func (s *RNGDMgmtTestSuite) TestParsePEOccupancyInvalidValue() {
	_, err := parsePEOccupancy("0\n2\n1")
	s.Error(err)
}

func (s *RNGDMgmtTestSuite) TestParsePEOccupancyRejectsBlankLine() {
	// Blank lines in the middle would silently shift core indices, so we
	// reject them explicitly instead of skipping.
	_, err := parsePEOccupancy("0\n\n1")
	s.Error(err)
}

func (s *RNGDMgmtTestSuite) TestParsePEOccupancySingleCore() {
	result, err := parsePEOccupancy("1")
	s.Require().NoError(err)
	s.Len(result, 1)
	s.Equal(device.CoreStatusOccupied, result[0])
}

func (s *RNGDMgmtTestSuite) TestParseCoreFrequencyBasic() {
	result, err := parseCoreFrequency("1500000000\n1600000000\n1700000000\n1800000000")
	s.Require().NoError(err)
	s.Len(result, 4)
	s.Equal(uint32(1500), result[0])
	s.Equal(uint32(1800), result[3])
}

func (s *RNGDMgmtTestSuite) TestParseCoreFrequencyTrailingNewline() {
	result, err := parseCoreFrequency("1500000000\n1600000000\n")
	s.Require().NoError(err)
	s.Len(result, 2)
	s.Equal(uint32(1600), result[1])
}

func (s *RNGDMgmtTestSuite) TestParseCoreFrequencyInvalidValue() {
	_, err := parseCoreFrequency("1500000000\nbogus\n1700000000")
	s.Error(err)
}

func (s *RNGDMgmtTestSuite) TestParseCoreFrequencyUsesUint32Range() {
	got, err := parseFrequencyLine(0, "4294967295")
	s.Require().NoError(err)
	s.Equal(uint32(4294), got)

	_, err = parseFrequencyLine(0, "4294967296")
	s.Error(err)
}

func (s *RNGDMgmtTestSuite) TestParseCoreFrequencyRejectsBlankLine() {
	// Same rationale as parsePEOccupancy: blank lines would silently shift
	// core indices.
	_, err := parseCoreFrequency("1500000000\n\n1700000000")
	s.Error(err)
}

func (s *RNGDMgmtTestSuite) TestParseGovernorProfilePerformance() {
	got, err := parseGovernorProfile("performance")
	s.Require().NoError(err)
	s.Equal(device.GovernorProfilePerformance, got)
}

func (s *RNGDMgmtTestSuite) TestParseGovernorProfilePowerSave() {
	got, err := parseGovernorProfile("powersave")
	s.Require().NoError(err)
	s.Equal(device.GovernorProfilePowerSave, got)
}

func (s *RNGDMgmtTestSuite) TestParseGovernorProfileUsesFirstLineOnly() {
	// Driver may emit multiple lines; only the first is inspected (matches
	// production).
	got, err := parseGovernorProfile("performance\npowersave")
	s.Require().NoError(err)
	s.Equal(device.GovernorProfilePerformance, got)
}

func (s *RNGDMgmtTestSuite) TestParseGovernorProfileInvalid() {
	_, err := parseGovernorProfile("bogus")
	s.Error(err)
}

func (s *RNGDMgmtTestSuite) TestWriteGovernorProfileRejectsInvalidValue() {
	mgmtPath := s.T().TempDir()
	attrPath := filepath.Join(mgmtPath, "npu_governor")
	s.Require().NoError(os.WriteFile(attrPath, []byte("performance\n"), 0o644))

	err := WriteGovernorProfile(mgmtPath, device.GovernorProfile(99))
	s.Require().Error(err)

	got, readErr := os.ReadFile(attrPath)
	s.Require().NoError(readErr)
	s.Equal("performance\n", string(got))
}

func (s *RNGDMgmtTestSuite) TestParseLivenessAlive() {
	s.True(parseLiveness("0"))
	s.True(parseLiveness("0\n"))
}

func (s *RNGDMgmtTestSuite) TestParseLivenessOthersDead() {
	s.False(parseLiveness("1"))
	s.False(parseLiveness("2"))
	s.False(parseLiveness("garbage"))
	s.False(parseLiveness(""))
}

func (s *RNGDMgmtTestSuite) TestParseThrottleReasonNone() {
	got, err := parseThrottleReasonBits("Throttle reason = 0x00000000")
	s.Require().NoError(err)
	s.Equal(uint32(0), got)
}

func (s *RNGDMgmtTestSuite) TestParseThrottleReasonBits() {
	got, err := parseThrottleReasonBits("Throttle reason = 0x00000003")
	s.Require().NoError(err)
	s.Equal(uint32(0x3), got)

	got, err = parseThrottleReasonBits("Throttle reason = 0x00000180")
	s.Require().NoError(err)
	s.Equal(uint32(0x180), got)
}

func (s *RNGDMgmtTestSuite) TestParseThrottleReasonEmpty() {
	_, err := parseThrottleReasonBits("")
	s.Error(err)
}

func (s *RNGDMgmtTestSuite) TestParseThrottleReasonInvalidHex() {
	_, err := parseThrottleReasonBits("Throttle reason = 0xzz")
	s.Error(err)
}

func TestRNGDMgmtSuite(t *testing.T) {
	suite.Run(t, new(RNGDMgmtTestSuite))
}

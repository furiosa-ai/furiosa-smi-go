package version

// furiosa-smi/src/version.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type VersionInfoTestSuite struct {
	suite.Suite
}

func (*VersionInfoTestSuite) SetupSuite() {
}

func (*VersionInfoTestSuite) TearDownSuite() {
}

func (*VersionInfoTestSuite) SetupTest() {
}

func (*VersionInfoTestSuite) TearDownTest() {
}

func (s *VersionInfoTestSuite) TestParseVersionInfoFull() {
	info, err := ParseVersionInfo("2025.3.1~rc0, abcdef0")
	s.Require().NoError(err)
	s.Equal(uint32(2025), info.Major())
	s.Equal(uint32(3), info.Minor())
	s.Equal(uint32(1), info.Patch())
	s.Require().NotNil(info.Prerelease())
	s.Equal("rc0", *info.Prerelease())
	s.Require().NotNil(info.Metadata())
	s.Equal("abcdef0", *info.Metadata())
}

func (s *VersionInfoTestSuite) TestParseVersionInfoMetadataOnly() {
	info, err := ParseVersionInfo("1.2.3, deadbee")
	s.Require().NoError(err)
	s.Equal(uint32(1), info.Major())
	s.Equal(uint32(2), info.Minor())
	s.Equal(uint32(3), info.Patch())
	s.Nil(info.Prerelease())
	s.Require().NotNil(info.Metadata())
	s.Equal("deadbee", *info.Metadata())
}

func (s *VersionInfoTestSuite) TestParseVersionInfoPrereleaseOnly() {
	info, err := ParseVersionInfo("1.2.3~dev")
	s.Require().NoError(err)
	s.Require().NotNil(info.Prerelease())
	s.Equal("dev", *info.Prerelease())
	s.Nil(info.Metadata())
}

func (s *VersionInfoTestSuite) TestParseVersionInfoPlain() {
	info, err := ParseVersionInfo("1.2.3")
	s.Require().NoError(err)
	s.Equal(uint32(1), info.Major())
	s.Equal(uint32(2), info.Minor())
	s.Equal(uint32(3), info.Patch())
	s.Nil(info.Prerelease())
	s.Nil(info.Metadata())
}

func (s *VersionInfoTestSuite) TestParseVersionInfoInvalid() {
	for _, input := range []string{"not-a-version", "1.2"} {
		_, err := ParseVersionInfo(input)
		s.Errorf(err, "ParseVersionInfo(%q)", input)
	}
}

func (s *VersionInfoTestSuite) TestVersionInfoDisplayWithPrereleaseAndMetadata() {
	info, err := ParseVersionInfo("2025.3.1~rc0, abcdef0")
	s.Require().NoError(err)
	s.Equal("2025.3.1(rc0), abcdef0", info.String())
}

func (s *VersionInfoTestSuite) TestVersionInfoDisplayWithoutPrerelease() {
	info, err := ParseVersionInfo("1.2.3, deadbee")
	s.Require().NoError(err)
	s.Equal("1.2.3, deadbee", info.String())
}

func (s *VersionInfoTestSuite) TestVersionInfoDisplayWithoutMetadata() {
	info, err := ParseVersionInfo("1.2.3~dev")
	s.Require().NoError(err)
	s.Equal("1.2.3(dev)", info.String())
}

func (s *VersionInfoTestSuite) TestVersionInfoDisplayPlain() {
	info, err := ParseVersionInfo("1.2.3")
	s.Require().NoError(err)
	s.Equal("1.2.3", info.String())
}

func TestVersionInfoSuite(t *testing.T) {
	suite.Run(t, new(VersionInfoTestSuite))
}

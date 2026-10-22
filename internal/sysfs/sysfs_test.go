package sysfs

// furiosa-smi/src/sysfs.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SysfsTestSuite struct {
	suite.Suite
}

func (*SysfsTestSuite) SetupSuite() {
}

func (*SysfsTestSuite) TearDownSuite() {
}

func (*SysfsTestSuite) SetupTest() {
}

func (*SysfsTestSuite) TearDownTest() {
}

func (s *SysfsTestSuite) TestParseHexU16WithPrefix() {
	got, err := parseHexU16("0x1ed2")
	s.Require().NoError(err)
	s.Equal(uint16(0x1ed2), got)
}

func (s *SysfsTestSuite) TestParseHexU16UppercasePrefix() {
	got, err := parseHexU16("0X1ED2")
	s.Require().NoError(err)
	s.Equal(uint16(0x1ed2), got)
}

func (s *SysfsTestSuite) TestParseHexU16NoPrefix() {
	got, err := parseHexU16("1ed2")
	s.Require().NoError(err)
	s.Equal(uint16(0x1ed2), got)
}

func (s *SysfsTestSuite) TestParseHexU16WithWhitespace() {
	got, err := parseHexU16("  0x1ed2  ")
	s.Require().NoError(err)
	s.Equal(uint16(0x1ed2), got)
}

func (s *SysfsTestSuite) TestParseHexU16Invalid() {
	_, err := parseHexU16("zzzz")
	s.Error(err)
	_, err = parseHexU16("")
	s.Error(err)
}

func TestSysfsSuite(t *testing.T) {
	suite.Run(t, new(SysfsTestSuite))
}

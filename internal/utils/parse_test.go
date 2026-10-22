package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParseTestSuite struct {
	suite.Suite
}

func (*ParseTestSuite) SetupSuite() {
}

func (*ParseTestSuite) TearDownSuite() {
}

func (*ParseTestSuite) SetupTest() {
}

func (*ParseTestSuite) TearDownTest() {
}

func (s *ParseTestSuite) TestParseHexU64AcceptsPrefixed() {
	got, err := ParseHexU64("0xff")
	s.Require().NoError(err)
	s.Equal(uint64(0xff), got)

	got, err = ParseHexU64("0XAB")
	s.Require().NoError(err)
	s.Equal(uint64(0xab), got)

	got, err = ParseHexU64("ff")
	s.Require().NoError(err)
	s.Equal(uint64(0xff), got)

	got, err = ParseHexU64(" 0x10 ")
	s.Require().NoError(err)
	s.Equal(uint64(0x10), got)
}

func (s *ParseTestSuite) TestParseHexU64RejectsGarbage() {
	_, err := ParseHexU64("zz")
	s.Error(err)
	_, err = ParseHexU64("")
	s.Error(err)
}

func TestParseSuite(t *testing.T) {
	suite.Run(t, new(ParseTestSuite))
}

package bdf

// furiosa-smi/src/device/bdf.rs

import (
	"fmt"
	"sort"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/stretchr/testify/suite"
)

type BDFTestSuite struct {
	suite.Suite
}

func (*BDFTestSuite) SetupSuite() {
}

func (*BDFTestSuite) TearDownSuite() {
}

func (*BDFTestSuite) SetupTest() {
}

func (*BDFTestSuite) TearDownTest() {
}

func (s *BDFTestSuite) TestParseBDFFull() {
	got, err := ParseBDF("0000:17:00.0")
	s.Require().NoError(err)
	s.Equal(uint16(0x0000), got.Domain())
	s.Equal(uint8(0x17), got.Bus())
	s.Equal(uint8(0x00), got.Device())
	s.Equal(uint8(0), got.Function())
}

func (s *BDFTestSuite) TestParseBDFWithoutDomain() {
	got, err := ParseBDF("17:00.0")
	s.Require().NoError(err)
	s.Equal(uint16(0), got.Domain())
	s.Equal(uint8(0x17), got.Bus())
}

func (s *BDFTestSuite) TestParseBDFMaxValues() {
	got, err := ParseBDF("ffff:ff:1f.7")
	s.Require().NoError(err)
	s.Equal(uint16(0xffff), got.Domain())
	s.Equal(uint8(0xff), got.Bus())
	s.Equal(uint8(0x1f), got.Device())
	s.Equal(uint8(7), got.Function())
}

func (s *BDFTestSuite) TestParseBDFInvalid() {
	for _, input := range []string{"invalid", "", "00:00", "0000:00:00:0"} {
		s.requireParseError(input)
	}
}

func (s *BDFTestSuite) TestParseBDFInvalidFunctionRange() {
	s.requireParseError("0000:00:00.8")
}

func (s *BDFTestSuite) requireParseError(input string) {
	s.T().Helper()
	_, err := ParseBDF(input)
	s.Require().Error(err, "ParseBDF(%q)", input)
	var smiErr smierrors.SMIError
	s.Require().ErrorAs(err, &smiErr)
	s.Equal(smierrors.ErrorKindParse, smiErr.Kind())
	s.Equal(fmt.Sprintf("parse error: invalid BDF: '%s'", input), err.Error())
}

func (s *BDFTestSuite) TestBDFDisplayRoundtrip() {
	const input = "0000:17:00.0"
	got, err := ParseBDF(input)
	s.Require().NoError(err)
	s.Equal(input, got.String())
}

func (s *BDFTestSuite) TestBDFDisplayFillsDomain() {
	got, err := ParseBDF("17:00.0")
	s.Require().NoError(err)
	s.Equal("0000:17:00.0", got.String())
}

func (s *BDFTestSuite) TestBDFOrdering() {
	a, err := ParseBDF("0000:01:00.0")
	s.Require().NoError(err)
	b, err := ParseBDF("0000:02:00.0")
	s.Require().NoError(err)
	c, err := ParseBDF("0001:01:00.0")
	s.Require().NoError(err)
	values := []BDF{c, b, a}
	sort.Slice(values, func(i, j int) bool {
		left, right := values[i], values[j]
		if left.Domain() != right.Domain() {
			return left.Domain() < right.Domain()
		}

		if left.Bus() != right.Bus() {
			return left.Bus() < right.Bus()
		}

		if left.Device() != right.Device() {
			return left.Device() < right.Device()
		}

		return left.Function() < right.Function()
	})
	s.Equal([]string{"0000:01:00.0", "0000:02:00.0", "0001:01:00.0"}, []string{
		values[0].String(), values[1].String(), values[2].String(),
	})
}

func (s *BDFTestSuite) TestBDFEquality() {
	a, err := ParseBDF("0000:17:00.0")
	s.Require().NoError(err)
	b, err := ParseBDF("0000:17:00.0")
	s.Require().NoError(err)
	s.Equal(a, b)
}

func TestBDFSuite(t *testing.T) {
	suite.Run(t, new(BDFTestSuite))
}

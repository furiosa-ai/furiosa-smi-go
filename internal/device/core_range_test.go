package device

// furiosa-smi/src/device/core_range.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DeviceCoreRangeTestSuite struct {
	suite.Suite
}

func (*DeviceCoreRangeTestSuite) SetupSuite() {
}

func (*DeviceCoreRangeTestSuite) TearDownSuite() {
}

func (*DeviceCoreRangeTestSuite) SetupTest() {
}

func (*DeviceCoreRangeTestSuite) TearDownTest() {
}

func (s *DeviceCoreRangeTestSuite) TestParseDeviceCoreRangeRange() {
	got, err := ParseDeviceCoreRange("npu0pe0-3")
	s.Require().NoError(err)
	s.Equal(uint8(0), got.DevNodeIndex())
	s.Equal(uint32(0), got.StartCore())
	s.Equal(uint32(3), got.EndCore())
}

func (s *DeviceCoreRangeTestSuite) TestParseDeviceCoreRangeSingleCore() {
	got, err := ParseDeviceCoreRange("npu1pe2")
	s.Require().NoError(err)
	s.Equal(uint8(1), got.DevNodeIndex())
	s.Equal(uint32(2), got.StartCore())
	s.Equal(uint32(2), got.EndCore())
}

func (s *DeviceCoreRangeTestSuite) TestParseDeviceCoreRangeBareDevice() {
	for _, test := range []struct {
		name         string
		input        string
		devNodeIndex uint8
	}{
		{name: "Rust vector", input: "npu0", devNodeIndex: 0},
		{name: "nonzero node", input: "npu3", devNodeIndex: 3},
	} {
		s.Run(test.name, func() {
			got, err := ParseDeviceCoreRange(test.input)
			s.Require().NoError(err)
			s.Equal(test.devNodeIndex, got.DevNodeIndex())
			s.Equal(uint32(0), got.StartCore())
			s.Equal(uint32(0), got.EndCore())
		})
	}
}

func (s *DeviceCoreRangeTestSuite) TestParseDeviceCoreRangeInvalid() {
	for _, input := range []string{"other", "", "npu0peabc", "npu0pe1-abc"} {
		_, err := ParseDeviceCoreRange(input)
		s.Error(err, "ParseDeviceCoreRange(%q)", input)
	}
}

func (s *DeviceCoreRangeTestSuite) TestNewDeviceCoreRangeAccessors() {
	got := NewDeviceCoreRange(3, 4, 7)
	s.Equal(uint8(3), got.DevNodeIndex())
	s.Equal(uint32(4), got.StartCore())
	s.Equal(uint32(7), got.EndCore())
}

func (s *DeviceCoreRangeTestSuite) TestDeviceCoreRangeCores() {
	s.Equal([]uint32{4, 5, 6, 7}, NewDeviceCoreRange(0, 4, 7).Cores())
	s.Equal([]uint32{2}, NewDeviceCoreRange(1, 2, 2).Cores())
	s.Empty(NewDeviceCoreRange(0, 5, 2).Cores())

	maxCore := ^uint32(0)
	s.Equal([]uint32{maxCore}, NewDeviceCoreRange(0, maxCore, maxCore).Cores())
}

func (s *DeviceCoreRangeTestSuite) TestDeviceCoreRangeString() {
	s.Equal("npu0pe0", NewDeviceCoreRange(0, 0, 0).String())
	s.Equal("npu1pe2", NewDeviceCoreRange(1, 2, 2).String())
	s.Equal("npu0pe0-3", NewDeviceCoreRange(0, 0, 3).String())
	s.Equal("npu3pe2-5", NewDeviceCoreRange(3, 2, 5).String())
}

// TestDeviceCoreRangeStringRoundTrips confirms ParseDeviceCoreRange(x.String())
// reconstructs x, mirroring Rust's Display/FromStr pairing. This isn't
// byte-for-byte round-tripping of the original input: a bare "npuN" parses
// to the same value as "npuNpe0", and String() always renders the latter.
func (s *DeviceCoreRangeTestSuite) TestDeviceCoreRangeStringRoundTrips() {
	for _, input := range []string{"npu0pe0", "npu1pe2", "npu0pe0-3", "npu3pe2-5", "npu0"} {
		want, err := ParseDeviceCoreRange(input)
		s.Require().NoError(err)

		got, err := ParseDeviceCoreRange(want.String())
		s.Require().NoError(err)
		s.Equal(want, got, "input %q", input)
	}
}

func TestDeviceCoreRangeSuite(t *testing.T) {
	suite.Run(t, new(DeviceCoreRangeTestSuite))
}

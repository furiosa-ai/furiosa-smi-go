package rngd

// furiosa-smi/src/arch/rngd.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
)

type RNGDTestSuite struct {
	suite.Suite
}

func (*RNGDTestSuite) SetupSuite() {
}

func (*RNGDTestSuite) TearDownSuite() {
}

func (*RNGDTestSuite) SetupTest() {
}

func (*RNGDTestSuite) TearDownTest() {
}

func (s *RNGDTestSuite) TestMgmtNodeIndexParsing() {
	cases := []struct {
		input string
		want  uint8
		ok    bool
	}{
		{"rngd!npu0mgmt", 0, true},
		{"rngd!npu7mgmt", 7, true},
		{"rngd!npu12mgmt", 12, true},
		{"rngd!npu0pe0", 0, false},
		{"rngd!npu0bar0", 0, false},
		{"rngd!npu0dmar", 0, false},
		{"other", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got := ParseMgmtNodeIndex(c.input)
		if c.ok {
			s.Require().NotNil(got, "input %q", c.input)
			s.Equal(c.want, *got, "input %q", c.input)
		} else {
			s.Nil(got, "input %q", c.input)
		}
	}
}

// TestDevNodeIndexParsing confirms parseDevNodeIndex accepts everything the
// public ParseDeviceCoreRange does -- including a bare "npuN" node with no
// "peS" suffix, which a real unpartitioned device could expose.
func (s *RNGDTestSuite) TestDevNodeIndexParsing() {
	cases := []struct {
		input string
		want  uint8
		ok    bool
	}{
		{"npu0pe0", 0, true},
		{"npu3pe2-5", 3, true},
		{"npu0", 0, true},
		{"npu7", 7, true},
		{"not_a_device", 0, false},
		{"rngd!npu0mgmt", 0, false},
	}
	for _, c := range cases {
		got := parseDevNodeIndex(c.input)
		if c.ok {
			s.Require().NotNil(got, "input %q", c.input)
			s.Equal(c.want, *got, "input %q", c.input)
		} else {
			s.Nil(got, "input %q", c.input)
		}
	}
}

// requireCoreRange asserts parseCoreRange matched name and returns the range.
func (s *RNGDTestSuite) requireCoreRange(devNodeIndex uint8, name string) device.DeviceCoreRange {
	s.T().Helper()
	got := parseCoreRange(devNodeIndex, name)
	s.Require().NotNil(got, "parseCoreRange(%d, %q)", devNodeIndex, name)

	return *got
}

func (s *RNGDTestSuite) TestCoreRangeSingle() {
	s.Equal(device.NewDeviceCoreRange(0, 0, 0), s.requireCoreRange(0, "npu0pe0"))
	s.Equal(device.NewDeviceCoreRange(0, 7, 7), s.requireCoreRange(0, "npu0pe7"))
	s.Equal(device.NewDeviceCoreRange(0, 0, 0), s.requireCoreRange(0, "npu0"))
	maxCore := ^uint32(0)
	s.Equal(device.NewDeviceCoreRange(0, maxCore, maxCore), s.requireCoreRange(0, "npu0pe4294967295"))
}

func (s *RNGDTestSuite) TestCoreRangeMulti() {
	s.Equal(device.NewDeviceCoreRange(0, 0, 3), s.requireCoreRange(0, "npu0pe0-3"))
	s.Equal(device.NewDeviceCoreRange(1, 2, 5), s.requireCoreRange(1, "npu1pe2-5"))
}

func (s *RNGDTestSuite) TestCoreRangeWrongNode() {
	s.Nil(parseCoreRange(0, "npu1pe0"))
	s.Nil(parseCoreRange(0, "other"))
}

func (s *RNGDTestSuite) TestCoreRangeInvalidName() {
	s.Nil(parseCoreRange(0, "npu0peabc"))
	s.Nil(parseCoreRange(0, "npu0pe1-abc"))
	// P2P DMA interface nodes share the "npuN" prefix but aren't PE ranges.
	s.Nil(parseCoreRange(0, "npu0bar0"))
	s.Nil(parseCoreRange(0, "npu0ch0r"))
	s.Nil(parseCoreRange(0, "npu0dmar"))
}

// TestCoreRangeInvertedStillParses matches Rust, where parse_core_range
// returns Some for "npu0pe5-2": the name is well-formed, the range just
// spans no cores.
func (s *RNGDTestSuite) TestCoreRangeInvertedStillParses() {
	got := s.requireCoreRange(0, "npu0pe5-2")
	s.Equal(uint32(5), got.StartCore())
	s.Equal(uint32(2), got.EndCore())
	s.Empty(got.Cores())
}

func TestRNGDSuite(t *testing.T) {
	suite.Run(t, new(RNGDTestSuite))
}

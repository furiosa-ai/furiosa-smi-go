package device

// furiosa-smi/src/device/throttle.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ThrottleReasonTestSuite struct {
	suite.Suite
}

func (*ThrottleReasonTestSuite) SetupSuite() {
}

func (*ThrottleReasonTestSuite) TearDownSuite() {
}

func (*ThrottleReasonTestSuite) SetupTest() {
}

func (*ThrottleReasonTestSuite) TearDownTest() {
}

func (s *ThrottleReasonTestSuite) TestFromBitsEmpty() {
	s.Empty(ThrottleReasonsFromBits(0))
}

func (s *ThrottleReasonTestSuite) TestFromBitsKnownBits() {
	s.Equal([]ThrottleReason{ThrottleReasonIdle, ThrottleReasonThermalSlowdown}, ThrottleReasonsFromBits(0x3))
	s.Equal([]ThrottleReason{ThrottleReasonHwPowerCap, ThrottleReasonGovClockCap}, ThrottleReasonsFromBits(0xC0))
}

func (s *ThrottleReasonTestSuite) TestFromBitsAllKnown() {
	s.Equal([]ThrottleReason{
		ThrottleReasonIdle,
		ThrottleReasonThermalSlowdown,
		ThrottleReasonAppPowerCap,
		ThrottleReasonAppClockCap,
		ThrottleReasonHwClockCap,
		ThrottleReasonHwBusLimit,
		ThrottleReasonHwPowerCap,
		ThrottleReasonGovClockCap,
		ThrottleReasonOtherReason,
	}, ThrottleReasonsFromBits(0x1ff))
}

func (s *ThrottleReasonTestSuite) TestFromBitsPreservesUnknownBits() {
	s.Equal([]ThrottleReason{ThrottleReason(0x200), ThrottleReason(0x400)}, ThrottleReasonsFromBits(0x600))
}

func (s *ThrottleReasonTestSuite) TestFromBitsMixedKnownAndUnknown() {
	s.Equal([]ThrottleReason{ThrottleReasonIdle, ThrottleReason(0x200)}, ThrottleReasonsFromBits(0x201))
}

func (s *ThrottleReasonTestSuite) TestFromBitsRoundTrips() {
	for _, bits := range []uint32{0x0, 0x3, 0x1ff, 0x201, 0x600, 0xffffffff} {
		var result uint32
		for _, reason := range ThrottleReasonsFromBits(bits) {
			result |= uint32(reason)
		}

		s.Equal(bits, result)
	}
}

func (s *ThrottleReasonTestSuite) TestDisplayNames() {
	s.Equal("Idle", ThrottleReasonIdle.String())
	s.Equal("Thermal Slowdown", ThrottleReasonThermalSlowdown.String())
	s.Equal("HW Bus Limit", ThrottleReasonHwBusLimit.String())
	s.Equal("Gov Clock Cap", ThrottleReasonGovClockCap.String())
	s.Equal("Unknown(0x00000200)", ThrottleReason(0x200).String())
}

func TestThrottleReasonSuite(t *testing.T) {
	suite.Run(t, new(ThrottleReasonTestSuite))
}

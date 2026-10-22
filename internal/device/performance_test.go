package device

// furiosa-smi/src/device/performance.rs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PePerformanceCounterTestSuite struct {
	suite.Suite
}

func (*PePerformanceCounterTestSuite) SetupSuite() {
}

func (*PePerformanceCounterTestSuite) TearDownSuite() {
}

func (*PePerformanceCounterTestSuite) SetupTest() {
}

func (*PePerformanceCounterTestSuite) TearDownTest() {
}

func (s *PePerformanceCounterTestSuite) TestPePerformanceCounterAccessors() {
	counter := NewPePerformanceCounter(time.Now(), 0x1000, 0x800)
	s.Equal(uint64(0x1000), counter.CycleCount())
	s.Equal(uint64(0x800), counter.TaskExecutionCycle())
}

func (s *PePerformanceCounterTestSuite) TestPePerformanceCounterDurationSince() {
	t0 := time.Now()
	prev := NewPePerformanceCounter(t0, 0, 0)
	curr := NewPePerformanceCounter(t0.Add(time.Second), 0, 0)
	s.Equal(time.Second, curr.DurationSince(prev))
	s.Zero(prev.DurationSince(curr))
}

func (s *PePerformanceCounterTestSuite) TestPePerformanceCounterUtilizationSince() {
	t0 := time.Now()
	prev := NewPePerformanceCounter(t0, 1000, 200)
	curr := NewPePerformanceCounter(t0, 2000, 700)
	s.Equal(0.5, curr.UtilizationSince(prev))
}

func (s *PePerformanceCounterTestSuite) TestPePerformanceCounterUtilizationSinceZeroCycleDelta() {
	t0 := time.Now()
	counter := NewPePerformanceCounter(t0, 1000, 200)
	s.Equal(0.0, counter.UtilizationSince(counter))
}

func TestPePerformanceCounterSuite(t *testing.T) {
	suite.Run(t, new(PePerformanceCounterTestSuite))
}

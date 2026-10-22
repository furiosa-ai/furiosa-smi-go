package device

// furiosa-smi/src/device/performance.rs

import "time"

// PePerformanceCounter is a performance counter snapshot.
type PePerformanceCounter struct {
	timestamp          time.Time
	cycleCount         uint64
	taskExecutionCycle uint64
}

func NewPePerformanceCounter(timestamp time.Time, cycleCount, taskExecutionCycle uint64) PePerformanceCounter {
	return PePerformanceCounter{timestamp: timestamp, cycleCount: cycleCount, taskExecutionCycle: taskExecutionCycle}
}

func (c PePerformanceCounter) CycleCount() uint64 {
	return c.cycleCount
}

func (c PePerformanceCounter) TaskExecutionCycle() uint64 {
	return c.taskExecutionCycle
}

// DurationSince returns 0, not a negative duration, if prev is actually the
// later snapshot.
func (c PePerformanceCounter) DurationSince(prev PePerformanceCounter) time.Duration {
	if c.timestamp.Before(prev.timestamp) {
		return 0
	}

	return c.timestamp.Sub(prev.timestamp)
}

// UtilizationSince is a ratio in [0.0, 1.0]; saturatingSubU64 guards
// against counter wrap-around or reversed snapshots.
func (c PePerformanceCounter) UtilizationSince(prev PePerformanceCounter) float64 {
	cycleDelta := saturatingSubU64(c.cycleCount, prev.cycleCount)
	if cycleDelta == 0 {
		return 0.0
	}

	taskDelta := saturatingSubU64(c.taskExecutionCycle, prev.taskExecutionCycle)

	return float64(taskDelta) / float64(cycleDelta)
}

func saturatingSubU64(a, b uint64) uint64 {
	if a < b {
		return 0
	}

	return a - b
}

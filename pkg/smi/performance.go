package smi

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// DeviceTemperature represents a temperature information of the device.
type DeviceTemperature interface {
	// SocPeak returns the highest temperature observed from SoC sensors.
	SocPeak() float64
	// Ambient returns the temperature observed from sensors attached to the board.
	Ambient() float64
}

var _ DeviceTemperature = new(FuriosaSmiDeviceTemperature)

type FuriosaSmiDeviceTemperature struct {
	socPeak float64
	ambient float64
}

func (d *FuriosaSmiDeviceTemperature) SocPeak() float64 {
	return d.socPeak
}

func (d *FuriosaSmiDeviceTemperature) Ambient() float64 {
	return d.ambient
}

// DevicePerformanceCounter represents a device performance counter.
type DevicePerformanceCounter interface {
	// PerformanceCounter returns a list of performance counters.
	PerformanceCounter() []PerformanceCounter
}
type FuriosaSmiDevicePerformanceCounter struct {
	PePerformanceCounters []FuriosaSmiPePerformanceCounter
}

var _ DevicePerformanceCounter = new(FuriosaSmiDevicePerformanceCounter)

func (d *FuriosaSmiDevicePerformanceCounter) PerformanceCounter() []PerformanceCounter {
	var ret []PerformanceCounter

	for i := 0; i < len(d.PePerformanceCounters); i++ {
		ret = append(ret, &d.PePerformanceCounters[i])
	}

	return ret
}

// PerformanceCounter represents a performance counter.
type PerformanceCounter interface {
	// Timestamp returns timestamp.
	Timestamp() time.Time
	// Core returns a core index.
	Core() uint32
	// CycleCount returns total cycle count in 64-bit unsigned int.
	CycleCount() uint64
	// TaskExecutionCycle returns cycle count used for task execution in 64-bit unsigned int.
	TaskExecutionCycle() uint64
}

var _ PerformanceCounter = new(FuriosaSmiPePerformanceCounter)

type FuriosaSmiPePerformanceCounter struct {
	timestamp          int64
	core               uint32
	cycleCount         uint64
	taskExecutionCycle uint64
}

func (p *FuriosaSmiPePerformanceCounter) Timestamp() time.Time {
	return time.Unix(p.timestamp, 0)
}

func (p *FuriosaSmiPePerformanceCounter) Core() uint32 {
	return p.core
}

func (p *FuriosaSmiPePerformanceCounter) CycleCount() uint64 {
	return p.cycleCount
}

func (p *FuriosaSmiPePerformanceCounter) TaskExecutionCycle() uint64 {
	return p.taskExecutionCycle
}

type performanceCounterMap struct {
	mu   sync.RWMutex
	data map[string]performanceCounterInfo
}

func (pcm *performanceCounterMap) get(dev Device) (performanceCounterInfo, bool) {
	pcm.mu.RLock()
	defer pcm.mu.RUnlock()

	deviceInfo, err := dev.DeviceInfo()
	if err != nil {
		return performanceCounterInfo{}, false
	}

	info, exists := pcm.data[deviceInfo.UUID()]
	return info, exists
}

func (pcm *performanceCounterMap) set(dev Device, info performanceCounterInfo) {
	pcm.mu.Lock()
	defer pcm.mu.Unlock()

	deviceInfo, err := dev.DeviceInfo()
	if err != nil {
		return
	}
	pcm.data[deviceInfo.UUID()] = info
}

func newPerformanceCounterMap() performanceCounterMap {
	return performanceCounterMap{
		mu:   sync.RWMutex{},
		data: make(map[string]performanceCounterInfo),
	}
}

type performanceCounterInfo struct {
	beforeCounter DevicePerformanceCounter
	afterCounter  DevicePerformanceCounter
}

type ObserverOpt struct {
	devices  []Device
	interval uint32
}

func NewOptForObserver() (ObserverOpt, error) {
	devices, err := ListDevices()

	if err != nil {
		return ObserverOpt{}, err
	}

	return ObserverOpt{
		devices:  devices,
		interval: 500,
	}, nil
}

func (o *ObserverOpt) SetDevices(devices []Device) {
	o.devices = devices
}

func (o *ObserverOpt) SetInterval(interval uint32) {
	o.interval = interval
}

type observer struct {
	once                  sync.Once
	performanceCounterMap performanceCounterMap
	stopCh                chan struct{}
}

var _ Observer = new(observer)

// Observer represents an observer instance to collect device information.
type Observer interface {
	// GetCoreUtilization returns the core utilization for the given device.
	GetCoreUtilization(device Device) ([]CoreUtilization, error)
	// Destroy stops the observer when it calls explicitly or observer is destroyed by GC.
	Destroy()
}

func newObserverWithOpt(opt ObserverOpt) (Observer, error) {
	devices := opt.devices
	interval := opt.interval

	o := &observer{
		performanceCounterMap: newPerformanceCounterMap(),
		stopCh:                make(chan struct{}),
	}

	o.start(devices, time.Duration(interval)*time.Millisecond)

	runtime.SetFinalizer(o, func(o Observer) {
		o.Destroy()
	})
	return o, nil
}

func (o *observer) isDestroyed() bool {
	select {
	case <-o.stopCh:
		return true
	default:
		return false
	}
}

func (o *observer) GetCoreUtilization(device Device) ([]CoreUtilization, error) {
	if o.isDestroyed() {
		return nil, fmt.Errorf("observer is already destroyed")
	}

	if device == nil {
		return nil, fmt.Errorf("device is nil")
	}

	utilization, err := o.CalculateUtilization(device)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate utilization: %w", err)
	}

	return utilization, nil
}

func (o *observer) start(devices []Device, interval time.Duration) {
	o.updateUtilization(devices)
	time.Sleep(100 * time.Millisecond)
	o.updateUtilization(devices)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				o.updateUtilization(devices)

			case <-o.stopCh:
				return
			}
		}
	}()
}

func (o *observer) Destroy() {
	o.once.Do(func() {
		close(o.stopCh)
	})
}

func (o *observer) updateUtilization(devices []Device) {
	for _, device := range devices {
		performanceCounter, err := device.DevicePerformanceCounter()

		pc, exists := o.performanceCounterMap.get(device)

		if exists {
			o.performanceCounterMap.set(device, performanceCounterInfo{
				beforeCounter: pc.afterCounter,
				afterCounter:  performanceCounter})
		} else {
			o.performanceCounterMap.set(device, performanceCounterInfo{
				beforeCounter: performanceCounter,
				afterCounter:  performanceCounter})
		}

		if err != nil {
			continue
		}
	}
}

// CoreUtilization represents a core utilization information.
type CoreUtilization interface {
	// Core returns a core index.
	Core() uint32
	// TimeWindowMil returns a time window in milliseconds.
	TimeWindowMil() uint32
	// PeUsagePercentage returns a percentage of PE usage.
	PeUsagePercentage() float64
}

var _ CoreUtilization = new(coreUtilization)

type coreUtilization struct {
	core              uint32
	timeWindowMil     uint32
	peUsagePercentage float64
}

func newCoreUtilization(core uint32, timeWindowMil uint32, peUsagePercentage float64) CoreUtilization {
	return &coreUtilization{
		core:              core,
		timeWindowMil:     timeWindowMil,
		peUsagePercentage: peUsagePercentage,
	}
}

func (c *coreUtilization) Core() uint32 {
	return c.core
}

func (c *coreUtilization) TimeWindowMil() uint32 {
	return c.timeWindowMil
}

func (c *coreUtilization) PeUsagePercentage() float64 {
	return c.peUsagePercentage
}

func (o *observer) CalculateUtilization(device Device) ([]CoreUtilization, error) {
	performanceCounterInfo, exists := o.performanceCounterMap.get(device)

	if !exists {
		return nil, fmt.Errorf("no performance counter info found for device %v", device)
	}

	beforeCounter := performanceCounterInfo.beforeCounter
	afterCounter := performanceCounterInfo.afterCounter

	utilizationResult := make([]CoreUtilization, 0)

	afterPerfCounter := afterCounter.PerformanceCounter()
	beforePerfCounter := beforeCounter.PerformanceCounter()

	for i, beforePeCounter := range beforePerfCounter {
		afterPeCounter := afterPerfCounter[i]

		if afterPeCounter.CycleCount() < beforePeCounter.CycleCount() {
			utilization := newCoreUtilization(beforePeCounter.Core(), 0, 0.0)

			utilizationResult = append(utilizationResult, utilization)
			continue
		}

		taskExecutionCycleDiff := afterPeCounter.TaskExecutionCycle() - beforePeCounter.TaskExecutionCycle()
		cycleCountDiff := afterPeCounter.CycleCount() - beforePeCounter.CycleCount()

		peUsagePercentage := safeUsizeDivide(taskExecutionCycleDiff, cycleCountDiff) * 100.0

		utilization := newCoreUtilization(beforePeCounter.Core(), uint32(afterPeCounter.Timestamp().Sub(beforePeCounter.Timestamp()).Milliseconds()), peUsagePercentage)

		utilizationResult = append(utilizationResult, utilization)
	}
	return utilizationResult, nil
}

// MemoryUtilization represents a memory utilization information.
type MemoryUtilization interface {
	Dram() Memory
	DramShared() Memory
	Sram() Memory
	Instruction() Memory
}
type FuriosaSmiMemoryUtilization struct {
	dram        FuriosaSmiMemory
	ramShared   FuriosaSmiMemory
	sram        FuriosaSmiMemory
	instruction FuriosaSmiMemory
}

var _ MemoryUtilization = new(FuriosaSmiMemoryUtilization)

func (m *FuriosaSmiMemoryUtilization) Dram() Memory {
	return &m.dram
}

func (m *FuriosaSmiMemoryUtilization) DramShared() Memory {
	return &m.ramShared
}

func (m *FuriosaSmiMemoryUtilization) Sram() Memory {
	return &m.sram
}

func (m *FuriosaSmiMemoryUtilization) Instruction() Memory {
	return &m.instruction
}

// Memory represent a total memory information.
type Memory interface {
	Memory() []MemoryBlock
}
type FuriosaSmiMemory struct {
	memory []FuriosaSmiMemoryBlock
}

var _ Memory = new(FuriosaSmiMemory)

func (m *FuriosaSmiMemory) Memory() []MemoryBlock {
	var ret []MemoryBlock

	for i := 0; i < len(m.memory); i++ {
		ret = append(ret, &m.memory[i])
	}

	return ret
}

// MemoryBlock represent a (fusioned) memory information.
type MemoryBlock interface {
	Core() []uint32
	TotalBytes() uint64
	InUseBytes() uint64
}

type FuriosaSmiMemoryBlock struct {
	core       []uint32
	totalBytes uint64
	inUseBytes uint64
}

var _ MemoryBlock = new(FuriosaSmiMemoryBlock)

func (m *FuriosaSmiMemoryBlock) Core() []uint32 {
	var ret []uint32
	for i := 0; i < len(m.core); i++ {
		ret = append(ret, m.core[i])
	}
	return ret
}

func (m *FuriosaSmiMemoryBlock) TotalBytes() uint64 {
	return m.totalBytes
}

func (m *FuriosaSmiMemoryBlock) InUseBytes() uint64 {
	return m.inUseBytes
}

func safeUsizeDivide(fst, snd uint64) float64 {
	if snd == 0 {
		return 0.0
	}
	return float64(fst) / float64(snd)
}

// A type for representing a throttle reason
type ThrottleReason uint32

const (
	// Throttling not active
	ThrottleReasonNone ThrottleReason = 0
	// Throttling in idle or unused state
	ThrottleReasonIdle ThrottleReason = 1 << 0
	// Throttling triggered by high temperature
	ThrottleReasonThermalSlowdown ThrottleReason = 1 << 1
	// FuriosaSmiThrottleReasonAppPowerCap as defined in smi/furiosa_smi.h:281
	ThrottleReasonAppPowerCap ThrottleReason = 1 << 2
	// FuriosaSmiThrottleReasonAppClockCap as defined in smi/furiosa_smi.h:284
	ThrottleReasonAppClockCap ThrottleReason = 1 << 3
	// FuriosaSmiThrottleReasonHwClockCap as defined in smi/furiosa_smi.h:287
	ThrottleReasonHwClockCap ThrottleReason = 1 << 4
	// FuriosaSmiThrottleReasonHwBusLimit as defined in smi/furiosa_smi.h:290
	ThrottleReasonHwBusLimit ThrottleReason = 1 << 5
	// FuriosaSmiThrottleReasonHwPowerCap as defined in smi/furiosa_smi.h:293
	ThrottleReasonHwPowerCap ThrottleReason = 1 << 6
	// FuriosaSmiThrottleReasonOtherReason as defined in smi/furiosa_smi.h:296
	ThrottleReasonOtherReason ThrottleReason = 1 << 7

	FuriosaSmiThrottleReasonNone            = ThrottleReasonNone
	FuriosaSmiThrottleReasonIdle            = ThrottleReasonIdle
	FuriosaSmiThrottleReasonThermalSlowdown = ThrottleReasonThermalSlowdown
	FuriosaSmiThrottleReasonAppPowerCap     = ThrottleReasonAppPowerCap
	FuriosaSmiThrottleReasonAppClockCap     = ThrottleReasonAppClockCap
	FuriosaSmiThrottleReasonHwClockCap      = ThrottleReasonHwClockCap
	FuriosaSmiThrottleReasonHwBusLimit      = ThrottleReasonHwBusLimit
	FuriosaSmiThrottleReasonHwPowerCap      = ThrottleReasonHwPowerCap
	FuriosaSmiThrottleReasonOtherReason     = ThrottleReasonOtherReason
)

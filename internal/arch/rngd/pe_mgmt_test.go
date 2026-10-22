package rngd

// furiosa-smi/src/sysfs/rngd/pe_mgmt.rs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

var testMemoryConfig = sysfs.MemoryConfig{
	DRAMTotalSize:        48 * 1024 * 1024 * 1024,
	FirmwareReservedSize: 512 * 1024 * 1024,
	PECount:              8,
}

type RNGDPEMgmtTestSuite struct {
	suite.Suite
}

func (*RNGDPEMgmtTestSuite) SetupSuite() {
}

func (*RNGDPEMgmtTestSuite) TearDownSuite() {
}

// SetupTest points internal/sysroot at a fresh mock directory, so the
// tests that read PE nodes off disk start from an empty sysroot. Per that
// package's rule, no test in this suite may call t.Parallel().
func (s *RNGDPEMgmtTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
}

func (*RNGDPEMgmtTestSuite) TearDownTest() {
}

// buildMock builds mock's sysfs fixture and returns its mgmt node path --
// the argument every exported pe_mgmt function takes.
func (s *RNGDPEMgmtTestSuite) buildMock(mock *RngdMock) string {
	s.T().Helper()
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{mock}})

	return filepath.Join(RNGDManagementRoot(), fmt.Sprintf("rngd!npu%dmgmt", mock.DevNodeIndex))
}

// accessibleCores returns the set of every core index below count, standing
// in for what DevfsOps.GetAccessibleCoreIndices reports for a device whose
// PEs are all visible.
func (s *RNGDPEMgmtTestSuite) accessibleCores(count uint32) map[uint32]struct{} {
	s.T().Helper()
	cores := make(map[uint32]struct{}, count)
	for core := range count {
		cores[core] = struct{}{}
	}

	return cores
}

func (s *RNGDPEMgmtTestSuite) TestParsePerfCounterBasic() {
	cycle, task, err := parsePerfCounter("CycleCount: 0x12345678\nTaskCycleCount: 0x9abcdef0")
	s.Require().NoError(err)
	s.Equal(uint64(0x12345678), cycle)
	s.Equal(uint64(0x9abcdef0), task)
}

func (s *RNGDPEMgmtTestSuite) TestParsePerfCounterWithTrailingNewline() {
	cycle, task, err := parsePerfCounter("CycleCount: 0x1\nTaskCycleCount: 0x2\n")
	s.Require().NoError(err)
	s.Equal(uint64(1), cycle)
	s.Equal(uint64(2), task)
}

func (s *RNGDPEMgmtTestSuite) TestParsePerfCounterZeroValues() {
	cycle, task, err := parsePerfCounter("CycleCount: 0x0\nTaskCycleCount: 0x0")
	s.Require().NoError(err)
	s.Equal(uint64(0), cycle)
	s.Equal(uint64(0), task)
}

func (s *RNGDPEMgmtTestSuite) TestParsePerfCounterMissingFieldErrors() {
	_, _, err := parsePerfCounter("CycleCount: 0x1")
	s.Error(err)
	_, _, err = parsePerfCounter("TaskCycleCount: 0x1")
	s.Error(err)
	_, _, err = parsePerfCounter("")
	s.Error(err)
}

func (s *RNGDPEMgmtTestSuite) TestDevNodeFromMgmtParsesNumericSuffix() {
	got, err := devNodeFromMgmt("/sys/class/rngd_mgmt/rngd!npu0mgmt")
	s.Require().NoError(err)
	s.Equal(uint8(0), got)

	got, err = devNodeFromMgmt("/sys/class/rngd_mgmt/rngd!npu7mgmt")
	s.Require().NoError(err)
	s.Equal(uint8(7), got)
}

func (s *RNGDPEMgmtTestSuite) TestDevNodeFromMgmtRejectsGarbage() {
	_, err := devNodeFromMgmt("/sys/class/rngd_mgmt/rngd!npu0pe0")
	s.Error(err)
	_, err = devNodeFromMgmt("garbage")
	s.Error(err)
}

func (s *RNGDPEMgmtTestSuite) TestPEPathResolvesSibling() {
	s.Equal("/sys/class/rngd_mgmt/rngd!npu0pe3", pePath("/sys/class/rngd_mgmt", 0, 3))
}

func (s *RNGDPEMgmtTestSuite) TestParseMemoryInfoSumsDRAMAndShared() {
	input := "dram_capacity: 0x100\ndram_usage: 0x40\nshared_capacity: 0x80\nshared_usage: 0x20"
	info, err := parseMemoryInfo(input)
	s.Require().NoError(err)
	s.Equal(uint64(0x180), info.TotalBytes())
	s.Equal(uint64(0x60), info.InUseBytes())
}

func (s *RNGDPEMgmtTestSuite) TestParseMemoryInfoMissingFieldsErrors() {
	_, err := parseMemoryInfo("dram_capacity: 0x1\ndram_usage: 0x1")
	s.Error(err)
}

func (s *RNGDPEMgmtTestSuite) TestFusionSizeSinglePE() {
	info := device.NewMemoryInfo(testMemoryConfig.DRAMPerPE(), 0)
	s.Equal(1, fusionSize(info, testMemoryConfig))
	s.Equal(1, fusionSize(device.MemoryInfo{}, testMemoryConfig))
}

func (s *RNGDPEMgmtTestSuite) TestFusionSizeMultiPE() {
	dramPerPE := testMemoryConfig.DRAMPerPE()
	s.Equal(2, fusionSize(device.NewMemoryInfo(dramPerPE*2, 0), testMemoryConfig))
	s.Equal(4, fusionSize(device.NewMemoryInfo(dramPerPE*4, 0), testMemoryConfig))
	s.Equal(8, fusionSize(device.NewMemoryInfo(dramPerPE*8, 0), testMemoryConfig))
}

func (s *RNGDPEMgmtTestSuite) TestBuildMemoryUsageKeysByFusionRange() {
	dramPerPE := testMemoryConfig.DRAMPerPE()
	peMemoryMap := map[uint32]device.MemoryInfo{
		0: device.NewMemoryInfo(dramPerPE, 0),
		1: device.NewMemoryInfo(dramPerPE, 0),
		2: device.NewMemoryInfo(dramPerPE*2, 100),
		3: device.NewMemoryInfo(dramPerPE*2, 100),
	}

	usage, err := buildMemoryUsage(peMemoryMap, 0, testMemoryConfig)
	s.Require().NoError(err)
	s.Len(usage, 3)
	s.Equal(dramPerPE, usage[device.NewDeviceCoreRange(0, 0, 0)].TotalBytes())
	s.Equal(dramPerPE, usage[device.NewDeviceCoreRange(0, 1, 1)].TotalBytes())
	s.Equal(dramPerPE*2, usage[device.NewDeviceCoreRange(0, 2, 3)].TotalBytes())
	s.Equal(uint64(100), usage[device.NewDeviceCoreRange(0, 2, 3)].InUseBytes())
}

func (s *RNGDPEMgmtTestSuite) TestBuildMemoryUsageErrorsOnPartialFusionGroup() {
	dramPerPE := testMemoryConfig.DRAMPerPE()
	peMemoryMap := map[uint32]device.MemoryInfo{
		6: device.NewMemoryInfo(dramPerPE*4, 0),
		7: device.NewMemoryInfo(dramPerPE*4, 0),
	}

	_, err := buildMemoryUsage(peMemoryMap, 0, testMemoryConfig)
	s.Error(err)
}

func (s *RNGDPEMgmtTestSuite) TestBuildMemoryUsageHandlesMixedFusionSizes() {
	dramPerPE := testMemoryConfig.DRAMPerPE()
	peMemoryMap := map[uint32]device.MemoryInfo{
		0: device.NewMemoryInfo(dramPerPE*2, 0),
		1: device.NewMemoryInfo(dramPerPE*2, 0),
		2: device.NewMemoryInfo(dramPerPE*4, 0),
		3: device.NewMemoryInfo(dramPerPE*4, 0),
		4: device.NewMemoryInfo(dramPerPE*4, 0),
		5: device.NewMemoryInfo(dramPerPE*4, 0),
		6: device.NewMemoryInfo(dramPerPE, 0),
		7: device.NewMemoryInfo(dramPerPE, 0),
	}

	usage, err := buildMemoryUsage(peMemoryMap, 0, testMemoryConfig)
	s.Require().NoError(err)
	s.Len(usage, 4)
	s.Contains(usage, device.NewDeviceCoreRange(0, 0, 1))
	s.Contains(usage, device.NewDeviceCoreRange(0, 2, 5))
	s.Contains(usage, device.NewDeviceCoreRange(0, 6, 6))
	s.Contains(usage, device.NewDeviceCoreRange(0, 7, 7))
}

func (s *RNGDPEMgmtTestSuite) TestBuildMemoryUsageArbitraryStartFusion() {
	dramPerPE := testMemoryConfig.DRAMPerPE()
	peMemoryMap := map[uint32]device.MemoryInfo{
		0: device.NewMemoryInfo(dramPerPE, 0),
		1: device.NewMemoryInfo(dramPerPE*2, 0),
		2: device.NewMemoryInfo(dramPerPE*2, 0),
		3: device.NewMemoryInfo(dramPerPE, 0),
	}

	usage, err := buildMemoryUsage(peMemoryMap, 0, testMemoryConfig)
	s.Require().NoError(err)
	s.Contains(usage, device.NewDeviceCoreRange(0, 1, 2))
}

// TestReadPerformanceCounterReadsOnlyOccupiedCores covers the read path's
// central rule: a PE's "metrics" node only exists while that PE is
// allocated, so an unoccupied core must report a zero counter rather than
// fail on the missing file.
func (s *RNGDPEMgmtTestSuite) TestReadPerformanceCounterReadsOnlyOccupiedCores() {
	mgmtPath := s.buildMock(&RngdMock{
		BDF:     "0000:27:00.0",
		UUID:    "test-uuid",
		Serial:  "test-serial",
		CoreNum: 4,
		OccupiedPEs: []PEMock{
			{Core: 1, CycleCount: 0x12345678, TaskCycleCount: 0x9abcdef0},
		},
	})

	coreStatus, err := ReadCoreStatus(mgmtPath)
	s.Require().NoError(err)

	counters, err := ReadPerformanceCounter(mgmtPath, s.accessibleCores(4), coreStatus)
	s.Require().NoError(err)
	s.Require().Len(counters, 4)

	s.Equal(uint64(0x12345678), counters[1].CycleCount())
	s.Equal(uint64(0x9abcdef0), counters[1].TaskExecutionCycle())

	for _, core := range []uint32{0, 2, 3} {
		s.Zero(counters[core].CycleCount(), "core %d", core)
		s.Zero(counters[core].TaskExecutionCycle(), "core %d", core)
	}
}

func (s *RNGDPEMgmtTestSuite) TestReadPerformanceCounterRejectsInvalidMgmtPath() {
	_, err := ReadPerformanceCounter("/sys/class/rngd_mgmt/garbage", s.accessibleCores(1), nil)
	s.Error(err)
}

// TestReadMemoryUsageDefaultsUnoccupiedCoresToPerPEShare pins the behavior
// an idle device shows: no PE reports an allocation, so each core stands
// alone and owns exactly its share of DRAM.
func (s *RNGDPEMgmtTestSuite) TestReadMemoryUsageDefaultsUnoccupiedCoresToPerPEShare() {
	mgmtPath := s.buildMock(&RngdMock{
		BDF:     "0000:27:00.0",
		UUID:    "test-uuid",
		Serial:  "test-serial",
		CoreNum: 4,
	})

	coreStatus, err := ReadCoreStatus(mgmtPath)
	s.Require().NoError(err)

	usage, err := ReadMemoryUsage(mgmtPath, s.accessibleCores(4), coreStatus, testMemoryConfig)
	s.Require().NoError(err)
	s.Require().Len(usage, 4)

	for core := range uint32(4) {
		info, ok := usage[device.NewDeviceCoreRange(0, core, core)]
		s.Require().True(ok, "core %d missing", core)
		s.Equal(testMemoryConfig.DRAMPerPE(), info.TotalBytes())
		s.Zero(info.InUseBytes())
	}
}

func (s *RNGDPEMgmtTestSuite) TestReadMemoryUsageReadsAllocStatusOfOccupiedCore() {
	dramPerPE := testMemoryConfig.DRAMPerPE()
	mgmtPath := s.buildMock(&RngdMock{
		BDF:     "0000:27:00.0",
		UUID:    "test-uuid",
		Serial:  "test-serial",
		CoreNum: 2,
		OccupiedPEs: []PEMock{
			{Core: 0, DRAMCapacity: dramPerPE, DRAMUsage: 0x400, SharedCapacity: 0, SharedUsage: 0x100},
		},
	})

	coreStatus, err := ReadCoreStatus(mgmtPath)
	s.Require().NoError(err)

	usage, err := ReadMemoryUsage(mgmtPath, s.accessibleCores(2), coreStatus, testMemoryConfig)
	s.Require().NoError(err)

	info, ok := usage[device.NewDeviceCoreRange(0, 0, 0)]
	s.Require().True(ok)
	s.Equal(dramPerPE, info.TotalBytes())
	// alloc_status' dram and shared figures are summed, not reported apart.
	s.Equal(uint64(0x500), info.InUseBytes())
}

// TestReadMemoryUsageGroupsFusedCoresIntoOneRange exercises fusion end to
// end through real files: every PE in a fusion reports the whole fusion's
// capacity, which is what collapses them into a single core range.
func (s *RNGDPEMgmtTestSuite) TestReadMemoryUsageGroupsFusedCoresIntoOneRange() {
	fusedCapacity := testMemoryConfig.DRAMPerPE() * 4
	occupied := make([]PEMock, 0, 4)
	for core := range uint32(4) {
		occupied = append(occupied, PEMock{Core: core, DRAMCapacity: fusedCapacity, DRAMUsage: 0x800})
	}

	mgmtPath := s.buildMock(&RngdMock{
		BDF:         "0000:27:00.0",
		UUID:        "test-uuid",
		Serial:      "test-serial",
		CoreNum:     4,
		OccupiedPEs: occupied,
	})

	coreStatus, err := ReadCoreStatus(mgmtPath)
	s.Require().NoError(err)

	usage, err := ReadMemoryUsage(mgmtPath, s.accessibleCores(4), coreStatus, testMemoryConfig)
	s.Require().NoError(err)
	s.Require().Len(usage, 1)

	info, ok := usage[device.NewDeviceCoreRange(0, 0, 3)]
	s.Require().True(ok)
	s.Equal(fusedCapacity, info.TotalBytes())
	s.Equal(uint64(0x800), info.InUseBytes())
}

// TestReadMemoryUsageFallsBackWhenAllocStatusIsZero covers the guard for a
// PE that reports occupied but hasn't been given any capacity yet: a zero
// total would otherwise divide into a fusion size of zero.
func (s *RNGDPEMgmtTestSuite) TestReadMemoryUsageFallsBackWhenAllocStatusIsZero() {
	mgmtPath := s.buildMock(&RngdMock{
		BDF:     "0000:27:00.0",
		UUID:    "test-uuid",
		Serial:  "test-serial",
		CoreNum: 1,
		OccupiedPEs: []PEMock{
			{Core: 0, DRAMCapacity: 1, SharedCapacity: 0},
		},
	})
	// DRAMCapacity of 1 still produces a non-zero total, so overwrite the
	// node with an all-zero report, which is the case under test.
	peDir := filepath.Join(RNGDManagementRoot(), "rngd!npu0pe0")
	sysroot.MustWriteFile(s.T(), filepath.Join(peDir, "alloc_status"),
		"dram_capacity: 0x0\ndram_usage: 0x0\nshared_capacity: 0x0\nshared_usage: 0x0\n")

	coreStatus, err := ReadCoreStatus(mgmtPath)
	s.Require().NoError(err)

	usage, err := ReadMemoryUsage(mgmtPath, s.accessibleCores(1), coreStatus, testMemoryConfig)
	s.Require().NoError(err)

	info, ok := usage[device.NewDeviceCoreRange(0, 0, 0)]
	s.Require().True(ok)
	s.Equal(testMemoryConfig.DRAMPerPE(), info.TotalBytes())
}

// TestReadMemoryUsagePropagatesMissingAllocStatus confirms a core that
// reports occupied but has no readable PE node surfaces the I/O error
// rather than silently falling back to the per-PE share.
func (s *RNGDPEMgmtTestSuite) TestReadMemoryUsagePropagatesMissingAllocStatus() {
	mgmtPath := s.buildMock(&RngdMock{
		BDF:         "0000:27:00.0",
		UUID:        "test-uuid",
		Serial:      "test-serial",
		CoreNum:     1,
		OccupiedPEs: []PEMock{{Core: 0}},
	})
	s.Require().NoError(os.RemoveAll(filepath.Join(RNGDManagementRoot(), "rngd!npu0pe0")))

	coreStatus, err := ReadCoreStatus(mgmtPath)
	s.Require().NoError(err)

	_, err = ReadMemoryUsage(mgmtPath, s.accessibleCores(1), coreStatus, testMemoryConfig)
	s.Error(err)
}

func TestRNGDPEMgmtSuite(t *testing.T) {
	suite.Run(t, new(RNGDPEMgmtTestSuite))
}

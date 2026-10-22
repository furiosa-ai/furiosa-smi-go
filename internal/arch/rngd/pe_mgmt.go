package rngd

// furiosa-smi/src/sysfs/rngd/pe_mgmt.rs

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/utils"
)

func ReadPerformanceCounter(mgmtPath string, accessible map[uint32]struct{}, coreStatus map[uint32]device.CoreStatus) (map[uint32]device.PePerformanceCounter, error) {
	classRoot, devNode, err := peClassContext(mgmtPath)
	if err != nil {
		return nil, err
	}

	sorted := make([]uint32, 0, len(accessible))
	for core := range accessible {
		sorted = append(sorted, core)
	}

	slices.Sort(sorted)

	result := make(map[uint32]device.PePerformanceCounter, len(sorted))
	for _, core := range sorted {
		counter, err := readPEPerformanceCounter(classRoot, devNode, core, coreStatusOf(coreStatus, core))
		if err != nil {
			return nil, err
		}

		result[core] = counter
	}

	return result, nil
}

func readPEPerformanceCounter(classRoot string, devNode uint8, core uint32, status *device.CoreStatus) (device.PePerformanceCounter, error) {
	timestamp := time.Now()
	var cycleCount, taskExecutionCycle uint64
	if status != nil && *status == device.CoreStatusOccupied {
		metricsPath := filepath.Join(pePath(classRoot, devNode, core), "metrics")
		text, err := sysfs.ReadSysfsAttr(metricsPath)
		if err != nil {
			return device.PePerformanceCounter{}, err
		}

		cycleCount, taskExecutionCycle, err = parsePerfCounter(text)
		if err != nil {
			return device.PePerformanceCounter{}, err
		}
	}

	return device.NewPePerformanceCounter(timestamp, cycleCount, taskExecutionCycle), nil
}

func parsePerfCounter(text string) (uint64, uint64, error) {
	fields := utils.ParseColonFields(text)
	cycleCountRaw, ok1 := fields["CycleCount"]
	taskCycleCountRaw, ok2 := fields["TaskCycleCount"]
	if !ok1 || !ok2 {
		return 0, 0, smierrors.Parse("performance counter has unexpected format")
	}

	cycleCount, err := utils.ParseHexU64(cycleCountRaw)
	if err != nil {
		return 0, 0, err
	}

	taskExecutionCycle, err := utils.ParseHexU64(taskCycleCountRaw)
	if err != nil {
		return 0, 0, err
	}

	return cycleCount, taskExecutionCycle, nil
}

func ReadMemoryUsage(mgmtPath string, accessible map[uint32]struct{}, coreStatus map[uint32]device.CoreStatus, memoryConfig sysfs.MemoryConfig) (map[device.DeviceCoreRange]device.MemoryInfo, error) {
	classRoot, devNode, err := peClassContext(mgmtPath)
	if err != nil {
		return nil, err
	}

	peMemoryMap := make(map[uint32]device.MemoryInfo, len(accessible))
	for core := range accessible {
		info, err := readPEMemoryInfo(classRoot, devNode, core, coreStatusOf(coreStatus, core), memoryConfig)
		if err != nil {
			return nil, err
		}

		peMemoryMap[core] = info
	}

	return buildMemoryUsage(peMemoryMap, devNode, memoryConfig)
}

func readPEMemoryInfo(classRoot string, devNode uint8, core uint32, status *device.CoreStatus, memoryConfig sysfs.MemoryConfig) (device.MemoryInfo, error) {
	if status == nil || *status != device.CoreStatusOccupied {
		return device.NewMemoryInfo(memoryConfig.DRAMPerPE(), 0), nil
	}

	info, err := fetchPEMemoryInfo(classRoot, devNode, core)
	if err != nil {
		return device.MemoryInfo{}, err
	}

	if info.TotalBytes() == 0 {
		return device.NewMemoryInfo(memoryConfig.DRAMPerPE(), 0), nil
	}

	return info, nil
}

func fetchPEMemoryInfo(classRoot string, devNode uint8, core uint32) (device.MemoryInfo, error) {
	path := filepath.Join(pePath(classRoot, devNode, core), "alloc_status")
	text, err := sysfs.ReadSysfsAttr(path)
	if err != nil {
		return device.MemoryInfo{}, err
	}

	return parseMemoryInfo(text)
}

func parseMemoryInfo(text string) (device.MemoryInfo, error) {
	fields := utils.ParseColonFields(text)
	dc, ok1 := fields["dram_capacity"]
	du, ok2 := fields["dram_usage"]
	sc, ok3 := fields["shared_capacity"]
	su, ok4 := fields["shared_usage"]
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return device.MemoryInfo{}, smierrors.Parse("alloc_status has unexpected format")
	}

	dramCapacity, err := utils.ParseHexU64(dc)
	if err != nil {
		return device.MemoryInfo{}, err
	}

	dramUsage, err := utils.ParseHexU64(du)
	if err != nil {
		return device.MemoryInfo{}, err
	}

	sharedCapacity, err := utils.ParseHexU64(sc)
	if err != nil {
		return device.MemoryInfo{}, err
	}

	sharedUsage, err := utils.ParseHexU64(su)
	if err != nil {
		return device.MemoryInfo{}, err
	}

	return device.NewMemoryInfo(dramCapacity+sharedCapacity, dramUsage+sharedUsage), nil
}

func buildMemoryUsage(peMemoryMap map[uint32]device.MemoryInfo, devNode uint8, memoryConfig sysfs.MemoryConfig) (map[device.DeviceCoreRange]device.MemoryInfo, error) {
	cores := make([]uint32, 0, len(peMemoryMap))
	for core := range peMemoryMap {
		cores = append(cores, core)
	}

	// Go maps are unordered, unlike Rust's BTreeMap, so cores must be sorted
	// explicitly before grouping.
	slices.Sort(cores)

	groups, err := fusionGroups(cores, peMemoryMap, memoryConfig)
	if err != nil {
		return nil, err
	}

	result := make(map[device.DeviceCoreRange]device.MemoryInfo, len(groups))
	for _, group := range groups {
		info := peMemoryMap[group[0]]
		result[device.NewDeviceCoreRange(devNode, group[0], group[len(group)-1])] = info
	}

	return result, nil
}

// fusionGroups splits cores into contiguous fusion groups; each group
// covers its head PE's fusion size.
func fusionGroups(cores []uint32, peMemoryMap map[uint32]device.MemoryInfo, memoryConfig sysfs.MemoryConfig) ([][]uint32, error) {
	var groups [][]uint32
	groupStart := 0
	for groupStart < len(cores) {
		groupSize := fusionSize(peMemoryMap[cores[groupStart]], memoryConfig)
		groupEnd := groupStart + groupSize
		if groupEnd > len(cores) {
			return nil, smierrors.UnexpectedValue(fmt.Sprintf(
				"fusion group at PE %d (size %d) exceeds accessible cores", cores[groupStart], groupSize))
		}

		groups = append(groups, cores[groupStart:groupEnd])
		groupStart = groupEnd
	}

	return groups, nil
}

// fusionSize returns the number of PEs spanned by a fusion group. Each PE
// in a fusion reports the entire fusion's DRAM size.
func fusionSize(info device.MemoryInfo, memoryConfig sysfs.MemoryConfig) int {
	size := max(info.TotalBytes()/memoryConfig.DRAMPerPE(), 1)

	return int(size)
}

// coreStatusOf mirrors Rust's core_status.get(&core).copied(): nil when
// the core has no reported status.
func coreStatusOf(coreStatus map[uint32]device.CoreStatus, core uint32) *device.CoreStatus {
	status, ok := coreStatus[core]
	if !ok {
		return nil
	}

	return &status
}

func peClassContext(mgmtPath string) (string, uint8, error) {
	devNode, err := devNodeFromMgmt(mgmtPath)
	if err != nil {
		return "", 0, err
	}

	return filepath.Dir(mgmtPath), devNode, nil
}

func pePath(classRoot string, devNode uint8, core uint32) string {
	return filepath.Join(classRoot, fmt.Sprintf("rngd!npu%dpe%d", devNode, core))
}

// devNodeFromMgmt parses the same "rngd!npuNmgmt" shape ParseMgmtNodeIndex
// does; reuses it instead of re-parsing with a different technique.
func devNodeFromMgmt(mgmtPath string) (uint8, error) {
	name := filepath.Base(mgmtPath)
	value := ParseMgmtNodeIndex(name)
	if value == nil {
		return 0, smierrors.Parse(fmt.Sprintf("invalid mgmt name: '%s'", name))
	}

	return *value, nil
}

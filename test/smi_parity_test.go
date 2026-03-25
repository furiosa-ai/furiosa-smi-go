package test

import (
	"math"
	"os"
	"reflect"
	"sort"
	"testing"

	legacysmi "github.com/furiosa-ai/furiosa-smi-go/pkg/legacy"
	newsmi "github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

const floatEpsilon = 1e-6

type parityState struct {
	oldDevices map[string]legacysmi.Device
	newDevices map[string]newsmi.Device
	allBDFs    []string
}

func TestSMIParity_PackageLevel(t *testing.T) {
	st := setupParityState(t)

	t.Run("DriverInfo", func(t *testing.T) {
		oldInfo, oldErr := legacysmi.DriverInfo()
		newInfo, newErr := newsmi.DriverInfo()
		assertErrorParity(t, oldErr, newErr)
		if oldErr != nil {
			return
		}
		if !reflect.DeepEqual(versionSnapshotFromLegacy(oldInfo), versionSnapshotFromNew(newInfo)) {
			t.Fatalf("driver info mismatch")
		}
	})

	t.Run("ListDevices", func(t *testing.T) {
		if !reflect.DeepEqual(st.allBDFs, sortedKeys(st.newDevices)) {
			t.Fatalf("list devices mismatch")
		}
	})

	t.Run("ListDisabledDevices", func(t *testing.T) {
		oldBDFs, oldErr := legacysmi.ListDisabledDevices()
		newBDFs, newErr := newsmi.ListDisabledDevices()
		assertErrorParity(t, oldErr, newErr)
		if oldErr != nil {
			return
		}
		sort.Strings(oldBDFs)
		sort.Strings(newBDFs)
		if !reflect.DeepEqual(oldBDFs, newBDFs) {
			t.Fatalf("list disabled devices mismatch: old=%v new=%v", oldBDFs, newBDFs)
		}
	})
}

func TestSMIParity_DeviceMethods(t *testing.T) {
	st := setupParityState(t)
	if len(st.allBDFs) == 0 {
		t.Skip("no devices discovered")
	}

	for _, bdf := range st.allBDFs {
		bdf := bdf
		t.Run("Device@"+bdf, func(t *testing.T) {
			oldDev := st.oldDevices[bdf]
			newDev := st.newDevices[bdf]

			t.Run("DeviceInfo", func(t *testing.T) {
				oldInfo, oldErr := oldDev.DeviceInfo()
				newInfo, newErr := newDev.DeviceInfo()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !reflect.DeepEqual(deviceInfoSnapshotFromLegacy(oldInfo), deviceInfoSnapshotFromNew(newInfo)) {
					t.Fatalf("device info mismatch")
				}
			})

			t.Run("DeviceFiles", func(t *testing.T) {
				oldFiles, oldErr := oldDev.DeviceFiles()
				newFiles, newErr := newDev.DeviceFiles()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !reflect.DeepEqual(deviceFilesSnapshotFromLegacy(oldFiles), deviceFilesSnapshotFromNew(newFiles)) {
					t.Fatalf("device files mismatch")
				}
			})

			t.Run("CoreStatus", func(t *testing.T) {
				oldStatus, oldErr := oldDev.CoreStatus()
				newStatus, newErr := newDev.CoreStatus()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !reflect.DeepEqual(coreStatusSnapshotFromLegacy(oldStatus), coreStatusSnapshotFromNew(newStatus)) {
					t.Fatalf("core status mismatch")
				}
			})

			t.Run("Liveness", func(t *testing.T) {
				oldLive, oldErr := oldDev.Liveness()
				newLive, newErr := newDev.Liveness()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if oldLive != newLive {
					t.Fatalf("liveness mismatch: old=%v new=%v", oldLive, newLive)
				}
			})

			t.Run("CoreFrequency", func(t *testing.T) {
				oldFreq, oldErr := oldDev.CoreFrequency()
				newFreq, newErr := newDev.CoreFrequency()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !reflect.DeepEqual(coreFrequencySnapshotFromLegacy(oldFreq), coreFrequencySnapshotFromNew(newFreq)) {
					t.Fatalf("core frequency mismatch")
				}
			})

			t.Run("MemoryFrequency", func(t *testing.T) {
				oldFreq, oldErr := oldDev.MemoryFrequency()
				newFreq, newErr := newDev.MemoryFrequency()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if oldFreq.Frequency() != newFreq.Frequency() {
					t.Fatalf("memory frequency mismatch: old=%d new=%d", oldFreq.Frequency(), newFreq.Frequency())
				}
			})

			t.Run("PowerConsumption", func(t *testing.T) {
				oldPower, oldErr := oldDev.PowerConsumption()
				newPower, newErr := newDev.PowerConsumption()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !floatAlmostEqual(oldPower, newPower) {
					t.Fatalf("power consumption mismatch: old=%f new=%f", oldPower, newPower)
				}
			})

			t.Run("DeviceTemperature", func(t *testing.T) {
				oldTemp, oldErr := oldDev.DeviceTemperature()
				newTemp, newErr := newDev.DeviceTemperature()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !floatAlmostEqual(oldTemp.SocPeak(), newTemp.SocPeak()) || !floatAlmostEqual(oldTemp.Ambient(), newTemp.Ambient()) {
					t.Fatalf("device temperature mismatch")
				}
			})

			t.Run("DevicePerformanceCounter", func(t *testing.T) {
				oldPerf, oldErr := oldDev.DevicePerformanceCounter()
				newPerf, newErr := newDev.DevicePerformanceCounter()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !reflect.DeepEqual(perfSnapshotFromLegacy(oldPerf), perfSnapshotFromNew(newPerf)) {
					t.Fatalf("device performance counter mismatch")
				}
			})

			t.Run("GovernorProfile", func(t *testing.T) {
				oldProfile, oldErr := oldDev.GovernorProfile()
				newProfile, newErr := newDev.GovernorProfile()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if uint32(oldProfile) != uint32(newProfile) {
					t.Fatalf("governor profile mismatch: old=%d new=%d", oldProfile, newProfile)
				}
			})

			t.Run("PcieInfo", func(t *testing.T) {
				oldPcie, oldErr := oldDev.PcieInfo()
				newPcie, newErr := newDev.PcieInfo()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !reflect.DeepEqual(pcieSnapshotFromLegacy(oldPcie), pcieSnapshotFromNew(newPcie)) {
					t.Fatalf("pcie info mismatch")
				}
			})

			t.Run("ThrottleReason", func(t *testing.T) {
				oldReason, oldErr := oldDev.ThrottleReason()
				newReason, newErr := newDev.ThrottleReason()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if uint32(oldReason) != uint32(newReason) {
					t.Fatalf("throttle reason mismatch: old=%d new=%d", oldReason, newReason)
				}
			})

			t.Run("MemoryUtilization", func(t *testing.T) {
				oldMem, oldErr := oldDev.MemoryUtilization()
				newMem, newErr := newDev.MemoryUtilization()
				assertErrorParity(t, oldErr, newErr)
				if oldErr != nil {
					return
				}
				if !reflect.DeepEqual(memoryUtilSnapshotFromLegacy(oldMem), memoryUtilSnapshotFromNew(newMem)) {
					t.Fatalf("memory utilization mismatch")
				}
			})
		})
	}
}

func TestSMIParity_DevicePairMethods(t *testing.T) {
	st := setupParityState(t)
	if len(st.allBDFs) < 2 {
		t.Skip("need at least two devices")
	}

	for _, srcBDF := range st.allBDFs {
		for _, dstBDF := range st.allBDFs {
			srcOld := st.oldDevices[srcBDF]
			dstOld := st.oldDevices[dstBDF]
			srcNew := st.newDevices[srcBDF]
			dstNew := st.newDevices[dstBDF]

			t.Run("pair_"+srcBDF+"_to_"+dstBDF, func(t *testing.T) {
				oldLink, oldLinkErr := srcOld.DeviceToDeviceLinkType(dstOld)
				newLink, newLinkErr := srcNew.DeviceToDeviceLinkType(dstNew)
				assertErrorParity(t, oldLinkErr, newLinkErr)
				if oldLinkErr == nil && uint32(oldLink) != uint32(newLink) {
					t.Fatalf("link type mismatch: old=%d new=%d", oldLink, newLink)
				}

				oldP2P, oldP2PErr := srcOld.P2PAccessible(dstOld)
				newP2P, newP2PErr := srcNew.P2PAccessible(dstNew)
				assertErrorParity(t, oldP2PErr, newP2PErr)
				if oldP2PErr == nil && oldP2P != newP2P {
					t.Fatalf("p2p mismatch: old=%v new=%v", oldP2P, newP2P)
				}
			})
		}
	}
}

func TestSMIParity_MutatingMethods_OptIn(t *testing.T) {
	if os.Getenv("FURIOSA_SMI_PARITY_MUTATING") != "1" {
		t.Skip("set FURIOSA_SMI_PARITY_MUTATING=1 to enable mutating parity tests")
	}

	st := setupParityState(t)
	if len(st.allBDFs) == 0 {
		t.Skip("no devices discovered")
	}

	bdf := st.allBDFs[0]
	oldDev := st.oldDevices[bdf]
	newDev := st.newDevices[bdf]

	t.Run("SetGovernorProfile", func(t *testing.T) {
		oldCurrent, oldErr := oldDev.GovernorProfile()
		newCurrent, newErr := newDev.GovernorProfile()
		assertErrorParity(t, oldErr, newErr)
		if oldErr != nil {
			return
		}

		oldSetErr := oldDev.SetGovernorProfile(oldCurrent)
		newSetErr := newDev.SetGovernorProfile(newCurrent)
		assertErrorParity(t, oldSetErr, newSetErr)
	})

	t.Run("EnableDisableByBDF", func(t *testing.T) {
		defer func() {
			_ = legacysmi.EnableDevice(bdf)
			_ = newsmi.EnableDevice(bdf)
		}()

		oldDisableErr := legacysmi.DisableDevice(bdf)
		newDisableErr := newsmi.DisableDevice(bdf)
		assertErrorParity(t, oldDisableErr, newDisableErr)

		oldEnableErr := legacysmi.EnableDevice(bdf)
		newEnableErr := newsmi.EnableDevice(bdf)
		assertErrorParity(t, oldEnableErr, newEnableErr)
	})
}

func setupParityState(t *testing.T) parityState {
	t.Helper()

	oldInitErr := legacysmi.Init()
	if oldInitErr != nil {
		t.Skipf("legacy backend unavailable: %v", oldInitErr)
	}

	if err := newsmi.Init(); err != nil {
		t.Fatalf("new smi Init failed while legacy succeeded: %v", err)
	}

	oldDevices, oldErr := legacyDeviceMapByBDF()
	newDevices, newErr := newDeviceMapByBDF()
	assertErrorParity(t, oldErr, newErr)
	if oldErr != nil {
		return parityState{}
	}

	oldBDFs := sortedKeys(oldDevices)
	newBDFs := sortedKeys(newDevices)
	if !reflect.DeepEqual(oldBDFs, newBDFs) {
		t.Fatalf("device identity mismatch: old=%v new=%v", oldBDFs, newBDFs)
	}

	return parityState{oldDevices: oldDevices, newDevices: newDevices, allBDFs: oldBDFs}
}

func legacyDeviceMapByBDF() (map[string]legacysmi.Device, error) {
	devices, err := legacysmi.ListDevices()
	if err != nil {
		return nil, err
	}
	out := make(map[string]legacysmi.Device, len(devices))
	for _, d := range devices {
		info, infoErr := d.DeviceInfo()
		if infoErr != nil {
			return nil, infoErr
		}
		out[info.BDF()] = d
	}
	return out, nil
}

func newDeviceMapByBDF() (map[string]newsmi.Device, error) {
	devices, err := newsmi.ListDevices()
	if err != nil {
		return nil, err
	}
	out := make(map[string]newsmi.Device, len(devices))
	for _, d := range devices {
		info, infoErr := d.DeviceInfo()
		if infoErr != nil {
			return nil, infoErr
		}
		out[info.BDF()] = d
	}
	return out, nil
}

type versionSnapshot struct {
	Major      uint32
	Minor      uint32
	Patch      uint32
	Metadata   string
	Prerelease string
}

func versionSnapshotFromLegacy(v legacysmi.VersionInfo) versionSnapshot {
	return versionSnapshot{Major: v.Major(), Minor: v.Minor(), Patch: v.Patch(), Metadata: v.Metadata(), Prerelease: v.Prerelease()}
}

func versionSnapshotFromNew(v newsmi.VersionInfo) versionSnapshot {
	return versionSnapshot{Major: v.Major(), Minor: v.Minor(), Patch: v.Patch(), Metadata: v.Metadata(), Prerelease: v.Prerelease()}
}

type deviceInfoSnapshot struct {
	Index    uint32
	Arch     string
	CoreNum  uint32
	NumaNode int32
	Name     string
	Serial   string
	UUID     string
	BDF      string
	Major    uint16
	Minor    uint16
	FW       versionSnapshot
}

func deviceInfoSnapshotFromLegacy(i legacysmi.DeviceInfo) deviceInfoSnapshot {
	return deviceInfoSnapshot{
		Index:    i.Index(),
		Arch:     i.Arch().ToString(),
		CoreNum:  i.CoreNum(),
		NumaNode: i.NumaNode(),
		Name:     i.Name(),
		Serial:   i.Serial(),
		UUID:     i.UUID(),
		BDF:      i.BDF(),
		Major:    i.Major(),
		Minor:    i.Minor(),
		FW:       versionSnapshotFromLegacy(i.FirmwareVersion()),
	}
}

func deviceInfoSnapshotFromNew(i newsmi.DeviceInfo) deviceInfoSnapshot {
	return deviceInfoSnapshot{
		Index:    i.Index(),
		Arch:     i.Arch().ToString(),
		CoreNum:  i.CoreNum(),
		NumaNode: i.NumaNode(),
		Name:     i.Name(),
		Serial:   i.Serial(),
		UUID:     i.UUID(),
		BDF:      i.BDF(),
		Major:    i.Major(),
		Minor:    i.Minor(),
		FW:       versionSnapshotFromNew(i.FirmwareVersion()),
	}
}

type deviceFileSnapshot struct {
	Cores []uint32
	Path  string
}

func deviceFilesSnapshotFromLegacy(files []legacysmi.DeviceFile) []deviceFileSnapshot {
	ret := make([]deviceFileSnapshot, 0, len(files))
	for _, f := range files {
		ret = append(ret, deviceFileSnapshot{Cores: f.Cores(), Path: f.Path()})
	}
	return ret
}

func deviceFilesSnapshotFromNew(files []newsmi.DeviceFile) []deviceFileSnapshot {
	ret := make([]deviceFileSnapshot, 0, len(files))
	for _, f := range files {
		ret = append(ret, deviceFileSnapshot{Cores: f.Cores(), Path: f.Path()})
	}
	return ret
}

type coreStatusSnapshot struct {
	Core   uint32
	Status uint32
}

func coreStatusSnapshotFromLegacy(s legacysmi.CoreStatuses) []coreStatusSnapshot {
	pe := s.PeStatus()
	ret := make([]coreStatusSnapshot, 0, len(pe))
	for _, p := range pe {
		ret = append(ret, coreStatusSnapshot{Core: p.Core(), Status: uint32(p.Status())})
	}
	return ret
}

func coreStatusSnapshotFromNew(s newsmi.CoreStatuses) []coreStatusSnapshot {
	pe := s.PeStatus()
	ret := make([]coreStatusSnapshot, 0, len(pe))
	for _, p := range pe {
		ret = append(ret, coreStatusSnapshot{Core: p.Core(), Status: uint32(p.Status())})
	}
	return ret
}

type peFrequencySnapshot struct {
	Core      uint32
	Frequency uint32
}

func coreFrequencySnapshotFromLegacy(f legacysmi.CoreFrequency) []peFrequencySnapshot {
	pe := f.PeFrequency()
	ret := make([]peFrequencySnapshot, 0, len(pe))
	for _, p := range pe {
		ret = append(ret, peFrequencySnapshot{Core: p.Core(), Frequency: p.Frequency()})
	}
	return ret
}

func coreFrequencySnapshotFromNew(f newsmi.CoreFrequency) []peFrequencySnapshot {
	pe := f.PeFrequency()
	ret := make([]peFrequencySnapshot, 0, len(pe))
	for _, p := range pe {
		ret = append(ret, peFrequencySnapshot{Core: p.Core(), Frequency: p.Frequency()})
	}
	return ret
}

type perfSnapshot struct {
	Timestamp int64
	Core      uint32
	Cycle     uint64
	TaskCycle uint64
}

func perfSnapshotFromLegacy(p legacysmi.DevicePerformanceCounter) []perfSnapshot {
	pcs := p.PerformanceCounter()
	ret := make([]perfSnapshot, 0, len(pcs))
	for _, pc := range pcs {
		ret = append(ret, perfSnapshot{Timestamp: pc.Timestamp().Unix(), Core: pc.Core(), Cycle: pc.CycleCount(), TaskCycle: pc.TaskExecutionCycle()})
	}
	return ret
}

func perfSnapshotFromNew(p newsmi.DevicePerformanceCounter) []perfSnapshot {
	pcs := p.PerformanceCounter()
	ret := make([]perfSnapshot, 0, len(pcs))
	for _, pc := range pcs {
		ret = append(ret, perfSnapshot{Timestamp: pc.Timestamp().Unix(), Core: pc.Core(), Cycle: pc.CycleCount(), TaskCycle: pc.TaskExecutionCycle()})
	}
	return ret
}

type pcieSnapshot struct {
	DeviceID      uint16
	VendorID      uint16
	SubsystemID   uint16
	RevisionID    uint8
	ClassID       uint8
	SubclassID    uint8
	GenStatus     uint8
	WidthStatus   uint32
	SpeedStatus   float64
	MaxWidthCap   uint32
	MaxSpeedCap   float64
	SriovTotal    uint32
	SriovEnabled  uint32
	RootDomain    uint16
	RootBus       uint8
	SwitchDomain  uint16
	SwitchBus     uint8
	SwitchDevice  uint8
	SwitchFunc    uint8
	HasSwitchInfo bool
}

func pcieSnapshotFromLegacy(p legacysmi.PcieInfo) pcieSnapshot {
	di := p.DeviceInfo()
	li := p.LinkInfo()
	si := p.SriovInfo()
	rc := p.RootComplexInfo()
	sw := p.SwitchInfo()
	s := pcieSnapshot{
		DeviceID:     di.DeviceId(),
		VendorID:     di.VendorId(),
		SubsystemID:  di.SubsystemId(),
		RevisionID:   di.RevisionId(),
		ClassID:      di.ClassId(),
		SubclassID:   di.SubClassId(),
		GenStatus:    li.PcieGenStatus(),
		WidthStatus:  li.LinkWidthStatus(),
		SpeedStatus:  li.LinkSpeedStatus(),
		MaxWidthCap:  li.MaxLinkWidthCapability(),
		MaxSpeedCap:  li.MaxLinkSpeedCapability(),
		SriovTotal:   si.SriovTotalVfs(),
		SriovEnabled: si.SriovEnabledVfs(),
		RootDomain:   rc.Domain(),
		RootBus:      rc.Bus(),
	}
	if sw != nil {
		s.HasSwitchInfo = true
		s.SwitchDomain = sw.Domain()
		s.SwitchBus = sw.Bus()
		s.SwitchDevice = sw.Device()
		s.SwitchFunc = sw.Function()
	}
	return s
}

func pcieSnapshotFromNew(p newsmi.PcieInfo) pcieSnapshot {
	di := p.DeviceInfo()
	li := p.LinkInfo()
	si := p.SriovInfo()
	rc := p.RootComplexInfo()
	sw := p.SwitchInfo()
	s := pcieSnapshot{
		DeviceID:     di.DeviceId(),
		VendorID:     di.VendorId(),
		SubsystemID:  di.SubsystemId(),
		RevisionID:   di.RevisionId(),
		ClassID:      di.ClassId(),
		SubclassID:   di.SubClassId(),
		GenStatus:    li.PcieGenStatus(),
		WidthStatus:  li.LinkWidthStatus(),
		SpeedStatus:  li.LinkSpeedStatus(),
		MaxWidthCap:  li.MaxLinkWidthCapability(),
		MaxSpeedCap:  li.MaxLinkSpeedCapability(),
		SriovTotal:   si.SriovTotalVfs(),
		SriovEnabled: si.SriovEnabledVfs(),
		RootDomain:   rc.Domain(),
		RootBus:      rc.Bus(),
	}
	if sw != nil {
		s.HasSwitchInfo = true
		s.SwitchDomain = sw.Domain()
		s.SwitchBus = sw.Bus()
		s.SwitchDevice = sw.Device()
		s.SwitchFunc = sw.Function()
	}
	return s
}

type memoryBlockSnapshot struct {
	Core  []uint32
	Total uint64
	InUse uint64
}

type memoryUtilSnapshot struct {
	Dram       []memoryBlockSnapshot
	DramShared []memoryBlockSnapshot
	Sram       []memoryBlockSnapshot
	Instr      []memoryBlockSnapshot
}

func memoryUtilSnapshotFromLegacy(m legacysmi.MemoryUtilization) memoryUtilSnapshot {
	return memoryUtilSnapshot{
		Dram:       memoryBlocksFromLegacy(m.Dram()),
		DramShared: memoryBlocksFromLegacy(m.DramShared()),
		Sram:       memoryBlocksFromLegacy(m.Sram()),
		Instr:      memoryBlocksFromLegacy(m.Instruction()),
	}
}

func memoryUtilSnapshotFromNew(m newsmi.MemoryUtilization) memoryUtilSnapshot {
	return memoryUtilSnapshot{
		Dram:       memoryBlocksFromNew(m.Dram()),
		DramShared: memoryBlocksFromNew(m.DramShared()),
		Sram:       memoryBlocksFromNew(m.Sram()),
		Instr:      memoryBlocksFromNew(m.Instruction()),
	}
}

func memoryBlocksFromLegacy(m legacysmi.Memory) []memoryBlockSnapshot {
	blocks := m.Memory()
	ret := make([]memoryBlockSnapshot, 0, len(blocks))
	for _, b := range blocks {
		ret = append(ret, memoryBlockSnapshot{Core: b.Core(), Total: b.TotalBytes(), InUse: b.InUseBytes()})
	}
	return ret
}

func memoryBlocksFromNew(m newsmi.Memory) []memoryBlockSnapshot {
	blocks := m.Memory()
	ret := make([]memoryBlockSnapshot, 0, len(blocks))
	for _, b := range blocks {
		ret = append(ret, memoryBlockSnapshot{Core: b.Core(), Total: b.TotalBytes(), InUse: b.InUseBytes()})
	}
	return ret
}

func assertErrorParity(t *testing.T, oldErr, newErr error) {
	t.Helper()
	if (oldErr == nil) != (newErr == nil) {
		t.Fatalf("error parity mismatch: oldErr=%v newErr=%v", oldErr, newErr)
	}
}

func floatAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatEpsilon
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

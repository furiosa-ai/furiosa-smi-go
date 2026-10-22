package device_test

// furiosa-smi/src/device.rs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch"
	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
	"github.com/furiosa-ai/furiosa-smi-go/internal/topology"
	"github.com/furiosa-ai/furiosa-smi-go/internal/version"
)

// stubMgmtOps/stubDevfsOps/stubHwmonOps are test doubles that satisfy
// device.MgmtOps/DevfsOps/HwmonOps directly, rather than routing through a
// real architecture's sysfs parsing (internal/arch/rngd doesn't exist yet
// in this module). This tests Device's own gating/delegation/filtering
// logic in isolation -- an architecture's own parsing correctness gets
// tested against that architecture's package once it exists, and again
// through Device once it's wired in via internal/arch/config.
type stubMgmtOps struct {
	firmwareVersion    version.VersionInfo
	driverVersion      version.VersionInfo
	major, minor       uint16
	coreStatus         map[uint32]device.CoreStatus
	coreFrequency      map[uint32]uint32
	governorProfile    device.GovernorProfile
	writeGovernorErr   error
	liveness           bool
	performanceCounter map[uint32]device.PePerformanceCounter
	memoryUsage        map[device.DeviceCoreRange]device.MemoryInfo
}

func (o stubMgmtOps) ReadFirmwareVersion(string) (version.VersionInfo, error) {
	return o.firmwareVersion, nil
}
func (o stubMgmtOps) ReadDriverVersion(string) (version.VersionInfo, error) {
	return o.driverVersion, nil
}
func (o stubMgmtOps) ReadMajorMinor(string) (uint16, uint16, error) {
	return o.major, o.minor, nil
}
func (o stubMgmtOps) ReadCoreStatus(string) (map[uint32]device.CoreStatus, error) {
	return o.coreStatus, nil
}
func (o stubMgmtOps) ReadCoreFrequency(string) (map[uint32]uint32, error) {
	return o.coreFrequency, nil
}
func (o stubMgmtOps) ReadGovernorProfile(string) (device.GovernorProfile, error) {
	return o.governorProfile, nil
}
func (o stubMgmtOps) WriteGovernorProfile(string, device.GovernorProfile) error {
	return o.writeGovernorErr
}
func (o stubMgmtOps) ReadLiveness(string) bool { return o.liveness }
func (o stubMgmtOps) ReadPerformanceCounter(string, map[uint32]struct{}, map[uint32]device.CoreStatus) (map[uint32]device.PePerformanceCounter, error) {
	return o.performanceCounter, nil
}
func (o stubMgmtOps) ReadMemoryUsage(string, map[uint32]struct{}, map[uint32]device.CoreStatus, sysfs.MemoryConfig) (map[device.DeviceCoreRange]device.MemoryInfo, error) {
	return o.memoryUsage, nil
}

// ReadThrottleReason/WritePowerLimit exist only to satisfy device.MgmtOps;
// dedicated test coverage lands in a follow-up.
func (o stubMgmtOps) ReadThrottleReason(string) ([]device.ThrottleReason, error) { return nil, nil }
func (o stubMgmtOps) WritePowerLimit(string, uint32) error                       { return nil }

type stubDevfsOps struct {
	accessibleCores map[uint32]struct{}
	coreCount       uint32
}

func (o stubDevfsOps) GetDevNodeIndices(string) map[uint8]struct{} { return nil }
func (o stubDevfsOps) GetCoreCount(string, uint8) uint32           { return o.coreCount }
func (o stubDevfsOps) GetAccessibleCoreIndices(string, uint8) map[uint32]struct{} {
	return o.accessibleCores
}

// GetDeviceFiles/HasP2PInterfaces exist only to satisfy device.DevfsOps;
// dedicated test coverage lands in a follow-up.
func (o stubDevfsOps) GetDeviceFiles(string, uint8) []device.DeviceFile { return nil }
func (o stubDevfsOps) HasP2PInterfaces(string, uint8) bool              { return false }

type stubHwmonOps struct {
	temperature device.DeviceTemperature
	power       float64
}

func (o stubHwmonOps) ReadDeviceTemperature(string) (device.DeviceTemperature, error) {
	return o.temperature, nil
}
func (o stubHwmonOps) ReadPowerConsumption(string) (float64, error) { return o.power, nil }

// ReadPowerLimit exists only to satisfy device.HwmonOps; dedicated test
// coverage lands in a follow-up.
func (o stubHwmonOps) ReadPowerLimit(string) (uint32, error) { return 0, nil }

// testOps builds a working Ops bundle over the three stubs above.
func testOps(mgmtPath string, mgmt stubMgmtOps, devfs stubDevfsOps, hwmon stubHwmonOps) device.Ops {
	return device.Ops{
		MgmtPath:     mgmtPath,
		DevRoot:      "/dev/test",
		DevfsOps:     devfs,
		MgmtOps:      mgmt,
		HwmonOps:     hwmon,
		MemoryConfig: sysfs.MemoryConfig{DRAMTotalSize: 48 * 1024 * 1024 * 1024, FirmwareReservedSize: 512 * 1024 * 1024, PECount: 8},
	}
}

type DeviceTestSuite struct {
	suite.Suite
	testDevice device.Device
}

func (s *DeviceTestSuite) SetupSuite() {
	deviceBDF, err := bdf.ParseBDF("0000:27:00.0")
	s.Require().NoError(err)
	pciDevice := device.NewPCIDevice(deviceBDF, 0x0001, arch.FuriosaVendorID)
	info := device.NewDeviceInfo(3, arch.RNGD, "npu3", "test-uuid", "test-serial", 8, 2)
	s.testDevice = device.NewDevice(pciDevice, info,
		testOps("/mgmt/npu3", stubMgmtOps{}, stubDevfsOps{}, stubHwmonOps{}))
}

func (*DeviceTestSuite) TearDownSuite() {
}

// SetupTest points internal/sysroot at a fresh mock directory. Per that
// package's rule, no test in this suite (or its subtests) may call
// t.Parallel().
func (s *DeviceTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
}

func (*DeviceTestSuite) TearDownTest() {
}

func (s *DeviceTestSuite) TestDeviceAccessors() {
	s.Equal("0000:27:00.0", s.testDevice.BDF().String())
	s.Equal(uint16(0x0001), s.testDevice.DeviceID())
	s.Equal(arch.FuriosaVendorID, s.testDevice.VendorID())
	s.Equal("test-uuid", s.testDevice.UUID())
	s.Equal("test-serial", s.testDevice.Serial())
	s.Equal(uint32(8), s.testDevice.CoreNum())
	s.Equal(int32(2), s.testDevice.NUMANode())
	s.Equal(uint8(3), s.testDevice.DevNodeIndex())
	s.Equal(arch.RNGD, s.testDevice.Arch())
	s.Equal("npu3", s.testDevice.DeviceName())
}

func (s *DeviceTestSuite) TestProductName() {
	deviceBDF, err := bdf.ParseBDF("0000:27:00.0")
	s.Require().NoError(err)

	info := device.NewDeviceInfo(0, arch.RNGD, "npu0", "", "", 0, 0)

	rngd := device.NewDevice(device.NewPCIDevice(deviceBDF, arch.RNGDDeviceID, arch.FuriosaVendorID), info, device.Ops{})
	s.Equal("RNGD", rngd.ProductName())

	rngdPlus := device.NewDevice(device.NewPCIDevice(deviceBDF, arch.RNGDPlusDeviceID, arch.FuriosaVendorID), info, device.Ops{})
	s.Equal("RNGD+", rngdPlus.ProductName())

	unknown := device.NewDevice(device.NewPCIDevice(deviceBDF, 0x9999, arch.FuriosaVendorID), info, device.Ops{})
	s.Equal("Unknown", unknown.ProductName())
}

func (s *DeviceTestSuite) TestDeviceToDeviceLinkTypeRejectsNil() {
	linkType, err := s.testDevice.DeviceToDeviceLinkType(nil)
	s.Require().Error(err)
	s.Equal(topology.LinkTypeUnknown, linkType)

	var smiErr smierrors.SMIError
	s.Require().ErrorAs(err, &smiErr)
	s.Equal(smierrors.ErrorKindUnexpectedValue, smiErr.Kind())
}

func (s *DeviceTestSuite) TestDeviceToDeviceLinkTypeSameBDFIsNoC() {
	linkType, err := s.testDevice.DeviceToDeviceLinkType(&s.testDevice)
	s.Require().NoError(err)
	s.Equal(topology.LinkTypeNoC, linkType)
}

// TestZeroValueP2PAccessibleDoesNotPanic covers a zero-value Device, which
// has no DevfsOps to consult.
func (s *DeviceTestSuite) TestZeroValueP2PAccessibleDoesNotPanic() {
	var zero device.Device
	s.NotPanics(func() {
		s.False(zero.P2PAccessible(&zero))
		s.False(zero.P2PAccessible(&s.testDevice))
		s.False(s.testDevice.P2PAccessible(&zero))
	})
}

// newUnavailableDevice has no "driver" subdirectory under the mock PCI
// root, so sysfs.IsDriverBound reports it as not available.
func (s *DeviceTestSuite) newUnavailableDevice() device.Device {
	deviceBDF, err := bdf.ParseBDF("0000:99:00.0")
	s.Require().NoError(err)
	pciDevice := device.NewPCIDevice(deviceBDF, arch.RNGDDeviceID, arch.FuriosaVendorID)
	info := device.NewDeviceInfo(0, arch.RNGD, "npu0", "test-uuid", "test-serial", 8, 0)

	return device.NewDevice(pciDevice, info,
		testOps("/mgmt/npu0", stubMgmtOps{}, stubDevfsOps{}, stubHwmonOps{}))
}

// newAvailableDevice creates a "driver" subdirectory (so it's considered
// available), returning the device built over the given stubs.
func (s *DeviceTestSuite) newAvailableDevice(mgmt stubMgmtOps, devfs stubDevfsOps, hwmon stubHwmonOps) device.Device {
	deviceBDF, err := bdf.ParseBDF("0000:27:00.0")
	s.Require().NoError(err)
	pciDir := filepath.Join(sysroot.PCIDevicesRoot(), deviceBDF.String())
	s.Require().NoError(os.MkdirAll(filepath.Join(pciDir, "driver"), 0o755))

	pciDevice := device.NewPCIDevice(deviceBDF, arch.RNGDDeviceID, arch.FuriosaVendorID)
	info := device.NewDeviceInfo(0, arch.RNGD, "npu0", "test-uuid", "test-serial", 8, 0)

	return device.NewDevice(pciDevice, info,
		testOps("/mgmt/npu0", mgmt, devfs, hwmon))
}

func (s *DeviceTestSuite) requireDeviceNotAvailable(err error) {
	s.T().Helper()
	s.Require().Error(err)
	var smiErr smierrors.SMIError
	s.Require().ErrorAs(err, &smiErr)
	s.Equal(smierrors.ErrorKindDeviceNotAvailable, smiErr.Kind())
}

// TestUnavailableDeviceRejectsAllMgmtAndHwmonMethods covers ensureAvailable's
// gate: every mgmt/hwmon method must fail with ErrorKindDeviceNotAvailable
// when the device isn't driver-bound, and Liveness must report false rather
// than propagate an error (matching Rust's unwrap_or(false)).
func (s *DeviceTestSuite) TestUnavailableDeviceRejectsAllMgmtAndHwmonMethods() {
	d := s.newUnavailableDevice()

	cases := []struct {
		name   string
		invoke func() error
	}{

		{"FirmwareVersion", func() error { _, err := d.FirmwareVersion(); return err }},

		{"DriverVersion", func() error { _, err := device.DriverVersion(d); return err }},

		{"Major", func() error { _, err := d.Major(); return err }},

		{"Minor", func() error { _, err := d.Minor(); return err }},

		{"DeviceTemperature", func() error { _, err := d.DeviceTemperature(); return err }},

		{"PowerConsumption", func() error { _, err := d.PowerConsumption(); return err }},

		{"CoreStatus", func() error { _, err := d.CoreStatus(); return err }},

		{"CoreFrequency", func() error { _, err := d.CoreFrequency(); return err }},

		{"GovernorProfile", func() error { _, err := d.GovernorProfile(); return err }},
		{"SetGovernorProfile", func() error { return d.SetGovernorProfile(device.GovernorProfilePerformance) }},

		{"PerformanceCounter", func() error { _, err := d.PerformanceCounter(); return err }},

		{"MemoryUsage", func() error { _, err := d.MemoryUsage(); return err }},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			s.requireDeviceNotAvailable(c.invoke())
		})
	}

	s.False(d.Liveness())
}

func (s *DeviceTestSuite) TestFirmwareAndDriverVersionDelegateToMgmtOps() {
	firmwareVersion, err := version.ParseVersionInfo("1.0.0")
	s.Require().NoError(err)
	driverVersion, err := version.ParseVersionInfo("2025.1.0")
	s.Require().NoError(err)

	d := s.newAvailableDevice(stubMgmtOps{firmwareVersion: firmwareVersion, driverVersion: driverVersion}, stubDevfsOps{}, stubHwmonOps{})

	firmware, err := d.FirmwareVersion()
	s.Require().NoError(err)
	s.Equal("1.0.0", firmware.String())

	driver, err := device.DriverVersion(d)
	s.Require().NoError(err)
	s.Equal("2025.1.0", driver.String())
}

func (s *DeviceTestSuite) TestMajorMinorDelegateToMgmtOps() {
	d := s.newAvailableDevice(stubMgmtOps{major: 240, minor: 3}, stubDevfsOps{}, stubHwmonOps{})

	major, err := d.Major()
	s.Require().NoError(err)
	s.Equal(uint16(240), major)

	minor, err := d.Minor()
	s.Require().NoError(err)
	s.Equal(uint16(3), minor)
}

func (s *DeviceTestSuite) TestGovernorProfileRoundTrip() {
	d := s.newAvailableDevice(stubMgmtOps{governorProfile: device.GovernorProfilePowerSave}, stubDevfsOps{}, stubHwmonOps{})

	profile, err := d.GovernorProfile()
	s.Require().NoError(err)
	s.Equal(device.GovernorProfilePowerSave, profile)
}

// TestSetGovernorProfilePropagatesUnderlyingError confirms Device doesn't
// swallow or mask whatever error the architecture's WriteGovernorProfile
// reports (e.g. an out-of-range profile) -- that validation itself belongs
// to the architecture implementation, not Device.
func (s *DeviceTestSuite) TestSetGovernorProfilePropagatesUnderlyingError() {
	d := s.newAvailableDevice(stubMgmtOps{writeGovernorErr: smierrors.UnexpectedValue("invalid governor profile")}, stubDevfsOps{}, stubHwmonOps{})

	err := d.SetGovernorProfile(device.GovernorProfile(99))
	s.Require().Error(err)
	var smiErr smierrors.SMIError
	s.Require().ErrorAs(err, &smiErr)
	s.Equal(smierrors.ErrorKindUnexpectedValue, smiErr.Kind())
}

func (s *DeviceTestSuite) TestLivenessDelegatesToMgmtOps() {
	d := s.newAvailableDevice(stubMgmtOps{liveness: true}, stubDevfsOps{}, stubHwmonOps{})
	s.True(d.Liveness())
}

func (s *DeviceTestSuite) TestDeviceTemperatureAndPowerConsumptionFailWithoutHwmonDirectory() {
	d := s.newAvailableDevice(stubMgmtOps{}, stubDevfsOps{}, stubHwmonOps{})

	_, err := d.DeviceTemperature()
	s.Error(err)

	_, err = d.PowerConsumption()
	s.Error(err)
}

func (s *DeviceTestSuite) TestDeviceTemperatureAndPowerConsumptionDelegateToHwmonOps() {
	deviceBDF, err := bdf.ParseBDF("0000:27:00.0")
	s.Require().NoError(err)
	pciDir := filepath.Join(sysroot.PCIDevicesRoot(), deviceBDF.String())
	s.Require().NoError(os.MkdirAll(filepath.Join(pciDir, "driver"), 0o755))
	s.Require().NoError(os.MkdirAll(filepath.Join(pciDir, "hwmon", "hwmon0"), 0o755))

	pciDevice := device.NewPCIDevice(deviceBDF, arch.RNGDDeviceID, arch.FuriosaVendorID)
	info := device.NewDeviceInfo(0, arch.RNGD, "npu0", "test-uuid", "test-serial", 8, 0)
	d := device.NewDevice(pciDevice, info,
		testOps("/mgmt/npu0", stubMgmtOps{}, stubDevfsOps{}, stubHwmonOps{
			temperature: device.NewDeviceTemperature(45.5, 30.2),
			power:       12.5,
		}))

	temp, err := d.DeviceTemperature()
	s.Require().NoError(err)
	s.Equal(45.5, temp.SoCPeak())
	s.Equal(30.2, temp.Ambient())

	power, err := d.PowerConsumption()
	s.Require().NoError(err)
	s.Equal(12.5, power)
}

// TestCoreStatusAndCoreFrequencyFilterToAccessibleCoresOnly closes a gap:
// a naive implementation might return every core the mgmt node reports,
// rather than filtering to the cores actually visible in this environment.
func (s *DeviceTestSuite) TestCoreStatusAndCoreFrequencyFilterToAccessibleCoresOnly() {
	mgmt := stubMgmtOps{
		coreStatus: map[uint32]device.CoreStatus{
			0: device.CoreStatusAvailable, 1: device.CoreStatusOccupied,
			2: device.CoreStatusAvailable, 3: device.CoreStatusOccupied,
		},
		coreFrequency: map[uint32]uint32{0: 1500, 1: 1600, 2: 1700, 3: 1800},
	}
	devfs := stubDevfsOps{accessibleCores: map[uint32]struct{}{0: {}, 2: {}}}
	d := s.newAvailableDevice(mgmt, devfs, stubHwmonOps{})

	status, err := d.CoreStatus()
	s.Require().NoError(err)
	s.Len(status, 2)
	s.Contains(status, uint32(0))
	s.Contains(status, uint32(2))
	s.NotContains(status, uint32(1))
	s.NotContains(status, uint32(3))

	freq, err := d.CoreFrequency()
	s.Require().NoError(err)
	s.Len(freq, 2)
	s.Equal(uint32(1500), freq[0])
	s.Equal(uint32(1700), freq[2])
}

func (s *DeviceTestSuite) TestPerformanceCounterAndMemoryUsageDelegateToMgmtOps() {
	mgmt := stubMgmtOps{
		coreStatus: map[uint32]device.CoreStatus{0: device.CoreStatusOccupied},
		performanceCounter: map[uint32]device.PePerformanceCounter{
			0: device.NewPePerformanceCounter(time.Now(), 0x1000, 0x800),
		},
		memoryUsage: map[device.DeviceCoreRange]device.MemoryInfo{
			device.NewDeviceCoreRange(0, 0, 0): device.NewMemoryInfo(1024, 256),
		},
	}
	devfs := stubDevfsOps{accessibleCores: map[uint32]struct{}{0: {}}}
	d := s.newAvailableDevice(mgmt, devfs, stubHwmonOps{})

	counters, err := d.PerformanceCounter()
	s.Require().NoError(err)
	s.Require().Contains(counters, uint32(0))
	s.Equal(uint64(0x1000), counters[0].CycleCount())

	usage, err := d.MemoryUsage()
	s.Require().NoError(err)
	s.Require().Contains(usage, device.NewDeviceCoreRange(0, 0, 0))
	s.Equal(uint64(1024), usage[device.NewDeviceCoreRange(0, 0, 0)].TotalBytes())
}

func TestDeviceSuite(t *testing.T) {
	suite.Run(t, new(DeviceTestSuite))
}

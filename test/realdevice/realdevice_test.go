package realdevice_test

// Cross-checks pkg/api against the Rust furiosa-smi CLI (src/bin/furiosa-smi)
// on whatever real Furiosa hardware happens to be attached to the machine
// running this test. Skips cleanly (never fails) when there's no such
// hardware, the Rust workspace can't be found, or `cargo build` fails --
// this test only asserts something when it can compare against a real,
// independently-computed source of truth.
//
// Static fields (identity, firmware, governor, ...) are compared exactly.
// Live sensor/clock readings (temperature, power, core clock) are compared
// with generous tolerances, since the CLI reads them at a slightly
// different instant than this test does.

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/device"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

// Generous tolerances for values that can drift between the CLI's read and
// this test's read.
const (
	temperatureToleranceC = 1.0
	powerToleranceW       = 10.0
	coreClockToleranceMHz = 100.0
)

// cliInfoEntry mirrors furiosa-smi/src/bin/furiosa-smi/info.rs's
// NpuDeviceInfo JSON shape (`furiosa-smi info --format json`).
type cliInfoEntry struct {
	Arch        string `json:"arch"`
	DevName     string `json:"dev_name"`
	DeviceUUID  string `json:"device_uuid"`
	DeviceSN    string `json:"device_sn"`
	Firmware    string `json:"firmware"`
	Temperature string `json:"temperature"`
	Power       string `json:"power"`
	CoreClock   string `json:"core_clock"`
	Governor    string `json:"governor"`
	PCIBDF      string `json:"pci_bdf"`
	PCIDev      string `json:"pci_dev"`
}

// cliVersionEntry mirrors version.rs's Version JSON shape (`furiosa-smi
// version --format json`).
type cliVersionEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type RealDeviceTestSuite struct {
	suite.Suite

	cliInfo    []cliInfoEntry
	cliVersion []cliVersionEntry
	goDevices  []device.Device
}

func (*RealDeviceTestSuite) TearDownSuite() {
}

func (*RealDeviceTestSuite) SetupTest() {
}

func (*RealDeviceTestSuite) TearDownTest() {
}

// SetupSuite locates (building if necessary) the real furiosa-smi CLI,
// runs it once, and fetches the Go API's view of the same machine
// immediately afterward so both sides observe as close to the same instant
// as practical. Skips the whole suite if any of that isn't possible.
func (s *RealDeviceTestSuite) SetupSuite() {
	t := s.T()

	cliPath, err := locateCLI()
	if err != nil {
		t.Skipf("furiosa-smi CLI unavailable, skipping real-device comparison: %v", err)
	}

	t.Logf("using furiosa-smi CLI at %s", cliPath)

	info, err := runCLIJSON[[]cliInfoEntry](cliPath, "info", "--format", "json")
	if err != nil || len(info) == 0 {
		t.Skipf("no real Furiosa device found via furiosa-smi CLI, skipping real-device comparison: %v", err)
	}

	s.cliInfo = info
	t.Logf("furiosa-smi info --format json: %+v", info)

	// The driver-version comparison is best-effort: some setups intentionally
	// have no driver bound, in which case the CLI simply omits this entry.
	if versionEntries, err := runCLIJSON[[]cliVersionEntry](cliPath, "version", "--format", "json"); err == nil {
		s.cliVersion = versionEntries
		t.Logf("furiosa-smi version --format json: %+v", versionEntries)
	} else {
		t.Logf("furiosa-smi version --format json failed (driver-version comparison will be skipped): %v", err)
	}

	goDevices, err := api.ListDevices()
	s.Require().NoError(err)
	s.goDevices = goDevices
	t.Logf("api.ListDevices() returned %d device(s)", len(goDevices))
}

// locateCLI finds (building if necessary) the furiosa-smi CLI binary.
//
// FURIOSA_SMI_CLI_BIN, if set, points directly at a built binary.
// FURIOSA_SMI_RUST_DIR, if set, points at the Rust Cargo workspace root to
// build from. Otherwise this assumes the default repo layout: this Go
// module is a sibling of the Rust workspace, both under a common parent
// (.../furiosa-smi/{furiosa-smi-go, furiosa-smi}).
func locateCLI() (string, error) {
	if path := os.Getenv("FURIOSA_SMI_CLI_BIN"); path != "" {
		if _, err := os.Stat(path); err != nil {
			return "", fmt.Errorf("FURIOSA_SMI_CLI_BIN=%s: %w", path, err)
		}

		return path, nil
	}

	rustDir := os.Getenv("FURIOSA_SMI_RUST_DIR")
	if rustDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		rustDir = filepath.Join(wd, "..", "..", "..")
	}

	if _, err := os.Stat(filepath.Join(rustDir, "Cargo.toml")); err != nil {
		return "", fmt.Errorf("rust workspace not found at %s (set FURIOSA_SMI_RUST_DIR or FURIOSA_SMI_CLI_BIN): %w", rustDir, err)
	}

	buildCmd := exec.Command("cargo", "build", "--bin", "furiosa-smi")
	buildCmd.Dir = rustDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("cargo build --bin furiosa-smi failed: %w\n%s", err, output)
	}

	for _, profile := range []string{"debug", "release"} {
		candidate := filepath.Join(rustDir, "target", profile, "furiosa-smi")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("built furiosa-smi but binary not found under %s/target/{debug,release}", rustDir)
}

func runCLIJSON[T any](cliPath string, args ...string) (T, error) {
	var zero T
	output, err := exec.Command(cliPath, args...).Output()
	if err != nil {
		return zero, fmt.Errorf("running %s %v: %w", cliPath, args, err)
	}

	var result T
	if err := json.Unmarshal(output, &result); err != nil {
		return zero, fmt.Errorf("parsing JSON from %s %v: %w", cliPath, args, err)
	}

	return result, nil
}

func parseSuffixedFloat(s, suffix string) (float64, bool) {
	trimmed := strings.TrimSuffix(s, suffix)
	value, err := strconv.ParseFloat(strings.TrimSpace(trimmed), 64)

	return value, err == nil
}

func parsePCIDev(s string) (major, minor uint64, ok bool) {
	before, after, found := strings.Cut(s, ":")
	if !found {
		return 0, 0, false
	}

	major, err1 := strconv.ParseUint(before, 10, 16)
	minor, err2 := strconv.ParseUint(after, 10, 16)

	return major, minor, err1 == nil && err2 == nil
}

// cliEntryFor returns the CLI's info entry for the given BDF, requiring
// that it exists (a device the Go API found but the CLI didn't, or vice
// versa, is a real discrepancy worth failing on).
func (s *RealDeviceTestSuite) cliEntryFor(bdf string) cliInfoEntry {
	s.T().Helper()
	for _, entry := range s.cliInfo {
		if entry.PCIBDF == bdf {
			return entry
		}
	}

	s.FailNow(fmt.Sprintf("furiosa-smi CLI reported no device at BDF %s that the Go API found", bdf))

	return cliInfoEntry{}
}

func (s *RealDeviceTestSuite) TestDeviceCountMatches() {
	s.T().Logf("device count: cli=%d go=%d", len(s.cliInfo), len(s.goDevices))
	s.Len(s.goDevices, len(s.cliInfo))
}

func (s *RealDeviceTestSuite) TestIdentityFieldsMatchExactly() {
	for _, d := range s.goDevices {
		cli := s.cliEntryFor(d.BDF().String())
		s.T().Logf("[%s] arch: cli=%q go=%q", d.BDF(), cli.Arch, d.Arch().String())
		s.Equal(cli.Arch, d.Arch().String(), "arch for %s", d.BDF())
		s.T().Logf("[%s] device name: cli=%q go=%q", d.BDF(), cli.DevName, d.DeviceName())
		s.Equal(cli.DevName, d.DeviceName(), "device name for %s", d.BDF())
		s.T().Logf("[%s] uuid: cli=%q go=%q", d.BDF(), cli.DeviceUUID, d.UUID())
		s.Equal(cli.DeviceUUID, d.UUID(), "uuid for %s", d.BDF())
		s.T().Logf("[%s] serial: cli=%q go=%q", d.BDF(), cli.DeviceSN, d.Serial())
		s.Equal(cli.DeviceSN, d.Serial(), "serial for %s", d.BDF())

		wantMajor, wantMinor, ok := parsePCIDev(cli.PCIDev)
		major, majorErr := d.Major()
		minor, minorErr := d.Minor()
		if ok {
			s.T().Logf("[%s] major:minor: cli=%d:%d go=%d:%d (errs: major=%v minor=%v)",
				d.BDF(), wantMajor, wantMinor, major, minor, majorErr, minorErr)
			s.NoError(majorErr, "major for %s", d.BDF())
			s.NoError(minorErr, "minor for %s", d.BDF())
			s.Equal(wantMajor, uint64(major), "major for %s", d.BDF())
			s.Equal(wantMinor, uint64(minor), "minor for %s", d.BDF())
		}
	}
}

func (s *RealDeviceTestSuite) TestFirmwareVersionMatches() {
	for _, d := range s.goDevices {
		cli := s.cliEntryFor(d.BDF().String())
		firmwareVersion, err := d.FirmwareVersion()
		s.T().Logf("[%s] firmware version: cli=%q go=%q (err=%v)", d.BDF(), cli.Firmware, firmwareVersion.String(), err)
		if cli.Firmware == "N/A" {
			s.Error(err, "expected FirmwareVersion to fail for %s since the CLI reported N/A", d.BDF())
			continue
		}

		s.Require().NoError(err, "firmware version for %s", d.BDF())
		s.Equal(cli.Firmware, firmwareVersion.String(), "firmware version for %s", d.BDF())
	}
}

func (s *RealDeviceTestSuite) TestGovernorProfileMatches() {
	for _, d := range s.goDevices {
		cli := s.cliEntryFor(d.BDF().String())
		profile, err := d.GovernorProfile()
		s.T().Logf("[%s] governor profile: cli=%q go=%q (err=%v)", d.BDF(), cli.Governor, profile.String(), err)
		if cli.Governor == "N/A" {
			s.Error(err, "expected GovernorProfile to fail for %s since the CLI reported N/A", d.BDF())
			continue
		}

		s.Require().NoError(err, "governor profile for %s", d.BDF())
		s.Equal(cli.Governor, profile.String(), "governor profile for %s", d.BDF())
	}
}

// TestLiveReadingsAreClose compares temperature, power, and the first
// core's clock speed with generous tolerances, since these are live sensor
// reads that can drift between the CLI's snapshot and this test's.
func (s *RealDeviceTestSuite) TestLiveReadingsAreClose() {
	for _, d := range s.goDevices {
		cli := s.cliEntryFor(d.BDF().String())

		if wantTemp, ok := parseSuffixedFloat(cli.Temperature, "°C"); ok {
			temp, err := d.DeviceTemperature()
			s.Require().NoError(err, "device temperature for %s", d.BDF())
			s.T().Logf("[%s] soc peak temperature: cli=%.2f°C go=%.2f°C (tolerance=±%.1f)",
				d.BDF(), wantTemp, temp.SoCPeak(), temperatureToleranceC)
			s.InDelta(wantTemp, temp.SoCPeak(), temperatureToleranceC, "soc peak temperature for %s", d.BDF())
		} else {
			s.T().Logf("[%s] soc peak temperature: cli=%q (not comparable)", d.BDF(), cli.Temperature)
		}

		if wantPower, ok := parseSuffixedFloat(cli.Power, "W"); ok {
			power, err := d.PowerConsumption()
			s.Require().NoError(err, "power consumption for %s", d.BDF())
			s.T().Logf("[%s] power consumption: cli=%.2fW go=%.2fW (tolerance=±%.1f)",
				d.BDF(), wantPower, power, powerToleranceW)
			s.InDelta(wantPower, power, powerToleranceW, "power consumption for %s", d.BDF())
		} else {
			s.T().Logf("[%s] power consumption: cli=%q (not comparable)", d.BDF(), cli.Power)
		}

		if wantClock, ok := parseSuffixedFloat(cli.CoreClock, "MHz"); ok {
			freqs, err := d.CoreFrequency()
			s.Require().NoError(err, "core frequency for %s", d.BDF())
			s.Require().NotEmpty(freqs, "core frequency for %s", d.BDF())
			// The CLI only ever reports the lowest-indexed core.
			var firstCore uint32
			isFirst := true
			for core := range freqs {
				if isFirst || core < firstCore {
					firstCore = core
					isFirst = false
				}
			}

			s.T().Logf("[%s] core %d clock: cli=%.0fMHz go=%dMHz (tolerance=±%.0f) all cores=%v",
				d.BDF(), firstCore, wantClock, freqs[firstCore], coreClockToleranceMHz, freqs)
			s.InDelta(wantClock, float64(freqs[firstCore]), coreClockToleranceMHz, "core clock for %s", d.BDF())
		} else {
			s.T().Logf("[%s] core clock: cli=%q (not comparable)", d.BDF(), cli.CoreClock)
		}
	}
}

func (s *RealDeviceTestSuite) TestDriverInfoMatches() {
	var wantDriverVersion string
	found := false
	for _, entry := range s.cliVersion {
		if entry.Name == "device driver" {
			wantDriverVersion = entry.Version
			found = true
			break
		}
	}

	if !found {
		s.T().Skip("furiosa-smi version reported no device driver entry")
	}

	got, err := api.DriverInfo()
	s.T().Logf("driver version: cli=%q go=%q (err=%v)", wantDriverVersion, got.String(), err)
	s.Require().NoError(err)
	s.Equal(wantDriverVersion, got.String())
}

func TestRealDeviceSuite(t *testing.T) {
	suite.Run(t, new(RealDeviceTestSuite))
}

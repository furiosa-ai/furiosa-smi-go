package rngd

// furiosa-smi/src/sysfs/rngd/hwmon.rs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

const hwmonTestBDF = "0000:27:00.0"

type RNGDHwmonTestSuite struct {
	suite.Suite
}

func (*RNGDHwmonTestSuite) SetupSuite() {
}

func (*RNGDHwmonTestSuite) TearDownSuite() {
}

// SetupTest points internal/sysroot at a fresh mock directory, so the tests
// that read sensors off disk start from an empty sysroot. Per that
// package's rule, no test in this suite may call t.Parallel().
func (s *RNGDHwmonTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
}

func (*RNGDHwmonTestSuite) TearDownTest() {
}

// buildHwmonMock builds a driver-bound mock carrying hwmon and returns the
// hwmon directory, resolved the same way production code resolves it.
func (s *RNGDHwmonTestSuite) buildHwmonMock(hwmon *HWMonMock) string {
	s.T().Helper()
	sysroot.BuildMock(s.T(), sysroot.MockSpec{Devices: []sysroot.MockDevice{&RngdMock{
		BDF:    hwmonTestBDF,
		UUID:   "test-uuid",
		Serial: "test-serial",
		HWMon:  hwmon,
	}}})

	target, err := bdf.ParseBDF(hwmonTestBDF)
	s.Require().NoError(err)
	hwmonDir, err := sysfs.FindHwmonDir(target)
	s.Require().NoError(err)

	return hwmonDir
}

func (s *RNGDHwmonTestSuite) TestParsePowerLimitConvertsMicrowattsToMilliwatts() {
	got, err := parsePowerLimit("300000000")
	s.Require().NoError(err)
	s.Equal(uint32(300000), got)

	got, err = parsePowerLimit("45000000\n")
	s.Require().NoError(err)
	s.Equal(uint32(45000), got)

	got, err = parsePowerLimit("4294967295000")
	s.Require().NoError(err)
	s.Equal(uint32(4294967295), got)
}

func (s *RNGDHwmonTestSuite) TestParsePowerLimitRejectsInvalidValues() {
	_, err := parsePowerLimit("")
	s.Error(err)

	_, err = parsePowerLimit("bogus")
	s.Error(err)

	_, err = parsePowerLimit("-5")
	s.Error(err)

	_, err = parsePowerLimit("4294967296000")
	s.Error(err)
}

// TestReadDeviceTemperatureLooksSensorsUpByLabel covers the lookup rule
// that makes hwmon readable at all: sensor indices are not fixed, so PEAK
// and AMBIENT are found through their "_label" files rather than by
// assuming temp1 is the SoC and temp2 the ambient probe.
func (s *RNGDHwmonTestSuite) TestReadDeviceTemperatureLooksSensorsUpByLabel() {
	hwmonDir := s.buildHwmonMock(&HWMonMock{SoCPeakCelsius: 61.25, AmbientCelsius: 28.5})

	temperature, err := ReadDeviceTemperature(hwmonDir)
	s.Require().NoError(err)
	s.InDelta(61.25, temperature.SoCPeak(), 0.001)
	s.InDelta(28.5, temperature.Ambient(), 0.001)
}

func (s *RNGDHwmonTestSuite) TestReadDeviceTemperatureFailsWhenLabelMissing() {
	hwmonDir := s.buildHwmonMock(&HWMonMock{})
	s.Require().NoError(os.Remove(filepath.Join(hwmonDir, "temp2_label")))

	_, err := ReadDeviceTemperature(hwmonDir)
	s.Error(err)
}

func (s *RNGDHwmonTestSuite) TestReadPowerConsumptionConvertsMicrowattsToWatts() {
	hwmonDir := s.buildHwmonMock(&HWMonMock{PowerConsumption: 87.5})

	watts, err := ReadPowerConsumption(hwmonDir)
	s.Require().NoError(err)
	s.InDelta(87.5, watts, 0.001)
}

func (s *RNGDHwmonTestSuite) TestReadPowerLimitReadsPower1Max() {
	hwmonDir := s.buildHwmonMock(&HWMonMock{PowerLimitMW: 275_000})

	limit, err := ReadPowerLimit(hwmonDir)
	s.Require().NoError(err)
	s.Equal(uint32(275_000), limit)
}

func (s *RNGDHwmonTestSuite) TestReadersFailOnMissingHwmonDirectory() {
	missing := filepath.Join(s.T().TempDir(), "absent")

	_, err := ReadDeviceTemperature(missing)
	s.Error(err)

	_, err = ReadPowerConsumption(missing)
	s.Error(err)

	_, err = ReadPowerLimit(missing)
	s.Error(err)
}

func TestRNGDHwmonSuite(t *testing.T) {
	suite.Run(t, new(RNGDHwmonTestSuite))
}

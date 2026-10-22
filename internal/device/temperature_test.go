package device

// furiosa-smi/src/sysfs.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DeviceTemperatureTestSuite struct {
	suite.Suite
}

func (*DeviceTemperatureTestSuite) SetupSuite() {
}

func (*DeviceTemperatureTestSuite) TearDownSuite() {
}

func (*DeviceTemperatureTestSuite) SetupTest() {
}

func (*DeviceTemperatureTestSuite) TearDownTest() {
}

func (s *DeviceTemperatureTestSuite) TestDeviceTemperatureAccessors() {
	temp := NewDeviceTemperature(45.5, 30.2)
	s.Equal(45.5, temp.SoCPeak())
	s.Equal(30.2, temp.Ambient())
}

func TestDeviceTemperatureSuite(t *testing.T) {
	suite.Run(t, new(DeviceTemperatureTestSuite))
}

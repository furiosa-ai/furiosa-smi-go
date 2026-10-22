package device

// furiosa-smi/src/device/core_status.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CoreStatusTestSuite struct {
	suite.Suite
}

func (*CoreStatusTestSuite) SetupSuite() {
}

func (*CoreStatusTestSuite) TearDownSuite() {
}

func (*CoreStatusTestSuite) SetupTest() {
}

func (*CoreStatusTestSuite) TearDownTest() {
}

func (s *CoreStatusTestSuite) TestCoreStatusDisplay() {
	s.Equal("Available", CoreStatusAvailable.String())
	s.Equal("Occupied", CoreStatusOccupied.String())
}

func TestCoreStatusSuite(t *testing.T) {
	suite.Run(t, new(CoreStatusTestSuite))
}

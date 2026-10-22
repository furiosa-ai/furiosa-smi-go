package device

// furiosa-smi/src/device/governor.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GovernorProfileTestSuite struct {
	suite.Suite
}

func (*GovernorProfileTestSuite) SetupSuite() {
}

func (*GovernorProfileTestSuite) TearDownSuite() {
}

func (*GovernorProfileTestSuite) SetupTest() {
}

func (*GovernorProfileTestSuite) TearDownTest() {
}

func (s *GovernorProfileTestSuite) TestGovernorProfileDisplay() {
	s.Equal("Performance", GovernorProfilePerformance.String())
	s.Equal("PowerSave", GovernorProfilePowerSave.String())
}

func TestGovernorProfileSuite(t *testing.T) {
	suite.Run(t, new(GovernorProfileTestSuite))
}

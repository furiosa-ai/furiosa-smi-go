package arch

// furiosa-smi/src/arch.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ArchTestSuite struct {
	suite.Suite
}

func (*ArchTestSuite) SetupSuite() {
}

func (*ArchTestSuite) TearDownSuite() {
}

func (*ArchTestSuite) SetupTest() {
}

func (*ArchTestSuite) TearDownTest() {
}

func (s *ArchTestSuite) TestArchDisplay() {
	s.Equal("RNGD", RNGD.String())
}

func TestArchSuite(t *testing.T) {
	suite.Run(t, new(ArchTestSuite))
}

package config

// furiosa-smi/src/arch.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/arch"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (*ConfigTestSuite) SetupSuite() {
}

func (*ConfigTestSuite) TearDownSuite() {
}

func (*ConfigTestSuite) SetupTest() {
}

func (*ConfigTestSuite) TearDownTest() {
}

func (s *ConfigTestSuite) TestRetrieveArchConfigKnown() {
	got, err := RetrieveArchConfig(arch.RNGDDeviceID)
	s.Require().NoError(err)
	s.Equal(arch.RNGD, got.Arch)
	s.NotEmpty(got.DevRoot)
	s.Require().NotNil(got.Mgmt)
	s.NotEmpty(got.Mgmt.ClassRoot())
	s.Require().NotNil(got.MgmtOps)
	s.Require().NotNil(got.HwmonOps)
	s.Require().NotNil(got.DevfsOps)
	s.Equal(uint64(48*1024*1024*1024), got.MemoryConfig.DRAMTotalSize)

	gotPlus, err := RetrieveArchConfig(arch.RNGDPlusDeviceID)
	s.Require().NoError(err)
	s.Equal(arch.RNGD, gotPlus.Arch)
	s.Equal(uint64(72*1024*1024*1024), gotPlus.MemoryConfig.DRAMTotalSize)
}

func (s *ConfigTestSuite) TestRetrieveArchConfigUnknown() {
	_, err := RetrieveArchConfig(0x9999)
	s.Error(err)
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

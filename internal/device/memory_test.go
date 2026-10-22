package device

// furiosa-smi/src/device/memory.rs

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MemoryInfoTestSuite struct {
	suite.Suite
}

func (*MemoryInfoTestSuite) SetupSuite() {
}

func (*MemoryInfoTestSuite) TearDownSuite() {
}

func (*MemoryInfoTestSuite) SetupTest() {
}

func (*MemoryInfoTestSuite) TearDownTest() {
}

func (s *MemoryInfoTestSuite) TestMemoryInfoAccessors() {
	info := NewMemoryInfo(1024, 256)
	s.Equal(uint64(1024), info.TotalBytes())
	s.Equal(uint64(256), info.InUseBytes())
}

func TestMemoryInfoSuite(t *testing.T) {
	suite.Run(t, new(MemoryInfoTestSuite))
}

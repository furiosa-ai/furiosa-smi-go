package discovery

// furiosa-smi/src/discovery.rs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

type DiscoveryTestSuite struct {
	suite.Suite
}

func (*DiscoveryTestSuite) SetupSuite() {
}

func (*DiscoveryTestSuite) TearDownSuite() {
}

func (*DiscoveryTestSuite) SetupTest() {
}

func (*DiscoveryTestSuite) TearDownTest() {
}

func (s *DiscoveryTestSuite) TestIsAssignedChecksMembership() {
	nodes := map[uint16]map[uint8]struct{}{
		0x0001: {0: {}, 1: {}, 2: {}},
	}

	s.True(isAssigned(nodes, 0x0001, 0))
	s.True(isAssigned(nodes, 0x0001, 2))
	s.False(isAssigned(nodes, 0x0001, 5))
	s.False(isAssigned(nodes, 0x9999, 0))
}

func (s *DiscoveryTestSuite) TestEmptyDiscoveryReturnsEmptySlices() {
	sysroot.SetMock(s.T(), s.T().TempDir())
	s.Require().NoError(os.MkdirAll(sysroot.PCIDevicesRoot(), 0o755))

	pciDevices, err := ListPCIDevices()
	s.Require().NoError(err)
	s.NotNil(pciDevices)
	s.Empty(pciDevices)

	devices, err := ListDevices()
	s.Require().NoError(err)
	s.NotNil(devices)
	s.Empty(devices)
}

func TestDiscoverySuite(t *testing.T) {
	suite.Run(t, new(DiscoveryTestSuite))
}

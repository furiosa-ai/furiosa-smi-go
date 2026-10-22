package topology

// furiosa-smi/src/topology.rs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

type TopologyTestSuite struct {
	suite.Suite
}

func (*TopologyTestSuite) SetupSuite() {
}

func (*TopologyTestSuite) TearDownSuite() {
}

// SetupTest points internal/sysroot at a fresh mock directory. Per that
// package's rule, no test in this suite may call t.Parallel().
func (s *TopologyTestSuite) SetupTest() {
	sysroot.SetMock(s.T(), s.T().TempDir())
}

func (*TopologyTestSuite) TearDownTest() {
}

func (s *TopologyTestSuite) mustParseBDF(value string) bdf.BDF {
	s.T().Helper()
	parsed, err := bdf.ParseBDF(value)
	s.Require().NoError(err)

	return parsed
}

func (s *TopologyTestSuite) TestLinkTypeDisplay() {
	cases := []struct {
		linkType LinkType
		want     string
	}{
		{LinkTypeUnknown, "Unknown"},
		{LinkTypeInterconnect, "Interconnect"},
		{LinkTypeCPU, "Cpu"},
		{LinkTypeBridge, "Bridge"},
		{LinkTypeNoC, "Noc"},
	}
	for _, c := range cases {
		s.Equal(c.want, c.linkType.String())
	}
}

func (s *TopologyTestSuite) TestSameBDFIsNoC() {
	a := s.mustParseBDF("0000:18:00.0")
	got, err := GetLinkType(a, a)
	s.Require().NoError(err)
	s.Equal(LinkTypeNoC, got)
}

// buildPCIDevice lays out device under upstream the way real sysfs does
// (/sys/bus/pci/devices/<device> -> /sys/devices/pciDDDD:BB/<upstream>/<device>),
// local to cpu, which sits on CPU package pkg. This is the shape
// internal/arch/rngd's mock builds; it's spelled out here because that
// package depends on this one.
func (s *TopologyTestSuite) buildPCIDevice(device, upstream string, cpu, pkg uint32) {
	s.T().Helper()

	hostBridge := "pci0000:" + upstream[5:7]
	deviceDir := sysroot.Join(filepath.Join("sys/devices", hostBridge, upstream, device))
	sysroot.MustMkdirAll(s.T(), deviceDir)
	sysroot.MustMkdirAll(s.T(), sysroot.PCIDevicesRoot())
	s.Require().NoError(os.Symlink(deviceDir, filepath.Join(sysroot.PCIDevicesRoot(), device)))

	mask := make([]string, cpu/32+1)
	for i := range mask {
		mask[i] = "00000000"
	}
	mask[0] = fmt.Sprintf("%08x", uint32(1)<<(cpu%32))
	sysroot.MustWriteFile(s.T(), filepath.Join(deviceDir, "local_cpus"), strings.Join(mask, ",")+"\n")
	sysroot.MustWriteFile(s.T(),
		filepath.Join(sysroot.CPUDevicesRoot(), fmt.Sprintf("cpu%d", cpu), "topology", "physical_package_id"),
		fmt.Sprintf("%d\n", pkg))
}

func (s *TopologyTestSuite) TestLinkTypeFromSysfs() {
	// 27 and 2a share bridge 26:00.0 on socket 0; 51 is on socket 0 under
	// its own bridge; 9e is on socket 1.
	s.buildPCIDevice("0000:27:00.0", "0000:26:00.0", 0, 0)
	s.buildPCIDevice("0000:2a:00.0", "0000:26:00.0", 0, 0)
	s.buildPCIDevice("0000:51:00.0", "0000:50:00.0", 0, 0)
	s.buildPCIDevice("0000:9e:00.0", "0000:9d:00.0", 32, 1)

	for _, c := range []struct {
		name          string
		first, second string
		want          LinkType
	}{
		{"same bridge", "0000:27:00.0", "0000:2a:00.0", LinkTypeBridge},
		{"same socket", "0000:27:00.0", "0000:51:00.0", LinkTypeCPU},
		{"cross socket", "0000:27:00.0", "0000:9e:00.0", LinkTypeInterconnect},
		{"cross socket reversed", "0000:9e:00.0", "0000:51:00.0", LinkTypeInterconnect},
	} {
		s.Run(c.name, func() {
			got, err := GetLinkType(s.mustParseBDF(c.first), s.mustParseBDF(c.second))
			s.Require().NoError(err)
			s.Equal(c.want, got)
		})
	}
}

func (s *TopologyTestSuite) TestLinkTypeUnknownWithoutSocket() {
	s.buildPCIDevice("0000:27:00.0", "0000:26:00.0", 0, 0)
	s.buildPCIDevice("0000:51:00.0", "0000:50:00.0", 0, 0)
	s.Require().NoError(os.Remove(filepath.Join(sysroot.PCIDevicesRoot(), "0000:51:00.0", "local_cpus")))

	got, err := GetLinkType(s.mustParseBDF("0000:27:00.0"), s.mustParseBDF("0000:51:00.0"))
	s.Require().NoError(err)
	s.Equal(LinkTypeUnknown, got)
}

func (s *TopologyTestSuite) TestFirstCPUFromSingleGroup() {
	got := firstCPUFromCPUMask("0000ffff")
	s.Require().NotNil(got)
	s.Equal(uint32(0), *got)

	got = firstCPUFromCPUMask("00010000")
	s.Require().NotNil(got)
	s.Equal(uint32(16), *got)
}

func (s *TopologyTestSuite) TestFirstCPUFromMultiGroup() {
	got := firstCPUFromCPUMask("00000000,0000ffff")
	s.Require().NotNil(got)
	s.Equal(uint32(0), *got)

	got = firstCPUFromCPUMask("00000001,00000000")
	s.Require().NotNil(got)
	s.Equal(uint32(32), *got)
}

func (s *TopologyTestSuite) TestFirstCPUFromZeroMask() {
	s.Nil(firstCPUFromCPUMask("00000000"))
	s.Nil(firstCPUFromCPUMask("00000000,00000000"))
}

func (s *TopologyTestSuite) TestComponentAsBDFExtractsPCIDevices() {
	const real = "/sys/devices/pci0000:16/0000:16:00.0/0000:17:00.0/0000:18:00.0"
	var got []bdf.BDF
	for component := range strings.SplitSeq(real, "/") {
		if value := componentAsBDF(component); value != nil {
			got = append(got, *value)
		}
	}

	s.Require().Len(got, 3)
	s.Equal("0000:16:00.0", got[0].String())
	s.Equal("0000:17:00.0", got[1].String())
	s.Equal("0000:18:00.0", got[2].String())
}

func (s *TopologyTestSuite) TestComponentAsBDFRejectsHostBridge() {
	s.Nil(componentAsBDF("pci0000:16"))
}

func TestTopologySuite(t *testing.T) {
	suite.Run(t, new(TopologyTestSuite))
}

package apitest_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api/apitest"
)

// StaticRNGDTestSuite pins the static fixture consumers test against. Per
// apitest's rule, no test in this suite may call t.Parallel().
type StaticRNGDTestSuite struct {
	suite.Suite
	devices []api.Device
}

func (*StaticRNGDTestSuite) SetupSuite() {}

func (*StaticRNGDTestSuite) TearDownSuite() {}

func (s *StaticRNGDTestSuite) SetupTest() {
	s.devices = apitest.StaticRNGDDevices(s.T())
}

func (*StaticRNGDTestSuite) TearDownTest() {}

func (s *StaticRNGDTestSuite) TestIdentity() {
	bdfs := []string{
		"0000:27:00.0", "0000:2a:00.0", "0000:51:00.0", "0000:57:00.0",
		"0000:9e:00.0", "0000:a4:00.0", "0000:c7:00.0", "0000:ca:00.0",
	}

	s.Require().Len(s.devices, apitest.StaticRNGDDeviceCount)
	for i := range s.devices {
		d := &s.devices[i]
		s.Equal(api.ArchRNGD, d.Arch())
		s.Equal(bdfs[i], d.BDF().String())
		s.Equal(uint8(i), d.DevNodeIndex())
		s.Equal(fmt.Sprintf("npu%d", i), d.DeviceName())
		s.Equal(int32(i/4), d.NUMANode())
		s.Equal(fmt.Sprintf("A76AAD68-6855-40B1-9E86-D080852D1C8%d", i), d.UUID())
		s.Equal(fmt.Sprintf("TEST0236FH505KRE%d", i), d.Serial())
		s.Equal(uint32(8), d.CoreNum())
		s.True(d.Liveness())

		firmware, err := d.FirmwareVersion()
		s.Require().NoError(err)
		s.Equal("1.6.0(dev0), c1bebfd", firmware.String())

		major, err := d.Major()
		s.Require().NoError(err)
		s.Equal(uint16(234+i), major)

		minor, err := d.Minor()
		s.Require().NoError(err)
		s.Equal(uint16(0), minor)

		temperature, err := d.DeviceTemperature()
		s.Require().NoError(err)
		s.InDelta(20.0, temperature.SoCPeak(), 1e-9)

		power, err := d.PowerConsumption()
		s.Require().NoError(err)
		s.InDelta(100.0, power, 1e-9)
	}
}

func (s *StaticRNGDTestSuite) TestDeviceFiles() {
	suffixes := []string{
		"pe0", "pe0-1", "pe0-3", "pe1", "pe2", "pe2-3", "pe3",
		"pe4", "pe4-5", "pe4-7", "pe5", "pe6", "pe6-7", "pe7",
	}

	for i := range s.devices {
		files, err := s.devices[i].DeviceFiles()
		s.Require().NoError(err)

		paths := make([]string, len(files))
		for j, file := range files {
			paths[j] = file.Path()
		}

		want := make([]string, len(suffixes))
		for j, suffix := range suffixes {
			want[j] = fmt.Sprintf("/dev/rngd/npu%d%s", i, suffix)
		}

		s.Equal(want, paths, "npu%d", i)
	}
}

func (s *StaticRNGDTestSuite) TestLinkTypeMatrix() {
	// Devices 2k and 2k+1 share a bridge; 0-3 and 4-7 share a socket.
	want := func(i, j int) api.LinkType {
		switch {
		case i == j:
			return api.LinkTypeNoC
		case i/2 == j/2:
			return api.LinkTypeBridge
		case i/4 == j/4:
			return api.LinkTypeCPU
		default:
			return api.LinkTypeInterconnect
		}
	}

	for i := range s.devices {
		for j := range s.devices {
			got, err := s.devices[i].DeviceToDeviceLinkType(&s.devices[j])
			s.Require().NoError(err)
			s.Equal(want(i, j), got, "npu%d -> npu%d", i, j)
		}
	}
}

func (s *StaticRNGDTestSuite) TestP2PAccessible() {
	for i := range s.devices {
		for j := range s.devices {
			s.True(s.devices[i].P2PAccessible(&s.devices[j]), "npu%d -> npu%d", i, j)
		}
	}
}

func (s *StaticRNGDTestSuite) TestStaticRNGDDevice() {
	d := apitest.StaticRNGDDevice(s.T(), 5)
	s.Equal("npu5", d.DeviceName())
	s.Equal("0000:a4:00.0", d.BDF().String())
}

func TestStaticRNGDSuite(t *testing.T) {
	suite.Run(t, new(StaticRNGDTestSuite))
}

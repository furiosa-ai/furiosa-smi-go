package smierrors

// furiosa-smi/src/error.rs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SMIErrorTestSuite struct {
	suite.Suite
}

func (*SMIErrorTestSuite) SetupSuite() {
}

func (*SMIErrorTestSuite) TearDownSuite() {
}

func (*SMIErrorTestSuite) SetupTest() {
}

func (*SMIErrorTestSuite) TearDownTest() {
}

func (s *SMIErrorTestSuite) TestErrorDisplayFormat() {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"unknown arch", UnknownArch(0x0001), "unknown architecture: device_id=0x0001"},
		{"parse", Parse("bad value"), "parse error: bad value"},
		{"io", IO("/sys/foo", errors.New("not found")), "io error: not found (/sys/foo)"},
		{"device not available", DeviceNotAvailable("0000:27:00.0"), "device not available: 0000:27:00.0"},
		{"device not found", DeviceNotFound("npu0"), "device not found: npu0"},
		{"unexpected value", UnexpectedValue("bad fusion group"), "unexpected value: bad fusion group"},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			s.Require().Error(c.err)
			s.Equal(c.want, c.err.Error())
		})
	}
}

func (s *SMIErrorTestSuite) TestIOErrorSupportsErrorAsAndUnwrap() {
	cause := errors.New("permission denied")
	err := IO("/sys/example", cause)
	s.Require().Error(err)

	var smiErr SMIError
	s.Require().ErrorAs(err, &smiErr)
	s.Equal(ErrorKindIO, smiErr.Kind())
	s.Equal("io error: permission denied (/sys/example)", err.Error())
	s.ErrorIs(err, cause)
}

func TestSMIErrorSuite(t *testing.T) {
	suite.Run(t, new(SMIErrorTestSuite))
}

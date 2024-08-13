package smi

import "github.com/furiosa-ai/furiosa-smi-go/pkg/smi/binding"

type DeviceInfo interface {
	Arch() Arch
	CoreNum() uint32
	NumaNode() uint32
	Name() string
	Serial() string
	UUID() string
	BDF() string
	Major() uint16
	Minor() uint16
	FirmwareVersion() VersionInfo
	DriverVersion() VersionInfo
}

var _ DeviceInfo = new(deviceInfo)

type deviceInfo struct {
	raw binding.FuriosaSmiDeviceInfo
}

func newDeviceInfo(raw binding.FuriosaSmiDeviceInfo) DeviceInfo {
	return &deviceInfo{
		raw: raw,
	}
}

func (d *deviceInfo) Arch() Arch {
	return Arch(d.raw.Arch)
}

func (d *deviceInfo) CoreNum() uint32 {
	return d.raw.CoreNum
}

func (d *deviceInfo) NumaNode() uint32 {
	return d.raw.NumaNode
}

func (d *deviceInfo) Name() string {
	return byteBufferToString(d.raw.Name[:])
}

func (d *deviceInfo) Serial() string {
	return byteBufferToString(d.raw.Serial[:])
}

func (d *deviceInfo) UUID() string {
	return byteBufferToString(d.raw.Uuid[:])
}

func (d *deviceInfo) BDF() string {
	return byteBufferToString(d.raw.Bdf[:])
}

func (d *deviceInfo) Major() uint16 {
	return d.raw.Major
}

func (d *deviceInfo) Minor() uint16 {
	return d.raw.Minor
}

func (d *deviceInfo) FirmwareVersion() VersionInfo {
	// TODO: 기존 `device_info.go` 에는 DriverVersion 으로 명시되어 있음. 의도된 것인지 확인 필요.
	return newVersionInfo(d.raw.FirmwareVersion)
}

func (d *deviceInfo) DriverVersion() VersionInfo {
	return newVersionInfo(d.raw.DriverVersion)
}

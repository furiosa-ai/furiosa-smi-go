package device

// furiosa-smi/src/device/device_file.rs

// DeviceFile is one devfs entry for a device: a host path together with
// the core range it exposes (e.g. "/dev/rngd/npu0pe0-3" -> cores 0-3).
type DeviceFile struct {
	coreRange DeviceCoreRange
	path      string
}

func NewDeviceFile(coreRange DeviceCoreRange, path string) DeviceFile {
	return DeviceFile{coreRange: coreRange, path: path}
}

// CoreRange returns the core range designated by the device file.
func (f DeviceFile) CoreRange() DeviceCoreRange {
	return f.coreRange
}

// Cores returns the cores designated by the device file.
func (f DeviceFile) Cores() []uint32 {
	return f.coreRange.Cores()
}

// Path returns the device file path.
func (f DeviceFile) Path() string {
	return f.path
}

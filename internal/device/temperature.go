package device

// furiosa-smi/src/sysfs.rs

// DeviceTemperature is the temperature information of a device.
type DeviceTemperature struct {
	socPeak float64
	ambient float64
}

func NewDeviceTemperature(socPeak, ambient float64) DeviceTemperature {
	return DeviceTemperature{socPeak: socPeak, ambient: ambient}
}

func (t DeviceTemperature) SoCPeak() float64 {
	return t.socPeak
}

func (t DeviceTemperature) Ambient() float64 {
	return t.ambient
}

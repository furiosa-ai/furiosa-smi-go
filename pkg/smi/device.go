package smi

// ListDevices lists all Furiosa NPU devices in the system.
func ListDevices() ([]Device, error) {
	//logic here

	return nil, nil
}

// ListDisabledDevices lists all disabled Furiosa NPU devices in the system. It returns a list of BDF strings representing the disabled devices.
func ListDisabledDevices() ([]string, error) {
	//logic here

	return nil, nil
}

// EnableDevice enables a Furiosa NPU device by bdf. This requires root privileges.
func EnableDevice(bdf string) error {
	//logic here

	return nil
}

// DisableDevice disables a Furiosa NPU device by bdf. This requires root privileges.
func DisableDevice(bdf string) error {
	//logic here

	return nil
}

// DriverInfo return a driver information of the device.
func DriverInfo() (VersionInfo, error) {
	//logic here
	return nil, nil
}

func CreateObserverWithOpt(opt ObserverOpt) (Observer, error) {
	return newObserverWithOpt(opt)
}

func CreateDefaultObserver() (Observer, error) {
	opt, err := NewOptForObserver()
	if err != nil {
		return nil, err
	}

	return newObserverWithOpt(opt)
}

// Device represents the abstraction for a single Furiosa NPU device.
type Device interface {
	// DeviceInfo returns `DeviceInfo` which contains information about NPU device. (e.g. arch, serial, ...)
	DeviceInfo() (DeviceInfo, error)
	// DeviceFiles list device files under this device.
	DeviceFiles() ([]DeviceFile, error)
	// CoreStatus examine each core of the device, whether it is occupied or available.
	CoreStatus() (CoreStatuses, error)
	// Liveness returns a liveness state of the device.
	Liveness() (bool, error)
	// CoreFrequency returns a core frequency (MHz) of the device.
	CoreFrequency() (CoreFrequency, error)
	// MemoryFrequency returns a memory frequency (MHz) of the device.
	MemoryFrequency() (MemoryFrequency, error)
	// PowerConsumption returns a power consumption of the device.
	PowerConsumption() (float64, error)
	// DeviceTemperature returns a temperature of the device.
	DeviceTemperature() (DeviceTemperature, error)
	// DeviceToDeviceLinkType returns a device link type between two devices.
	DeviceToDeviceLinkType(target Device) (LinkType, error)
	// P2PAccessible returns whether two devices are p2p accessible each other or not.
	P2PAccessible(target Device) (bool, error)
	// DevicePerformanceCounter returns a performance counter of the device.
	DevicePerformanceCounter() (DevicePerformanceCounter, error)
	// GovernorProfile returns a governor profile of the device.
	GovernorProfile() (GovernorProfile, error)
	// SetGovernorProfile set a governor profile of the device.
	SetGovernorProfile(governorProfile GovernorProfile) error
	// PcieInfo returns a PCIe information of the device.
	PcieInfo() (PcieInfo, error)
	// ThrottleReason returns a throttle reason of the device.
	ThrottleReason() (ThrottleReason, error)
	// MemoryUtilization returns a memory utilization of the device.
	MemoryUtilization() (MemoryUtilization, error)
}

var _ Device = new(FuriosaSmiDevice)

type FuriosaSmiDevice struct {
	info FuriosaSmiDeviceInfo
}

/*
func newDevice(handle impl.FuriosaSmiDeviceHandle) Device {
	return &device{
		handle: handle,
	}
}*/

func (d *FuriosaSmiDevice) DeviceInfo() (DeviceInfo, error) {
	return &d.info, nil
}

func (d *FuriosaSmiDevice) DeviceFiles() ([]DeviceFile, error) {
	out, err := FuriosaSmiGetDeviceFiles(d)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return out.DeviceFiles(), nil
}

func (d *FuriosaSmiDevice) CoreStatus() (CoreStatuses, error) {
	//logic here
	return nil, nil
}
func (d *FuriosaSmiDevice) Liveness() (bool, error) {
	//logic here
	return false, nil
}

func (d *FuriosaSmiDevice) CoreFrequency() (CoreFrequency, error) {
	//logic here
	return nil, nil
}

func (d *FuriosaSmiDevice) MemoryFrequency() (MemoryFrequency, error) {
	//logic here

	return nil, nil
}

func (d *FuriosaSmiDevice) PowerConsumption() (float64, error) {
	//logic here

	return 0, nil
}

func (d *FuriosaSmiDevice) DeviceTemperature() (DeviceTemperature, error) {
	//logic here

	return nil, nil
}

func (d *FuriosaSmiDevice) DeviceToDeviceLinkType(target Device) (LinkType, error) {
	//logic here

	return FuriosaSmiLinkTypeUnknown, nil
}

func (d *FuriosaSmiDevice) P2PAccessible(target Device) (bool, error) {
	//logic here

	return false, nil
}

func (d *FuriosaSmiDevice) DevicePerformanceCounter() (DevicePerformanceCounter, error) {
	//logic here
	return nil, nil
}

func (d *FuriosaSmiDevice) GovernorProfile() (GovernorProfile, error) {
	//logic here
	return 0, nil
}

func (d *FuriosaSmiDevice) SetGovernorProfile(profile GovernorProfile) error {
	//logic here
	return nil
}

func (d *FuriosaSmiDevice) PcieInfo() (PcieInfo, error) {
	var outPcieDeviceInfo FuriosaSmiPcieDeviceInfo
	var outPcieLinkInfo FuriosaSmiPcieLinkInfo
	var outSriovInfo FuriosaSmiSriovInfo
	var outPcieRootComplexInfo FuriosaSmiPcieRootComplexInfo
	var outPcieSwitchInfo FuriosaSmiPcieSwitchInfo

	if err := FuriosaSmiGetPcieDeviceInfo(d, &outPcieDeviceInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetPcieLinkInfo(d, &outPcieLinkInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetSriovInfo(d, &outSriovInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetPcieRootComplexInfo(d, &outPcieRootComplexInfo); err != nil {
		return nil, err
	}

	if err := FuriosaSmiGetPcieSwitchInfo(d, &outPcieSwitchInfo); err != nil {
		return nil, err
	}

	return newPcieInfo(&outPcieDeviceInfo, &outPcieLinkInfo, &outSriovInfo, &outPcieRootComplexInfo, &outPcieSwitchInfo), nil
}

func (d *FuriosaSmiDevice) ThrottleReason() (ThrottleReason, error) {
	/*var out impl.FuriosaSmiThrottleReason

	if err := impl.FuriosaSmiGetThrottleReason(d.handle, &out); err != nil {
		return 0, err
	}*/

	return 0, nil
}

func (d *FuriosaSmiDevice) MemoryUtilization() (MemoryUtilization, error) {
	/*var out impl.FuriosaSmiMemoryUtilization

	if err := impl.FuriosaSmiGetMemoryUtilization(d.handle, &out); err != nil {
		return nil, err
	}*/

	return nil, nil
}

func FuriosaSmiGetDeviceInfo(device Device, outDeviceInfo *FuriosaSmiDeviceInfo) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetDeviceCoreStatus(device Device, outCoreStatus *FuriosaSmiCoreStatuses) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetDeviceLiveness(device Device, outLiveness *bool) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetDeviceToDeviceLinkType(device1 Device, device2 Device, outLinkType *LinkType) error {
	// TODO: Implement this function
	return nil
}

func FuriosaSmiGetP2pAccessible(device1 Device, device2 Device, outAccessible *bool) error {
	// TODO: Implement this function
	return nil
}

package main

import (
	"fmt"
	"os"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

func main() {
	if err := smi.Init(); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	disabledDevices, err := smi.ListDisabledDevices()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("There are %d disabled device(s).\n", len(disabledDevices))

	devices, err := smi.ListDevices()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	for _, disabledDevice := range disabledDevices {
		fmt.Printf("  Disabled Device BDF: %s\n", disabledDevice)
	}

	fmt.Printf("found %d device(s)\n", len(devices))

	driverInfo, err := smi.DriverInfo()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("driver info: %s\n", driverInfo)

	obs, err := smi.CreateDefaultObserver()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	for _, device := range devices {
		deviceInfo, err := device.DeviceInfo()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Device Arch: %v\n", deviceInfo.Arch())
		fmt.Printf("Device CoreNum: %d\n", deviceInfo.CoreNum())
		fmt.Printf("Device NumaNode: %d\n", deviceInfo.NumaNode())
		fmt.Printf("Device Name: %s\n", deviceInfo.Name())
		fmt.Printf("Device Serial: %s\n", deviceInfo.Serial())
		fmt.Printf("Device UUID: %s\n", deviceInfo.UUID())
		fmt.Printf("Device BDF: %s\n", deviceInfo.BDF())
		fmt.Printf("Device Major: %d\n", deviceInfo.Major())
		fmt.Printf("Device Minor: %d\n", deviceInfo.Minor())
		fmt.Printf("Device FirmwareVersion")
		fmt.Printf("  Major: %d\n", deviceInfo.FirmwareVersion().Major())
		fmt.Printf("  Minor: %d\n", deviceInfo.FirmwareVersion().Minor())
		fmt.Printf("  Patch: %d\n", deviceInfo.FirmwareVersion().Patch())
		fmt.Printf("  Meta: %s\n", deviceInfo.FirmwareVersion().Metadata())

		liveness, err := device.Liveness()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Liveness: %v\n", liveness)

		coreStatus, err := device.CoreStatus()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Core Status:\n")
		for _, peStatus := range coreStatus.PeStatus() {
			fmt.Printf("  Core %d: %v\n", peStatus.Core(), peStatus.Status())
		}

		deviceFiles, err := device.DeviceFiles()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Device Files:\n")
		for _, deviceFile := range deviceFiles {
			fmt.Printf("  Cores: %v\n", deviceFile.Cores())
			fmt.Printf("  Path: %s\n", deviceFile.Path())
		}

		performanceCounter, err := device.DevicePerformanceCounter()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Device Performance Counter:\n")
		for _, counter := range performanceCounter.PerformanceCounter() {
			fmt.Printf("  Core %d Performance Counter:\n", counter.Core())
			fmt.Printf("    Timestamp: %v\n", counter.Timestamp())
			fmt.Printf("    Cycle Count: %d\n", counter.CycleCount())
			fmt.Printf("    Task Execution Cycle: %d\n", counter.TaskExecutionCycle())
		}

		memoryFrequency, err := device.MemoryFrequency()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Memory Frequency: %d MHz\n", memoryFrequency.Frequency())

		coreFrequency, err := device.CoreFrequency()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Core Frequency:\n")
		for _, frequency := range coreFrequency.PeFrequency() {
			fmt.Printf("  Core %d Frequency: %d MHz\n", frequency.Core(), frequency.Frequency())
		}

		temperature, err := device.DeviceTemperature()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Device Temperature:\n")
		fmt.Printf("  Soc Peak: %f\n", temperature.SocPeak())
		fmt.Printf("  Ambient: %f\n", temperature.Ambient())

		powerConsumption, err := device.PowerConsumption()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Power Consumption: %f\n", powerConsumption)

		governor, err := device.GovernorProfile()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Governor Profile: %s\n", governor)

		pcieInfo, err := device.PcieInfo()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		pcieDeviceInfo := pcieInfo.DeviceInfo()
		fmt.Printf("PCIe Device Info\n")
		fmt.Printf("  Device id: 0x%04x\n", pcieDeviceInfo.DeviceId())
		fmt.Printf("  Device subsystem vendor id: 0x%04x\n", pcieDeviceInfo.VendorId())
		fmt.Printf("  Device subsystem device id: 0x%04x\n", pcieDeviceInfo.SubsystemId())
		fmt.Printf("  Device revision id: 0x%02x\n", pcieDeviceInfo.RevisionId())
		fmt.Printf("  Device class id: 0x%02x\n", pcieDeviceInfo.ClassId())
		fmt.Printf("  Device sub class id: 0x%02x\n", pcieDeviceInfo.SubClassId())

		pcieLinkInfo := pcieInfo.LinkInfo()
		fmt.Printf("PCIe Link Info\n")
		fmt.Printf("  Device pcie gen status: %d\n", pcieLinkInfo.PcieGenStatus())
		fmt.Printf("  Device pcie link width status: %d\n", pcieLinkInfo.LinkWidthStatus())
		fmt.Printf("  Device pcie link speed: %.2f\n", pcieLinkInfo.LinkSpeedStatus())
		fmt.Printf("  Device pcie max link width capability: %d\n", pcieLinkInfo.MaxLinkWidthCapability())
		fmt.Printf("  Device pcie max link speed capability: %.2f\n", pcieLinkInfo.MaxLinkSpeedCapability())

		sriovInfo := pcieInfo.SriovInfo()
		fmt.Printf("SR-IOV Info\n")
		fmt.Printf("  Device sr-iov total vfs: %d\n", sriovInfo.SriovTotalVfs())
		fmt.Printf("  Device sr-iov enabled: %d\n", sriovInfo.SriovEnabledVfs())

		pcieRootComplexInfo := pcieInfo.RootComplexInfo()
		fmt.Printf("PCIe Root Complex Info\n")
		fmt.Printf("  Device root complex: %s\n", pcieRootComplexInfo.String())

		pcieSwitchInfo := pcieInfo.SwitchInfo()
		if pcieSwitchInfo == nil {
			fmt.Printf("  Device pcie switch: Not available\n")
		} else {
			fmt.Printf("  Device pcie switch: %s\n", pcieSwitchInfo.String())
		}

		utilization, err := obs.GetCoreUtilization(device)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Core Utilization:\n")
		for _, peUtilization := range utilization {
			fmt.Printf("  Core %d: %.2f%%\n", peUtilization.Core, peUtilization.PeUsagePercentage)
		}

		throttleReason, err := device.ThrottleReason()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Throttle Reason: 0x%08x\n", throttleReason)
	}
}

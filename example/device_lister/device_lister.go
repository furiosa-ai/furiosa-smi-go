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

	devices, err := smi.ListDevices()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("found %d device(s)\n", len(devices))

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
		for core, status := range coreStatus {
			fmt.Printf("  Core %d: %v\n", core, status)
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

		coreUtilization, err := device.CoreUtilization()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Core Utilization:\n")
		for _, peUtilization := range coreUtilization.PeUtilization() {
			fmt.Printf("  PE Utilization:\n")
			fmt.Printf("    Core: %v\n", peUtilization.Core())
			fmt.Printf("    Time Window Mill: %d\n", peUtilization.TimeWindowMill())
			fmt.Printf("    PE Usage Percentage: %f\n", peUtilization.PeUsagePercentage())
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
	}
}

// Command list_devices prints every Furiosa PCI device and then the full
// details of each available NPU device. It is a port of the Rust crate's
// examples/list_devices.rs. Run it on a host with Furiosa NPUs:
//
//	go run ./example/list_devices
package main

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

func main() {
	pciDevices, err := api.ListPCIDevices()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Found %d PCI device(s)\n", len(pciDevices))
	for _, pci := range pciDevices {
		fmt.Printf("  BDF: %v, Device ID: 0x%04x, Vendor ID: 0x%04x\n", pci.BDF(), pci.DeviceID(), pci.VendorID())
	}

	fmt.Println()

	devices, err := api.ListDevices()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Found %d available device(s)\n\n", len(devices))

	for i := range devices {
		if err := printDevice(&devices[i]); err != nil {
			fmt.Printf("BDF: %v (error: %v)\n\n", devices[i].BDF(), err)
		}
	}
}

func printDevice(d *api.Device) error {
	product := d.ProductName()
	name := d.DeviceName()
	bdf := d.BDF()
	arch := d.Arch()
	uuid := d.UUID()
	serial := d.Serial()
	coreNum := d.CoreNum()
	numaNode := d.NUMANode()

	major, err := d.Major()
	if err != nil {
		return err
	}

	minor, err := d.Minor()
	if err != nil {
		return err
	}

	fwVersion, err := d.FirmwareVersion()
	if err != nil {
		return err
	}

	temp, err := d.DeviceTemperature()
	if err != nil {
		return err
	}

	coreStatuses, err := d.CoreStatus()
	if err != nil {
		return err
	}

	pcie, err := d.PcieInfo()
	if err != nil {
		return err
	}

	deviceInfo := pcie.DeviceInfo()
	linkInfo := pcie.LinkInfo()
	sriovInfo := pcie.SriovInfo()
	rootComplexInfo := pcie.RootComplexInfo()
	switchBDF := "N/A"
	if switchInfo := pcie.SwitchInfo(); switchInfo != nil {
		switchBDF = switchInfo.String()
	}

	fmt.Printf("%s (%v) - %s\n", name, bdf, product)
	fmt.Printf("  Arch:       %v\n", arch)
	fmt.Printf("  UUID:       %s\n", uuid)
	fmt.Printf("  Serial:     %s\n", serial)
	fmt.Printf("  Cores:      %d\n", coreNum)
	fmt.Printf("  NUMA node:  %d\n", numaNode)
	fmt.Printf("  Dev:        %d:%d\n", major, minor)
	fmt.Printf("  FW version: %v\n", fwVersion)
	fmt.Printf("  PCIe IDs:   dev=0x%04x vendor=0x%04x subsystem=0x%04x rev=0x%02x class=0x%02x:0x%02x\n",
		deviceInfo.DeviceID(), deviceInfo.VendorID(), deviceInfo.SubsystemID(),
		deviceInfo.RevisionID(), deviceInfo.ClassID(), deviceInfo.SubClassID())
	fmt.Printf("  PCIe Link:  gen%d x%d @ %.1f GT/s (max x%d @ %.1f GT/s)\n",
		linkInfo.PCIeGenStatus(), linkInfo.LinkWidthStatus(), linkInfo.LinkSpeedStatus(),
		linkInfo.MaxLinkWidthCapability(), linkInfo.MaxLinkSpeedCapability())
	fmt.Printf("  PCIe SR-IOV: total=%d enabled=%d\n", sriovInfo.SriovTotalVFs(), sriovInfo.SriovEnabledVFs())
	fmt.Printf("  PCIe Topo:  root=%04x:%02x switch=%s\n", rootComplexInfo.Domain(), rootComplexInfo.Bus(), switchBDF)
	fmt.Printf("  SoC Peak:   %.1f°C\n", temp.SoCPeak())
	fmt.Printf("  Ambient:    %.1f°C\n", temp.Ambient())
	fmt.Printf("  Liveness:   %t\n", d.Liveness())

	throttleReasons, err := d.ThrottleReason()
	if err != nil {
		return err
	}

	fmt.Printf("  Throttle Reason: %s\n", throttleText(throttleReasons))

	governor, err := d.GovernorProfile()
	if err != nil {
		return err
	}

	fmt.Printf("  Governor:   %v\n", governor)

	powerLimit, err := d.PowerLimit()
	if err != nil {
		return err
	}

	fmt.Printf("  Power limit: %d mW\n", powerLimit)

	fmt.Println("  Core status:")
	for _, core := range slices.Sorted(maps.Keys(coreStatuses)) {
		fmt.Printf("    PE %d: %v\n", core, coreStatuses[core])
	}

	fmt.Println("  Performance Counter:")
	counters, err := d.PerformanceCounter()
	if err != nil {
		return err
	}

	for _, core := range slices.Sorted(maps.Keys(counters)) {
		pe := counters[core]
		fmt.Printf("    Core %d: cycle_count=0x%x task_execution_cycle=0x%x\n", core, pe.CycleCount(), pe.TaskExecutionCycle())
	}

	fmt.Println("  Memory Usage:")
	memoryUsage, err := d.MemoryUsage()
	if err != nil {
		return err
	}

	for _, coreRange := range slices.SortedFunc(maps.Keys(memoryUsage), api.DeviceCoreRange.Compare) {
		info := memoryUsage[coreRange]
		fmt.Printf("    %s: %.2f / %.2f GiB\n", rangeLabel(coreRange), gib(info.InUseBytes()), gib(info.TotalBytes()))
	}

	fmt.Println("  Device files:")
	files, err := d.DeviceFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		fmt.Printf("    %s (cores %s)\n", file.Path(), coreList(file.Cores()))
	}

	fmt.Println()

	return nil
}

func throttleText(reasons []api.ThrottleReason) string {
	if len(reasons) == 0 {
		return "None"
	}

	parts := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		parts = append(parts, reason.String())
	}

	return strings.Join(parts, ", ")
}

func rangeLabel(coreRange api.DeviceCoreRange) string {
	start, end := coreRange.StartCore(), coreRange.EndCore()
	if start == end {
		return fmt.Sprintf("PE %d", start)
	}

	return fmt.Sprintf("PE %d-%d", start, end)
}

// coreList renders cores the way Rust's {:?} prints a Vec<u32>: "[0, 1, 2]".
func coreList(cores []uint32) string {
	parts := make([]string, 0, len(cores))
	for _, core := range cores {
		parts = append(parts, fmt.Sprintf("%d", core))
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func gib(bytes uint64) float64 {
	return float64(bytes) / (1024.0 * 1024.0 * 1024.0)
}

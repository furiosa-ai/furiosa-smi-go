// Command example calls pkg/api's read-only functions in order and prints
// what each one returns. Run it on a host with Furiosa NPUs:
//
//	go run ./example
package main

import (
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

func main() {
	example(os.Stdout)
}

func example(w io.Writer) {
	p := printer{w: w}

	showParsingHelpers(p)

	p.section("Driver")
	driver, err := api.DriverInfo()
	p.result("DriverInfo()", driver, err)

	showPCIDevices(p)

	p.section("Devices")
	devices, err := api.ListDevices()
	p.result("ListDevices()", fmt.Sprintf("%d devices", len(devices)), err)
	if len(devices) == 0 {
		p.detail("no Furiosa NPU found on this host")
	}
	for i := range devices {
		showDevice(p, &devices[i])
	}

	if len(devices) > 1 {
		showDevicePairs(p, devices)
	}

}

func showParsingHelpers(p printer) {
	p.section("Parsing helpers")
	full, err := api.ParseBDF("0000:27:00.0")
	p.result(`ParseBDF("0000:27:00.0")`, full, err)
	short, err := api.ParseBDF("27:00.0")
	p.result(`ParseBDF("27:00.0")`, fmt.Sprintf("%v (domain defaults to 0000)", short), err)
	_, err = api.ParseBDF("not-a-bdf")
	p.result(`ParseBDF("not-a-bdf")`, nil, err)

	built := api.NewDeviceCoreRange(0, 0, 3)
	p.result("NewDeviceCoreRange(0, 0, 3)", fmt.Sprintf("%v, Cores()=%v", built, built.Cores()), nil)
	parsed, err := api.ParseDeviceCoreRange("npu1pe2-5")
	p.result(`ParseDeviceCoreRange("npu1pe2-5")`,
		fmt.Sprintf("node=%d start=%d end=%d", parsed.DevNodeIndex(), parsed.StartCore(), parsed.EndCore()), err)
	bare, err := api.ParseDeviceCoreRange("npu1")
	p.result(`ParseDeviceCoreRange("npu1")`, fmt.Sprintf("%v (a bare node is PE 0)", bare), err)
}

func showPCIDevices(p printer) {
	p.section("PCI devices")
	pciDevices, err := api.ListPCIDevices()
	p.result("ListPCIDevices()", fmt.Sprintf("%d devices", len(pciDevices)), err)
	for _, d := range pciDevices {
		p.detail(fmt.Sprintf("%v  device=0x%04x vendor=0x%04x", d.BDF(), d.DeviceID(), d.VendorID()))
	}
	if len(pciDevices) > 0 {
		target := pciDevices[0].BDF()
		got, err := api.GetPCIDevice(target)
		p.result(fmt.Sprintf("GetPCIDevice(%v)", target), got.BDF(), err)
	}
	disabled, err := api.ListDisabledDevices()
	p.result("ListDisabledDevices()", bdfList(disabled), err)
}

func showDevice(p printer, d *api.Device) {
	p.subsection(fmt.Sprintf("%s (%v)", d.DeviceName(), d.BDF()))

	p.result("Arch() / ProductName()", fmt.Sprintf("%v / %s", d.Arch(), d.ProductName()), nil)
	p.result("DeviceID() / VendorID()", fmt.Sprintf("0x%04x / 0x%04x", d.DeviceID(), d.VendorID()), nil)
	p.result("UUID() / Serial()", fmt.Sprintf("%s / %s", d.UUID(), d.Serial()), nil)
	p.result("DevNodeIndex() / CoreNum() / NUMANode()",
		fmt.Sprintf("%d / %d / %d", d.DevNodeIndex(), d.CoreNum(), d.NUMANode()), nil)

	firmware, err := d.FirmwareVersion()
	p.result("FirmwareVersion()", firmware, err)
	major, err := d.Major()
	p.result("Major()", major, err)
	minor, err := d.Minor()
	p.result("Minor()", minor, err)
	p.result("Liveness()", d.Liveness(), nil)

	governor, err := d.GovernorProfile()
	p.result("GovernorProfile()", governor, err)
	status, err := d.CoreStatus()
	p.result("CoreStatus()", byCore(status), err)
	frequency, err := d.CoreFrequency()
	p.result("CoreFrequency() [MHz]", byCore(frequency), err)
	tcp, err := d.TCPFrequency()
	p.result("TCPFrequency() [MHz]", byCore(tcp), err)

	temperature, err := d.DeviceTemperature()
	p.result("DeviceTemperature()",
		fmt.Sprintf("SoC peak %.1f°C, ambient %.1f°C", temperature.SoCPeak(), temperature.Ambient()), err)
	power, err := d.PowerConsumption()
	p.result("PowerConsumption()", fmt.Sprintf("%.1f W", power), err)
	limit, err := d.PowerLimit()
	p.result("PowerLimit()", fmt.Sprintf("%d mW", limit), err)
	reasons, err := d.ThrottleReason()
	p.result("ThrottleReason()", throttleList(reasons), err)

	memory, err := d.MemoryUsage()
	p.result("MemoryUsage()", memoryList(memory), err)
	showPerformanceCounter(p, d)

	files, err := d.DeviceFiles()
	p.result("DeviceFiles()", fmt.Sprintf("%d files", len(files)), err)
	for _, f := range files {
		p.detail(fmt.Sprintf("%-10v %s", f.CoreRange(), f.Path()))
	}

	showPcieInfo(p, d)
}

// showPerformanceCounter takes two snapshots, since a single counter value
// only means something relative to an earlier one.
func showPerformanceCounter(p printer, d *api.Device) {
	before, err := d.PerformanceCounter()
	if err != nil {
		p.result("PerformanceCounter()", nil, err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	after, err := d.PerformanceCounter()
	if err != nil {
		p.result("PerformanceCounter()", nil, err)
		return
	}
	p.result("PerformanceCounter()", fmt.Sprintf("%d cores, sampled twice", len(after)), nil)
	for _, core := range slices.Sorted(maps.Keys(after)) {
		now, prev := after[core], before[core]
		p.detail(fmt.Sprintf("pe%d cycles=%d task=%d UtilizationSince=%.1f%% over %v",
			core, now.CycleCount(), now.TaskExecutionCycle(),
			now.UtilizationSince(prev)*100, now.DurationSince(prev).Round(time.Millisecond)))
	}
}

func showPcieInfo(p printer, d *api.Device) {
	info, err := d.PcieInfo()
	if err != nil {
		p.result("PcieInfo()", nil, err)
		return
	}
	p.result("PcieInfo()", nil, nil)
	dev, link, sriov, root := info.DeviceInfo(), info.LinkInfo(), info.SriovInfo(), info.RootComplexInfo()
	p.detail(fmt.Sprintf("device: id=0x%04x vendor=0x%04x subsystem=0x%04x rev=0x%02x class=0x%02x/0x%02x",
		dev.DeviceID(), dev.VendorID(), dev.SubsystemID(), dev.RevisionID(), dev.ClassID(), dev.SubClassID()))
	p.detail(fmt.Sprintf("link:   Gen%d x%d @ %.1f GT/s (max x%d @ %.1f GT/s)",
		link.PCIeGenStatus(), link.LinkWidthStatus(), link.LinkSpeedStatus(),
		link.MaxLinkWidthCapability(), link.MaxLinkSpeedCapability()))
	p.detail(fmt.Sprintf("sriov:  %d of %d VFs enabled", sriov.SriovEnabledVFs(), sriov.SriovTotalVFs()))
	p.detail(fmt.Sprintf("root complex: domain %04x bus %02x", root.Domain(), root.Bus()))
	if upstream := info.SwitchInfo(); upstream != nil {
		p.detail(fmt.Sprintf("switch: %v", *upstream))
	} else {
		p.detail("switch: none")
	}
}

func showDevicePairs(p printer, devices []api.Device) {
	p.section("Between devices")
	for i := range devices {
		for j := range devices {
			if i >= j {
				continue
			}
			a, b := &devices[i], &devices[j]
			linkType, err := a.DeviceToDeviceLinkType(b)
			pair := fmt.Sprintf("%s <-> %s", a.DeviceName(), b.DeviceName())
			p.result(pair+" DeviceToDeviceLinkType()", linkType, err)
			p.result(pair+" P2PAccessible()", a.P2PAccessible(b), nil)
		}
	}
}

// printer writes one aligned line per call: the call, then its value or its
// error.
type printer struct {
	w io.Writer
}

// printf ignores write errors: the example only ever prints to stdout, where a
// failed write has nowhere better to be reported.
func (p printer) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, format, args...)
}

func (p printer) section(title string) {
	p.printf("\n== %s ==\n", title)
}

func (p printer) subsection(title string) {
	p.printf("\n-- %s --\n", title)
}

func (p printer) result(call string, value any, err error) {
	switch {
	case err != nil:
		p.printf("  %-42s error: %v\n", call, err)
	case value == nil:
		p.printf("  %-42s ok\n", call)
	default:
		p.printf("  %-42s %v\n", call, value)
	}
}

func (p printer) detail(line string) {
	p.printf("      %s\n", line)
}

// byCore renders a per-core map in core order; Go maps have no order of
// their own, unlike the BTreeMap the Rust API returns.
func byCore[V any](m map[uint32]V) string {
	parts := make([]string, 0, len(m))
	for _, core := range slices.Sorted(maps.Keys(m)) {
		parts = append(parts, fmt.Sprintf("pe%d=%v", core, m[core]))
	}
	return strings.Join(parts, " ")
}

func memoryList(m map[api.DeviceCoreRange]api.MemoryInfo) string {
	ranges := slices.SortedFunc(maps.Keys(m), api.DeviceCoreRange.Compare)
	parts := make([]string, 0, len(ranges))
	for _, r := range ranges {
		parts = append(parts, fmt.Sprintf("%v=%.1f/%.1f GiB", r, gib(m[r].InUseBytes()), gib(m[r].TotalBytes())))
	}
	return strings.Join(parts, " ")
}

func gib(bytes uint64) float64 {
	return float64(bytes) / (1 << 30)
}

func throttleList(reasons []api.ThrottleReason) string {
	if len(reasons) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(reasons))
	for _, r := range reasons {
		parts = append(parts, r.String())
	}
	return strings.Join(parts, ", ")
}

func bdfList(devices []api.PCIDevice) string {
	if len(devices) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(devices))
	for _, d := range devices {
		parts = append(parts, d.BDF().String())
	}
	return strings.Join(parts, ", ")
}

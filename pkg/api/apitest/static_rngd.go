package apitest

import (
	"fmt"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

// StaticRNGDDeviceCount is the number of devices StaticRNGDSpec describes.
const StaticRNGDDeviceCount = 8

// staticRNGDBDFs are the BDFs of the static devices, indexed by device node
// index. Devices 2k and 2k+1 sit under the same PCIe bridge.
var staticRNGDBDFs = [StaticRNGDDeviceCount]string{
	"0000:27:00.0", "0000:2a:00.0", "0000:51:00.0", "0000:57:00.0",
	"0000:9e:00.0", "0000:a4:00.0", "0000:c7:00.0", "0000:ca:00.0",
}

// staticRNGDBridge is the PCIe bridge shared by one pair of static devices.
type staticRNGDBridge struct {
	bdf                       string
	secondaryBus, subordinate uint32
}

var staticRNGDBridges = [StaticRNGDDeviceCount / 2]staticRNGDBridge{
	{bdf: "0000:26:00.0", secondaryBus: 0x27, subordinate: 0x2a},
	{bdf: "0000:50:00.0", secondaryBus: 0x51, subordinate: 0x57},
	{bdf: "0000:9d:00.0", secondaryBus: 0x9e, subordinate: 0xa4},
	{bdf: "0000:c6:00.0", secondaryBus: 0xc7, subordinate: 0xca},
}

// StaticRNGDSpec describes a fixed 8-device, 2-socket RNGD server:
//
//	socket 0: bridge 0000:26:00.0 {npu0, npu1}, bridge 0000:50:00.0 {npu2, npu3}
//	socket 1: bridge 0000:9d:00.0 {npu4, npu5}, bridge 0000:c6:00.0 {npu6, npu7}
//
// so a device's link to itself is NoC, to its bridge mate Bridge, to the
// other devices on its socket Cpu, and to the other socket Interconnect.
// Every device has 8 cores, the fused partitions pe0-1, pe2-3, pe0-3,
// pe4-5, pe6-7 and pe4-7 (14 device files in all), P2P interfaces and
// hwmon sensors.
func StaticRNGDSpec() MockSpec {
	devices := make([]MockDevice, StaticRNGDDeviceCount)
	for i := range StaticRNGDDeviceCount {
		bridge := staticRNGDBridges[i/2]
		socket := &SocketMock{FirstCPU: 0, PackageID: 0}
		if i >= StaticRNGDDeviceCount/2 {
			socket = &SocketMock{FirstCPU: 32, PackageID: 1}
		}

		devices[i] = &RNGDMock{
			BDF:             staticRNGDBDFs[i],
			DevNodeIndex:    uint8(i),
			NUMANode:        int32(i / 4),
			UUID:            fmt.Sprintf("A76AAD68-6855-40B1-9E86-D080852D1C8%d", i),
			Serial:          fmt.Sprintf("TEST0236FH505KRE%d", i),
			CoreNum:         8,
			FirmwareVersion: "1.6.0~dev0, c1bebfd",
			DevNumber:       fmt.Sprintf("%d:0", 234+i),
			FusedPEs:        [][2]uint32{{0, 1}, {2, 3}, {0, 3}, {4, 5}, {6, 7}, {4, 7}},
			P2P:             true,
			Socket:          socket,
			PCIe: &PCIeMock{
				Upstream:               bridge.bdf,
				UpstreamSecondaryBus:   bridge.secondaryBus,
				UpstreamSubordinateBus: bridge.subordinate,
			},
			HWMon: &HWMonMock{SoCPeakCelsius: 20, AmbientCelsius: 10, PowerConsumption: 100},
		}
	}

	return MockSpec{Devices: devices}
}

// StaticRNGDDevices installs a fresh mock root under t.TempDir(), builds
// StaticRNGDSpec into it and returns api.ListDevices(), where devices[i]
// is npu{i}. The mock root stays installed until t completes, so the
// returned devices stay readable throughout the test; tests using it must
// not call t.Parallel().
func StaticRNGDDevices(t *testing.T) []api.Device {
	t.Helper()

	SetMockRoot(t, t.TempDir())
	Build(t, StaticRNGDSpec())

	devices, err := api.ListDevices()
	if err != nil {
		t.Fatalf("api.ListDevices: %v", err)
	}

	if len(devices) != StaticRNGDDeviceCount {
		t.Fatalf("api.ListDevices: got %d devices, want %d", len(devices), StaticRNGDDeviceCount)
	}

	for i := range devices {
		if got := devices[i].DevNodeIndex(); got != uint8(i) {
			t.Fatalf("api.ListDevices: devices[%d] is npu%d, want npu%d", i, got, i)
		}
	}

	return devices
}

// StaticRNGDDevice returns npu{index} of StaticRNGDDevices(t).
func StaticRNGDDevice(t *testing.T, index int) api.Device {
	t.Helper()

	if index < 0 || index >= StaticRNGDDeviceCount {
		t.Fatalf("StaticRNGDDevice: index %d out of range [0, %d)", index, StaticRNGDDeviceCount)
	}

	return StaticRNGDDevices(t)[index]
}

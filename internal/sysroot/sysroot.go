package sysroot

import "io/fs"

// These are suffixes (no leading slash) so both the real root (prefixed
// with "/") and a mock root (joined via filepath.Join) can share one
// literal per path.
const (
	pciDevicesSuffix      = "sys/bus/pci/devices"
	pciDriversProbeSuffix = "sys/bus/pci/drivers_probe"
	cpuDevicesSuffix      = "sys/devices/system/cpu"
)

// root joins a suffix onto, and lists directory entries under, whichever
// directory is currently active as the filesystem root. This is the only
// place any mock-vs-real decision is made -- Join, ReadDir, PCIDevicesRoot,
// and architecture packages (e.g. internal/arch/rngd's
// management/devfs/debugfs roots) all just call through whichever
// implementation is currently active, with no mock-awareness of their own.
type root interface {
	join(suffix string) string
	readDir(name string) ([]fs.DirEntry, error)
	hostPath(path string) string
}

var currentRoot root

func init() {
	currentRoot = defaultRoot{}
}

func PCIDevicesRoot() string { return Join(pciDevicesSuffix) }

func PCIDriversProbePath() string { return Join(pciDriversProbeSuffix) }

func CPUDevicesRoot() string { return Join(cpuDevicesSuffix) }

// Join resolves suffix (no leading slash) against whichever root is
// currently active. Architecture-specific packages use this for paths
// outside the three fixed roots above.
func Join(suffix string) string { return currentRoot.join(suffix) }

// ReadDir reads the named directory. Under a mock root, symlinked entries
// are reported as character devices, since mocks use symlinks to stand in
// for real character devices that can't be created without root.
func ReadDir(name string) ([]fs.DirEntry, error) { return currentRoot.readDir(name) }

// HostPath returns path as it would appear on the host, i.e. with any mock
// root prefix stripped (e.g. "<mockdir>/dev/rngd/npu0pe0" ->
// "/dev/rngd/npu0pe0"). Under the default root it returns path unchanged.
// Use it for paths handed to callers for presentation (such as
// DeviceFile.Path), never for paths this module goes on to open.
func HostPath(path string) string { return currentRoot.hostPath(path) }

package smi

// DeviceFile represents a device file.
type DeviceFile interface {
	// Cores returns a list of core for device file.
	Cores() []uint32
	// Path returns a device file path.
	Path() string
}

type FuriosaSmiDeviceFiles struct {
	count       uint32
	deviceFiles []DeviceFile
}

var _ DeviceFile = new(FuriosaSmiDeviceFile)

type FuriosaSmiDeviceFile struct {
	coreStart uint32
	coreEnd   uint32
	path      string
}

func newDeviceFile(coreStart, coreEnd uint32, path string) DeviceFile {
	return &FuriosaSmiDeviceFile{
		coreStart: coreStart,
		coreEnd:   coreEnd,
		path:      path,
	}
}

func (fdf *FuriosaSmiDeviceFile) Cores() []uint32 {
	var cores []uint32

	for i := fdf.coreStart; i <= fdf.coreEnd; i++ {
		cores = append(cores, i)
	}

	return cores
}

func (fdf *FuriosaSmiDeviceFile) Path() string {
	return fdf.path
}

func FuriosaSmiGetDeviceFiles(device Device) (*FuriosaSmiDeviceFiles, error) {
	// TODO: Implement this function
	return nil, nil
}

func (f *FuriosaSmiDeviceFiles) DeviceFiles() []DeviceFile {
	return f.deviceFiles
}

package smi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

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

// FuriosaSmiGetDeviceFiles returns all device files accessible for the given
// device by scanning /dev/rngd for character device files that match the
// device's node index. Mirrors parse_device_files in device_info.rs.
func FuriosaSmiGetDeviceFiles(device Device) (*FuriosaSmiDeviceFiles, error) {
	info, err := device.DeviceInfo()
	if err != nil {
		return nil, err
	}
	name := info.Name()
	s := strings.TrimPrefix(name, "npu")
	if s == name {
		return nil, fmt.Errorf("%w: unexpected device name %q", ErrParse, name)
	}
	nodeIdxU, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}
	nodeIdx := uint32(nodeIdxU)

	entries, err := os.ReadDir(rngdDevRootPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &FuriosaSmiDeviceFiles{}, nil
		}
		return nil, err
	}

	var files []DeviceFile
	for _, e := range entries {
		fi, ferr := e.Info()
		if e.IsDir() || ferr != nil || !isRngdDeviceFile(fi) {
			continue
		}
		parsed, ok := parseDeviceFilename(e.Name())
		if !ok || parsed.nodeIdx != nodeIdx {
			continue
		}
		absPath := filepath.Join(rngdDevRootPath(), e.Name())
		if real, eerr := filepath.EvalSymlinks(absPath); eerr == nil {
			absPath = real
		}
		files = append(files, newDeviceFile(parsed.coreStart, parsed.coreEnd, absPath))
	}

	sort.Slice(files, func(i, j int) bool {
		return filepath.Base(files[i].Path()) < filepath.Base(files[j].Path())
	})

	return &FuriosaSmiDeviceFiles{count: uint32(len(files)), deviceFiles: files}, nil
}

func (f *FuriosaSmiDeviceFiles) DeviceFiles() []DeviceFile {
	return f.deviceFiles
}

package device

// furiosa-smi/src/device.rs
// furiosa-smi/src/drain.rs

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/furiosa-ai/furiosa-smi-go/internal/bdf"
	"github.com/furiosa-ai/furiosa-smi-go/internal/smierrors"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysfs"
	"github.com/furiosa-ai/furiosa-smi-go/internal/sysroot"
)

// PCIDevice is a PCI device recognized with Furiosa's vendor ID.
type PCIDevice struct {
	bdf      bdf.BDF
	deviceID uint16
	vendorID uint16
}

func NewPCIDevice(deviceBDF bdf.BDF, deviceID, vendorID uint16) PCIDevice {
	return PCIDevice{bdf: deviceBDF, deviceID: deviceID, vendorID: vendorID}
}

// BDF returns the PCI BDF address of the device.
func (d PCIDevice) BDF() bdf.BDF {
	return d.bdf
}

// DeviceID returns the PCI device ID.
func (d PCIDevice) DeviceID() uint16 {
	return d.deviceID
}

// VendorID returns the PCI vendor ID.
func (d PCIDevice) VendorID() uint16 {
	return d.vendorID
}

// Compare orders PCIDevice values by vendor ID, then device ID, then BDF --
// the ordering internal/discovery uses to make ListPCIDevices' output
// deterministic.
func (d PCIDevice) Compare(other PCIDevice) int {
	return cmp.Or(
		cmp.Compare(d.vendorID, other.vendorID),
		cmp.Compare(d.deviceID, other.deviceID),
		d.bdf.Compare(other.bdf),
	)
}

// Disable requires root privileges.
func (d PCIDevice) Disable() error {
	devDir, err := pciDeviceDir(d.bdf)
	if err != nil {
		return err
	}

	if !sysfs.IsDriverBound(d.bdf) {
		return nil
	}

	unbindPath := filepath.Join(devDir, "driver", "unbind")
	if err := os.WriteFile(unbindPath, []byte(d.bdf.String()+"\n"), 0o644); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return smierrors.IO(unbindPath, err)
	}

	return nil
}

// Enable requires root privileges.
func (d PCIDevice) Enable() error {
	if _, err := pciDeviceDir(d.bdf); err != nil {
		return err
	}

	if sysfs.IsDriverBound(d.bdf) {
		return nil
	}

	probePath := sysroot.PCIDriversProbePath()
	if err := os.WriteFile(probePath, []byte(d.bdf.String()+"\n"), 0o644); err != nil {
		return smierrors.IO(probePath, err)
	}

	if sysfs.IsDriverBound(d.bdf) {
		return nil
	}

	return smierrors.DeviceNotAvailable(d.bdf.String())
}

func pciDeviceDir(target bdf.BDF) (string, error) {
	devDir := filepath.Join(sysroot.PCIDevicesRoot(), target.String())
	if _, err := os.Stat(devDir); err != nil {
		return "", smierrors.DeviceNotFound(fmt.Sprintf("PCI device %s not found", target.String()))
	}

	return devDir, nil
}

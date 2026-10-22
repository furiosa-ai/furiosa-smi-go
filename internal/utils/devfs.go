package utils

import "os"

// IsCharDevice reports whether entry is a character device. Mock fixtures
// stand in for real devfs character devices using plain symlinks (real
// ones can't be created without root), so this is what makes ReadDir's
// symlink-to-char-device wrapping actually get recognized by callers.
func IsCharDevice(entry os.DirEntry) bool {
	return entry.Type()&os.ModeCharDevice != 0
}

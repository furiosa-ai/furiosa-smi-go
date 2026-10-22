package sysroot

// Mock-fixture support for tests. Everything in this file is test-support
// surface, not production code -- see IsMock's doc comment for the rule on
// who's allowed to call what in this package.

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// IsMock reports whether a mock root is currently active. This exists for
// test setup to assert against (see SetMock), not for production code to
// branch on -- nothing outside a test-support caller should ever need to
// ask this question.
func IsMock() bool {
	_, ok := currentRoot.(mockRoot)

	return ok
}

// ApplyMockRoot redirects every root under this package to subdirectories
// of root, until the returned restore func is called. ApplyMockRoot
// rewrites shared package-level state, so callers using it concurrently
// (or across parallel subtests) would race -- SetMock, its sole caller,
// documents that constraint for test authors.
func ApplyMockRoot(root string) (restore func()) {
	prev := currentRoot
	currentRoot = mockRoot{dir: root}

	return func() {
		currentRoot = prev
	}
}

// mockRoot joins onto, and lists directory entries under, a fixed
// directory, substituted in for tests via ApplyMockRoot.
type mockRoot struct {
	dir string
}

func (r mockRoot) join(suffix string) string { return filepath.Join(r.dir, suffix) }

// hostPath strips r.dir from path, so a path built via join reads like its
// host counterpart. A path outside r.dir is returned unchanged.
func (r mockRoot) hostPath(p string) string {
	rel, err := filepath.Rel(r.dir, p)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return p
	}

	return path.Join("/", filepath.ToSlash(rel))
}

func (mockRoot) readDir(name string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return entries, err
	}

	for index, entry := range entries {
		if entry.Type()&fs.ModeSymlink == 0 {
			continue
		}

		if info, statErr := os.Stat(filepath.Join(name, entry.Name())); statErr == nil {
			entries[index] = &charDeviceDirEntry{DirEntry: entry, info: info}
		}
	}

	return entries, nil
}

// charDeviceDirEntry wraps a real fs.DirEntry so it reports as a character
// device. Mock fixtures stand in for real /dev/rngd character devices using
// plain symlinks, since character devices can't be created without root;
// this wrapper is what makes ReadDir report those symlinks as char devices
// to callers that inspect entry.Type().
type charDeviceDirEntry struct {
	fs.DirEntry
	info fs.FileInfo
}

func (e *charDeviceDirEntry) IsDir() bool       { return false }
func (e *charDeviceDirEntry) Type() fs.FileMode { return fs.ModeCharDevice }
func (e *charDeviceDirEntry) Info() (fs.FileInfo, error) {
	return &charDeviceFileInfo{FileInfo: e.info}, nil
}

// charDeviceFileInfo is charDeviceDirEntry's fs.FileInfo counterpart, used
// by charDeviceDirEntry.Info() so a mock symlink reports os.ModeCharDevice
// however a caller observes it via the DirEntry ReadDir returns.
type charDeviceFileInfo struct {
	fs.FileInfo
}

func (i *charDeviceFileInfo) Mode() fs.FileMode {
	return i.FileInfo.Mode().Perm() | fs.ModeCharDevice
}

// SetMock redirects every root this package exposes to subdirectories of
// root for the lifetime of t, restoring the real defaults when t
// completes. SetMock rewrites shared package-level state, so no test using
// it (or its subtests) may call t.Parallel() -- go test -race would report
// the resulting data race.
func SetMock(t *testing.T, root string) {
	t.Helper()
	t.Cleanup(ApplyMockRoot(root))
	if !IsMock() {
		t.Fatalf("SetMock(%q): IsMock() is false after ApplyMockRoot", root)
	}
}

// MockDevice is implemented by architecture-specific mock builders (e.g.
// internal/arch/rngd's RngdMock), so BuildMock can build the right kind of
// fixture for whichever device a test needs without this package needing
// to know anything architecture-specific -- not every device implementation
// necessarily touches every root this package exposes.
type MockDevice interface {
	Build(t *testing.T)
}

// MockSpec is a set of mock devices (of any architecture) to build
// together via BuildMock.
type MockSpec struct {
	Devices []MockDevice
}

// BuildMock builds the fixture for every device in spec under the
// currently active mock roots (see SetMock).
func BuildMock(t *testing.T, spec MockSpec) {
	t.Helper()
	for _, d := range spec.Devices {
		d.Build(t)
	}
}

// OrDefault returns fallback if value is empty, otherwise value.
func OrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

// MustMkdirAll creates path (and any missing parents), failing t on error.
func MustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

// MustWriteFile writes content to path, creating parent directories as
// needed, failing t on error.
func MustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

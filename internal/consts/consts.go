package consts

// MiB and GiB are generic byte-size units, not specific to any package or
// architecture -- anything computing a byte count (e.g.
// internal/sysfs.MemoryConfig) can reuse them instead of redefining its
// own mib/gib constants.
const (
	MiB = uint64(1024 * 1024)
	GiB = MiB * 1024
)

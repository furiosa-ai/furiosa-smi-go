package device

// furiosa-smi/src/device/memory.rs

// MemoryInfo is the DRAM usage of a PE fusion group.
type MemoryInfo struct {
	totalBytes uint64
	inUseBytes uint64
}

func NewMemoryInfo(totalBytes, inUseBytes uint64) MemoryInfo {
	return MemoryInfo{totalBytes: totalBytes, inUseBytes: inUseBytes}
}

// TotalBytes returns the total memory size in bytes.
func (m MemoryInfo) TotalBytes() uint64 {
	return m.totalBytes
}

// InUseBytes returns the memory size in use in bytes.
func (m MemoryInfo) InUseBytes() uint64 {
	return m.inUseBytes
}

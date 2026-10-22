package sysroot

import (
	"io/fs"
	"os"
)

// defaultRoot resolves against the actual host filesystem.
type defaultRoot struct{}

func (defaultRoot) join(suffix string) string { return "/" + suffix }

func (defaultRoot) readDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }

func (defaultRoot) hostPath(path string) string { return path }

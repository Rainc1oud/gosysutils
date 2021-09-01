package gosysutils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/disk"
	"golang.org/x/sys/unix"
)

func FileFallocate(filepath string, size int64, mode os.FileMode, force bool) error {
	if _, err := os.Stat(filepath); err == nil && !force {
		return fmt.Errorf("not overwriting existing file %s, use force=true to force", filepath)
	}

	// TODO: does O_TRUNC overwrite? (then we could change force logic)
	fo, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	var offset int64 = 0 // do we need to be able to specify this?
	// the mode field is not the same as permissions, but is a specific fallocate mode that we ignore (don't need) for now
	return unix.Fallocate(int(fo.Fd()), 0, offset, size)
}

// we can use github.com/minio/minio/pkg/disk or github.com/shirou/gopsutil/disk,
// the latter may be useful for much more so we go with that
func FsStatFromPath(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}

// DirSize returns the total size in bytes of the dir contents (recursively scanned)
// Adapted from https://ispycode.com/Blog/golang/2017-01/DU-estimate-file-space-usage
// work-alike with linux/unix du util
func DirSize(currentPath string, info os.FileInfo) (int64, error) {
	var err error

	if info == nil {
		info, err = os.Lstat(currentPath)
		if err != nil {
			return -1, err
		}
	}
	size := info.Size()

	if !info.IsDir() {
		return size, nil
	}

	dir, err := os.Open(currentPath)
	if err != nil {
		return size, err
	}
	defer dir.Close()

	fis, err := dir.Readdir(-1)
	if err != nil {
		return -1, err
	}

	for _, fi := range fis {
		if fi.Name() == "." || fi.Name() == ".." {
			continue
		}
		sizeinc, err := DirSize(filepath.Join(currentPath, fi.Name()), fi)
		if err != nil {
			return -1, err
		}
		size += sizeinc
	}

	return size, nil
}

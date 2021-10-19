package gosysutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/disk"
	"golang.org/x/sys/unix"
)

// FileExists checks whether filename exists and is a (regular) file (it returns (somehwat peculiar?) true, error if exists but is a dir)
func FileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, err
	}
	if info.IsDir() {
		return true, fmt.Errorf("%s is a directory", filename)
	}
	return true, nil
}

// DirExists checks whether dirname exists and is a dir (it returns (somehwat peculiar?) true, error if exists but is not a dir)
func DirExists(dirname string) (bool, error) {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false, err
	}
	if !info.IsDir() {
		return true, fmt.Errorf("%s is not a directory", dirname)
	}
	return true, nil
}

func IsSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

// ResolveSymlinks takes a list of paths and returns the resolved symlinks
func ResolveSymlinks(paths []string) ([]string, error) {
	var errl []string
	rls := make([]string, len(paths))
	for i, l := range paths {
		if rl, err := filepath.EvalSymlinks(l); err != nil {
			errl = append(errl, fmt.Sprintf("%s: %s", l, err.Error()))
			rls[i] = l
		} else {
			rls[i] = rl
		}
	}
	if len(errl) > 0 {
		return rls, fmt.Errorf(strings.Join(errl, "\n"))
	}
	return rls, nil
}

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

// LsDirs returns a string slice of all directory names found in the dir argument
func LsDirs(dir string) ([]string, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, nil
	}
	res := make([]string, len(des)) // first init res with upper bound, to avoid append inefficiency (at the cost of more mem)
	c := 0
	for _, de := range des {
		if de.IsDir() {
			res[c] = de.Name()
			c += 1
		}
	}
	if c < 1 {
		return []string{}, nil
	}
	return res[:c], nil
}

// LsNames returns a string slice of all (file) names found in the dir argument, excluding "." and ".."
func LsNames(dir string) ([]string, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, nil
	}
	res := make([]string, len(des)) // first init res with upper bound, to avoid append inefficiency (at the cost of more mem)
	c := 0
	for _, de := range des {
		if de.Name() != "." && de.Name() != ".." {
			res[c] = de.Name()
			c += 1
		}
	}
	if c < 1 {
		return []string{}, nil
	}
	return res[:c], nil
}

// LsNames returns a string slice with the absolute path of all (file) names found in the dir argument, excluding "." and ".."
func LsNamesAbs(dir string) ([]string, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, nil
	}
	res := make([]string, len(des)) // first init res with upper bound, to avoid append inefficiency (at the cost of more mem)
	c := 0
	for _, de := range des {
		if de.Name() != "." && de.Name() != ".." {
			res[c] = filepath.Join(dir, de.Name())
			c += 1
		}
	}
	if c < 1 {
		return []string{}, nil
	}
	return res[:c], nil
}

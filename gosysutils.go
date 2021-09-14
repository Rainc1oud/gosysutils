package gosysutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// MountBind bind mounts src on target
// The calling process needs to run under root for this!
func MountBind(src, tgt string) error {
	// mount point can:
	//	(1) not exist => create;
	//	(2) exist and not be a dir => error;
	//	(3) be already mounted => Mount returns busy error? // otherwise we should check
	var (
		fi  os.FileInfo
		err error
	)

	fi, err = os.Stat(src)
	if os.IsNotExist(err) || !fi.IsDir() {
		return fmt.Errorf("source dir %s doesn't exist or is not a directory", src)
	}

	fi, err = os.Stat(tgt)
	if os.IsNotExist(err) { // mkdir (we don't handle other errors, we'll see further down the line)
		// fmt.Printf("mountpoint %s doesn't exist, creating...\n", tgt)
		if err := os.MkdirAll(tgt, 0700); err != nil { // restrictive mode, should suffice since we execute rcnode always as the same user
			return fmt.Errorf("couldn't create mountpoint %s: %s", tgt, err.Error())
		}
	} else if !fi.IsDir() { // something already existed, is it a dir?
		return fmt.Errorf("couldn't create mountpoint %s: a non-directory with the same name already exists", tgt)
	}
	// fmt.Printf("calling Mount(%s, %s, ...)\n", src, tgt)
	return unix.Mount(src, tgt, "", unix.MS_BIND, "") // TBC: no constant for fstype?
}

// MountBindAll bind mounts the source dirs in args[:nargs-2] in args[nargs-1], creating mount points based on the source dir names as needed
func MountBindAll(dirs ...string) error {
	if len(dirs) < 2 {
		return fmt.Errorf("at least two arguments required")
	}
	var errstrs []string
	mpr := dirs[len(dirs)-1]
	for _, src := range dirs[:len(dirs)-1] { // attention: go index ranges don't include last index itself
		mp := filepath.Base(src)
		if mp == "." || mp == "/" {
			errstrs = append(errstrs, fmt.Sprintf("refusing to mount on mountpoint %s", mp))
		} else {
			// fmt.Printf("calling MountBind(%s, %s)\n", src, filepath.Join(mpr, mp))
			if err := MountBind(src, filepath.Join(mpr, mp)); err != nil {
				// fmt.Printf("error: %s\n", err.Error())
				errstrs = append(errstrs, err.Error())
			}
		}
	}
	if len(errstrs) == 0 {
		return nil
	}
	return fmt.Errorf("errors occurred:\n%s", strings.Join(errstrs, "\n"))
}

func Unmount(mountpoint string) error {
	return unix.Unmount(mountpoint, 0)
}

// UmountAll unmounts all mountpoints under mount point root mpr
// "not mounted" errors are ignored, other errors are collected and returned
func UmountAll(mpr string) error {
	var errstrs []string
	des, err := os.ReadDir(mpr)
	if err != nil {
		return err
	}

	for _, de := range des {
		// fmt.Printf("evaluating dir entry: %+v\n", de)
		if de.IsDir() {
			err := Unmount(filepath.Join(mpr, de.Name()))
			if err != nil {
				if err == unix.EINVAL {
					// fmt.Printf("ignoring EINVAL (not mounted?) for dir: %s\n", de.Name())
				} else {
					// ignore "invalid argument" which covers "not mounted" (and more???)
					errstrs = append(errstrs, err.Error())
				}
			}
		}
	}

	if len(errstrs) == 0 {
		return nil
	}
	return fmt.Errorf("errors occurred:\n%s", strings.Join(errstrs, "\n"))
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

func IsSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

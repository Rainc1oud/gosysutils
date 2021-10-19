package gosysutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

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

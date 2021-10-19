package gosysutils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBindMountDidMount(t *testing.T) {
	assert := assert.New(t)

	dirtgt := mktempdir(t)
	defer os.RemoveAll(dirtgt)
	dirsrc := mktempdir(t)
	defer os.RemoveAll(dirsrc)

	fn := filepath.Join(dirsrc, "somefile.txt")
	testcont := []byte("Some content")

	err := os.WriteFile(fn, testcont, 0644)
	assert.Nil(err)
	if err := MountBind(dirsrc, dirtgt); err != nil {
		t.Fatal(err)
	}
	txt, err := os.ReadFile(fn)
	assert.Nil(err)
	assert.Equal(testcont, txt)

	err = Unmount(dirtgt)
	assert.Nil(err)
}

func TestBindMountAll(t *testing.T) {
	assert := assert.New(t)

	dirtgt := mktempdir(t)
	defer os.RemoveAll(dirtgt)
	dirsrc := mktempdir(t)
	defer os.RemoveAll(dirsrc)

	dirs := []string{"somedir", "anotherdir", "dir3", "dirfour"}
	// define and make source dirs
	fullpaths := make([]string, len(dirs))
	for i, s := range dirs { // prepend source root to dir names
		fullpaths[i] = filepath.Join(dirsrc, s)
		if err := os.Mkdir(fullpaths[i], 0700); err != nil {
			t.Fatal(err)
		}
	}

	// do the mount
	if err := MountBindAll(append(fullpaths, dirtgt)...); err != nil {
		t.Fatal(err)
	}

	// write dummy files in each source dir and check whether they are indeed accessible from the mount
	for i, d := range dirs {
		wcont := []byte(fmt.Sprintf("Some content in dir %s", d))
		fn := fmt.Sprintf("somefile%d.txt", i+1)
		if err := (os.WriteFile(filepath.Join(dirsrc, d, fn), wcont, 0666)); err != nil {
			t.Fatal(err)
		}
		rcont, err := os.ReadFile(filepath.Join(dirtgt, d, fn))
		if err != nil {
			t.Fatal(err)
		}
		assert.EqualValues(wcont, rcont)
	}

	// first unmount one of the dirs, to test if UmountAll correctly ignores "not mounted" errors
	if err := Unmount(filepath.Join(dirtgt, dirs[len(dirs)-1])); err != nil {
		t.Fatal(err)
	}

	if err := UmountAll(dirtgt); err != nil {
		t.Fatal(err)
	}
}

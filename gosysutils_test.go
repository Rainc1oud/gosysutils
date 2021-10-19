package gosysutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mktemp returns a new temp dir or exits with a fatal error
func mktempdir(t *testing.T) string {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir(pwd, ".test-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestFsStat(t *testing.T) {
	assert := assert.New(t)
	pwd, err := os.Getwd()
	assert.Nil(err)
	stat, err := FsStatFromPath(pwd)
	assert.Nil(err)
	fmt.Printf("\nSome stats obtained:\nTotal %d, \nFree %d\n\n", stat.Total, stat.Free)
}

func TestFileFalloc(t *testing.T) {
	var sz int64 = 100000 // bytes
	assert := assert.New(t)
	dir := mktempdir(t)
	defer os.RemoveAll(dir)
	fn := path.Join(dir, "reserved")
	if err := FileFallocate(fn, sz, 0664, true); err != nil {
		t.Fatal(err)
	}
	// assert.Nil(err)
	// fmt.Println(err.Error())
	fi, err := os.Stat(fn)
	assert.Nil(err)
	fmt.Printf("File size: specified: %d, actual: %d\n", sz, fi.Size())
	assert.Equal(sz, fi.Size())
}

func TestLsNames(t *testing.T) {
	assert := assert.New(t)

	dirtgt := mktempdir(t)
	defer os.RemoveAll(dirtgt)
	// make a few dummy files and dirs
	srcnms := []string{"somefile.txt", "somefile2.txt", "somedir1", "somedir2", "symlinktofile", "symlinktodir"}
	assert.Nil(os.WriteFile(filepath.Join(dirtgt, srcnms[0]), []byte("Some content in the file\n"), 0644))
	assert.Nil(os.WriteFile(filepath.Join(dirtgt, srcnms[1]), []byte("Some content in the file2\n"), 0644))
	os.Mkdir(filepath.Join(dirtgt, srcnms[2]), 0755)
	os.Mkdir(filepath.Join(dirtgt, srcnms[3]), 0755)
	os.Symlink(filepath.Join(dirtgt, srcnms[0]), filepath.Join(dirtgt, srcnms[4]))
	os.Symlink(filepath.Join(dirtgt, srcnms[3]), filepath.Join(dirtgt, srcnms[5]))

	nms, err := LsNames(dirtgt)
	fmt.Printf("LsNames(%s) => %v\n", dirtgt, nms)
	assert.Nil(err)
	assert.ElementsMatch(nms, srcnms)
}

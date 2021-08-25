package gosysutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsStat(t *testing.T) {
	assert := assert.New(t)

	pwd, err := os.Getwd()
	assert.Nil(err)
	stat, err := FsStatFromPath(pwd)
	assert.Nil(err)
	fmt.Printf("some stats obtained:\nTotal %d, \nFree %d", stat.Total, stat.Free)
}

func TestFileFalloc(t *testing.T) {
	var sz int64 = 100000 // bytes
	assert := assert.New(t)
	pwd, err := os.Getwd()
	assert.Nil(err)
	dir, err := ioutil.TempDir(pwd, ".test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fn := path.Join(dir, "reserved")
	err = FileFallocate(fn, sz, 0664, true)
	t.Fatal(err)
	// assert.Nil(err)
	// fmt.Println(err.Error())
	fi, err := os.Stat(fn)
	assert.Nil(err)
	assert.Equal(sz, fi.Size())
}

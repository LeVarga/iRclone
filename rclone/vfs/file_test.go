package vfs

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/rclone/rclone/fstest"
	"github.com/rclone/rclone/fstest/mockfs"
	"github.com/rclone/rclone/fstest/mockobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fileCreate(t *testing.T, r *fstest.Run) (*VFS, *File, fstest.Item) {
	vfs := New(r.Fremote, nil)

	file1 := r.WriteObject(context.Background(), "dir/file1", "file1 contents", t1)
	fstest.CheckItems(t, r.Fremote, file1)

	node, err := vfs.Stat("dir/file1")
	require.NoError(t, err)
	require.True(t, node.IsFile())

	return vfs, node.(*File), file1
}

func TestFileMethods(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs, file, _ := fileCreate(t, r)

	// String
	assert.Equal(t, "dir/file1", file.String())
	assert.Equal(t, "<nil *File>", (*File)(nil).String())

	// IsDir
	assert.Equal(t, false, file.IsDir())

	// IsFile
	assert.Equal(t, true, file.IsFile())

	// Mode
	assert.Equal(t, vfs.Opt.FilePerms, file.Mode())

	// Name
	assert.Equal(t, "file1", file.Name())

	// Path
	assert.Equal(t, "dir/file1", file.Path())

	// Sys
	assert.Equal(t, nil, file.Sys())

	// Inode
	assert.NotEqual(t, uint64(0), file.Inode())

	// Node
	assert.Equal(t, file, file.Node())

	// ModTime
	assert.WithinDuration(t, t1, file.ModTime(), r.Fremote.Precision())

	// Size
	assert.Equal(t, int64(14), file.Size())

	// Sync
	assert.NoError(t, file.Sync())

	// DirEntry
	assert.Equal(t, file.o, file.DirEntry())

	// Dir
	assert.Equal(t, file.d, file.Dir())

	// VFS
	assert.Equal(t, vfs, file.VFS())
}

func TestFileSetModTime(t *testing.T) {
	r := fstest.NewRun(t)
	if !canSetModTime(t, r) {
		return
	}
	defer r.Finalise()
	vfs, file, file1 := fileCreate(t, r)

	err := file.SetModTime(t2)
	require.NoError(t, err)

	file1.ModTime = t2
	fstest.CheckItems(t, r.Fremote, file1)

	vfs.Opt.ReadOnly = true
	err = file.SetModTime(t2)
	assert.Equal(t, EROFS, err)
}

func TestFileOpenRead(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	_, file, _ := fileCreate(t, r)

	fd, err := file.openRead()
	require.NoError(t, err)

	contents, err := ioutil.ReadAll(fd)
	require.NoError(t, err)
	assert.Equal(t, "file1 contents", string(contents))

	require.NoError(t, fd.Close())
}

func TestFileOpenReadUnknownSize(t *testing.T) {
	var (
		contents = []byte("file contents")
		remote   = "file.txt"
		ctx      = context.Background()
	)

	// create a mock object which returns size -1
	o := mockobject.New(remote).WithContent(contents, mockobject.SeekModeNone)
	o.SetUnknownSize(true)
	assert.Equal(t, int64(-1), o.Size())

	// add it to a mock fs
	f := mockfs.NewFs("test", "root")
	f.AddObject(o)
	testObj, err := f.NewObject(ctx, remote)
	require.NoError(t, err)
	assert.Equal(t, int64(-1), testObj.Size())

	// create a VFS from that mockfs
	vfs := New(f, nil)

	// find the file
	node, err := vfs.Stat(remote)
	require.NoError(t, err)
	require.True(t, node.IsFile())
	file := node.(*File)

	// open it
	fd, err := file.openRead()
	require.NoError(t, err)
	assert.Equal(t, int64(0), fd.Size())

	// check the contents are not empty even though size is empty
	gotContents, err := ioutil.ReadAll(fd)
	require.NoError(t, err)
	assert.Equal(t, contents, gotContents)
	t.Logf("gotContents = %q", gotContents)

	// check that file size has been updated
	assert.Equal(t, int64(len(contents)), fd.Size())

	require.NoError(t, fd.Close())
}

func TestFileOpenWrite(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs, file, _ := fileCreate(t, r)

	fd, err := file.openWrite(os.O_WRONLY | os.O_TRUNC)
	require.NoError(t, err)

	newContents := []byte("this is some new contents")
	n, err := fd.Write(newContents)
	require.NoError(t, err)
	assert.Equal(t, len(newContents), n)
	require.NoError(t, fd.Close())

	assert.Equal(t, int64(25), file.Size())

	vfs.Opt.ReadOnly = true
	_, err = file.openWrite(os.O_WRONLY | os.O_TRUNC)
	assert.Equal(t, EROFS, err)
}

func TestFileRemove(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs, file, _ := fileCreate(t, r)

	err := file.Remove()
	require.NoError(t, err)

	fstest.CheckItems(t, r.Fremote)

	vfs.Opt.ReadOnly = true
	err = file.Remove()
	assert.Equal(t, EROFS, err)
}

func TestFileRemoveAll(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs, file, _ := fileCreate(t, r)

	err := file.RemoveAll()
	require.NoError(t, err)

	fstest.CheckItems(t, r.Fremote)

	vfs.Opt.ReadOnly = true
	err = file.RemoveAll()
	assert.Equal(t, EROFS, err)
}

func TestFileOpen(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	_, file, _ := fileCreate(t, r)

	fd, err := file.Open(os.O_RDONLY)
	require.NoError(t, err)
	_, ok := fd.(*ReadFileHandle)
	assert.True(t, ok)
	require.NoError(t, fd.Close())

	fd, err = file.Open(os.O_WRONLY)
	assert.NoError(t, err)
	_, ok = fd.(*WriteFileHandle)
	assert.True(t, ok)
	require.NoError(t, fd.Close())

	fd, err = file.Open(os.O_RDWR)
	assert.NoError(t, err)
	_, ok = fd.(*WriteFileHandle)
	assert.True(t, ok)

	fd, err = file.Open(3)
	assert.Equal(t, EPERM, err)
}

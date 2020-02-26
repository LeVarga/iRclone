// Test suite for vfs

package vfs

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/pkg/errors"
	_ "github.com/rclone/rclone/backend/all" // import all the backends
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Some times used in the tests
var (
	t1 = fstest.Time("2001-02-03T04:05:06.499999999Z")
	t2 = fstest.Time("2011-12-25T12:59:59.123456789Z")
	t3 = fstest.Time("2011-12-30T12:59:59.000000000Z")
)

// TestMain drives the tests
func TestMain(m *testing.M) {
	fstest.TestMain(m)
}

// Check baseHandle performs as advertised
func TestVFSbaseHandle(t *testing.T) {
	fh := baseHandle{}

	err := fh.Chdir()
	assert.Equal(t, ENOSYS, err)

	err = fh.Chmod(0)
	assert.Equal(t, ENOSYS, err)

	err = fh.Chown(0, 0)
	assert.Equal(t, ENOSYS, err)

	err = fh.Close()
	assert.Equal(t, ENOSYS, err)

	fd := fh.Fd()
	assert.Equal(t, uintptr(0), fd)

	name := fh.Name()
	assert.Equal(t, "", name)

	_, err = fh.Read(nil)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.ReadAt(nil, 0)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.Readdir(0)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.Readdirnames(0)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.Seek(0, io.SeekStart)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.Stat()
	assert.Equal(t, ENOSYS, err)

	err = fh.Sync()
	assert.Equal(t, nil, err)

	err = fh.Truncate(0)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.Write(nil)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.WriteAt(nil, 0)
	assert.Equal(t, ENOSYS, err)

	_, err = fh.WriteString("")
	assert.Equal(t, ENOSYS, err)

	err = fh.Flush()
	assert.Equal(t, ENOSYS, err)

	err = fh.Release()
	assert.Equal(t, ENOSYS, err)

	node := fh.Node()
	assert.Nil(t, node)
}

// TestNew sees if the New command works properly
func TestVFSNew(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()

	// Check making a VFS with nil options
	vfs := New(r.Fremote, nil)
	var defaultOpt = DefaultOpt
	defaultOpt.DirPerms |= os.ModeDir
	assert.Equal(t, vfs.Opt, defaultOpt)
	assert.Equal(t, vfs.f, r.Fremote)

	// Check the initialisation
	var opt = DefaultOpt
	opt.DirPerms = 0777
	opt.FilePerms = 0666
	opt.Umask = 0002
	vfs = New(r.Fremote, &opt)
	assert.Equal(t, os.FileMode(0775)|os.ModeDir, vfs.Opt.DirPerms)
	assert.Equal(t, os.FileMode(0664), vfs.Opt.FilePerms)
}

// TestRoot checks root directory is present and correct
func TestVFSRoot(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs := New(r.Fremote, nil)

	root, err := vfs.Root()
	require.NoError(t, err)
	assert.Equal(t, vfs.root, root)
	assert.True(t, root.IsDir())
	assert.Equal(t, vfs.Opt.DirPerms.Perm(), root.Mode().Perm())
}

func TestVFSStat(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs := New(r.Fremote, nil)

	file1 := r.WriteObject(context.Background(), "file1", "file1 contents", t1)
	file2 := r.WriteObject(context.Background(), "dir/file2", "file2 contents", t2)
	fstest.CheckItems(t, r.Fremote, file1, file2)

	node, err := vfs.Stat("file1")
	require.NoError(t, err)
	assert.True(t, node.IsFile())
	assert.Equal(t, "file1", node.Name())

	node, err = vfs.Stat("dir")
	require.NoError(t, err)
	assert.True(t, node.IsDir())
	assert.Equal(t, "dir", node.Name())

	node, err = vfs.Stat("dir/file2")
	require.NoError(t, err)
	assert.True(t, node.IsFile())
	assert.Equal(t, "file2", node.Name())

	node, err = vfs.Stat("not found")
	assert.Equal(t, os.ErrNotExist, err)

	node, err = vfs.Stat("dir/not found")
	assert.Equal(t, os.ErrNotExist, err)

	node, err = vfs.Stat("not found/not found")
	assert.Equal(t, os.ErrNotExist, err)

	node, err = vfs.Stat("file1/under a file")
	assert.Equal(t, os.ErrNotExist, err)
}

func TestVFSStatParent(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs := New(r.Fremote, nil)

	file1 := r.WriteObject(context.Background(), "file1", "file1 contents", t1)
	file2 := r.WriteObject(context.Background(), "dir/file2", "file2 contents", t2)
	fstest.CheckItems(t, r.Fremote, file1, file2)

	node, leaf, err := vfs.StatParent("file1")
	require.NoError(t, err)
	assert.True(t, node.IsDir())
	assert.Equal(t, "/", node.Name())
	assert.Equal(t, "file1", leaf)

	node, leaf, err = vfs.StatParent("dir/file2")
	require.NoError(t, err)
	assert.True(t, node.IsDir())
	assert.Equal(t, "dir", node.Name())
	assert.Equal(t, "file2", leaf)

	node, leaf, err = vfs.StatParent("not found")
	require.NoError(t, err)
	assert.True(t, node.IsDir())
	assert.Equal(t, "/", node.Name())
	assert.Equal(t, "not found", leaf)

	_, _, err = vfs.StatParent("not found dir/not found")
	assert.Equal(t, os.ErrNotExist, err)

	_, _, err = vfs.StatParent("file1/under a file")
	assert.Equal(t, os.ErrExist, err)
}

func TestVFSOpenFile(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs := New(r.Fremote, nil)

	file1 := r.WriteObject(context.Background(), "file1", "file1 contents", t1)
	file2 := r.WriteObject(context.Background(), "dir/file2", "file2 contents", t2)
	fstest.CheckItems(t, r.Fremote, file1, file2)

	fd, err := vfs.OpenFile("file1", os.O_RDONLY, 0777)
	require.NoError(t, err)
	assert.NotNil(t, fd)
	require.NoError(t, fd.Close())

	fd, err = vfs.OpenFile("dir", os.O_RDONLY, 0777)
	require.NoError(t, err)
	assert.NotNil(t, fd)
	require.NoError(t, fd.Close())

	fd, err = vfs.OpenFile("dir/new_file.txt", os.O_RDONLY, 0777)
	assert.Equal(t, os.ErrNotExist, err)
	assert.Nil(t, fd)

	fd, err = vfs.OpenFile("dir/new_file.txt", os.O_WRONLY|os.O_CREATE, 0777)
	require.NoError(t, err)
	assert.NotNil(t, fd)
	err = fd.Close()
	if errors.Cause(err) != fs.ErrorCantUploadEmptyFiles {
		require.NoError(t, err)
	}

	fd, err = vfs.OpenFile("not found/new_file.txt", os.O_WRONLY|os.O_CREATE, 0777)
	assert.Equal(t, os.ErrNotExist, err)
	assert.Nil(t, fd)
}

func TestVFSRename(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	features := r.Fremote.Features()
	if features.Move == nil && features.Copy == nil {
		return // skip as can't rename files
	}
	vfs := New(r.Fremote, nil)

	file1 := r.WriteObject(context.Background(), "dir/file2", "file2 contents", t2)
	fstest.CheckItems(t, r.Fremote, file1)

	err := vfs.Rename("dir/file2", "dir/file1")
	require.NoError(t, err)
	file1.Path = "dir/file1"
	fstest.CheckItems(t, r.Fremote, file1)

	err = vfs.Rename("dir/file1", "file0")
	require.NoError(t, err)
	file1.Path = "file0"
	fstest.CheckItems(t, r.Fremote, file1)

	err = vfs.Rename("not found/file0", "file0")
	assert.Equal(t, os.ErrNotExist, err)

	err = vfs.Rename("file0", "not found/file0")
	assert.Equal(t, os.ErrNotExist, err)
}

func TestVFSStatfs(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	vfs := New(r.Fremote, nil)

	// pre-conditions
	assert.Nil(t, vfs.usage)
	assert.True(t, vfs.usageTime.IsZero())

	aboutSupported := r.Fremote.Features().About != nil

	// read
	total, used, free := vfs.Statfs()
	if !aboutSupported {
		assert.Equal(t, int64(-1), total)
		assert.Equal(t, int64(-1), free)
		assert.Equal(t, int64(-1), used)
		return // can't test anything else if About not supported
	}
	require.NotNil(t, vfs.usage)
	assert.False(t, vfs.usageTime.IsZero())
	if vfs.usage.Total != nil {
		assert.Equal(t, *vfs.usage.Total, total)
	} else {
		assert.Equal(t, int64(-1), total)
	}
	if vfs.usage.Free != nil {
		assert.Equal(t, *vfs.usage.Free, free)
	} else {
		assert.Equal(t, int64(-1), free)
	}
	if vfs.usage.Used != nil {
		assert.Equal(t, *vfs.usage.Used, used)
	} else {
		assert.Equal(t, int64(-1), used)
	}

	// read cached
	oldUsage := vfs.usage
	oldTime := vfs.usageTime
	total2, used2, free2 := vfs.Statfs()
	assert.Equal(t, oldUsage, vfs.usage)
	assert.Equal(t, total, total2)
	assert.Equal(t, used, used2)
	assert.Equal(t, free, free2)
	assert.Equal(t, oldTime, vfs.usageTime)
}

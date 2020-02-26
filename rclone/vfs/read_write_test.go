package vfs

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fstest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanup(t *testing.T, r *fstest.Run, vfs *VFS) {
	assert.NoError(t, vfs.CleanUp())
	vfs.Shutdown()
	r.Finalise()
}

// Open a file for write
func rwHandleCreateReadOnly(t *testing.T, r *fstest.Run) (*VFS, *RWFileHandle) {
	opt := DefaultOpt
	opt.CacheMode = CacheModeFull
	vfs := New(r.Fremote, &opt)

	file1 := r.WriteObject(context.Background(), "dir/file1", "0123456789abcdef", t1)
	fstest.CheckItems(t, r.Fremote, file1)

	h, err := vfs.OpenFile("dir/file1", os.O_RDONLY, 0777)
	require.NoError(t, err)
	fh, ok := h.(*RWFileHandle)
	require.True(t, ok)

	return vfs, fh
}

// Open a file for write
func rwHandleCreateWriteOnly(t *testing.T, r *fstest.Run) (*VFS, *RWFileHandle) {
	opt := DefaultOpt
	opt.CacheMode = CacheModeFull
	vfs := New(r.Fremote, &opt)

	h, err := vfs.OpenFile("file1", os.O_WRONLY|os.O_CREATE, 0777)
	require.NoError(t, err)
	fh, ok := h.(*RWFileHandle)
	require.True(t, ok)

	return vfs, fh
}

// read data from the string
func rwReadString(t *testing.T, fh *RWFileHandle, n int) string {
	buf := make([]byte, n)
	n, err := fh.Read(buf)
	if err != io.EOF {
		assert.NoError(t, err)
	}
	return string(buf[:n])
}

func TestRWFileHandleMethodsRead(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateReadOnly(t, r)
	defer cleanup(t, r, vfs)

	// String
	assert.Equal(t, "dir/file1 (rw)", fh.String())
	assert.Equal(t, "<nil *RWFileHandle>", (*RWFileHandle)(nil).String())
	assert.Equal(t, "<nil *RWFileHandle.file>", new(RWFileHandle).String())

	// Node
	node := fh.Node()
	assert.Equal(t, "file1", node.Name())

	// Size
	assert.Equal(t, int64(16), fh.Size())

	// No opens yet
	assert.Equal(t, 0, fh.file.rwOpens())

	// Read 1
	assert.Equal(t, "0", rwReadString(t, fh, 1))

	// Open after the read
	assert.Equal(t, 1, fh.file.rwOpens())

	// Read remainder
	assert.Equal(t, "123456789abcdef", rwReadString(t, fh, 256))

	// Read EOF
	buf := make([]byte, 16)
	_, err := fh.Read(buf)
	assert.Equal(t, io.EOF, err)

	// Sync
	err = fh.Sync()
	assert.NoError(t, err)

	// Stat
	var fi os.FileInfo
	fi, err = fh.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(16), fi.Size())
	assert.Equal(t, "file1", fi.Name())

	// Close
	assert.False(t, fh.closed)
	assert.Equal(t, nil, fh.Close())
	assert.True(t, fh.closed)

	// No opens again
	assert.Equal(t, 0, fh.file.rwOpens())

	// Close again
	assert.Equal(t, ECLOSED, fh.Close())
}

func TestRWFileHandleSeek(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateReadOnly(t, r)
	defer cleanup(t, r, vfs)

	assert.Equal(t, fh.opened, false)

	// Check null seeks don't open the file
	n, err := fh.Seek(0, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, fh.opened, false)
	n, err = fh.Seek(0, io.SeekCurrent)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, fh.opened, false)

	assert.Equal(t, "0", rwReadString(t, fh, 1))

	// 0 means relative to the origin of the file,
	n, err = fh.Seek(5, io.SeekStart)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), n)
	assert.Equal(t, "5", rwReadString(t, fh, 1))

	// 1 means relative to the current offset
	n, err = fh.Seek(-3, io.SeekCurrent)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), n)
	assert.Equal(t, "3", rwReadString(t, fh, 1))

	// 2 means relative to the end.
	n, err = fh.Seek(-3, io.SeekEnd)
	assert.NoError(t, err)
	assert.Equal(t, int64(13), n)
	assert.Equal(t, "d", rwReadString(t, fh, 1))

	// Seek off the end
	_, err = fh.Seek(100, io.SeekStart)
	assert.NoError(t, err)

	// Get the error on read
	buf := make([]byte, 16)
	l, err := fh.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, l)

	// Close
	assert.Equal(t, nil, fh.Close())
}

func TestRWFileHandleReadAt(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateReadOnly(t, r)
	defer cleanup(t, r, vfs)

	// read from start
	buf := make([]byte, 1)
	n, err := fh.ReadAt(buf, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, "0", string(buf[:n]))

	// seek forwards
	n, err = fh.ReadAt(buf, 5)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, "5", string(buf[:n]))

	// seek backwards
	n, err = fh.ReadAt(buf, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, "1", string(buf[:n]))

	// read exactly to the end
	buf = make([]byte, 6)
	n, err = fh.ReadAt(buf, 10)
	require.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "abcdef", string(buf[:n]))

	// read off the end
	buf = make([]byte, 256)
	n, err = fh.ReadAt(buf, 10)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "abcdef", string(buf[:n]))

	// read starting off the end
	n, err = fh.ReadAt(buf, 100)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)

	// Properly close the file
	assert.NoError(t, fh.Close())

	// check reading on closed file
	_, err = fh.ReadAt(buf, 100)
	assert.Equal(t, ECLOSED, err)
}

func TestRWFileHandleFlushRead(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateReadOnly(t, r)
	defer cleanup(t, r, vfs)

	// Check Flush does nothing if read not called
	err := fh.Flush()
	assert.NoError(t, err)
	assert.False(t, fh.closed)

	// Read data
	buf := make([]byte, 256)
	n, err := fh.Read(buf)
	assert.True(t, err == io.EOF || err == nil)
	assert.Equal(t, 16, n)

	// Check Flush does nothing if read called
	err = fh.Flush()
	assert.NoError(t, err)
	assert.False(t, fh.closed)

	// Check flush does nothing if called again
	err = fh.Flush()
	assert.NoError(t, err)
	assert.False(t, fh.closed)

	// Properly close the file
	assert.NoError(t, fh.Close())
}

func TestRWFileHandleReleaseRead(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateReadOnly(t, r)
	defer cleanup(t, r, vfs)

	// Read data
	buf := make([]byte, 256)
	n, err := fh.Read(buf)
	assert.True(t, err == io.EOF || err == nil)
	assert.Equal(t, 16, n)

	// Check Release closes file
	err = fh.Release()
	assert.NoError(t, err)
	assert.True(t, fh.closed)

	// Check Release does nothing if called again
	err = fh.Release()
	assert.NoError(t, err)
	assert.True(t, fh.closed)
}

/// ------------------------------------------------------------

func TestRWFileHandleMethodsWrite(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateWriteOnly(t, r)
	defer cleanup(t, r, vfs)

	// 1 opens since we opened with O_CREATE and the file didn't
	// exist in the cache
	assert.Equal(t, 1, fh.file.rwOpens())

	// String
	assert.Equal(t, "file1 (rw)", fh.String())
	assert.Equal(t, "<nil *RWFileHandle>", (*RWFileHandle)(nil).String())
	assert.Equal(t, "<nil *RWFileHandle.file>", new(RWFileHandle).String())

	// Node
	node := fh.Node()
	assert.Equal(t, "file1", node.Name())

	offset := func() int64 {
		n, err := fh.Seek(0, io.SeekCurrent)
		require.NoError(t, err)
		return n
	}

	// Offset #1
	assert.Equal(t, int64(0), offset())
	assert.Equal(t, int64(0), node.Size())

	// Size #1
	assert.Equal(t, int64(0), fh.Size())

	// Write
	n, err := fh.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	// Open after the write
	assert.Equal(t, 1, fh.file.rwOpens())

	// Offset #2
	assert.Equal(t, int64(5), offset())
	assert.Equal(t, int64(5), node.Size())

	// Size #2
	assert.Equal(t, int64(5), fh.Size())

	// WriteString
	n, err = fh.WriteString(" world!")
	assert.NoError(t, err)
	assert.Equal(t, 7, n)

	// Sync
	err = fh.Sync()
	assert.NoError(t, err)

	// Stat
	var fi os.FileInfo
	fi, err = fh.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(12), fi.Size())
	assert.Equal(t, "file1", fi.Name())

	// Truncate
	err = fh.Truncate(11)
	assert.NoError(t, err)

	// Close
	assert.NoError(t, fh.Close())

	// No opens again
	assert.Equal(t, 0, fh.file.rwOpens())

	// Check double close
	err = fh.Close()
	assert.Equal(t, ECLOSED, err)

	// check vfs
	root, err := vfs.Root()
	require.NoError(t, err)
	checkListing(t, root, []string{"file1,11,false"})

	// check the underlying r.Fremote but not the modtime
	file1 := fstest.NewItem("file1", "hello world", t1)
	fstest.CheckListingWithPrecision(t, r.Fremote, []fstest.Item{file1}, []string{}, fs.ModTimeNotSupported)
}

func TestRWFileHandleWriteAt(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateWriteOnly(t, r)
	defer cleanup(t, r, vfs)

	offset := func() int64 {
		n, err := fh.Seek(0, io.SeekCurrent)
		require.NoError(t, err)
		return n
	}

	// Preconditions
	assert.Equal(t, int64(0), offset())
	assert.True(t, fh.opened)
	assert.False(t, fh.writeCalled)
	assert.True(t, fh.changed)

	// Write the data
	n, err := fh.WriteAt([]byte("hello**"), 0)
	assert.NoError(t, err)
	assert.Equal(t, 7, n)

	// After write
	assert.Equal(t, int64(0), offset())
	assert.True(t, fh.writeCalled)

	// Write more data
	n, err = fh.WriteAt([]byte(" world"), 5)
	assert.NoError(t, err)
	assert.Equal(t, 6, n)

	// Close
	assert.NoError(t, fh.Close())

	// Check can't write on closed handle
	n, err = fh.WriteAt([]byte("hello"), 0)
	assert.Equal(t, ECLOSED, err)
	assert.Equal(t, 0, n)

	// check vfs
	root, err := vfs.Root()
	require.NoError(t, err)
	checkListing(t, root, []string{"file1,11,false"})

	// check the underlying r.Fremote but not the modtime
	file1 := fstest.NewItem("file1", "hello world", t1)
	fstest.CheckListingWithPrecision(t, r.Fremote, []fstest.Item{file1}, []string{}, fs.ModTimeNotSupported)
}

func TestRWFileHandleWriteNoWrite(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateWriteOnly(t, r)
	defer cleanup(t, r, vfs)

	// Close the file without writing to it
	err := fh.Close()
	if errors.Cause(err) == fs.ErrorCantUploadEmptyFiles {
		t.Logf("skipping test: %v", err)
		return
	}
	assert.NoError(t, err)

	// Create a different file (not in the cache)
	h, err := vfs.OpenFile("file2", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	require.NoError(t, err)

	// Close it with Flush and Release
	err = h.Flush()
	assert.NoError(t, err)
	err = h.Release()
	assert.NoError(t, err)

	// check vfs
	root, err := vfs.Root()
	require.NoError(t, err)
	checkListing(t, root, []string{"file1,0,false", "file2,0,false"})

	// check the underlying r.Fremote but not the modtime
	file1 := fstest.NewItem("file1", "", t1)
	file2 := fstest.NewItem("file2", "", t1)
	fstest.CheckListingWithPrecision(t, r.Fremote, []fstest.Item{file1, file2}, []string{}, fs.ModTimeNotSupported)
}

func TestRWFileHandleFlushWrite(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateWriteOnly(t, r)
	defer cleanup(t, r, vfs)

	// Check that the file has been create and is open
	assert.True(t, fh.opened)

	// Write some data
	n, err := fh.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	// Check Flush does not close file if write called
	err = fh.Flush()
	assert.NoError(t, err)
	assert.False(t, fh.closed)

	// Check flush does nothing if called again
	err = fh.Flush()
	assert.NoError(t, err)
	assert.False(t, fh.closed)

	// Check that Close closes the file
	err = fh.Close()
	assert.NoError(t, err)
	assert.True(t, fh.closed)
}

func TestRWFileHandleReleaseWrite(t *testing.T) {
	r := fstest.NewRun(t)
	vfs, fh := rwHandleCreateWriteOnly(t, r)
	defer cleanup(t, r, vfs)

	// Write some data
	n, err := fh.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)

	// Check Release closes file
	err = fh.Release()
	assert.NoError(t, err)
	assert.True(t, fh.closed)

	// Check Release does nothing if called again
	err = fh.Release()
	assert.NoError(t, err)
	assert.True(t, fh.closed)
}

func testRWFileHandleOpenTest(t *testing.T, vfs *VFS, test *openTest) {
	fileName := "open-test-file"

	// first try with file not existing
	_, err := vfs.Stat(fileName)
	require.True(t, os.IsNotExist(err), test.what)

	f, openNonExistentErr := vfs.OpenFile(fileName, test.flags, 0666)

	var readNonExistentErr error
	var writeNonExistentErr error
	if openNonExistentErr == nil {
		// read some bytes
		buf := []byte{0, 0}
		_, readNonExistentErr = f.Read(buf)

		// write some bytes
		_, writeNonExistentErr = f.Write([]byte("hello"))

		// close
		err = f.Close()
		require.NoError(t, err, test.what)
	}

	// write the file
	f, err = vfs.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	require.NoError(t, err, test.what)
	_, err = f.Write([]byte("hello"))
	require.NoError(t, err, test.what)
	err = f.Close()
	require.NoError(t, err, test.what)

	// then open file and try with file existing

	f, openExistingErr := vfs.OpenFile(fileName, test.flags, 0666)
	var readExistingErr error
	var writeExistingErr error
	if openExistingErr == nil {
		// read some bytes
		buf := []byte{0, 0}
		_, readExistingErr = f.Read(buf)

		// write some bytes
		_, writeExistingErr = f.Write([]byte("HEL"))

		// close
		err = f.Close()
		require.NoError(t, err, test.what)
	}

	// read the file
	f, err = vfs.OpenFile(fileName, os.O_RDONLY, 0)
	require.NoError(t, err, test.what)
	buf, err := ioutil.ReadAll(f)
	require.NoError(t, err, test.what)
	err = f.Close()
	require.NoError(t, err, test.what)
	contents := string(buf)

	// remove file
	node, err := vfs.Stat(fileName)
	require.NoError(t, err, test.what)
	err = node.Remove()
	require.NoError(t, err, test.what)

	// check
	assert.Equal(t, test.openNonExistentErr, openNonExistentErr, "openNonExistentErr: %s: want=%v, got=%v", test.what, test.openNonExistentErr, openNonExistentErr)
	assert.Equal(t, test.readNonExistentErr, readNonExistentErr, "readNonExistentErr: %s: want=%v, got=%v", test.what, test.readNonExistentErr, readNonExistentErr)
	assert.Equal(t, test.writeNonExistentErr, writeNonExistentErr, "writeNonExistentErr: %s: want=%v, got=%v", test.what, test.writeNonExistentErr, writeNonExistentErr)
	assert.Equal(t, test.openExistingErr, openExistingErr, "openExistingErr: %s: want=%v, got=%v", test.what, test.openExistingErr, openExistingErr)
	assert.Equal(t, test.readExistingErr, readExistingErr, "readExistingErr: %s: want=%v, got=%v", test.what, test.readExistingErr, readExistingErr)
	assert.Equal(t, test.writeExistingErr, writeExistingErr, "writeExistingErr: %s: want=%v, got=%v", test.what, test.writeExistingErr, writeExistingErr)
	assert.Equal(t, test.contents, contents, test.what)
}

func TestRWFileHandleOpenTests(t *testing.T) {
	r := fstest.NewRun(t)
	opt := DefaultOpt
	opt.CacheMode = CacheModeFull
	vfs := New(r.Fremote, &opt)
	defer cleanup(t, r, vfs)

	for _, test := range openTests {
		testRWFileHandleOpenTest(t, vfs, &test)
	}
}

// tests mod time on open files
func TestRWFileModTimeWithOpenWriters(t *testing.T) {
	r := fstest.NewRun(t)
	defer r.Finalise()
	if !canSetModTime(t, r) {
		return
	}
	vfs, fh := rwHandleCreateWriteOnly(t, r)

	mtime := time.Date(2012, time.November, 18, 17, 32, 31, 0, time.UTC)

	_, err := fh.Write([]byte{104, 105})
	require.NoError(t, err)

	err = fh.Node().SetModTime(mtime)
	require.NoError(t, err)

	err = fh.Close()
	require.NoError(t, err)

	info, err := vfs.Stat("file1")
	require.NoError(t, err)

	if r.Fremote.Precision() != fs.ModTimeNotSupported {
		// avoid errors because of timezone differences
		assert.Equal(t, info.ModTime().Unix(), mtime.Unix())
	}
}

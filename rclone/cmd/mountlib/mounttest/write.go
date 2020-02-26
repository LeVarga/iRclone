package mounttest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rclone/rclone/vfs"
)

// TestWriteFileNoWrite tests writing a file with no write()'s to it
func TestWriteFileNoWrite(t *testing.T) {
	run.skipIfNoFUSE(t)

	fd, err := osCreate(run.path("testnowrite"))
	assert.NoError(t, err)

	err = fd.Close()
	assert.NoError(t, err)

	run.waitForWriters()

	run.checkDir(t, "testnowrite 0")

	run.rm(t, "testnowrite")
}

// FIXMETestWriteOpenFileInDirListing tests open file in directory listing
func FIXMETestWriteOpenFileInDirListing(t *testing.T) {
	run.skipIfNoFUSE(t)

	fd, err := osCreate(run.path("testnowrite"))
	assert.NoError(t, err)

	run.checkDir(t, "testnowrite 0")

	err = fd.Close()
	assert.NoError(t, err)

	run.waitForWriters()

	run.rm(t, "testnowrite")
}

// TestWriteFileWrite tests writing a file and reading it back
func TestWriteFileWrite(t *testing.T) {
	run.skipIfNoFUSE(t)

	run.createFile(t, "testwrite", "data")
	run.checkDir(t, "testwrite 4")
	contents := run.readFile(t, "testwrite")
	assert.Equal(t, "data", contents)
	run.rm(t, "testwrite")
}

// TestWriteFileOverwrite tests overwriting a file
func TestWriteFileOverwrite(t *testing.T) {
	run.skipIfNoFUSE(t)

	run.createFile(t, "testwrite", "data")
	run.checkDir(t, "testwrite 4")
	run.createFile(t, "testwrite", "potato")
	contents := run.readFile(t, "testwrite")
	assert.Equal(t, "potato", contents)
	run.rm(t, "testwrite")
}

// TestWriteFileFsync tests Fsync
//
// NB the code for this is in file.go rather than write.go
func TestWriteFileFsync(t *testing.T) {
	run.skipIfNoFUSE(t)

	filepath := run.path("to be synced")
	fd, err := osCreate(filepath)
	require.NoError(t, err)
	_, err = fd.Write([]byte("hello"))
	require.NoError(t, err)
	err = fd.Sync()
	require.NoError(t, err)
	err = fd.Close()
	require.NoError(t, err)
	run.waitForWriters()
	run.rm(t, "to be synced")
}

// TestWriteFileDup tests behavior of mmap() in Python by using dup() on a file handle
func TestWriteFileDup(t *testing.T) {
	run.skipIfNoFUSE(t)

	if run.vfs.Opt.CacheMode < vfs.CacheModeWrites {
		t.Skip("not supported on vfs-cache-mode < writes")
		return
	}

	filepath := run.path("to be synced")
	fh, err := osCreate(filepath)
	require.NoError(t, err)

	testData := []byte("0123456789")

	err = fh.Truncate(int64(len(testData) + 2))
	require.NoError(t, err)

	err = fh.Sync()
	require.NoError(t, err)

	var dupFd uintptr
	dupFd, err = writeTestDup(fh.Fd())
	require.NoError(t, err)

	dupFile := os.NewFile(dupFd, fh.Name())
	_, err = dupFile.Write(testData)
	require.NoError(t, err)

	err = dupFile.Close()
	require.NoError(t, err)

	_, err = fh.Seek(int64(len(testData)), 0)
	require.NoError(t, err)

	_, err = fh.Write([]byte("10"))
	require.NoError(t, err)

	err = fh.Close()
	require.NoError(t, err)

	run.waitForWriters()
	run.rm(t, "to be synced")
}

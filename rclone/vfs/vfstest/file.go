package vfstest

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/rclone/rclone/vfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileModTime tests mod times on files
func TestFileModTime(t *testing.T) {
	run.skipIfNoFUSE(t)

	run.createFile(t, "file", "123")

	mtime := time.Date(2012, time.November, 18, 17, 32, 31, 0, time.UTC)
	err := run.os.Chtimes(run.path("file"), mtime, mtime)
	require.NoError(t, err)

	info, err := run.os.Stat(run.path("file"))
	require.NoError(t, err)

	// avoid errors because of timezone differences
	assert.Equal(t, info.ModTime().Unix(), mtime.Unix())

	run.rm(t, "file")
}

// run.os.Create without opening for write too
func osCreate(name string) (vfs.OsFiler, error) {
	return run.os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}

// run.os.Create with append
func osAppend(name string) (vfs.OsFiler, error) {
	return run.os.OpenFile(name, os.O_WRONLY|os.O_APPEND, 0666)
}

// TestFileModTimeWithOpenWriters tests mod time on open files
func TestFileModTimeWithOpenWriters(t *testing.T) {
	run.skipIfNoFUSE(t)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	mtime := time.Date(2012, time.November, 18, 17, 32, 31, 0, time.UTC)
	filepath := run.path("cp-archive-test")

	f, err := osCreate(filepath)
	require.NoError(t, err)

	_, err = f.Write([]byte{104, 105})
	require.NoError(t, err)

	err = run.os.Chtimes(filepath, mtime, mtime)
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

	run.waitForWriters()

	info, err := run.os.Stat(filepath)
	require.NoError(t, err)

	// avoid errors because of timezone differences
	assert.Equal(t, info.ModTime().Unix(), mtime.Unix())

	run.rm(t, "cp-archive-test")
}

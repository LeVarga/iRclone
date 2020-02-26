// Test suite for rclonefs

package mounttest

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "github.com/rclone/rclone/backend/all" // import all the backends
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/walk"
	"github.com/rclone/rclone/fstest"
	"github.com/rclone/rclone/vfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	// UnmountFn is called to unmount the file system
	UnmountFn func() error
	// MountFn is called to mount the file system
	MountFn func(f fs.Fs, mountpoint string) (*vfs.VFS, <-chan error, func() error, error)
)

var (
	mountFn MountFn
)

// RunTests runs all the tests against all the VFS cache modes
func RunTests(t *testing.T, fn MountFn) {
	mountFn = fn
	flag.Parse()
	cacheModes := []vfs.CacheMode{
		vfs.CacheModeOff,
		vfs.CacheModeMinimal,
		vfs.CacheModeWrites,
		vfs.CacheModeFull,
	}
	run = newRun()
	for _, cacheMode := range cacheModes {
		run.cacheMode(cacheMode)
		log.Printf("Starting test run with cache mode %v", cacheMode)
		ok := t.Run(fmt.Sprintf("CacheMode=%v", cacheMode), func(t *testing.T) {
			t.Run("TestTouchAndDelete", TestTouchAndDelete)
			t.Run("TestRenameOpenHandle", TestRenameOpenHandle)
			t.Run("TestDirLs", TestDirLs)
			t.Run("TestDirCreateAndRemoveDir", TestDirCreateAndRemoveDir)
			t.Run("TestDirCreateAndRemoveFile", TestDirCreateAndRemoveFile)
			t.Run("TestDirRenameFile", TestDirRenameFile)
			t.Run("TestDirRenameEmptyDir", TestDirRenameEmptyDir)
			t.Run("TestDirRenameFullDir", TestDirRenameFullDir)
			t.Run("TestDirModTime", TestDirModTime)
			t.Run("TestDirCacheFlush", TestDirCacheFlush)
			t.Run("TestDirCacheFlushOnDirRename", TestDirCacheFlushOnDirRename)
			t.Run("TestFileModTime", TestFileModTime)
			t.Run("TestFileModTimeWithOpenWriters", TestFileModTimeWithOpenWriters)
			t.Run("TestMount", TestMount)
			t.Run("TestRoot", TestRoot)
			t.Run("TestReadByByte", TestReadByByte)
			t.Run("TestReadChecksum", TestReadChecksum)
			t.Run("TestReadFileDoubleClose", TestReadFileDoubleClose)
			t.Run("TestReadSeek", TestReadSeek)
			t.Run("TestWriteFileNoWrite", TestWriteFileNoWrite)
			t.Run("TestWriteFileWrite", TestWriteFileWrite)
			t.Run("TestWriteFileOverwrite", TestWriteFileOverwrite)
			t.Run("TestWriteFileDoubleClose", TestWriteFileDoubleClose)
			t.Run("TestWriteFileFsync", TestWriteFileFsync)
			t.Run("TestWriteFileDup", TestWriteFileDup)
		})
		log.Printf("Finished test run with cache mode %v (ok=%v)", cacheMode, ok)
		if !ok {
			break
		}
	}
	run.Finalise()
}

// Run holds the remotes for a test run
type Run struct {
	vfs          *vfs.VFS
	mountPath    string
	fremote      fs.Fs
	fremoteName  string
	cleanRemote  func()
	umountResult <-chan error
	umountFn     UnmountFn
	skip         bool
}

// run holds the master Run data
var run *Run

// newRun initialise the remote mount for testing and returns a run
// object.
//
// r.fremote is an empty remote Fs
//
// Finalise() will tidy them away when done.
func newRun() *Run {
	r := &Run{
		umountResult: make(chan error, 1),
	}

	fstest.Initialise()

	var err error
	r.fremote, r.fremoteName, r.cleanRemote, err = fstest.RandomRemote()
	if err != nil {
		log.Fatalf("Failed to open remote %q: %v", *fstest.RemoteName, err)
	}

	err = r.fremote.Mkdir(context.Background(), "")
	if err != nil {
		log.Fatalf("Failed to open mkdir %q: %v", *fstest.RemoteName, err)
	}

	r.mountPath = findMountPath()
	// Mount it up
	r.mount()

	return r
}

func findMountPath() string {
	if runtime.GOOS != "windows" {
		mountPath, err := ioutil.TempDir("", "rclonefs-mount")
		if err != nil {
			log.Fatalf("Failed to create mount dir: %v", err)
		}
		return mountPath
	}

	// Find a free drive letter
	drive := ""
	for letter := 'E'; letter <= 'Z'; letter++ {
		drive = string(letter) + ":"
		_, err := os.Stat(drive + "\\")
		if os.IsNotExist(err) {
			goto found
		}
	}
	log.Fatalf("Couldn't find free drive letter for test")
found:
	return drive
}

func (r *Run) mount() {
	log.Printf("mount %q %q", r.fremote, r.mountPath)
	var err error
	r.vfs, r.umountResult, r.umountFn, err = mountFn(r.fremote, r.mountPath)
	if err != nil {
		log.Printf("mount FAILED: %v", err)
		r.skip = true
	} else {
		log.Printf("mount OK")
	}
}

func (r *Run) umount() {
	if r.skip {
		log.Printf("FUSE not found so skipping umount")
		return
	}
	/*
		log.Printf("Calling fusermount -u %q", r.mountPath)
		err := exec.Command("fusermount", "-u", r.mountPath).Run()
		if err != nil {
			log.Printf("fusermount failed: %v", err)
		}
	*/
	log.Printf("Unmounting %q", r.mountPath)
	err := r.umountFn()
	if err != nil {
		log.Printf("signal to umount failed - retrying: %v", err)
		time.Sleep(3 * time.Second)
		err = r.umountFn()
	}
	if err != nil {
		log.Fatalf("signal to umount failed: %v", err)
	}
	log.Printf("Waiting for umount")
	err = <-r.umountResult
	if err != nil {
		log.Fatalf("umount failed: %v", err)
	}

	// Cleanup the VFS cache - umount has called Shutdown
	err = r.vfs.CleanUp()
	if err != nil {
		log.Printf("Failed to cleanup the VFS cache: %v", err)
	}
}

// cacheMode flushes the VFS and changes the CacheMode
func (r *Run) cacheMode(cacheMode vfs.CacheMode) {
	if r.skip {
		log.Printf("FUSE not found so skipping cacheMode")
		return
	}
	// Wait for writers to finish
	r.vfs.WaitForWriters(30 * time.Second)
	// Empty and remake the remote
	r.cleanRemote()
	err := r.fremote.Mkdir(context.Background(), "")
	if err != nil {
		log.Fatalf("Failed to open mkdir %q: %v", *fstest.RemoteName, err)
	}
	// Empty the cache
	err = r.vfs.CleanUp()
	if err != nil {
		log.Printf("Failed to cleanup the VFS cache: %v", err)
	}
	// Reset the cache mode
	r.vfs.SetCacheMode(cacheMode)
	// Flush the directory cache
	r.vfs.FlushDirCache()

}

func (r *Run) skipIfNoFUSE(t *testing.T) {
	if r.skip {
		t.Skip("FUSE not found so skipping test")
	}
}

// Finalise cleans the remote and unmounts
func (r *Run) Finalise() {
	r.umount()
	r.cleanRemote()
	err := os.RemoveAll(r.mountPath)
	if err != nil {
		log.Printf("Failed to clean mountPath %q: %v", r.mountPath, err)
	}
}

// path returns an OS local path for filepath
func (r *Run) path(filePath string) string {
	// return windows drive letter root as E:\
	if filePath == "" && runtime.GOOS == "windows" {
		return run.mountPath + `\`
	}
	return filepath.Join(run.mountPath, filepath.FromSlash(filePath))
}

type dirMap map[string]struct{}

// Create a dirMap from a string
func newDirMap(dirString string) (dm dirMap) {
	dm = make(dirMap)
	for _, entry := range strings.Split(dirString, "|") {
		if entry != "" {
			dm[entry] = struct{}{}
		}
	}
	return dm
}

// Returns a dirmap with only the files in
func (dm dirMap) filesOnly() dirMap {
	newDm := make(dirMap)
	for name := range dm {
		if !strings.HasSuffix(name, "/") {
			newDm[name] = struct{}{}
		}
	}
	return newDm
}

// reads the local tree into dir
func (r *Run) readLocal(t *testing.T, dir dirMap, filePath string) {
	realPath := r.path(filePath)
	files, err := ioutil.ReadDir(realPath)
	require.NoError(t, err)
	for _, fi := range files {
		name := path.Join(filePath, fi.Name())
		if fi.IsDir() {
			dir[name+"/"] = struct{}{}
			r.readLocal(t, dir, name)
			assert.Equal(t, run.vfs.Opt.DirPerms&os.ModePerm, fi.Mode().Perm())
		} else {
			dir[fmt.Sprintf("%s %d", name, fi.Size())] = struct{}{}
			assert.Equal(t, run.vfs.Opt.FilePerms&os.ModePerm, fi.Mode().Perm())
		}
	}
}

// reads the remote tree into dir
func (r *Run) readRemote(t *testing.T, dir dirMap, filepath string) {
	objs, dirs, err := walk.GetAll(context.Background(), r.fremote, filepath, true, 1)
	if err == fs.ErrorDirNotFound {
		return
	}
	require.NoError(t, err)
	for _, obj := range objs {
		dir[fmt.Sprintf("%s %d", obj.Remote(), obj.Size())] = struct{}{}
	}
	for _, d := range dirs {
		name := d.Remote()
		dir[name+"/"] = struct{}{}
		r.readRemote(t, dir, name)
	}
}

// checkDir checks the local and remote against the string passed in
func (r *Run) checkDir(t *testing.T, dirString string) {
	var retries = *fstest.ListRetries
	sleep := time.Second / 5
	var remoteOK, fuseOK bool
	var dm, localDm, remoteDm dirMap
	for i := 1; i <= retries; i++ {
		dm = newDirMap(dirString)
		localDm = make(dirMap)
		r.readLocal(t, localDm, "")
		remoteDm = make(dirMap)
		r.readRemote(t, remoteDm, "")
		// Ignore directories for remote compare
		remoteOK = reflect.DeepEqual(dm.filesOnly(), remoteDm.filesOnly())
		fuseOK = reflect.DeepEqual(dm, localDm)
		if remoteOK && fuseOK {
			return
		}
		sleep *= 2
		t.Logf("Sleeping for %v for list eventual consistency: %d/%d", sleep, i, retries)
		time.Sleep(sleep)
	}
	assert.Equal(t, dm.filesOnly(), remoteDm.filesOnly(), "expected vs remote")
	assert.Equal(t, dm, localDm, "expected vs fuse mount")
}

// wait for any files being written to be released by fuse
func (r *Run) waitForWriters() {
	run.vfs.WaitForWriters(10 * time.Second)
}

func (r *Run) createFile(t *testing.T, filepath string, contents string) {
	filepath = r.path(filepath)
	err := ioutil.WriteFile(filepath, []byte(contents), 0600)
	require.NoError(t, err)
	r.waitForWriters()
}

func (r *Run) readFile(t *testing.T, filepath string) string {
	filepath = r.path(filepath)
	result, err := ioutil.ReadFile(filepath)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond) // FIXME wait for Release
	return string(result)
}

func (r *Run) mkdir(t *testing.T, filepath string) {
	filepath = r.path(filepath)
	err := os.Mkdir(filepath, 0700)
	require.NoError(t, err)
}

func (r *Run) rm(t *testing.T, filepath string) {
	filepath = r.path(filepath)
	err := os.Remove(filepath)
	require.NoError(t, err)

	// Wait for file to disappear from listing
	for i := 0; i < 100; i++ {
		_, err := os.Stat(filepath)
		if os.IsNotExist(err) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.Fail(t, "failed to delete file", filepath)
}

func (r *Run) rmdir(t *testing.T, filepath string) {
	filepath = r.path(filepath)
	err := os.Remove(filepath)
	require.NoError(t, err)
}

// TestMount checks that the Fs is mounted by seeing if the mountpoint
// is in the mount output
func TestMount(t *testing.T) {
	run.skipIfNoFUSE(t)
	if runtime.GOOS == "windows" {
		t.Skip("not running on windows")
	}

	out, err := exec.Command("mount").Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), run.mountPath)
}

// TestRoot checks root directory is present and correct
func TestRoot(t *testing.T) {
	run.skipIfNoFUSE(t)

	fi, err := os.Lstat(run.mountPath)
	require.NoError(t, err)
	assert.True(t, fi.IsDir())
	assert.Equal(t, run.vfs.Opt.DirPerms&os.ModePerm, fi.Mode().Perm())
}

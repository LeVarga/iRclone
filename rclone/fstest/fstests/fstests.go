// Package fstests provides generic integration tests for the Fs and
// Object interfaces.
//
// These tests are concerned with the basic functionality of a
// backend.  The tests in fs/sync and fs/operations tests more
// cornercases that these tests don't.
package fstests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/bits"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config"
	"github.com/rclone/rclone/fs/fserrors"
	"github.com/rclone/rclone/fs/fspath"
	"github.com/rclone/rclone/fs/hash"
	"github.com/rclone/rclone/fs/object"
	"github.com/rclone/rclone/fs/operations"
	"github.com/rclone/rclone/fs/walk"
	"github.com/rclone/rclone/fstest"
	"github.com/rclone/rclone/lib/encoder"
	"github.com/rclone/rclone/lib/random"
	"github.com/rclone/rclone/lib/readers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// InternalTester is an optional interface for Fs which allows to execute internal tests
//
// This interface should be implemented in 'backend'_internal_test.go and not in 'backend'.go
type InternalTester interface {
	InternalTest(*testing.T)
}

// ChunkedUploadConfig contains the values used by TestFsPutChunked
// to determine the limits of chunked uploading
type ChunkedUploadConfig struct {
	// Minimum allowed chunk size
	MinChunkSize fs.SizeSuffix
	// Maximum allowed chunk size, 0 is no limit
	MaxChunkSize fs.SizeSuffix
	// Rounds the given chunk size up to the next valid value
	// nil will disable rounding
	// e.g. the next power of 2
	CeilChunkSize func(fs.SizeSuffix) fs.SizeSuffix
	// More than one chunk is required on upload
	NeedMultipleChunks bool
}

// SetUploadChunkSizer is a test only interface to change the upload chunk size at runtime
type SetUploadChunkSizer interface {
	// Change the configured UploadChunkSize.
	// Will only be called while no transfer is in progress.
	SetUploadChunkSize(fs.SizeSuffix) (fs.SizeSuffix, error)
}

// SetUploadCutoffer is a test only interface to change the upload cutoff size at runtime
type SetUploadCutoffer interface {
	// Change the configured UploadCutoff.
	// Will only be called while no transfer is in progress.
	SetUploadCutoff(fs.SizeSuffix) (fs.SizeSuffix, error)
}

// NextPowerOfTwo returns the current or next bigger power of two.
// All values less or equal 0 will return 0
func NextPowerOfTwo(i fs.SizeSuffix) fs.SizeSuffix {
	return 1 << uint(64-bits.LeadingZeros64(uint64(i)-1))
}

// NextMultipleOf returns a function that can be used as a CeilChunkSize function.
// This function will return the next multiple of m that is equal or bigger than i.
// All values less or equal 0 will return 0.
func NextMultipleOf(m fs.SizeSuffix) func(fs.SizeSuffix) fs.SizeSuffix {
	if m <= 0 {
		panic(fmt.Sprintf("invalid multiplier %s", m))
	}
	return func(i fs.SizeSuffix) fs.SizeSuffix {
		if i <= 0 {
			return 0
		}

		return (((i - 1) / m) + 1) * m
	}
}

// dirsToNames returns a sorted list of names
func dirsToNames(dirs []fs.Directory) []string {
	names := []string{}
	for _, dir := range dirs {
		names = append(names, fstest.Normalize(dir.Remote()))
	}
	sort.Strings(names)
	return names
}

// objsToNames returns a sorted list of object names
func objsToNames(objs []fs.Object) []string {
	names := []string{}
	for _, obj := range objs {
		names = append(names, fstest.Normalize(obj.Remote()))
	}
	sort.Strings(names)
	return names
}

// findObject finds the object on the remote
func findObject(ctx context.Context, t *testing.T, f fs.Fs, Name string) fs.Object {
	var obj fs.Object
	var err error
	sleepTime := 1 * time.Second
	for i := 1; i <= *fstest.ListRetries; i++ {
		obj, err = f.NewObject(ctx, Name)
		if err == nil {
			break
		}
		t.Logf("Sleeping for %v for findObject eventual consistency: %d/%d (%v)", sleepTime, i, *fstest.ListRetries, err)
		time.Sleep(sleepTime)
		sleepTime = (sleepTime * 3) / 2
	}
	require.NoError(t, err)
	return obj
}

// retry f() until no retriable error
func retry(t *testing.T, what string, f func() error) {
	const maxTries = 10
	var err error
	for tries := 1; tries <= maxTries; tries++ {
		err = f()
		// exit if no error, or error is not retriable
		if err == nil || !fserrors.IsRetryError(err) {
			break
		}
		t.Logf("%s error: %v - low level retry %d/%d", what, err, tries, maxTries)
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err, what)
}

// testPut puts file with random contents to the remote
func testPut(ctx context.Context, t *testing.T, f fs.Fs, file *fstest.Item) (string, fs.Object) {
	return PutTestContents(ctx, t, f, file, random.String(100), true)
}

// PutTestContents puts file with given contents to the remote and checks it but unlike TestPutLarge doesn't remove
func PutTestContents(ctx context.Context, t *testing.T, f fs.Fs, file *fstest.Item, contents string, check bool) (string, fs.Object) {
	var (
		err        error
		obj        fs.Object
		uploadHash *hash.MultiHasher
	)
	retry(t, "Put", func() error {
		buf := bytes.NewBufferString(contents)
		uploadHash = hash.NewMultiHasher()
		in := io.TeeReader(buf, uploadHash)

		file.Size = int64(buf.Len())
		obji := object.NewStaticObjectInfo(file.Path, file.ModTime, file.Size, true, nil, nil)
		obj, err = f.Put(ctx, in, obji)
		return err
	})
	file.Hashes = uploadHash.Sums()
	if check {
		file.Check(t, obj, f.Precision())
		// Re-read the object and check again
		obj = findObject(ctx, t, f, file.Path)
		file.Check(t, obj, f.Precision())
	}
	return contents, obj
}

// TestPutLarge puts file to the remote, checks it and removes it on success.
func TestPutLarge(ctx context.Context, t *testing.T, f fs.Fs, file *fstest.Item) {
	var (
		err        error
		obj        fs.Object
		uploadHash *hash.MultiHasher
	)
	retry(t, "PutLarge", func() error {
		r := readers.NewPatternReader(file.Size)
		uploadHash = hash.NewMultiHasher()
		in := io.TeeReader(r, uploadHash)

		obji := object.NewStaticObjectInfo(file.Path, file.ModTime, file.Size, true, nil, nil)
		obj, err = f.Put(ctx, in, obji)
		if file.Size == 0 && err == fs.ErrorCantUploadEmptyFiles {
			t.Skip("Can't upload zero length files")
		}
		return err
	})
	file.Hashes = uploadHash.Sums()
	file.Check(t, obj, f.Precision())

	// Re-read the object and check again
	obj = findObject(ctx, t, f, file.Path)
	file.Check(t, obj, f.Precision())

	// Download the object and check it is OK
	downloadHash := hash.NewMultiHasher()
	download, err := obj.Open(ctx)
	require.NoError(t, err)
	n, err := io.Copy(downloadHash, download)
	require.NoError(t, err)
	assert.Equal(t, file.Size, n)
	require.NoError(t, download.Close())
	assert.Equal(t, file.Hashes, downloadHash.Sums())

	// Remove the object
	require.NoError(t, obj.Remove(ctx))
}

// errorReader just returns an error on Read
type errorReader struct {
	err error
}

// Read returns an error immediately
func (er errorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

// read the contents of an object as a string
func readObject(ctx context.Context, t *testing.T, obj fs.Object, limit int64, options ...fs.OpenOption) string {
	what := fmt.Sprintf("readObject(%q) limit=%d, options=%+v", obj, limit, options)
	in, err := obj.Open(ctx, options...)
	require.NoError(t, err, what)
	var r io.Reader = in
	if limit >= 0 {
		r = &io.LimitedReader{R: r, N: limit}
	}
	contents, err := ioutil.ReadAll(r)
	require.NoError(t, err, what)
	err = in.Close()
	require.NoError(t, err, what)
	return string(contents)
}

// ExtraConfigItem describes a config item for the tests
type ExtraConfigItem struct{ Name, Key, Value string }

// Opt is options for Run
type Opt struct {
	RemoteName                   string
	NilObject                    fs.Object
	ExtraConfig                  []ExtraConfigItem
	SkipBadWindowsCharacters     bool     // skips unusable characters for windows if set
	SkipFsMatch                  bool     // if set skip exact matching of Fs value
	TiersToTest                  []string // List of tiers which can be tested in setTier test
	ChunkedUpload                ChunkedUploadConfig
	UnimplementableFsMethods     []string // List of methods which can't be implemented in this wrapping Fs
	UnimplementableObjectMethods []string // List of methods which can't be implemented in this wrapping Fs
	SkipFsCheckWrap              bool     // if set skip FsCheckWrap
	SkipObjectCheckWrap          bool     // if set skip ObjectCheckWrap
	SkipInvalidUTF8              bool     // if set skip invalid UTF-8 checks
}

// returns true if x is found in ss
func stringsContains(x string, ss []string) bool {
	for _, s := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// Run runs the basic integration tests for a remote using the options passed in.
//
// They are structured in a hierarchical way so that dependencies for the tests can be created.
//
// For example some tests require the directory to be created - these
// are inside the "FsMkdir" test.  Some tests require some tests files
// - these are inside the "FsPutFiles" test.
func Run(t *testing.T, opt *Opt) {
	var (
		remote        fs.Fs
		remoteName    = opt.RemoteName
		subRemoteName string
		subRemoteLeaf string
		file1         = fstest.Item{
			ModTime: fstest.Time("2001-02-03T04:05:06.499999999Z"),
			Path:    "file name.txt",
		}
		file1Contents string
		file2         = fstest.Item{
			ModTime: fstest.Time("2001-02-03T04:05:10.123123123Z"),
			Path:    `hello? sausage/êé/Hello, 世界/ " ' @ < > & ? + ≠/z.txt`,
		}
		isLocalRemote bool
		purged        bool // whether the dir has been purged or not
		ctx           = context.Background()
	)

	// Skip the test if the remote isn't configured
	skipIfNotOk := func(t *testing.T) {
		if remote == nil {
			t.Skipf("WARN: %q not configured", remoteName)
		}
	}

	// Skip if remote is not ListR capable, otherwise set the useListR
	// flag, returning a function to restore its value
	skipIfNotListR := func(t *testing.T) func() {
		skipIfNotOk(t)
		if remote.Features().ListR == nil {
			t.Skip("FS has no ListR interface")
		}
		previous := fs.Config.UseListR
		fs.Config.UseListR = true
		return func() {
			fs.Config.UseListR = previous
		}
	}

	// Skip if remote is not SetTier and GetTier capable
	skipIfNotSetTier := func(t *testing.T) {
		skipIfNotOk(t)
		if remote.Features().SetTier == false ||
			remote.Features().GetTier == false {
			t.Skip("FS has no SetTier & GetTier interfaces")
		}
	}

	// Return true if f (or any of the things it wraps) is bucket
	// based but not at the root.
	isBucketBasedButNotRoot := func(f fs.Fs) bool {
		return fs.UnWrapFs(f).Features().BucketBased && strings.Contains(strings.Trim(f.Root(), "/"), "/")
	}

	// Initialise the remote
	fstest.Initialise()

	// Set extra config if supplied
	for _, item := range opt.ExtraConfig {
		config.FileSet(item.Name, item.Key, item.Value)
	}
	if *fstest.RemoteName != "" {
		remoteName = *fstest.RemoteName
	}
	fstest.RemoteName = &remoteName
	t.Logf("Using remote %q", remoteName)
	var err error
	if remoteName == "" {
		remoteName, err = fstest.LocalRemote()
		require.NoError(t, err)
		isLocalRemote = true
	}

	// Make the Fs we are testing with, initialising the local variables
	// subRemoteName - name of the remote after the TestRemote:
	// subRemoteLeaf - a subdirectory to use under that
	// remote - the result of  fs.NewFs(TestRemote:subRemoteName)
	subRemoteName, subRemoteLeaf, err = fstest.RandomRemoteName(remoteName)
	require.NoError(t, err)
	remote, err = fs.NewFs(subRemoteName)
	if err == fs.ErrorNotFoundInConfigFile {
		t.Logf("Didn't find %q in config file - skipping tests", remoteName)
		return
	}
	require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))

	// Skip the rest if it failed
	skipIfNotOk(t)

	// Check to see if Fs that wrap other Fs implement all the optional methods
	t.Run("FsCheckWrap", func(t *testing.T) {
		skipIfNotOk(t)
		if opt.SkipFsCheckWrap {
			t.Skip("Skipping FsCheckWrap on this Fs")
		}
		ft := new(fs.Features).Fill(remote)
		if ft.UnWrap == nil {
			t.Skip("Not a wrapping Fs")
		}
		v := reflect.ValueOf(ft).Elem()
		vType := v.Type()
		for i := 0; i < v.NumField(); i++ {
			vName := vType.Field(i).Name
			if stringsContains(vName, opt.UnimplementableFsMethods) {
				continue
			}
			field := v.Field(i)
			// skip the bools
			if field.Type().Kind() == reflect.Bool {
				continue
			}
			if field.IsNil() {
				t.Errorf("Missing Fs wrapper for %s", vName)
			}
		}
	})

	// TestFsRmdirNotFound tests deleting a non existent directory
	t.Run("FsRmdirNotFound", func(t *testing.T) {
		skipIfNotOk(t)
		if isBucketBasedButNotRoot(remote) {
			t.Skip("Skipping test as non root bucket based remote")
		}
		err := remote.Rmdir(ctx, "")
		assert.Error(t, err, "Expecting error on Rmdir non existent")
	})

	// Make the directory
	err = remote.Mkdir(ctx, "")
	require.NoError(t, err)
	fstest.CheckListing(t, remote, []fstest.Item{})

	// TestFsString tests the String method
	t.Run("FsString", func(t *testing.T) {
		skipIfNotOk(t)
		str := remote.String()
		require.NotEqual(t, "", str)
	})

	// TestFsName tests the Name method
	t.Run("FsName", func(t *testing.T) {
		skipIfNotOk(t)
		got := remote.Name()
		want := remoteName[:strings.LastIndex(remoteName, ":")+1]
		if isLocalRemote {
			want = "local:"
		}
		require.Equal(t, want, got+":")
	})

	// TestFsRoot tests the Root method
	t.Run("FsRoot", func(t *testing.T) {
		skipIfNotOk(t)
		name := remote.Name() + ":"
		root := remote.Root()
		if isLocalRemote {
			// only check last path element on local
			require.Equal(t, filepath.Base(subRemoteName), filepath.Base(root))
		} else {
			require.Equal(t, subRemoteName, name+root)
		}
	})

	// TestFsRmdirEmpty tests deleting an empty directory
	t.Run("FsRmdirEmpty", func(t *testing.T) {
		skipIfNotOk(t)
		err := remote.Rmdir(ctx, "")
		require.NoError(t, err)
	})

	// TestFsMkdir tests making a directory
	//
	// Tests that require the directory to be made are within this
	t.Run("FsMkdir", func(t *testing.T) {
		skipIfNotOk(t)

		err := remote.Mkdir(ctx, "")
		require.NoError(t, err)
		fstest.CheckListing(t, remote, []fstest.Item{})

		err = remote.Mkdir(ctx, "")
		require.NoError(t, err)

		// TestFsMkdirRmdirSubdir tests making and removing a sub directory
		t.Run("FsMkdirRmdirSubdir", func(t *testing.T) {
			skipIfNotOk(t)
			dir := "dir/subdir"
			err := operations.Mkdir(ctx, remote, dir)
			require.NoError(t, err)
			fstest.CheckListingWithPrecision(t, remote, []fstest.Item{}, []string{"dir", "dir/subdir"}, fs.GetModifyWindow(remote))

			err = operations.Rmdir(ctx, remote, dir)
			require.NoError(t, err)
			fstest.CheckListingWithPrecision(t, remote, []fstest.Item{}, []string{"dir"}, fs.GetModifyWindow(remote))

			err = operations.Rmdir(ctx, remote, "dir")
			require.NoError(t, err)
			fstest.CheckListingWithPrecision(t, remote, []fstest.Item{}, []string{}, fs.GetModifyWindow(remote))
		})

		// TestFsListEmpty tests listing an empty directory
		t.Run("FsListEmpty", func(t *testing.T) {
			skipIfNotOk(t)
			fstest.CheckListing(t, remote, []fstest.Item{})
		})

		// TestFsListDirEmpty tests listing the directories from an empty directory
		TestFsListDirEmpty := func(t *testing.T) {
			skipIfNotOk(t)
			objs, dirs, err := walk.GetAll(ctx, remote, "", true, 1)
			if !remote.Features().CanHaveEmptyDirectories {
				if err != fs.ErrorDirNotFound {
					require.NoError(t, err)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, []string{}, objsToNames(objs))
			assert.Equal(t, []string{}, dirsToNames(dirs))
		}
		t.Run("FsListDirEmpty", TestFsListDirEmpty)

		// TestFsListRDirEmpty tests listing the directories from an empty directory using ListR
		t.Run("FsListRDirEmpty", func(t *testing.T) {
			defer skipIfNotListR(t)()
			TestFsListDirEmpty(t)
		})

		// TestFsListDirNotFound tests listing the directories from an empty directory
		TestFsListDirNotFound := func(t *testing.T) {
			skipIfNotOk(t)
			objs, dirs, err := walk.GetAll(ctx, remote, "does not exist", true, 1)
			if !remote.Features().CanHaveEmptyDirectories {
				if err != fs.ErrorDirNotFound {
					assert.NoError(t, err)
					assert.Equal(t, 0, len(objs)+len(dirs))
				}
			} else {
				assert.Equal(t, fs.ErrorDirNotFound, err)
			}
		}
		t.Run("FsListDirNotFound", TestFsListDirNotFound)

		// TestFsListRDirNotFound tests listing the directories from an empty directory using ListR
		t.Run("FsListRDirNotFound", func(t *testing.T) {
			defer skipIfNotListR(t)()
			TestFsListDirNotFound(t)
		})

		// FsEncoding tests that file name encodings are
		// working by uploading a series of unusual files
		// Must be run in an empty directory
		t.Run("FsEncoding", func(t *testing.T) {
			skipIfNotOk(t)

			// check no files or dirs as pre-requisite
			fstest.CheckListingWithPrecision(t, remote, []fstest.Item{}, []string{}, fs.GetModifyWindow(remote))

			for _, test := range []struct {
				name string
				path string
			}{
				// See lib/encoder/encoder.go for list of things that go here
				{"control chars", "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1A\x1B\x1C\x1D\x1E\x1F\x7F"},
				{"dot", "."},
				{"dot dot", ".."},
				{"punctuation", "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"},
				{"leading space", " leading space"},
				{"leading tilde", "~leading tilde"},
				{"leading CR", "\rleading CR"},
				{"leading LF", "\nleading LF"},
				{"leading HT", "\tleading HT"},
				{"leading VT", "\vleading VT"},
				{"leading dot", ".leading dot"},
				{"trailing space", "trailing space "},
				{"trailing CR", "trailing CR\r"},
				{"trailing LF", "trailing LF\n"},
				{"trailing HT", "trailing HT\t"},
				{"trailing VT", "trailing VT\v"},
				{"trailing dot", "trailing dot."},
				{"invalid UTF-8", "invalid utf-8\xfe"},
			} {
				t.Run(test.name, func(t *testing.T) {
					if opt.SkipInvalidUTF8 && test.name == "invalid UTF-8" {
						t.Skip("Skipping " + test.name)
					}
					// turn raw strings into Standard encoding
					fileName := encoder.Standard.Encode(test.path)
					dirName := fileName
					t.Logf("testing %q", fileName)
					assert.NoError(t, remote.Mkdir(ctx, dirName))
					file := fstest.Item{
						ModTime: time.Now(),
						Path:    dirName + "/" + fileName, // test creating a file and dir with that name
					}
					_, o := testPut(context.Background(), t, remote, &file)
					fstest.CheckListingWithPrecision(t, remote, []fstest.Item{file}, []string{dirName}, fs.GetModifyWindow(remote))
					assert.NoError(t, o.Remove(ctx))
					assert.NoError(t, remote.Rmdir(ctx, dirName))
					fstest.CheckListingWithPrecision(t, remote, []fstest.Item{}, []string{}, fs.GetModifyWindow(remote))
				})
			}
		})

		// TestFsNewObjectNotFound tests not finding a object
		t.Run("FsNewObjectNotFound", func(t *testing.T) {
			skipIfNotOk(t)
			// Object in an existing directory
			o, err := remote.NewObject(ctx, "potato")
			assert.Nil(t, o)
			assert.Equal(t, fs.ErrorObjectNotFound, err)
			// Now try an object in a non existing directory
			o, err = remote.NewObject(ctx, "directory/not/found/potato")
			assert.Nil(t, o)
			assert.Equal(t, fs.ErrorObjectNotFound, err)
		})

		// TestFsPutError tests uploading a file where there is an error
		//
		// It makes sure that aborting a file half way through does not create
		// a file on the remote.
		//
		// go test -v -run 'TestIntegration/Test(Setup|Init|FsMkdir|FsPutError)$'
		t.Run("FsPutError", func(t *testing.T) {
			skipIfNotOk(t)

			var N int64 = 5 * 1024
			if *fstest.SizeLimit > 0 && N > *fstest.SizeLimit {
				N = *fstest.SizeLimit
				t.Logf("Reduce file size due to limit %d", N)
			}

			// Read N bytes then produce an error
			contents := random.String(int(N))
			buf := bytes.NewBufferString(contents)
			er := &errorReader{errors.New("potato")}
			in := io.MultiReader(buf, er)

			obji := object.NewStaticObjectInfo(file2.Path, file2.ModTime, 2*N, true, nil, nil)
			_, err := remote.Put(ctx, in, obji)
			// assert.Nil(t, obj) - FIXME some remotes return the object even on nil
			assert.NotNil(t, err)

			obj, err := remote.NewObject(ctx, file2.Path)
			assert.Nil(t, obj)
			assert.Equal(t, fs.ErrorObjectNotFound, err)
		})

		t.Run("FsPutZeroLength", func(t *testing.T) {
			skipIfNotOk(t)

			TestPutLarge(ctx, t, remote, &fstest.Item{
				ModTime: fstest.Time("2001-02-03T04:05:06.499999999Z"),
				Path:    fmt.Sprintf("zero-length-file"),
				Size:    int64(0),
			})
		})

		t.Run("FsOpenWriterAt", func(t *testing.T) {
			skipIfNotOk(t)
			openWriterAt := remote.Features().OpenWriterAt
			if openWriterAt == nil {
				t.Skip("FS has no OpenWriterAt interface")
			}
			path := "writer-at-subdir/writer-at-file"
			out, err := openWriterAt(ctx, path, -1)
			require.NoError(t, err)

			var n int
			n, err = out.WriteAt([]byte("def"), 3)
			assert.NoError(t, err)
			assert.Equal(t, 3, n)
			n, err = out.WriteAt([]byte("ghi"), 6)
			assert.NoError(t, err)
			assert.Equal(t, 3, n)
			n, err = out.WriteAt([]byte("abc"), 0)
			assert.NoError(t, err)
			assert.Equal(t, 3, n)

			assert.NoError(t, out.Close())

			obj := findObject(ctx, t, remote, path)
			assert.Equal(t, "abcdefghi", readObject(ctx, t, obj, -1), "contents of file differ")

			assert.NoError(t, obj.Remove(ctx))
			assert.NoError(t, remote.Rmdir(ctx, "writer-at-subdir"))
		})

		// TestFsChangeNotify tests that changes are properly
		// propagated
		//
		// go test -v -remote TestDrive: -run '^Test(Setup|Init|FsChangeNotify)$' -verbose
		t.Run("FsChangeNotify", func(t *testing.T) {
			skipIfNotOk(t)

			// Check have ChangeNotify
			doChangeNotify := remote.Features().ChangeNotify
			if doChangeNotify == nil {
				t.Skip("FS has no ChangeNotify interface")
			}

			err := operations.Mkdir(ctx, remote, "dir")
			require.NoError(t, err)

			pollInterval := make(chan time.Duration)
			dirChanges := map[string]struct{}{}
			objChanges := map[string]struct{}{}
			doChangeNotify(ctx, func(x string, e fs.EntryType) {
				fs.Debugf(nil, "doChangeNotify(%q, %+v)", x, e)
				if strings.HasPrefix(x, file1.Path[:5]) || strings.HasPrefix(x, file2.Path[:5]) {
					fs.Debugf(nil, "Ignoring notify for file1 or file2: %q, %v", x, e)
					return
				}
				if e == fs.EntryDirectory {
					dirChanges[x] = struct{}{}
				} else if e == fs.EntryObject {
					objChanges[x] = struct{}{}
				}
			}, pollInterval)
			defer func() { close(pollInterval) }()
			pollInterval <- time.Second

			var dirs []string
			for _, idx := range []int{1, 3, 2} {
				dir := fmt.Sprintf("dir/subdir%d", idx)
				err = operations.Mkdir(ctx, remote, dir)
				require.NoError(t, err)
				dirs = append(dirs, dir)
			}

			var objs []fs.Object
			for _, idx := range []int{2, 4, 3} {
				file := fstest.Item{
					ModTime: time.Now(),
					Path:    fmt.Sprintf("dir/file%d", idx),
				}
				_, o := testPut(ctx, t, remote, &file)
				objs = append(objs, o)
			}

			// Looks for each item in wants in changes -
			// if they are all found it returns true
			contains := func(changes map[string]struct{}, wants []string) bool {
				for _, want := range wants {
					_, ok := changes[want]
					if !ok {
						return false
					}
				}
				return true
			}

			// Wait a little while for the changes to come in
			wantDirChanges := []string{"dir/subdir1", "dir/subdir3", "dir/subdir2"}
			wantObjChanges := []string{"dir/file2", "dir/file4", "dir/file3"}
			ok := false
			for tries := 1; tries < 10; tries++ {
				ok = contains(dirChanges, wantDirChanges) && contains(objChanges, wantObjChanges)
				if ok {
					break
				}
				t.Logf("Try %d/10 waiting for dirChanges and objChanges", tries)
				time.Sleep(3 * time.Second)
			}
			if !ok {
				t.Errorf("%+v does not contain %+v or \n%+v does not contain %+v", dirChanges, wantDirChanges, objChanges, wantObjChanges)
			}

			// tidy up afterwards
			for _, o := range objs {
				assert.NoError(t, o.Remove(ctx))
			}
			dirs = append(dirs, "dir")
			for _, dir := range dirs {
				assert.NoError(t, remote.Rmdir(ctx, dir))
			}
		})

		// TestFsPut files writes file1, file2 and tests an update
		//
		// Tests that require file1, file2 are within this
		t.Run("FsPutFiles", func(t *testing.T) {
			skipIfNotOk(t)
			file1Contents, _ = testPut(ctx, t, remote, &file1)
			/* file2Contents = */ testPut(ctx, t, remote, &file2)
			file1Contents, _ = testPut(ctx, t, remote, &file1)
			// Note that the next test will check there are no duplicated file names

			// TestFsListDirFile2 tests the files are correctly uploaded by doing
			// Depth 1 directory listings
			TestFsListDirFile2 := func(t *testing.T) {
				skipIfNotOk(t)
				list := func(dir string, expectedDirNames, expectedObjNames []string) {
					var objNames, dirNames []string
					for i := 1; i <= *fstest.ListRetries; i++ {
						objs, dirs, err := walk.GetAll(ctx, remote, dir, true, 1)
						if errors.Cause(err) == fs.ErrorDirNotFound {
							objs, dirs, err = walk.GetAll(ctx, remote, dir, true, 1)
						}
						require.NoError(t, err)
						objNames = objsToNames(objs)
						dirNames = dirsToNames(dirs)
						if len(objNames) >= len(expectedObjNames) && len(dirNames) >= len(expectedDirNames) {
							break
						}
						t.Logf("Sleeping for 1 second for TestFsListDirFile2 eventual consistency: %d/%d", i, *fstest.ListRetries)
						time.Sleep(1 * time.Second)
					}
					assert.Equal(t, expectedDirNames, dirNames)
					assert.Equal(t, expectedObjNames, objNames)
				}
				dir := file2.Path
				deepest := true
				for dir != "" {
					expectedObjNames := []string{}
					expectedDirNames := []string{}
					child := dir
					dir = path.Dir(dir)
					if dir == "." {
						dir = ""
						expectedObjNames = append(expectedObjNames, file1.Path)
					}
					if deepest {
						expectedObjNames = append(expectedObjNames, file2.Path)
						deepest = false
					} else {
						expectedDirNames = append(expectedDirNames, child)
					}
					list(dir, expectedDirNames, expectedObjNames)
				}
			}
			t.Run("FsListDirFile2", TestFsListDirFile2)

			// TestFsListRDirFile2 tests the files are correctly uploaded by doing
			// Depth 1 directory listings using ListR
			t.Run("FsListRDirFile2", func(t *testing.T) {
				defer skipIfNotListR(t)()
				TestFsListDirFile2(t)
			})

			// Test the files are all there with walk.ListR recursive listings
			t.Run("FsListR", func(t *testing.T) {
				skipIfNotOk(t)
				objs, dirs, err := walk.GetAll(ctx, remote, "", true, -1)
				require.NoError(t, err)
				assert.Equal(t, []string{
					"hello? sausage",
					"hello? sausage/êé",
					"hello? sausage/êé/Hello, 世界",
					"hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠",
				}, dirsToNames(dirs))
				assert.Equal(t, []string{
					"file name.txt",
					"hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠/z.txt",
				}, objsToNames(objs))
			})

			// Test the files are all there with
			// walk.ListR recursive listings on a sub dir
			t.Run("FsListRSubdir", func(t *testing.T) {
				skipIfNotOk(t)
				objs, dirs, err := walk.GetAll(ctx, remote, path.Dir(path.Dir(path.Dir(path.Dir(file2.Path)))), true, -1)
				require.NoError(t, err)
				assert.Equal(t, []string{
					"hello? sausage/êé",
					"hello? sausage/êé/Hello, 世界",
					"hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠",
				}, dirsToNames(dirs))
				assert.Equal(t, []string{
					"hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠/z.txt",
				}, objsToNames(objs))
			})

			// TestFsListDirRoot tests that DirList works in the root
			TestFsListDirRoot := func(t *testing.T) {
				skipIfNotOk(t)
				rootRemote, err := fs.NewFs(remoteName)
				require.NoError(t, err)
				_, dirs, err := walk.GetAll(ctx, rootRemote, "", true, 1)
				require.NoError(t, err)
				assert.Contains(t, dirsToNames(dirs), subRemoteLeaf, "Remote leaf not found")
			}
			t.Run("FsListDirRoot", TestFsListDirRoot)

			// TestFsListRDirRoot tests that DirList works in the root using ListR
			t.Run("FsListRDirRoot", func(t *testing.T) {
				defer skipIfNotListR(t)()
				TestFsListDirRoot(t)
			})

			// TestFsListSubdir tests List works for a subdirectory
			TestFsListSubdir := func(t *testing.T) {
				skipIfNotOk(t)
				fileName := file2.Path
				var err error
				var objs []fs.Object
				var dirs []fs.Directory
				for i := 0; i < 2; i++ {
					dir, _ := path.Split(fileName)
					dir = dir[:len(dir)-1]
					objs, dirs, err = walk.GetAll(ctx, remote, dir, true, -1)
				}
				require.NoError(t, err)
				require.Len(t, objs, 1)
				assert.Equal(t, fileName, objs[0].Remote())
				require.Len(t, dirs, 0)
			}
			t.Run("FsListSubdir", TestFsListSubdir)

			// TestFsListRSubdir tests List works for a subdirectory using ListR
			t.Run("FsListRSubdir", func(t *testing.T) {
				defer skipIfNotListR(t)()
				TestFsListSubdir(t)
			})

			// TestFsListLevel2 tests List works for 2 levels
			TestFsListLevel2 := func(t *testing.T) {
				skipIfNotOk(t)
				objs, dirs, err := walk.GetAll(ctx, remote, "", true, 2)
				if err == fs.ErrorLevelNotSupported {
					return
				}
				require.NoError(t, err)
				assert.Equal(t, []string{file1.Path}, objsToNames(objs))
				assert.Equal(t, []string{"hello? sausage", "hello? sausage/êé"}, dirsToNames(dirs))
			}
			t.Run("FsListLevel2", TestFsListLevel2)

			// TestFsListRLevel2 tests List works for 2 levels using ListR
			t.Run("FsListRLevel2", func(t *testing.T) {
				defer skipIfNotListR(t)()
				TestFsListLevel2(t)
			})

			// TestFsListFile1 tests file present
			t.Run("FsListFile1", func(t *testing.T) {
				skipIfNotOk(t)
				fstest.CheckListing(t, remote, []fstest.Item{file1, file2})
			})

			// TestFsNewObject tests NewObject
			t.Run("FsNewObject", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				file1.Check(t, obj, remote.Precision())
			})

			// TestFsListFile1and2 tests two files present
			t.Run("FsListFile1and2", func(t *testing.T) {
				skipIfNotOk(t)
				fstest.CheckListing(t, remote, []fstest.Item{file1, file2})
			})

			// TestFsNewObjectDir tests NewObject on a directory which should produce an error
			t.Run("FsNewObjectDir", func(t *testing.T) {
				skipIfNotOk(t)
				dir := path.Dir(file2.Path)
				obj, err := remote.NewObject(ctx, dir)
				assert.Nil(t, obj)
				assert.NotNil(t, err)
			})

			// TestFsCopy tests Copy
			t.Run("FsCopy", func(t *testing.T) {
				skipIfNotOk(t)

				// Check have Copy
				doCopy := remote.Features().Copy
				if doCopy == nil {
					t.Skip("FS has no Copier interface")
				}

				// Test with file2 so have + and ' ' in file name
				var file2Copy = file2
				file2Copy.Path += "-copy"

				// do the copy
				src := findObject(ctx, t, remote, file2.Path)
				dst, err := doCopy(ctx, src, file2Copy.Path)
				if err == fs.ErrorCantCopy {
					t.Skip("FS can't copy")
				}
				require.NoError(t, err, fmt.Sprintf("Error: %#v", err))

				// check file exists in new listing
				fstest.CheckListing(t, remote, []fstest.Item{file1, file2, file2Copy})

				// Check dst lightly - list above has checked ModTime/Hashes
				assert.Equal(t, file2Copy.Path, dst.Remote())

				// Delete copy
				err = dst.Remove(ctx)
				require.NoError(t, err)

			})

			// TestFsMove tests Move
			t.Run("FsMove", func(t *testing.T) {
				skipIfNotOk(t)

				// Check have Move
				doMove := remote.Features().Move
				if doMove == nil {
					t.Skip("FS has no Mover interface")
				}

				// state of files now:
				// 1: file name.txt
				// 2: hello sausage?/../z.txt

				var file1Move = file1
				var file2Move = file2

				// check happy path, i.e. no naming conflicts when rename and move are two
				// separate operations
				file2Move.Path = "other.txt"
				src := findObject(ctx, t, remote, file2.Path)
				dst, err := doMove(ctx, src, file2Move.Path)
				if err == fs.ErrorCantMove {
					t.Skip("FS can't move")
				}
				require.NoError(t, err)
				// check file exists in new listing
				fstest.CheckListing(t, remote, []fstest.Item{file1, file2Move})
				// Check dst lightly - list above has checked ModTime/Hashes
				assert.Equal(t, file2Move.Path, dst.Remote())
				// 1: file name.txt
				// 2: other.txt

				// Check conflict on "rename, then move"
				file1Move.Path = "moveTest/other.txt"
				src = findObject(ctx, t, remote, file1.Path)
				_, err = doMove(ctx, src, file1Move.Path)
				require.NoError(t, err)
				fstest.CheckListing(t, remote, []fstest.Item{file1Move, file2Move})
				// 1: moveTest/other.txt
				// 2: other.txt

				// Check conflict on "move, then rename"
				src = findObject(ctx, t, remote, file1Move.Path)
				_, err = doMove(ctx, src, file1.Path)
				require.NoError(t, err)
				fstest.CheckListing(t, remote, []fstest.Item{file1, file2Move})
				// 1: file name.txt
				// 2: other.txt

				src = findObject(ctx, t, remote, file2Move.Path)
				_, err = doMove(ctx, src, file2.Path)
				require.NoError(t, err)
				fstest.CheckListing(t, remote, []fstest.Item{file1, file2})
				// 1: file name.txt
				// 2: hello sausage?/../z.txt

				// Tidy up moveTest directory
				require.NoError(t, remote.Rmdir(ctx, "moveTest"))
			})

			// Move src to this remote using server side move operations.
			//
			// Will only be called if src.Fs().Name() == f.Name()
			//
			// If it isn't possible then return fs.ErrorCantDirMove
			//
			// If destination exists then return fs.ErrorDirExists

			// TestFsDirMove tests DirMove
			//
			// go test -v -run 'TestIntegration/Test(Setup|Init|FsMkdir|FsPutFile1|FsPutFile2|FsUpdateFile1|FsDirMove)$
			t.Run("FsDirMove", func(t *testing.T) {
				skipIfNotOk(t)

				// Check have DirMove
				doDirMove := remote.Features().DirMove
				if doDirMove == nil {
					t.Skip("FS has no DirMover interface")
				}

				// Check it can't move onto itself
				err := doDirMove(ctx, remote, "", "")
				require.Equal(t, fs.ErrorDirExists, err)

				// new remote
				newRemote, _, removeNewRemote, err := fstest.RandomRemote()
				require.NoError(t, err)
				defer removeNewRemote()

				const newName = "new_name/sub_new_name"
				// try the move
				err = newRemote.Features().DirMove(ctx, remote, "", newName)
				require.NoError(t, err)

				// check remotes
				// remote should not exist here
				_, err = remote.List(ctx, "")
				assert.Equal(t, fs.ErrorDirNotFound, errors.Cause(err))
				//fstest.CheckListingWithPrecision(t, remote, []fstest.Item{}, []string{}, remote.Precision())
				file1Copy := file1
				file1Copy.Path = path.Join(newName, file1.Path)
				file2Copy := file2
				file2Copy.Path = path.Join(newName, file2.Path)
				fstest.CheckListingWithPrecision(t, newRemote, []fstest.Item{file2Copy, file1Copy}, []string{
					"new_name",
					"new_name/sub_new_name",
					"new_name/sub_new_name/hello? sausage",
					"new_name/sub_new_name/hello? sausage/êé",
					"new_name/sub_new_name/hello? sausage/êé/Hello, 世界",
					"new_name/sub_new_name/hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠",
				}, newRemote.Precision())

				// move it back
				err = doDirMove(ctx, newRemote, newName, "")
				require.NoError(t, err)

				// check remotes
				fstest.CheckListingWithPrecision(t, remote, []fstest.Item{file2, file1}, []string{
					"hello? sausage",
					"hello? sausage/êé",
					"hello? sausage/êé/Hello, 世界",
					"hello? sausage/êé/Hello, 世界/ \" ' @ < > & ? + ≠",
				}, remote.Precision())
				fstest.CheckListingWithPrecision(t, newRemote, []fstest.Item{}, []string{
					"new_name",
				}, newRemote.Precision())
			})

			// TestFsRmdirFull tests removing a non empty directory
			t.Run("FsRmdirFull", func(t *testing.T) {
				skipIfNotOk(t)
				if isBucketBasedButNotRoot(remote) {
					t.Skip("Skipping test as non root bucket based remote")
				}
				err := remote.Rmdir(ctx, "")
				require.Error(t, err, "Expecting error on RMdir on non empty remote")
			})

			// TestFsPrecision tests the Precision of the Fs
			t.Run("FsPrecision", func(t *testing.T) {
				skipIfNotOk(t)
				precision := remote.Precision()
				if precision == fs.ModTimeNotSupported {
					return
				}
				if precision > time.Second || precision < 0 {
					t.Fatalf("Precision out of range %v", precision)
				}
				// FIXME check expected precision
			})

			// TestObjectString tests the Object String method
			t.Run("ObjectString", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				assert.Equal(t, file1.Path, obj.String())
				if opt.NilObject != nil {
					assert.Equal(t, "<nil>", opt.NilObject.String())
				}
			})

			// TestObjectFs tests the object can be found
			t.Run("ObjectFs", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				// If this is set we don't do the direct comparison of
				// the Fs from the object as it may be different
				if opt.SkipFsMatch {
					return
				}
				testRemote := remote
				if obj.Fs() != testRemote {
					// Check to see if this wraps something else
					if doUnWrap := testRemote.Features().UnWrap; doUnWrap != nil {
						testRemote = doUnWrap()
					}
				}
				assert.Equal(t, obj.Fs(), testRemote)
			})

			// TestObjectRemote tests the Remote is correct
			t.Run("ObjectRemote", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				assert.Equal(t, file1.Path, obj.Remote())
			})

			// TestObjectHashes checks all the hashes the object supports
			t.Run("ObjectHashes", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				file1.CheckHashes(t, obj)
			})

			// TestObjectModTime tests the ModTime of the object is correct
			TestObjectModTime := func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				file1.CheckModTime(t, obj, obj.ModTime(ctx), remote.Precision())
			}
			t.Run("ObjectModTime", TestObjectModTime)

			// TestObjectMimeType tests the MimeType of the object is correct
			t.Run("ObjectMimeType", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				do, ok := obj.(fs.MimeTyper)
				if !ok {
					t.Skip("MimeType method not supported")
				}
				mimeType := do.MimeType(ctx)
				if strings.ContainsRune(mimeType, ';') {
					assert.Equal(t, "text/plain; charset=utf-8", mimeType)
				} else {
					assert.Equal(t, "text/plain", mimeType)
				}
			})

			// TestObjectSetModTime tests that SetModTime works
			t.Run("ObjectSetModTime", func(t *testing.T) {
				skipIfNotOk(t)
				newModTime := fstest.Time("2011-12-13T14:15:16.999999999Z")
				obj := findObject(ctx, t, remote, file1.Path)
				err := obj.SetModTime(ctx, newModTime)
				if err == fs.ErrorCantSetModTime || err == fs.ErrorCantSetModTimeWithoutDelete {
					t.Log(err)
					return
				}
				require.NoError(t, err)
				file1.ModTime = newModTime
				file1.CheckModTime(t, obj, obj.ModTime(ctx), remote.Precision())
				// And make a new object and read it from there too
				TestObjectModTime(t)
			})

			// TestObjectSize tests that Size works
			t.Run("ObjectSize", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				assert.Equal(t, file1.Size, obj.Size())
			})

			// TestObjectOpen tests that Open works
			t.Run("ObjectOpen", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				assert.Equal(t, file1Contents, readObject(ctx, t, obj, -1), "contents of file1 differ")
			})

			// TestObjectOpenSeek tests that Open works with SeekOption
			t.Run("ObjectOpenSeek", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				assert.Equal(t, file1Contents[50:], readObject(ctx, t, obj, -1, &fs.SeekOption{Offset: 50}), "contents of file1 differ after seek")
			})

			// TestObjectOpenRange tests that Open works with RangeOption
			//
			// go test -v -run 'TestIntegration/Test(Setup|Init|FsMkdir|FsPutFile1|FsPutFile2|FsUpdateFile1|ObjectOpenRange)$'
			t.Run("ObjectOpenRange", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				for _, test := range []struct {
					ro                 fs.RangeOption
					wantStart, wantEnd int
				}{
					{fs.RangeOption{Start: 5, End: 15}, 5, 16},
					{fs.RangeOption{Start: 80, End: -1}, 80, 100},
					{fs.RangeOption{Start: 81, End: 100000}, 81, 100},
					{fs.RangeOption{Start: -1, End: 20}, 80, 100}, // if start is omitted this means get the final bytes
					// {fs.RangeOption{Start: -1, End: -1}, 0, 100}, - this seems to work but the RFC doesn't define it
				} {
					got := readObject(ctx, t, obj, -1, &test.ro)
					foundAt := strings.Index(file1Contents, got)
					help := fmt.Sprintf("%#v failed want [%d:%d] got [%d:%d]", test.ro, test.wantStart, test.wantEnd, foundAt, foundAt+len(got))
					assert.Equal(t, file1Contents[test.wantStart:test.wantEnd], got, help)
				}
			})

			// TestObjectPartialRead tests that reading only part of the object does the correct thing
			t.Run("ObjectPartialRead", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				assert.Equal(t, file1Contents[:50], readObject(ctx, t, obj, 50), "contents of file1 differ after limited read")
			})

			// TestObjectUpdate tests that Update works
			t.Run("ObjectUpdate", func(t *testing.T) {
				skipIfNotOk(t)
				contents := random.String(200)
				buf := bytes.NewBufferString(contents)
				hash := hash.NewMultiHasher()
				in := io.TeeReader(buf, hash)

				file1.Size = int64(buf.Len())
				obj := findObject(ctx, t, remote, file1.Path)
				obji := object.NewStaticObjectInfo(file1.Path, file1.ModTime, int64(len(contents)), true, nil, obj.Fs())
				err := obj.Update(ctx, in, obji)
				require.NoError(t, err)
				file1.Hashes = hash.Sums()

				// check the object has been updated
				file1.Check(t, obj, remote.Precision())

				// Re-read the object and check again
				obj = findObject(ctx, t, remote, file1.Path)
				file1.Check(t, obj, remote.Precision())

				// check contents correct
				assert.Equal(t, contents, readObject(ctx, t, obj, -1), "contents of updated file1 differ")
				file1Contents = contents
			})

			// TestObjectStorable tests that Storable works
			t.Run("ObjectStorable", func(t *testing.T) {
				skipIfNotOk(t)
				obj := findObject(ctx, t, remote, file1.Path)
				require.NotNil(t, !obj.Storable(), "Expecting object to be storable")
			})

			// TestFsIsFile tests that an error is returned along with a valid fs
			// which points to the parent directory.
			t.Run("FsIsFile", func(t *testing.T) {
				skipIfNotOk(t)
				remoteName := subRemoteName + "/" + file2.Path
				file2Copy := file2
				file2Copy.Path = "z.txt"
				fileRemote, err := fs.NewFs(remoteName)
				require.NotNil(t, fileRemote)
				assert.Equal(t, fs.ErrorIsFile, err)

				if strings.HasPrefix(remoteName, "TestChunker") && strings.Contains(remoteName, "Nometa") {
					// TODO fix chunker and remove this bypass
					t.Logf("Skip listing check -- chunker can't yet handle this tricky case")
					return
				}
				fstest.CheckListing(t, fileRemote, []fstest.Item{file2Copy})
			})

			// TestFsIsFileNotFound tests that an error is not returned if no object is found
			t.Run("FsIsFileNotFound", func(t *testing.T) {
				skipIfNotOk(t)
				remoteName := subRemoteName + "/not found.txt"
				fileRemote, err := fs.NewFs(remoteName)
				require.NoError(t, err)
				fstest.CheckListing(t, fileRemote, []fstest.Item{})
			})

			// Test that things work from the root
			t.Run("FromRoot", func(t *testing.T) {
				if features := remote.Features(); features.BucketBased && !features.BucketBasedRootOK {
					t.Skip("Can't list from root on this remote")
				}

				configName, configLeaf, err := fspath.Parse(subRemoteName)
				require.NoError(t, err)
				if configName == "" {
					configName, configLeaf = path.Split(subRemoteName)
				} else {
					configName += ":"
				}
				t.Logf("Opening root remote %q path %q from %q", configName, configLeaf, subRemoteName)
				rootRemote, err := fs.NewFs(configName)
				require.NoError(t, err)

				file1Root := file1
				file1Root.Path = path.Join(configLeaf, file1Root.Path)
				file2Root := file2
				file2Root.Path = path.Join(configLeaf, file2Root.Path)
				var dirs []string
				dir := file2.Path
				for {
					dir = path.Dir(dir)
					if dir == "" || dir == "." || dir == "/" {
						break
					}
					dirs = append(dirs, path.Join(configLeaf, dir))
				}

				// Check that we can see file1 and file2 from the root
				t.Run("List", func(t *testing.T) {
					fstest.CheckListingWithRoot(t, rootRemote, configLeaf, []fstest.Item{file1Root, file2Root}, dirs, rootRemote.Precision())
				})

				// Check that that listing the entries is OK
				t.Run("ListEntries", func(t *testing.T) {
					entries, err := rootRemote.List(context.Background(), configLeaf)
					require.NoError(t, err)
					fstest.CompareItems(t, entries, []fstest.Item{file1Root}, dirs[len(dirs)-1:], rootRemote.Precision(), "ListEntries")
				})

				// List the root with ListR
				t.Run("ListR", func(t *testing.T) {
					doListR := rootRemote.Features().ListR
					if doListR == nil {
						t.Skip("FS has no ListR interface")
					}
					file1Found, file2Found := false, false
					stopTime := time.Now().Add(10 * time.Second)
					errTooMany := errors.New("too many files")
					errFound := errors.New("found")
					err := doListR(context.Background(), "", func(entries fs.DirEntries) error {
						for _, entry := range entries {
							remote := entry.Remote()
							if remote == file1Root.Path {
								file1Found = true
							}
							if remote == file2Root.Path {
								file2Found = true
							}
							if file1Found && file2Found {
								return errFound
							}
						}
						if time.Now().After(stopTime) {
							return errTooMany
						}
						return nil
					})
					if err != errFound && err != errTooMany {
						assert.NoError(t, err)
					}
					if err != errTooMany {
						assert.True(t, file1Found, "file1Root not found")
						assert.True(t, file2Found, "file2Root not found")
					} else {
						t.Logf("Too many files to list - giving up")
					}
				})

				// Create a new file
				t.Run("Put", func(t *testing.T) {
					file3Root := fstest.Item{
						ModTime: time.Now(),
						Path:    path.Join(configLeaf, "created from root.txt"),
					}
					_, file3Obj := testPut(ctx, t, rootRemote, &file3Root)
					fstest.CheckListingWithRoot(t, rootRemote, configLeaf, []fstest.Item{file1Root, file2Root, file3Root}, nil, rootRemote.Precision())

					// And then remove it
					t.Run("Remove", func(t *testing.T) {
						require.NoError(t, file3Obj.Remove(context.Background()))
						fstest.CheckListingWithRoot(t, rootRemote, configLeaf, []fstest.Item{file1Root, file2Root}, nil, rootRemote.Precision())
					})
				})
			})

			// TestPublicLink tests creation of sharable, public links
			// go test -v -run 'TestIntegration/Test(Setup|Init|FsMkdir|FsPutFile1|FsPutFile2|FsUpdateFile1|PublicLink)$'
			t.Run("PublicLink", func(t *testing.T) {
				skipIfNotOk(t)

				doPublicLink := remote.Features().PublicLink
				if doPublicLink == nil {
					t.Skip("FS has no PublicLinker interface")
				}

				// if object not found
				link, err := doPublicLink(ctx, file1.Path+"_does_not_exist")
				require.Error(t, err, "Expected to get error when file doesn't exist")
				require.Equal(t, "", link, "Expected link to be empty on error")

				// sharing file for the first time
				link1, err := doPublicLink(ctx, file1.Path)
				require.NoError(t, err)
				require.NotEqual(t, "", link1, "Link should not be empty")

				link2, err := doPublicLink(ctx, file2.Path)
				require.NoError(t, err)
				require.NotEqual(t, "", link2, "Link should not be empty")

				require.NotEqual(t, link1, link2, "Links to different files should differ")

				// sharing file for the 2nd time
				link1, err = doPublicLink(ctx, file1.Path)
				require.NoError(t, err)
				require.NotEqual(t, "", link1, "Link should not be empty")

				// sharing directory for the first time
				path := path.Dir(file2.Path)
				link3, err := doPublicLink(ctx, path)
				if err != nil && errors.Cause(err) == fs.ErrorCantShareDirectories {
					t.Log("skipping directory tests as not supported on this backend")
				} else {
					require.NoError(t, err)
					require.NotEqual(t, "", link3, "Link should not be empty")

					// sharing directory for the second time
					link3, err = doPublicLink(ctx, path)
					require.NoError(t, err)
					require.NotEqual(t, "", link3, "Link should not be empty")

					// sharing the "root" directory in a subremote
					subRemote, _, removeSubRemote, err := fstest.RandomRemote()
					require.NoError(t, err)
					defer removeSubRemote()
					// ensure sub remote isn't empty
					buf := bytes.NewBufferString("somecontent")
					obji := object.NewStaticObjectInfo("somefile", time.Now(), int64(buf.Len()), true, nil, nil)
					_, err = subRemote.Put(ctx, buf, obji)
					require.NoError(t, err)

					link4, err := subRemote.Features().PublicLink(ctx, "")
					require.NoError(t, err, "Sharing root in a sub-remote should work")
					require.NotEqual(t, "", link4, "Link should not be empty")
				}
			})

			// TestSetTier tests SetTier and GetTier functionality
			t.Run("SetTier", func(t *testing.T) {
				skipIfNotSetTier(t)
				obj := findObject(ctx, t, remote, file1.Path)
				setter, ok := obj.(fs.SetTierer)
				assert.NotNil(t, ok)
				getter, ok := obj.(fs.GetTierer)
				assert.NotNil(t, ok)
				// If interfaces are supported TiersToTest should contain
				// at least one entry
				supportedTiers := opt.TiersToTest
				assert.NotEmpty(t, supportedTiers)
				// test set tier changes on supported storage classes or tiers
				for _, tier := range supportedTiers {
					err := setter.SetTier(tier)
					assert.Nil(t, err)
					got := getter.GetTier()
					assert.Equal(t, tier, got)
				}
			})

			// Check to see if Fs that wrap other Objects implement all the optional methods
			t.Run("ObjectCheckWrap", func(t *testing.T) {
				skipIfNotOk(t)
				if opt.SkipObjectCheckWrap {
					t.Skip("Skipping FsCheckWrap on this Fs")
				}
				ft := new(fs.Features).Fill(remote)
				if ft.UnWrap == nil {
					t.Skip("Not a wrapping Fs")
				}
				obj := findObject(ctx, t, remote, file1.Path)
				_, unsupported := fs.ObjectOptionalInterfaces(obj)
				for _, name := range unsupported {
					if !stringsContains(name, opt.UnimplementableObjectMethods) {
						t.Errorf("Missing Object wrapper for %s", name)
					}
				}
			})

			// TestObjectRemove tests Remove
			t.Run("ObjectRemove", func(t *testing.T) {
				skipIfNotOk(t)
				// remove file1
				obj := findObject(ctx, t, remote, file1.Path)
				err := obj.Remove(ctx)
				require.NoError(t, err)
				// check listing without modtime as TestPublicLink may change the modtime
				fstest.CheckListingWithPrecision(t, remote, []fstest.Item{file2}, nil, fs.ModTimeNotSupported)
			})

			// TestAbout tests the About optional interface
			t.Run("ObjectAbout", func(t *testing.T) {
				skipIfNotOk(t)

				// Check have About
				doAbout := remote.Features().About
				if doAbout == nil {
					t.Skip("FS does not support About")
				}

				// Can't really check the output much!
				usage, err := doAbout(context.Background())
				require.NoError(t, err)
				require.NotNil(t, usage)
				assert.NotEqual(t, int64(0), usage.Total)
			})

			// Just file2 remains for Purge to clean up

			// TestFsPutStream tests uploading files when size isn't known in advance.
			// This may trigger large buffer allocation in some backends, keep it
			// close to the end of suite. (See fs/operations/xtra_operations_test.go)
			t.Run("FsPutStream", func(t *testing.T) {
				skipIfNotOk(t)
				if remote.Features().PutStream == nil {
					t.Skip("FS has no PutStream interface")
				}

				file := fstest.Item{
					ModTime: fstest.Time("2001-02-03T04:05:06.499999999Z"),
					Path:    "piped data.txt",
					Size:    -1, // use unknown size during upload
				}

				var (
					err         error
					obj         fs.Object
					uploadHash  *hash.MultiHasher
					contentSize = 100
				)
				retry(t, "PutStream", func() error {
					contents := random.String(contentSize)
					buf := bytes.NewBufferString(contents)
					uploadHash = hash.NewMultiHasher()
					in := io.TeeReader(buf, uploadHash)

					file.Size = -1
					obji := object.NewStaticObjectInfo(file.Path, file.ModTime, file.Size, true, nil, nil)
					obj, err = remote.Features().PutStream(ctx, in, obji)
					return err
				})
				file.Hashes = uploadHash.Sums()
				file.Size = int64(contentSize) // use correct size when checking
				file.Check(t, obj, remote.Precision())
				// Re-read the object and check again
				obj = findObject(ctx, t, remote, file.Path)
				file.Check(t, obj, remote.Precision())
				require.NoError(t, obj.Remove(ctx))
			})

			// TestInternal calls InternalTest() on the Fs
			t.Run("Internal", func(t *testing.T) {
				skipIfNotOk(t)
				if it, ok := remote.(InternalTester); ok {
					it.InternalTest(t)
				} else {
					t.Skipf("%T does not implement InternalTester", remote)
				}
			})

		})

		// TestFsPutChunked may trigger large buffer allocation with
		// some backends (see fs/operations/xtra_operations_test.go),
		// keep it closer to the end of suite.
		t.Run("FsPutChunked", func(t *testing.T) {
			skipIfNotOk(t)
			if testing.Short() {
				t.Skip("not running with -short")
			}

			setUploadChunkSizer, _ := remote.(SetUploadChunkSizer)
			if setUploadChunkSizer == nil {
				t.Skipf("%T does not implement SetUploadChunkSizer", remote)
			}

			setUploadCutoffer, _ := remote.(SetUploadCutoffer)

			minChunkSize := opt.ChunkedUpload.MinChunkSize
			if minChunkSize < 100 {
				minChunkSize = 100
			}
			if opt.ChunkedUpload.CeilChunkSize != nil {
				minChunkSize = opt.ChunkedUpload.CeilChunkSize(minChunkSize)
			}

			maxChunkSize := 2 * fs.MebiByte
			if maxChunkSize < 2*minChunkSize {
				maxChunkSize = 2 * minChunkSize
			}
			if opt.ChunkedUpload.MaxChunkSize > 0 && maxChunkSize > opt.ChunkedUpload.MaxChunkSize {
				maxChunkSize = opt.ChunkedUpload.MaxChunkSize
			}
			if opt.ChunkedUpload.CeilChunkSize != nil {
				maxChunkSize = opt.ChunkedUpload.CeilChunkSize(maxChunkSize)
			}

			next := func(f func(fs.SizeSuffix) fs.SizeSuffix) fs.SizeSuffix {
				s := f(minChunkSize)
				if s > maxChunkSize {
					s = minChunkSize
				}
				return s
			}

			chunkSizes := fs.SizeSuffixList{
				minChunkSize,
				minChunkSize + (maxChunkSize-minChunkSize)/3,
				next(NextPowerOfTwo),
				next(NextMultipleOf(100000)),
				next(NextMultipleOf(100001)),
				maxChunkSize,
			}
			chunkSizes.Sort()

			// Set the minimum chunk size, upload cutoff and reset it at the end
			oldChunkSize, err := setUploadChunkSizer.SetUploadChunkSize(minChunkSize)
			require.NoError(t, err)
			var oldUploadCutoff fs.SizeSuffix
			if setUploadCutoffer != nil {
				oldUploadCutoff, err = setUploadCutoffer.SetUploadCutoff(minChunkSize)
				require.NoError(t, err)
			}
			defer func() {
				_, err := setUploadChunkSizer.SetUploadChunkSize(oldChunkSize)
				assert.NoError(t, err)
				if setUploadCutoffer != nil {
					_, err := setUploadCutoffer.SetUploadCutoff(oldUploadCutoff)
					assert.NoError(t, err)
				}
			}()

			var lastCs fs.SizeSuffix
			for _, cs := range chunkSizes {
				if cs <= lastCs {
					continue
				}
				if opt.ChunkedUpload.CeilChunkSize != nil {
					cs = opt.ChunkedUpload.CeilChunkSize(cs)
				}
				lastCs = cs

				t.Run(cs.String(), func(t *testing.T) {
					_, err := setUploadChunkSizer.SetUploadChunkSize(cs)
					require.NoError(t, err)
					if setUploadCutoffer != nil {
						_, err = setUploadCutoffer.SetUploadCutoff(cs)
						require.NoError(t, err)
					}

					var testChunks []fs.SizeSuffix
					if opt.ChunkedUpload.NeedMultipleChunks {
						// If NeedMultipleChunks is set then test with > cs
						testChunks = []fs.SizeSuffix{cs + 1, 2 * cs, 2*cs + 1}
					} else {
						testChunks = []fs.SizeSuffix{cs - 1, cs, 2*cs + 1}
					}

					for _, fileSize := range testChunks {
						t.Run(fmt.Sprintf("%d", fileSize), func(t *testing.T) {
							TestPutLarge(ctx, t, remote, &fstest.Item{
								ModTime: fstest.Time("2001-02-03T04:05:06.499999999Z"),
								Path:    fmt.Sprintf("chunked-%s-%s.bin", cs.String(), fileSize.String()),
								Size:    int64(fileSize),
							})
						})
					}
				})
			}
		})

		// TestFsUploadUnknownSize ensures Fs.Put() and Object.Update() don't panic when
		// src.Size() == -1
		//
		// This may trigger large buffer allocation in some backends, keep it
		// closer to the suite end. (See fs/operations/xtra_operations_test.go)
		t.Run("FsUploadUnknownSize", func(t *testing.T) {
			skipIfNotOk(t)

			t.Run("FsPutUnknownSize", func(t *testing.T) {
				defer func() {
					assert.Nil(t, recover(), "Fs.Put() should not panic when src.Size() == -1")
				}()

				contents := random.String(100)
				in := bytes.NewBufferString(contents)

				obji := object.NewStaticObjectInfo("unknown-size-put.txt", fstest.Time("2002-02-03T04:05:06.499999999Z"), -1, true, nil, nil)
				obj, err := remote.Put(ctx, in, obji)
				if err == nil {
					require.NoError(t, obj.Remove(ctx), "successfully uploaded unknown-sized file but failed to remove")
				}
				// if err != nil: it's okay as long as no panic
			})

			t.Run("FsUpdateUnknownSize", func(t *testing.T) {
				unknownSizeUpdateFile := fstest.Item{
					ModTime: fstest.Time("2002-02-03T04:05:06.499999999Z"),
					Path:    "unknown-size-update.txt",
				}

				testPut(ctx, t, remote, &unknownSizeUpdateFile)

				defer func() {
					assert.Nil(t, recover(), "Object.Update() should not panic when src.Size() == -1")
				}()

				newContents := random.String(200)
				in := bytes.NewBufferString(newContents)

				obj := findObject(ctx, t, remote, unknownSizeUpdateFile.Path)
				obji := object.NewStaticObjectInfo(unknownSizeUpdateFile.Path, unknownSizeUpdateFile.ModTime, -1, true, nil, obj.Fs())
				err := obj.Update(ctx, in, obji)
				if err == nil {
					require.NoError(t, obj.Remove(ctx), "successfully updated object with unknown-sized source but failed to remove")
				}
				// if err != nil: it's okay as long as no panic
			})

		})

		// TestFsRootCollapse tests if the root of an fs "collapses" to the
		// absolute root. It creates a new fs of the same backend type with its
		// root set to a *non-existent* folder, and attempts to read the info of
		// an object in that folder, whose name is taken from a directory that
		// exists in the absolute root.
		// This test is added after
		// https://github.com/rclone/rclone/issues/3164.
		t.Run("FsRootCollapse", func(t *testing.T) {
			deepRemoteName := subRemoteName + "/deeper/nonexisting/directory"
			deepRemote, err := fs.NewFs(deepRemoteName)
			require.NoError(t, err)

			colonIndex := strings.IndexRune(deepRemoteName, ':')
			firstSlashIndex := strings.IndexRune(deepRemoteName, '/')
			firstDir := deepRemoteName[colonIndex+1 : firstSlashIndex]
			_, err = deepRemote.NewObject(ctx, firstDir)
			require.Equal(t, fs.ErrorObjectNotFound, err)
			// If err is not fs.ErrorObjectNotFound, it means the backend is
			// somehow confused about root and absolute root.
		})

		// Purge the folder
		err = operations.Purge(ctx, remote, "")
		if errors.Cause(err) != fs.ErrorDirNotFound {
			require.NoError(t, err)
		}
		purged = true
		fstest.CheckListing(t, remote, []fstest.Item{})

		// Check purging again if not bucket based
		if !isBucketBasedButNotRoot(remote) {
			err = operations.Purge(ctx, remote, "")
			assert.Error(t, err, "Expecting error after on second purge")
		}

	})

	// Check directory is purged
	if !purged {
		_ = operations.Purge(ctx, remote, "")
	}

	// Remove the local directory so we don't clutter up /tmp
	if strings.HasPrefix(remoteName, "/") {
		t.Log("remoteName", remoteName)
		// Remove temp directory
		err := os.Remove(remoteName)
		require.NoError(t, err)
	}
}

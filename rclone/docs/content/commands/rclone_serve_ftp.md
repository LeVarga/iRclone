---
date: 2019-11-19T16:02:36Z
title: "rclone serve ftp"
slug: rclone_serve_ftp
url: /commands/rclone_serve_ftp/
---
## rclone serve ftp

Serve remote:path over FTP.

### Synopsis


rclone serve ftp implements a basic ftp server to serve the
remote over FTP protocol. This can be viewed with a ftp client
or you can make a remote of type ftp to read and write it.

### Server options

Use --addr to specify which IP address and port the server should
listen on, eg --addr 1.2.3.4:8000 or --addr :8080 to listen to all
IPs.  By default it only listens on localhost.  You can use port
:0 to let the OS choose an available port.

If you set --addr to listen on a public or LAN accessible IP address
then using Authentication is advised - see the next section for info.

#### Authentication

By default this will serve files without needing a login.

You can set a single username and password with the --user and --pass flags.

### Directory Cache

Using the `--dir-cache-time` flag, you can set how long a
directory should be considered up to date and not refreshed from the
backend. Changes made locally in the mount may appear immediately or
invalidate the cache. However, changes done on the remote will only
be picked up once the cache expires.

Alternatively, you can send a `SIGHUP` signal to rclone for
it to flush all directory caches, regardless of how old they are.
Assuming only one rclone instance is running, you can reset the cache
like this:

    kill -SIGHUP $(pidof rclone)

If you configure rclone with a [remote control](/rc) then you can use
rclone rc to flush the whole directory cache:

    rclone rc vfs/forget

Or individual files or directories:

    rclone rc vfs/forget file=path/to/file dir=path/to/dir

### File Buffering

The `--buffer-size` flag determines the amount of memory,
that will be used to buffer data in advance.

Each open file descriptor will try to keep the specified amount of
data in memory at all times. The buffered data is bound to one file
descriptor and won't be shared between multiple open file descriptors
of the same file.

This flag is a upper limit for the used memory per file descriptor.
The buffer will only use memory for data that is downloaded but not
not yet read. If the buffer is empty, only a small amount of memory
will be used.
The maximum memory used by rclone for buffering can be up to
`--buffer-size * open files`.

### File Caching

These flags control the VFS file caching options.  The VFS layer is
used by rclone mount to make a cloud storage system work more like a
normal file system.

You'll need to enable VFS caching if you want, for example, to read
and write simultaneously to a file.  See below for more details.

Note that the VFS cache works in addition to the cache backend and you
may find that you need one or the other or both.

    --cache-dir string                   Directory rclone will use for caching.
    --vfs-cache-max-age duration         Max age of objects in the cache. (default 1h0m0s)
    --vfs-cache-mode string              Cache mode off|minimal|writes|full (default "off")
    --vfs-cache-poll-interval duration   Interval to poll the cache for stale objects. (default 1m0s)
    --vfs-cache-max-size int             Max total size of objects in the cache. (default off)

If run with `-vv` rclone will print the location of the file cache.  The
files are stored in the user cache file area which is OS dependent but
can be controlled with `--cache-dir` or setting the appropriate
environment variable.

The cache has 4 different modes selected by `--vfs-cache-mode`.
The higher the cache mode the more compatible rclone becomes at the
cost of using disk space.

Note that files are written back to the remote only when they are
closed so if rclone is quit or dies with open files then these won't
get written back to the remote.  However they will still be in the on
disk cache.

If using --vfs-cache-max-size note that the cache may exceed this size
for two reasons.  Firstly because it is only checked every
--vfs-cache-poll-interval.  Secondly because open files cannot be
evicted from the cache.

#### --vfs-cache-mode off

In this mode the cache will read directly from the remote and write
directly to the remote without caching anything on disk.

This will mean some operations are not possible

  * Files can't be opened for both read AND write
  * Files opened for write can't be seeked
  * Existing files opened for write must have O_TRUNC set
  * Files open for read with O_TRUNC will be opened write only
  * Files open for write only will behave as if O_TRUNC was supplied
  * Open modes O_APPEND, O_TRUNC are ignored
  * If an upload fails it can't be retried

#### --vfs-cache-mode minimal

This is very similar to "off" except that files opened for read AND
write will be buffered to disks.  This means that files opened for
write will be a lot more compatible, but uses the minimal disk space.

These operations are not possible

  * Files opened for write only can't be seeked
  * Existing files opened for write must have O_TRUNC set
  * Files opened for write only will ignore O_APPEND, O_TRUNC
  * If an upload fails it can't be retried

#### --vfs-cache-mode writes

In this mode files opened for read only are still read directly from
the remote, write only and read/write files are buffered to disk
first.

This mode should support all normal file system operations.

If an upload fails it will be retried up to --low-level-retries times.

#### --vfs-cache-mode full

In this mode all reads and writes are buffered to and from disk.  When
a file is opened for read it will be downloaded in its entirety first.

This may be appropriate for your needs, or you may prefer to look at
the cache backend which does a much more sophisticated job of caching,
including caching directory hierarchies and chunks of files.

In this mode, unlike the others, when a file is written to the disk,
it will be kept on the disk after it is written to the remote.  It
will be purged on a schedule according to `--vfs-cache-max-age`.

This mode should support all normal file system operations.

If an upload or download fails it will be retried up to
--low-level-retries times.

### Auth Proxy

If you supply the parameter `--auth-proxy /path/to/program` then
rclone will use that program to generate backends on the fly which
then are used to authenticate incoming requests.  This uses a simple
JSON based protocl with input on STDIN and output on STDOUT.

There is an example program
[bin/test_proxy.py](https://github.com/rclone/rclone/blob/master/test_proxy.py)
in the rclone source code.

The program's job is to take a `user` and `pass` on the input and turn
those into the config for a backend on STDOUT in JSON format.  This
config will have any default parameters for the backend added, but it
won't use configuration from environment variables or command line
options - it is the job of the proxy program to make a complete
config.

This config generated must have this extra parameter
- `_root` - root to use for the backend

And it may have this parameter
- `_obscure` - comma separated strings for parameters to obscure

For example the program might take this on STDIN

```
{
	"user": "me",
	"pass": "mypassword"
}
```

And return this on STDOUT

```
{
	"type": "sftp",
	"_root": "",
	"_obscure": "pass",
	"user": "me",
	"pass": "mypassword",
	"host": "sftp.example.com"
}
```

This would mean that an SFTP backend would be created on the fly for
the `user` and `pass` returned in the output to the host given.  Note
that since `_obscure` is set to `pass`, rclone will obscure the `pass`
parameter before creating the backend (which is required for sftp
backends).

The progam can manipulate the supplied `user` in any way, for example
to make proxy to many different sftp backends, you could make the
`user` be `user@example.com` and then set the `host` to `example.com`
in the output and the user to `user`. For security you'd probably want
to restrict the `host` to a limited list.

Note that an internal cache is keyed on `user` so only use that for
configuration, don't use `pass`.  This also means that if a user's
password is changed the cache will need to expire (which takes 5 mins)
before it takes effect.

This can be used to build general purpose proxies to any kind of
backend that rclone supports.  


```
rclone serve ftp remote:path [flags]
```

### Options

```
      --addr string                            IPaddress:Port or :Port to bind server to. (default "localhost:2121")
      --auth-proxy string                      A program to use to create the backend from the auth.
      --dir-cache-time duration                Time to cache directory entries for. (default 5m0s)
      --dir-perms FileMode                     Directory permissions (default 0777)
      --file-perms FileMode                    File permissions (default 0666)
      --gid uint32                             Override the gid field set by the filesystem. (default 1000)
  -h, --help                                   help for ftp
      --no-checksum                            Don't compare checksums on up/download.
      --no-modtime                             Don't read/write the modification time (can speed things up).
      --no-seek                                Don't allow seeking in files.
      --pass string                            Password for authentication. (empty value allow every password)
      --passive-port string                    Passive port range to use. (default "30000-32000")
      --poll-interval duration                 Time to wait between polling for changes. Must be smaller than dir-cache-time. Only on supported remotes. Set to 0 to disable. (default 1m0s)
      --public-ip string                       Public IP address to advertise for passive connections.
      --read-only                              Mount read-only.
      --uid uint32                             Override the uid field set by the filesystem. (default 1000)
      --umask int                              Override the permission bits set by the filesystem. (default 2)
      --user string                            User name for authentication. (default "anonymous")
      --vfs-cache-max-age duration             Max age of objects in the cache. (default 1h0m0s)
      --vfs-cache-max-size SizeSuffix          Max total size of objects in the cache. (default off)
      --vfs-cache-mode CacheMode               Cache mode off|minimal|writes|full (default off)
      --vfs-cache-poll-interval duration       Interval to poll the cache for stale objects. (default 1m0s)
      --vfs-case-insensitive                   If a file name not found, find a case insensitive match.
      --vfs-read-chunk-size SizeSuffix         Read the source objects in chunks. (default 128M)
      --vfs-read-chunk-size-limit SizeSuffix   If greater than --vfs-read-chunk-size, double the chunk size after each chunk read, until the limit is reached. 'off' is unlimited. (default off)
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone serve](/commands/rclone_serve/)	 - Serve a remote over a protocol.


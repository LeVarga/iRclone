---
title: "Documentation"
description: "Rclone Usage"
date: "2019-02-25"
---

Configure
---------

First, you'll need to configure rclone.  As the object storage systems
have quite complicated authentication these are kept in a config file.
(See the `--config` entry for how to find the config file and choose
its location.)

The easiest way to make the config is to run rclone with the config
option:

    rclone config

See the following for detailed instructions for

  * [1Fichier](/fichier/)
  * [Alias](/alias/)
  * [Amazon Drive](/amazonclouddrive/)
  * [Amazon S3](/s3/)
  * [Backblaze B2](/b2/)
  * [Box](/box/)
  * [Cache](/cache/)
  * [Chunker](/chunker/) - transparently splits large files for other remotes
  * [Citrix ShareFile](/sharefile/)
  * [Crypt](/crypt/) - to encrypt other remotes
  * [DigitalOcean Spaces](/s3/#digitalocean-spaces)
  * [Dropbox](/dropbox/)
  * [FTP](/ftp/)
  * [Google Cloud Storage](/googlecloudstorage/)
  * [Google Drive](/drive/)
  * [Google Photos](/googlephotos/)
  * [HTTP](/http/)
  * [Hubic](/hubic/)
  * [Jottacloud](/jottacloud/)
  * [Koofr](/koofr/)
  * [Mail.ru Cloud](/mailru/)
  * [Mega](/mega/)
  * [Microsoft Azure Blob Storage](/azureblob/)
  * [Microsoft OneDrive](/onedrive/)
  * [Openstack Swift / Rackspace Cloudfiles / Memset Memstore](/swift/)
  * [OpenDrive](/opendrive/)
  * [Pcloud](/pcloud/)
  * [premiumize.me](/premiumizeme/)
  * [put.io](/putio/)
  * [QingStor](/qingstor/)
  * [SFTP](/sftp/)
  * [Union](/union/)
  * [WebDAV](/webdav/)
  * [Yandex Disk](/yandex/)
  * [The local filesystem](/local/)

Usage
-----

Rclone syncs a directory tree from one storage system to another.

Its syntax is like this

    Syntax: [options] subcommand <parameters> <parameters...>

Source and destination paths are specified by the name you gave the
storage system in the config file then the sub path, eg
"drive:myfolder" to look at "myfolder" in Google drive.

You can define as many storage paths as you like in the config file.

Subcommands
-----------

rclone uses a system of subcommands.  For example

    rclone ls remote:path # lists a remote
    rclone copy /local/path remote:path # copies /local/path to the remote
    rclone sync /local/path remote:path # syncs /local/path to the remote

The main rclone commands with most used first

* [rclone config](/commands/rclone_config/)	- Enter an interactive configuration session.
* [rclone copy](/commands/rclone_copy/)		- Copy files from source to dest, skipping already copied.
* [rclone sync](/commands/rclone_sync/)		- Make source and dest identical, modifying destination only.
* [rclone move](/commands/rclone_move/)		- Move files from source to dest.
* [rclone delete](/commands/rclone_delete/)	- Remove the contents of path.
* [rclone purge](/commands/rclone_purge/)	- Remove the path and all of its contents.
* [rclone mkdir](/commands/rclone_mkdir/)	- Make the path if it doesn't already exist.
* [rclone rmdir](/commands/rclone_rmdir/)	- Remove the path.
* [rclone rmdirs](/commands/rclone_rmdirs/)	- Remove any empty directories under the path.
* [rclone check](/commands/rclone_check/)	- Check if the files in the source and destination match.
* [rclone ls](/commands/rclone_ls/)		- List all the objects in the path with size and path.
* [rclone lsd](/commands/rclone_lsd/)		- List all directories/containers/buckets in the path.
* [rclone lsl](/commands/rclone_lsl/)		- List all the objects in the path with size, modification time and path.
* [rclone md5sum](/commands/rclone_md5sum/)	- Produce an md5sum file for all the objects in the path.
* [rclone sha1sum](/commands/rclone_sha1sum/)	- Produce a sha1sum file for all the objects in the path.
* [rclone size](/commands/rclone_size/)		- Return the total size and number of objects in remote:path.
* [rclone version](/commands/rclone_version/)	- Show the version number.
* [rclone cleanup](/commands/rclone_cleanup/)	- Clean up the remote if possible.
* [rclone dedupe](/commands/rclone_dedupe/)	- Interactively find duplicate files and delete/rename them.
* [rclone authorize](/commands/rclone_authorize/)	- Remote authorization.
* [rclone cat](/commands/rclone_cat/)		- Concatenate any files and send them to stdout.
* [rclone copyto](/commands/rclone_copyto/)	- Copy files from source to dest, skipping already copied.
* [rclone genautocomplete](/commands/rclone_genautocomplete/)	- Output shell completion scripts for rclone.
* [rclone gendocs](/commands/rclone_gendocs/)	- Output markdown docs for rclone to the directory supplied.
* [rclone listremotes](/commands/rclone_listremotes/)	- List all the remotes in the config file.
* [rclone mount](/commands/rclone_mount/)	- Mount the remote as a mountpoint.
* [rclone moveto](/commands/rclone_moveto/)	- Move file or directory from source to dest.
* [rclone obscure](/commands/rclone_obscure/)	- Obscure password for use in the rclone.conf
* [rclone cryptcheck](/commands/rclone_cryptcheck/)	- Check the integrity of a crypted remote.
* [rclone about](/commands/rclone_about/)	- Get quota information from the remote.

See the [commands index](/commands/) for the full list.

Copying single files
--------------------

rclone normally syncs or copies directories.  However, if the source
remote points to a file, rclone will just copy that file.  The
destination remote must point to a directory - rclone will give the
error `Failed to create file system for "remote:file": is a file not a
directory` if it isn't.

For example, suppose you have a remote with a file in called
`test.jpg`, then you could copy just that file like this

    rclone copy remote:test.jpg /tmp/download

The file `test.jpg` will be placed inside `/tmp/download`.

This is equivalent to specifying

    rclone copy --files-from /tmp/files remote: /tmp/download

Where `/tmp/files` contains the single line

    test.jpg

It is recommended to use `copy` when copying individual files, not `sync`.
They have pretty much the same effect but `copy` will use a lot less
memory.

Syntax of remote paths
----------------------

The syntax of the paths passed to the rclone command are as follows.

### /path/to/dir

This refers to the local file system.

On Windows only `\` may be used instead of `/` in local paths
**only**, non local paths must use `/`.

These paths needn't start with a leading `/` - if they don't then they
will be relative to the current directory.

### remote:path/to/dir

This refers to a directory `path/to/dir` on `remote:` as defined in
the config file (configured with `rclone config`).

### remote:/path/to/dir

On most backends this is refers to the same directory as
`remote:path/to/dir` and that format should be preferred.  On a very
small number of remotes (FTP, SFTP, Dropbox for business) this will
refer to a different directory.  On these, paths without a leading `/`
will refer to your "home" directory and paths with a leading `/` will
refer to the root.

### :backend:path/to/dir

This is an advanced form for creating remotes on the fly.  `backend`
should be the name or prefix of a backend (the `type` in the config
file) and all the configuration for the backend should be provided on
the command line (or in environment variables).

Here are some examples:

    rclone lsd --http-url https://pub.rclone.org :http:

To list all the directories in the root of `https://pub.rclone.org/`.

    rclone lsf --http-url https://example.com :http:path/to/dir

To list files and directories in `https://example.com/path/to/dir/`

    rclone copy --http-url https://example.com :http:path/to/dir /tmp/dir

To copy files and directories in `https://example.com/path/to/dir` to `/tmp/dir`.

    rclone copy --sftp-host example.com :sftp:path/to/dir /tmp/dir

To copy files and directories from `example.com` in the relative
directory `path/to/dir` to `/tmp/dir` using sftp.

Quoting and the shell
---------------------

When you are typing commands to your computer you are using something
called the command line shell.  This interprets various characters in
an OS specific way.

Here are some gotchas which may help users unfamiliar with the shell rules

### Linux / OSX ###

If your names have spaces or shell metacharacters (eg `*`, `?`, `$`,
`'`, `"` etc) then you must quote them.  Use single quotes `'` by default.

    rclone copy 'Important files?' remote:backup

If you want to send a `'` you will need to use `"`, eg

    rclone copy "O'Reilly Reviews" remote:backup

The rules for quoting metacharacters are complicated and if you want
the full details you'll have to consult the manual page for your
shell.

### Windows ###

If your names have spaces in you need to put them in `"`, eg

    rclone copy "E:\folder name\folder name\folder name" remote:backup

If you are using the root directory on its own then don't quote it
(see [#464](https://github.com/rclone/rclone/issues/464) for why), eg

    rclone copy E:\ remote:backup

Copying files or directories with `:` in the names
--------------------------------------------------

rclone uses `:` to mark a remote name.  This is, however, a valid
filename component in non-Windows OSes.  The remote name parser will
only search for a `:` up to the first `/` so if you need to act on a
file or directory like this then use the full path starting with a
`/`, or use `./` as a current directory prefix.

So to sync a directory called `sync:me` to a remote called `remote:` use

    rclone sync ./sync:me remote:path

or

    rclone sync /full/path/to/sync:me remote:path

Server Side Copy
----------------

Most remotes (but not all - see [the
overview](/overview/#optional-features)) support server side copy.

This means if you want to copy one folder to another then rclone won't
download all the files and re-upload them; it will instruct the server
to copy them in place.

Eg

    rclone copy s3:oldbucket s3:newbucket

Will copy the contents of `oldbucket` to `newbucket` without
downloading and re-uploading.

Remotes which don't support server side copy **will** download and
re-upload in this case.

Server side copies are used with `sync` and `copy` and will be
identified in the log when using the `-v` flag.  The `move` command
may also use them if remote doesn't support server side move directly.
This is done by issuing a server side copy then a delete which is much
quicker than a download and re-upload.

Server side copies will only be attempted if the remote names are the
same.

This can be used when scripting to make aged backups efficiently, eg

    rclone sync remote:current-backup remote:previous-backup
    rclone sync /path/to/files remote:current-backup

Options
-------

Rclone has a number of options to control its behaviour.

Options that take parameters can have the values passed in two ways,
`--option=value` or `--option value`. However boolean (true/false)
options behave slightly differently to the other options in that
`--boolean` sets the option to `true` and the absence of the flag sets
it to `false`.  It is also possible to specify `--boolean=false` or
`--boolean=true`.  Note that `--boolean false` is not valid - this is
parsed as `--boolean` and the `false` is parsed as an extra command
line argument for rclone.

Options which use TIME use the go time parser.  A duration string is a
possibly signed sequence of decimal numbers, each with optional
fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid
time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

Options which use SIZE use kByte by default.  However, a suffix of `b`
for bytes, `k` for kBytes, `M` for MBytes, `G` for GBytes, `T` for
TBytes and `P` for PBytes may be used.  These are the binary units, eg
1, 2\*\*10, 2\*\*20, 2\*\*30 respectively.

### --backup-dir=DIR ###

When using `sync`, `copy` or `move` any files which would have been
overwritten or deleted are moved in their original hierarchy into this
directory.

If `--suffix` is set, then the moved files will have the suffix added
to them.  If there is a file with the same path (after the suffix has
been added) in DIR, then it will be overwritten.

The remote in use must support server side move or copy and you must
use the same remote as the destination of the sync.  The backup
directory must not overlap the destination directory.

For example

    rclone sync /path/to/local remote:current --backup-dir remote:old

will sync `/path/to/local` to `remote:current`, but for any files
which would have been updated or deleted will be stored in
`remote:old`.

If running rclone from a script you might want to use today's date as
the directory name passed to `--backup-dir` to store the old files, or
you might want to pass `--suffix` with today's date.

See `--compare-dest` and `--copy-dest`.

### --bind string ###

Local address to bind to for outgoing connections.  This can be an
IPv4 address (1.2.3.4), an IPv6 address (1234::789A) or host name.  If
the host name doesn't resolve or resolves to more than one IP address
it will give an error.

### --bwlimit=BANDWIDTH_SPEC ###

This option controls the bandwidth limit. Limits can be specified
in two ways: As a single limit, or as a timetable.

Single limits last for the duration of the session. To use a single limit,
specify the desired bandwidth in kBytes/s, or use a suffix b|k|M|G.  The
default is `0` which means to not limit bandwidth.

For example, to limit bandwidth usage to 10 MBytes/s use `--bwlimit 10M`

It is also possible to specify a "timetable" of limits, which will cause
certain limits to be applied at certain times. To specify a timetable, format your
entries as "WEEKDAY-HH:MM,BANDWIDTH WEEKDAY-HH:MM,BANDWIDTH..." where:
WEEKDAY is optional element.
It could be written as whole world or only using 3 first characters.
HH:MM is an hour from 00:00 to 23:59.

An example of a typical timetable to avoid link saturation during daytime
working hours could be:

`--bwlimit "08:00,512 12:00,10M 13:00,512 18:00,30M 23:00,off"`

In this example, the transfer bandwidth will be every day set to 512kBytes/sec at 8am.
At noon, it will raise to 10Mbytes/s, and drop back to 512kBytes/sec at 1pm.
At 6pm, the bandwidth limit will be set to 30MBytes/s, and at 11pm it will be
completely disabled (full speed). Anything between 11pm and 8am will remain
unlimited.

An example of timetable with WEEKDAY could be:

`--bwlimit "Mon-00:00,512 Fri-23:59,10M Sat-10:00,1M Sun-20:00,off"`

It mean that, the transfer bandwidth will be set to 512kBytes/sec on Monday.
It will raise to 10Mbytes/s before the end of Friday. 
At 10:00 on Sunday it will be set to 1Mbyte/s.
From 20:00 at Sunday will be unlimited.

Timeslots without weekday are extended to whole week.
So this one example:

`--bwlimit "Mon-00:00,512 12:00,1M Sun-20:00,off"`

Is equal to this:

`--bwlimit "Mon-00:00,512Mon-12:00,1M Tue-12:00,1M Wed-12:00,1M Thu-12:00,1M Fri-12:00,1M Sat-12:00,1M Sun-12:00,1M Sun-20:00,off"`

Bandwidth limits only apply to the data transfer. They don't apply to the
bandwidth of the directory listings etc.

Note that the units are Bytes/s, not Bits/s.  Typically connections are
measured in Bits/s - to convert divide by 8.  For example, let's say
you have a 10 Mbit/s connection and you wish rclone to use half of it
- 5 Mbit/s.  This is 5/8 = 0.625MByte/s so you would use a `--bwlimit
0.625M` parameter for rclone.

On Unix systems (Linux, MacOS, …) the bandwidth limiter can be toggled by
sending a `SIGUSR2` signal to rclone. This allows to remove the limitations
of a long running rclone transfer and to restore it back to the value specified
with `--bwlimit` quickly when needed. Assuming there is only one rclone instance
running, you can toggle the limiter like this:

    kill -SIGUSR2 $(pidof rclone)

If you configure rclone with a [remote control](/rc) then you can use
change the bwlimit dynamically:

    rclone rc core/bwlimit rate=1M

### --buffer-size=SIZE ###

Use this sized buffer to speed up file transfers.  Each `--transfer`
will use this much memory for buffering.

When using `mount` or `cmount` each open file descriptor will use this much
memory for buffering.
See the [mount](/commands/rclone_mount/#file-buffering) documentation for more details.

Set to 0 to disable the buffering for the minimum memory usage.

Note that the memory allocation of the buffers is influenced by the
[--use-mmap](#use-mmap) flag.

### --checkers=N ###

The number of checkers to run in parallel.  Checkers do the equality
checking of files during a sync.  For some storage systems (eg S3,
Swift, Dropbox) this can take a significant amount of time so they are
run in parallel.

The default is to run 8 checkers in parallel.

### -c, --checksum ###

Normally rclone will look at modification time and size of files to
see if they are equal.  If you set this flag then rclone will check
the file hash and size to determine if files are equal.

This is useful when the remote doesn't support setting modified time
and a more accurate sync is desired than just checking the file size.

This is very useful when transferring between remotes which store the
same hash type on the object, eg Drive and Swift. For details of which
remotes support which hash type see the table in the [overview
section](/overview/).

Eg `rclone --checksum sync s3:/bucket swift:/bucket` would run much
quicker than without the `--checksum` flag.

When using this flag, rclone won't update mtimes of remote files if
they are incorrect as it would normally.

### --compare-dest=DIR ###

When using `sync`, `copy` or `move` DIR is checked in addition to the 
destination for files. If a file identical to the source is found that 
file is NOT copied from source. This is useful to copy just files that 
have changed since the last backup.

You must use the same remote as the destination of the sync.  The 
compare directory must not overlap the destination directory.

See `--copy-dest` and `--backup-dir`.

### --config=CONFIG_FILE ###

Specify the location of the rclone config file.

Normally the config file is in your home directory as a file called
`.config/rclone/rclone.conf` (or `.rclone.conf` if created with an
older version). If `$XDG_CONFIG_HOME` is set it will be at
`$XDG_CONFIG_HOME/rclone/rclone.conf`.

If there is a file `rclone.conf` in the same directory as the rclone
executable it will be preferred. This file must be created manually
for Rclone to use it, it will never be created automatically.

If you run `rclone config file` you will see where the default
location is for you.

Use this flag to override the config location, eg `rclone
--config=".myconfig" .config`.

### --contimeout=TIME ###

Set the connection timeout. This should be in go time format which
looks like `5s` for 5 seconds, `10m` for 10 minutes, or `3h30m`.

The connection timeout is the amount of time rclone will wait for a
connection to go through to a remote object storage system.  It is
`1m` by default.

### --copy-dest=DIR ###

When using `sync`, `copy` or `move` DIR is checked in addition to the 
destination for files. If a file identical to the source is found that 
file is server side copied from DIR to the destination. This is useful 
for incremental backup.

The remote in use must support server side copy and you must
use the same remote as the destination of the sync.  The compare
directory must not overlap the destination directory.

See `--compare-dest` and `--backup-dir`.

### --dedupe-mode MODE ###

Mode to run dedupe command in.  One of `interactive`, `skip`, `first`, `newest`, `oldest`, `rename`.  The default is `interactive`.  See the dedupe command for more information as to what these options mean.

### --disable FEATURE,FEATURE,... ###

This disables a comma separated list of optional features. For example
to disable server side move and server side copy use:

    --disable move,copy

The features can be put in in any case.

To see a list of which features can be disabled use:

    --disable help

See the overview [features](/overview/#features) and
[optional features](/overview/#optional-features) to get an idea of
which feature does what.

This flag can be useful for debugging and in exceptional circumstances
(eg Google Drive limiting the total volume of Server Side Copies to
100GB/day).

### -n, --dry-run ###

Do a trial run with no permanent changes.  Use this to see what rclone
would do without actually doing it.  Useful when setting up the `sync`
command which deletes files in the destination.

### --ignore-case-sync ###

Using this option will cause rclone to ignore the case of the files 
when synchronizing so files will not be copied/synced when the
existing filenames are the same, even if the casing is different.

### --ignore-checksum ###

Normally rclone will check that the checksums of transferred files
match, and give an error "corrupted on transfer" if they don't.

You can use this option to skip that check.  You should only use it if
you have had the "corrupted on transfer" error message and you are
sure you might want to transfer potentially corrupted data.

### --ignore-existing ###

Using this option will make rclone unconditionally skip all files
that exist on the destination, no matter the content of these files.

While this isn't a generally recommended option, it can be useful
in cases where your files change due to encryption. However, it cannot
correct partial transfers in case a transfer was interrupted.

### --ignore-size ###

Normally rclone will look at modification time and size of files to
see if they are equal.  If you set this flag then rclone will check
only the modification time.  If `--checksum` is set then it only
checks the checksum.

It will also cause rclone to skip verifying the sizes are the same
after transfer.

This can be useful for transferring files to and from OneDrive which
occasionally misreports the size of image files (see
[#399](https://github.com/rclone/rclone/issues/399) for more info).

### -I, --ignore-times ###

Using this option will cause rclone to unconditionally upload all
files regardless of the state of files on the destination.

Normally rclone would skip any files that have the same
modification time and are the same size (or have the same checksum if
using `--checksum`).

### --immutable ###

Treat source and destination files as immutable and disallow
modification.

With this option set, files will be created and deleted as requested,
but existing files will never be updated.  If an existing file does
not match between the source and destination, rclone will give the error
`Source and destination exist but do not match: immutable file modified`.

Note that only commands which transfer files (e.g. `sync`, `copy`,
`move`) are affected by this behavior, and only modification is
disallowed.  Files may still be deleted explicitly (e.g. `delete`,
`purge`) or implicitly (e.g. `sync`, `move`).  Use `copy --immutable`
if it is desired to avoid deletion as well as modification.

This can be useful as an additional layer of protection for immutable
or append-only data sets (notably backup archives), where modification
implies corruption and should not be propagated.

## --leave-root ###

During rmdirs it will not remove root directory, even if it's empty.

### --log-file=FILE ###

Log all of rclone's output to FILE.  This is not active by default.
This can be useful for tracking down problems with syncs in
combination with the `-v` flag.  See the [Logging section](#logging)
for more info.

Note that if you are using the `logrotate` program to manage rclone's
logs, then you should use the `copytruncate` option as rclone doesn't
have a signal to rotate logs.

### --log-format LIST ###

Comma separated list of log format options. `date`, `time`, `microseconds`, `longfile`, `shortfile`, `UTC`.  The default is "`date`,`time`". 

### --log-level LEVEL ###

This sets the log level for rclone.  The default log level is `NOTICE`.

`DEBUG` is equivalent to `-vv`. It outputs lots of debug info - useful
for bug reports and really finding out what rclone is doing.

`INFO` is equivalent to `-v`. It outputs information about each transfer
and prints stats once a minute by default.

`NOTICE` is the default log level if no logging flags are supplied. It
outputs very little when things are working normally. It outputs
warnings and significant events.

`ERROR` is equivalent to `-q`. It only outputs error messages.

### --use-json-log ###

This switches the log format to JSON for rclone. The fields of json log 
are level, msg, source, time.

### --low-level-retries NUMBER ###

This controls the number of low level retries rclone does.

A low level retry is used to retry a failing operation - typically one
HTTP request.  This might be uploading a chunk of a big file for
example.  You will see low level retries in the log with the `-v`
flag.

This shouldn't need to be changed from the default in normal operations.
However, if you get a lot of low level retries you may wish
to reduce the value so rclone moves on to a high level retry (see the
`--retries` flag) quicker.

Disable low level retries with `--low-level-retries 1`.

### --max-backlog=N ###

This is the maximum allowable backlog of files in a sync/copy/move
queued for being checked or transferred.

This can be set arbitrarily large.  It will only use memory when the
queue is in use.  Note that it will use in the order of N kB of memory
when the backlog is in use.

Setting this large allows rclone to calculate how many files are
pending more accurately and give a more accurate estimated finish
time.

Setting this small will make rclone more synchronous to the listings
of the remote which may be desirable.

### --max-delete=N ###

This tells rclone not to delete more than N files.  If that limit is
exceeded then a fatal error will be generated and rclone will stop the
operation in progress.

### --max-depth=N ###

This modifies the recursion depth for all the commands except purge.

So if you do `rclone --max-depth 1 ls remote:path` you will see only
the files in the top level directory.  Using `--max-depth 2` means you
will see all the files in first two directory levels and so on.

For historical reasons the `lsd` command defaults to using a
`--max-depth` of 1 - you can override this with the command line flag.

You can use this command to disable recursion (with `--max-depth 1`).

Note that if you use this with `sync` and `--delete-excluded` the
files not recursed through are considered excluded and will be deleted
on the destination.  Test first with `--dry-run` if you are not sure
what will happen.

### --max-transfer=SIZE ###

Rclone will stop transferring when it has reached the size specified.
Defaults to off.

When the limit is reached all transfers will stop immediately.

Rclone will exit with exit code 8 if the transfer limit is reached.

### --modify-window=TIME ###

When checking whether a file has been modified, this is the maximum
allowed time difference that a file can have and still be considered
equivalent.

The default is `1ns` unless this is overridden by a remote.  For
example OS X only stores modification times to the nearest second so
if you are reading and writing to an OS X filing system this will be
`1s` by default.

This command line flag allows you to override that computed default.

### --multi-thread-cutoff=SIZE ###

When downloading files to the local backend above this size, rclone
will use multiple threads to download the file. (default 250M)

Rclone preallocates the file (using `fallocate(FALLOC_FL_KEEP_SIZE)`
on unix or `NTSetInformationFile` on Windows both of which takes no
time) then each thread writes directly into the file at the correct
place.  This means that rclone won't create fragmented or sparse files
and there won't be any assembly time at the end of the transfer.

The number of threads used to dowload is controlled by
`--multi-thread-streams`.

Use `-vv` if you wish to see info about the threads.

This will work with the `sync`/`copy`/`move` commands and friends
`copyto`/`moveto`.  Multi thread downloads will be used with `rclone
mount` and `rclone serve` if `--vfs-cache-mode` is set to `writes` or
above.

**NB** that this **only** works for a local destination but will work
with any source.

**NB** that multi thread copies are disabled for local to local copies
as they are faster without unless `--multi-thread-streams` is set
explicitly.

### --multi-thread-streams=N ###

When using multi thread downloads (see above `--multi-thread-cutoff`)
this sets the maximum number of streams to use.  Set to `0` to disable
multi thread downloads. (Default 4)

Exactly how many streams rclone uses for the download depends on the
size of the file. To calculate the number of download streams Rclone
divides the size of the file by the `--multi-thread-cutoff` and rounds
up, up to the maximum set with `--multi-thread-streams`.

So if `--multi-thread-cutoff 250MB` and `--multi-thread-streams 4` are
in effect (the defaults):

- 0MB.250MB files will be downloaded with 1 stream
- 250MB..500MB files will be downloaded with 2 streams
- 500MB..750MB files will be downloaded with 3 streams
- 750MB+ files will be downloaded with 4 streams

### --no-gzip-encoding ###

Don't set `Accept-Encoding: gzip`.  This means that rclone won't ask
the server for compressed files automatically. Useful if you've set
the server to return files with `Content-Encoding: gzip` but you
uploaded compressed files.

There is no need to set this in normal operation, and doing so will
decrease the network transfer efficiency of rclone.

### --no-traverse ###

The `--no-traverse` flag controls whether the destination file system
is traversed when using the `copy` or `move` commands.
`--no-traverse` is not compatible with `sync` and will be ignored if
you supply it with `sync`.

If you are only copying a small number of files (or are filtering most
of the files) and/or have a large number of files on the destination
then `--no-traverse` will stop rclone listing the destination and save
time.

However, if you are copying a large number of files, especially if you
are doing a copy where lots of the files under consideration haven't
changed and won't need copying then you shouldn't use `--no-traverse`.

See [rclone copy](/commands/rclone_copy/) for an example of how to use it.

### --no-update-modtime ###

When using this flag, rclone won't update modification times of remote
files if they are incorrect as it would normally.

This can be used if the remote is being synced with another tool also
(eg the Google Drive client).

### -P, --progress ###

This flag makes rclone update the stats in a static block in the
terminal providing a realtime overview of the transfer.

Any log messages will scroll above the static block.  Log messages
will push the static block down to the bottom of the terminal where it
will stay.

Normally this is updated every 500mS but this period can be overridden
with the `--stats` flag.

This can be used with the `--stats-one-line` flag for a simpler
display.

Note: On Windows until[this bug](https://github.com/Azure/go-ansiterm/issues/26)
is fixed all non-ASCII characters will be replaced with `.` when
`--progress` is in use.

### -q, --quiet ###

Normally rclone outputs stats and a completion message.  If you set
this flag it will make as little output as possible.

### --retries int ###

Retry the entire sync if it fails this many times it fails (default 3).

Some remotes can be unreliable and a few retries help pick up the
files which didn't get transferred because of errors.

Disable retries with `--retries 1`.

### --retries-sleep=TIME ###

This sets the interval between each retry specified by `--retries` 

The default is 0. Use 0 to disable.

### --size-only ###

Normally rclone will look at modification time and size of files to
see if they are equal.  If you set this flag then rclone will check
only the size.

This can be useful transferring files from Dropbox which have been
modified by the desktop sync client which doesn't set checksums of
modification times in the same way as rclone.

### --stats=TIME ###

Commands which transfer data (`sync`, `copy`, `copyto`, `move`,
`moveto`) will print data transfer stats at regular intervals to show
their progress.

This sets the interval.

The default is `1m`. Use 0 to disable.

If you set the stats interval then all commands can show stats.  This
can be useful when running other commands, `check` or `mount` for
example.

Stats are logged at `INFO` level by default which means they won't
show at default log level `NOTICE`.  Use `--stats-log-level NOTICE` or
`-v` to make them show.  See the [Logging section](#logging) for more
info on log levels.

Note that on macOS you can send a SIGINFO (which is normally ctrl-T in
the terminal) to make the stats print immediately.

### --stats-file-name-length integer ###
By default, the `--stats` output will truncate file names and paths longer 
than 40 characters.  This is equivalent to providing 
`--stats-file-name-length 40`. Use `--stats-file-name-length 0` to disable 
any truncation of file names printed by stats.

### --stats-log-level string ###

Log level to show `--stats` output at.  This can be `DEBUG`, `INFO`,
`NOTICE`, or `ERROR`.  The default is `INFO`.  This means at the
default level of logging which is `NOTICE` the stats won't show - if
you want them to then use `--stats-log-level NOTICE`.  See the [Logging
section](#logging) for more info on log levels.

### --stats-one-line ###

When this is specified, rclone condenses the stats into a single line
showing the most important stats only.

### --stats-one-line-date ###

When this is specified, rclone enables the single-line stats and prepends
the display with a date string. The default is `2006/01/02 15:04:05 - `

### --stats-one-line-date-format ###

When this is specified, rclone enables the single-line stats and prepends
the display with a user-supplied date string. The date string MUST be
enclosed in quotes. Follow [golang specs](https://golang.org/pkg/time/#Time.Format) for
date formatting syntax.

### --stats-unit=bits|bytes ###

By default, data transfer rates will be printed in bytes/second.

This option allows the data rate to be printed in bits/second.

Data transfer volume will still be reported in bytes.

The rate is reported as a binary unit, not SI unit. So 1 Mbit/s
equals 1,048,576 bits/s and not 1,000,000 bits/s.

The default is `bytes`.

### --suffix=SUFFIX ###

When using `sync`, `copy` or `move` any files which would have been
overwritten or deleted will have the suffix added to them.  If there 
is a file with the same path (after the suffix has been added), then 
it will be overwritten.

The remote in use must support server side move or copy and you must
use the same remote as the destination of the sync.

This is for use with files to add the suffix in the current directory 
or with `--backup-dir`. See `--backup-dir` for more info.

For example

    rclone sync /path/to/local/file remote:current --suffix .bak

will sync `/path/to/local` to `remote:current`, but for any files
which would have been updated or deleted have .bak added.

### --suffix-keep-extension ###

When using `--suffix`, setting this causes rclone put the SUFFIX
before the extension of the files that it backs up rather than after.

So let's say we had `--suffix -2019-01-01`, without the flag `file.txt`
would be backed up to `file.txt-2019-01-01` and with the flag it would
be backed up to `file-2019-01-01.txt`.  This can be helpful to make
sure the suffixed files can still be opened.

### --syslog ###

On capable OSes (not Windows or Plan9) send all log output to syslog.

This can be useful for running rclone in a script or `rclone mount`.

### --syslog-facility string ###

If using `--syslog` this sets the syslog facility (eg `KERN`, `USER`).
See `man syslog` for a list of possible facilities.  The default
facility is `DAEMON`.

### --tpslimit float ###

Limit HTTP transactions per second to this. Default is 0 which is used
to mean unlimited transactions per second.

For example to limit rclone to 10 HTTP transactions per second use
`--tpslimit 10`, or to 1 transaction every 2 seconds use `--tpslimit
0.5`.

Use this when the number of transactions per second from rclone is
causing a problem with the cloud storage provider (eg getting you
banned or rate limited).

This can be very useful for `rclone mount` to control the behaviour of
applications using it.

See also `--tpslimit-burst`.

### --tpslimit-burst int ###

Max burst of transactions for `--tpslimit`. (default 1)

Normally `--tpslimit` will do exactly the number of transaction per
second specified.  However if you supply `--tps-burst` then rclone can
save up some transactions from when it was idle giving a burst of up
to the parameter supplied.

For example if you provide `--tpslimit-burst 10` then if rclone has
been idle for more than 10*`--tpslimit` then it can do 10 transactions
very quickly before they are limited again.

This may be used to increase performance of `--tpslimit` without
changing the long term average number of transactions per second.

### --track-renames ###

By default, rclone doesn't keep track of renamed files, so if you
rename a file locally then sync it to a remote, rclone will delete the
old file on the remote and upload a new copy.

If you use this flag, and the remote supports server side copy or
server side move, and the source and destination have a compatible
hash, then this will track renames during `sync`
operations and perform renaming server-side.

Files will be matched by size and hash - if both match then a rename
will be considered.

If the destination does not support server-side copy or move, rclone
will fall back to the default behaviour and log an error level message
to the console. Note: Encrypted destinations are not supported
by `--track-renames`.

Note that `--track-renames` is incompatible with `--no-traverse` and
that it uses extra memory to keep track of all the rename candidates.

Note also that `--track-renames` is incompatible with
`--delete-before` and will select `--delete-after` instead of
`--delete-during`.

### --delete-(before,during,after) ###

This option allows you to specify when files on your destination are
deleted when you sync folders.

Specifying the value `--delete-before` will delete all files present
on the destination, but not on the source *before* starting the
transfer of any new or updated files. This uses two passes through the
file systems, one for the deletions and one for the copies.

Specifying `--delete-during` will delete files while checking and
uploading files. This is the fastest option and uses the least memory.

Specifying `--delete-after` (the default value) will delay deletion of
files until all new/updated files have been successfully transferred.
The files to be deleted are collected in the copy pass then deleted
after the copy pass has completed successfully.  The files to be
deleted are held in memory so this mode may use more memory.  This is
the safest mode as it will only delete files if there have been no
errors subsequent to that.  If there have been errors before the
deletions start then you will get the message `not deleting files as
there were IO errors`.

### --fast-list ###

When doing anything which involves a directory listing (eg `sync`,
`copy`, `ls` - in fact nearly every command), rclone normally lists a
directory and processes it before using more directory lists to
process any subdirectories.  This can be parallelised and works very
quickly using the least amount of memory.

However, some remotes have a way of listing all files beneath a
directory in one (or a small number) of transactions.  These tend to
be the bucket based remotes (eg S3, B2, GCS, Swift, Hubic).

If you use the `--fast-list` flag then rclone will use this method for
listing directories.  This will have the following consequences for
the listing:

  * It **will** use fewer transactions (important if you pay for them)
  * It **will** use more memory.  Rclone has to load the whole listing into memory.
  * It *may* be faster because it uses fewer transactions
  * It *may* be slower because it can't be parallelized

rclone should always give identical results with and without
`--fast-list`.

If you pay for transactions and can fit your entire sync listing into
memory then `--fast-list` is recommended.  If you have a very big sync
to do then don't use `--fast-list` otherwise you will run out of
memory.

If you use `--fast-list` on a remote which doesn't support it, then
rclone will just ignore it.

### --timeout=TIME ###

This sets the IO idle timeout.  If a transfer has started but then
becomes idle for this long it is considered broken and disconnected.

The default is `5m`.  Set to 0 to disable.

### --transfers=N ###

The number of file transfers to run in parallel.  It can sometimes be
useful to set this to a smaller number if the remote is giving a lot
of timeouts or bigger if you have lots of bandwidth and a fast remote.

The default is to run 4 file transfers in parallel.

### -u, --update ###

This forces rclone to skip any files which exist on the destination
and have a modified time that is newer than the source file.

This can be useful when transferring to a remote which doesn't support
mod times directly (or when using `--use-server-modtime` to avoid extra
API calls) as it is more accurate than a `--size-only` check and faster
than using `--checksum`.

If an existing destination file has a modification time equal (within
the computed modify window precision) to the source file's, it will be
updated if the sizes are different.  If `--checksum` is set then
rclone will update the destination if the checksums differ too.

If an existing destination file is older than the source file then
it will be updated if the size or checksum differs from the source file.

On remotes which don't support mod time directly (or when using
`--use-server-modtime`) the time checked will be the uploaded time.
This means that if uploading to one of these remotes, rclone will skip
any files which exist on the destination and have an uploaded time that
is newer than the modification time of the source file.

### --use-mmap ###

If this flag is set then rclone will use anonymous memory allocated by
mmap on Unix based platforms and VirtualAlloc on Windows for its
transfer buffers (size controlled by `--buffer-size`).  Memory
allocated like this does not go on the Go heap and can be returned to
the OS immediately when it is finished with.

If this flag is not set then rclone will allocate and free the buffers
using the Go memory allocator which may use more memory as memory
pages are returned less aggressively to the OS.

It is possible this does not work well on all platforms so it is
disabled by default; in the future it may be enabled by default.

### --use-server-modtime ###

Some object-store backends (e.g, Swift, S3) do not preserve file modification
times (modtime). On these backends, rclone stores the original modtime as
additional metadata on the object. By default it will make an API call to
retrieve the metadata when the modtime is needed by an operation.

Use this flag to disable the extra API call and rely instead on the server's
modified time. In cases such as a local to remote sync using `--update`,
knowing the local file is newer than the time it was last uploaded to the
remote is sufficient. In those cases, this flag can speed up the process and
reduce the number of API calls necessary.

Using this flag on a sync operation without also using `--update` would cause
all files modified at any time other than the last upload time to be uploaded
again, which is probably not what you want.

### -v, -vv, --verbose ###

With `-v` rclone will tell you about each file that is transferred and
a small number of significant events.

With `-vv` rclone will become very verbose telling you about every
file it considers and transfers.  Please send bug reports with a log
with this setting.

### -V, --version ###

Prints the version number

SSL/TLS options
---------------

The outoing SSL/TLS connections rclone makes can be controlled with
these options.  For example this can be very useful with the HTTP or
WebDAV backends. Rclone HTTP servers have their own set of
configuration for SSL/TLS which you can find in their documentation.

### --ca-cert string

This loads the PEM encoded certificate authority certificate and uses
it to verify the certificates of the servers rclone connects to.

If you have generated certificates signed with a local CA then you
will need this flag to connect to servers using those certificates.

### --client-cert string

This loads the PEM encoded client side certificate.

This is used for [mutual TLS authentication](https://en.wikipedia.org/wiki/Mutual_authentication).

The `--client-key` flag is required too when using this.

### --client-key string

This loads the PEM encoded client side private key used for mutual TLS
authentication.  Used in conjunction with `--client-cert`.

### --no-check-certificate=true/false ###

`--no-check-certificate` controls whether a client verifies the
server's certificate chain and host name.
If `--no-check-certificate` is true, TLS accepts any certificate
presented by the server and any host name in that certificate.
In this mode, TLS is susceptible to man-in-the-middle attacks.

This option defaults to `false`.

**This should be used only for testing.**

Configuration Encryption
------------------------
Your configuration file contains information for logging in to 
your cloud services. This means that you should keep your 
`.rclone.conf` file in a secure location.

If you are in an environment where that isn't possible, you can
add a password to your configuration. This means that you will
have to enter the password every time you start rclone.

To add a password to your rclone configuration, execute `rclone config`.

```
>rclone config
Current remotes:

e) Edit existing remote
n) New remote
d) Delete remote
s) Set configuration password
q) Quit config
e/n/d/s/q>
```

Go into `s`, Set configuration password:
```
e/n/d/s/q> s
Your configuration is not encrypted.
If you add a password, you will protect your login information to cloud services.
a) Add Password
q) Quit to main menu
a/q> a
Enter NEW configuration password:
password:
Confirm NEW password:
password:
Password set
Your configuration is encrypted.
c) Change Password
u) Unencrypt configuration
q) Quit to main menu
c/u/q>
```

Your configuration is now encrypted, and every time you start rclone
you will now be asked for the password. In the same menu, you can
change the password or completely remove encryption from your
configuration.

There is no way to recover the configuration if you lose your password.

rclone uses [nacl secretbox](https://godoc.org/golang.org/x/crypto/nacl/secretbox) 
which in turn uses XSalsa20 and Poly1305 to encrypt and authenticate 
your configuration with secret-key cryptography.
The password is SHA-256 hashed, which produces the key for secretbox.
The hashed password is not stored.

While this provides very good security, we do not recommend storing
your encrypted rclone configuration in public if it contains sensitive
information, maybe except if you use a very strong password.

If it is safe in your environment, you can set the `RCLONE_CONFIG_PASS`
environment variable to contain your password, in which case it will be
used for decrypting the configuration.

You can set this for a session from a script.  For unix like systems
save this to a file called `set-rclone-password`:

```
#!/bin/echo Source this file don't run it

read -s RCLONE_CONFIG_PASS
export RCLONE_CONFIG_PASS
```

Then source the file when you want to use it.  From the shell you
would do `source set-rclone-password`.  It will then ask you for the
password and set it in the environment variable.

If you are running rclone inside a script, you might want to disable 
password prompts. To do that, pass the parameter 
`--ask-password=false` to rclone. This will make rclone fail instead
of asking for a password if `RCLONE_CONFIG_PASS` doesn't contain
a valid password.


Developer options
-----------------

These options are useful when developing or debugging rclone.  There
are also some more remote specific options which aren't documented
here which are used for testing.  These start with remote name eg
`--drive-test-option` - see the docs for the remote in question.

### --cpuprofile=FILE ###

Write CPU profile to file.  This can be analysed with `go tool pprof`.

#### --dump flag,flag,flag ####

The `--dump` flag takes a comma separated list of flags to dump info
about.  These are:

#### --dump headers ####

Dump HTTP headers with `Authorization:` lines removed. May still
contain sensitive info.  Can be very verbose.  Useful for debugging
only.

Use `--dump auth` if you do want the `Authorization:` headers.

#### --dump bodies ####

Dump HTTP headers and bodies - may contain sensitive info.  Can be
very verbose.  Useful for debugging only.

Note that the bodies are buffered in memory so don't use this for
enormous files.

#### --dump requests ####

Like `--dump bodies` but dumps the request bodies and the response
headers.  Useful for debugging download problems.

#### --dump responses ####

Like `--dump bodies` but dumps the response bodies and the request
headers. Useful for debugging upload problems.

#### --dump auth ####

Dump HTTP headers - will contain sensitive info such as
`Authorization:` headers - use `--dump headers` to dump without
`Authorization:` headers.  Can be very verbose.  Useful for debugging
only.

#### --dump filters ####

Dump the filters to the output.  Useful to see exactly what include
and exclude options are filtering on.

#### --dump goroutines ####

This dumps a list of the running go-routines at the end of the command
to standard output.

#### --dump openfiles ####

This dumps a list of the open files at the end of the command.  It
uses the `lsof` command to do that so you'll need that installed to
use it.

### --memprofile=FILE ###

Write memory profile to file. This can be analysed with `go tool pprof`.

Filtering
---------

For the filtering options

  * `--delete-excluded`
  * `--filter`
  * `--filter-from`
  * `--exclude`
  * `--exclude-from`
  * `--include`
  * `--include-from`
  * `--files-from`
  * `--min-size`
  * `--max-size`
  * `--min-age`
  * `--max-age`
  * `--dump filters`

See the [filtering section](/filtering/).

Remote control
--------------

For the remote control options and for instructions on how to remote control rclone

  * `--rc`
  * and anything starting with `--rc-`

See [the remote control section](/rc/).

Logging
-------

rclone has 4 levels of logging, `ERROR`, `NOTICE`, `INFO` and `DEBUG`.

By default, rclone logs to standard error.  This means you can redirect
standard error and still see the normal output of rclone commands (eg
`rclone ls`).

By default, rclone will produce `Error` and `Notice` level messages.

If you use the `-q` flag, rclone will only produce `Error` messages.

If you use the `-v` flag, rclone will produce `Error`, `Notice` and
`Info` messages.

If you use the `-vv` flag, rclone will produce `Error`, `Notice`,
`Info` and `Debug` messages.

You can also control the log levels with the `--log-level` flag.

If you use the `--log-file=FILE` option, rclone will redirect `Error`,
`Info` and `Debug` messages along with standard error to FILE.

If you use the `--syslog` flag then rclone will log to syslog and the
`--syslog-facility` control which facility it uses.

Rclone prefixes all log messages with their level in capitals, eg INFO
which makes it easy to grep the log file for different kinds of
information.

Exit Code
---------

If any errors occur during the command execution, rclone will exit with a
non-zero exit code.  This allows scripts to detect when rclone
operations have failed.

During the startup phase, rclone will exit immediately if an error is
detected in the configuration.  There will always be a log message
immediately before exiting.

When rclone is running it will accumulate errors as it goes along, and
only exit with a non-zero exit code if (after retries) there were
still failed transfers.  For every error counted there will be a high
priority log message (visible with `-q`) showing the message and
which file caused the problem. A high priority message is also shown
when starting a retry so the user can see that any previous error
messages may not be valid after the retry. If rclone has done a retry
it will log a high priority message if the retry was successful.

### List of exit codes ###
  * `0` - success
  * `1` - Syntax or usage error
  * `2` - Error not otherwise categorised
  * `3` - Directory not found
  * `4` - File not found
  * `5` - Temporary error (one that more retries might fix) (Retry errors)
  * `6` - Less serious errors (like 461 errors from dropbox) (NoRetry errors)
  * `7` - Fatal error (one that more retries won't fix, like account suspended) (Fatal errors)
  * `8` - Transfer exceeded - limit set by --max-transfer reached

Environment Variables
---------------------

Rclone can be configured entirely using environment variables.  These
can be used to set defaults for options or config file entries.

### Options ###

Every option in rclone can have its default set by environment
variable.

To find the name of the environment variable, first, take the long
option name, strip the leading `--`, change `-` to `_`, make
upper case and prepend `RCLONE_`.

For example, to always set `--stats 5s`, set the environment variable
`RCLONE_STATS=5s`.  If you set stats on the command line this will
override the environment variable setting.

Or to always use the trash in drive `--drive-use-trash`, set
`RCLONE_DRIVE_USE_TRASH=true`.

The same parser is used for the options and the environment variables
so they take exactly the same form.

### Config file ###

You can set defaults for values in the config file on an individual
remote basis.  If you want to use this feature, you will need to
discover the name of the config items that you want.  The easiest way
is to run through `rclone config` by hand, then look in the config
file to see what the values are (the config file can be found by
looking at the help for `--config` in `rclone help`).

To find the name of the environment variable, you need to set, take
`RCLONE_CONFIG_` + name of remote + `_` + name of config file option
and make it all uppercase.

For example, to configure an S3 remote named `mys3:` without a config
file (using unix ways of setting environment variables):

```
$ export RCLONE_CONFIG_MYS3_TYPE=s3
$ export RCLONE_CONFIG_MYS3_ACCESS_KEY_ID=XXX
$ export RCLONE_CONFIG_MYS3_SECRET_ACCESS_KEY=XXX
$ rclone lsd MYS3:
          -1 2016-09-21 12:54:21        -1 my-bucket
$ rclone listremotes | grep mys3
mys3:
```

Note that if you want to create a remote using environment variables
you must create the `..._TYPE` variable as above.

### Other environment variables ###

  * RCLONE_CONFIG_PASS` set to contain your config file password (see [Configuration Encryption](#configuration-encryption) section)
  * HTTP_PROXY, HTTPS_PROXY and NO_PROXY (or the lowercase versions thereof).
    * HTTPS_PROXY takes precedence over HTTP_PROXY for https requests.
    * The environment values may be either a complete URL or a "host[:port]" for, in which case the "http" scheme is assumed.

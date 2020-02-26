---
date: 2019-11-19T16:02:36Z
title: "rclone copyto"
slug: rclone_copyto
url: /commands/rclone_copyto/
---
## rclone copyto

Copy files from source to dest, skipping already copied

### Synopsis


If source:path is a file or directory then it copies it to a file or
directory named dest:path.

This can be used to upload single files to other than their current
name.  If the source is a directory then it acts exactly like the copy
command.

So

    rclone copyto src dst

where src and dst are rclone paths, either remote:path or
/path/to/local or C:\windows\path\if\on\windows.

This will:

    if src is file
        copy it to dst, overwriting an existing file if it exists
    if src is directory
        copy it to dst, overwriting existing files if they exist
        see copy command for full details

This doesn't transfer unchanged files, testing by size and
modification time or MD5SUM.  It doesn't delete files from the
destination.

**Note**: Use the `-P`/`--progress` flag to view real-time transfer statistics


```
rclone copyto source:path dest:path [flags]
```

### Options

```
  -h, --help   help for copyto
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


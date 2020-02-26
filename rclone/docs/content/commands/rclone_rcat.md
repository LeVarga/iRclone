---
date: 2019-11-19T16:02:36Z
title: "rclone rcat"
slug: rclone_rcat
url: /commands/rclone_rcat/
---
## rclone rcat

Copies standard input to file on remote.

### Synopsis


rclone rcat reads from standard input (stdin) and copies it to a
single remote file.

    echo "hello world" | rclone rcat remote:path/to/file
    ffmpeg - | rclone rcat remote:path/to/file

If the remote file already exists, it will be overwritten.

rcat will try to upload small files in a single request, which is
usually more efficient than the streaming/chunked upload endpoints,
which use multiple requests. Exact behaviour depends on the remote.
What is considered a small file may be set through
`--streaming-upload-cutoff`. Uploading only starts after
the cutoff is reached or if the file ends before that. The data
must fit into RAM. The cutoff needs to be small enough to adhere
the limits of your remote, please see there. Generally speaking,
setting this cutoff too high will decrease your performance.

Note that the upload can also not be retried because the data is
not kept around until the upload succeeds. If you need to transfer
a lot of data, you're better off caching locally and then
`rclone move` it to the destination.

```
rclone rcat remote:path [flags]
```

### Options

```
  -h, --help   help for rcat
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


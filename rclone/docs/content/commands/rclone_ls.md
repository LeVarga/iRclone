---
date: 2019-11-19T16:02:36Z
title: "rclone ls"
slug: rclone_ls
url: /commands/rclone_ls/
---
## rclone ls

List the objects in the path with size and path.

### Synopsis


Lists the objects in the source path to standard output in a human
readable format with size and path. Recurses by default.

Eg

    $ rclone ls swift:bucket
        60295 bevajer5jef
        90613 canole
        94467 diwogej7
        37600 fubuwic


Any of the filtering options can be applied to this command.

There are several related list commands

  * `ls` to list size and path of objects only
  * `lsl` to list modification time, size and path of objects only
  * `lsd` to list directories only
  * `lsf` to list objects and directories in easy to parse format
  * `lsjson` to list objects and directories in JSON format

`ls`,`lsl`,`lsd` are designed to be human readable.
`lsf` is designed to be human and machine readable.
`lsjson` is designed to be machine readable.

Note that `ls` and `lsl` recurse by default - use "--max-depth 1" to stop the recursion.

The other list commands `lsd`,`lsf`,`lsjson` do not recurse by default - use "-R" to make them recurse.

Listing a non existent directory will produce an error except for
remotes which can't have empty directories (eg s3, swift, gcs, etc -
the bucket based remotes).


```
rclone ls remote:path [flags]
```

### Options

```
  -h, --help   help for ls
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


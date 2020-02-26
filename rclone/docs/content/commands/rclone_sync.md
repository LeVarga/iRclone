---
date: 2019-11-19T16:02:36Z
title: "rclone sync"
slug: rclone_sync
url: /commands/rclone_sync/
---
## rclone sync

Make source and dest identical, modifying destination only.

### Synopsis


Sync the source to the destination, changing the destination
only.  Doesn't transfer unchanged files, testing by size and
modification time or MD5SUM.  Destination is updated to match
source, including deleting files if necessary.

**Important**: Since this can cause data loss, test first with the
`--dry-run` flag to see exactly what would be copied and deleted.

Note that files in the destination won't be deleted if there were any
errors at any point.

It is always the contents of the directory that is synced, not the
directory so when source:path is a directory, it's the contents of
source:path that are copied, not the directory name and contents.  See
extended explanation in the `copy` command above if unsure.

If dest:path doesn't exist, it is created and the source:path contents
go there.

**Note**: Use the `-P`/`--progress` flag to view real-time transfer statistics


```
rclone sync source:path dest:path [flags]
```

### Options

```
      --create-empty-src-dirs   Create empty source dirs on destination after sync
  -h, --help                    help for sync
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


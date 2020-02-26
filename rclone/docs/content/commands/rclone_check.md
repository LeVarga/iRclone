---
date: 2019-11-19T16:02:36Z
title: "rclone check"
slug: rclone_check
url: /commands/rclone_check/
---
## rclone check

Checks the files in the source and destination match.

### Synopsis


Checks the files in the source and destination match.  It compares
sizes and hashes (MD5 or SHA1) and logs a report of files which don't
match.  It doesn't alter the source or destination.

If you supply the --size-only flag, it will only compare the sizes not
the hashes as well.  Use this for a quick check.

If you supply the --download flag, it will download the data from
both remotes and check them against each other on the fly.  This can
be useful for remotes that don't support hashes or if you really want
to check all the data.

If you supply the --one-way flag, it will only check that files in source
match the files in destination, not the other way around. Meaning extra files in
destination that are not in the source will not trigger an error.


```
rclone check source:path dest:path [flags]
```

### Options

```
      --download   Check by downloading rather than with hash.
  -h, --help       help for check
      --one-way    Check one way only, source files must exist on remote
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


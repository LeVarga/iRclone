---
date: 2019-11-19T16:02:36Z
title: "rclone about"
slug: rclone_about
url: /commands/rclone_about/
---
## rclone about

Get quota information from the remote.

### Synopsis


Get quota information from the remote, like bytes used/free/quota and bytes
used in the trash. Not supported by all remotes.

This will print to stdout something like this:

    Total:   17G
    Used:    7.444G
    Free:    1.315G
    Trashed: 100.000M
    Other:   8.241G

Where the fields are:

  * Total: total size available.
  * Used: total size used
  * Free: total amount this user could upload.
  * Trashed: total amount in the trash
  * Other: total amount in other storage (eg Gmail, Google Photos)
  * Objects: total number of objects in the storage

Note that not all the backends provide all the fields - they will be
missing if they are not known for that backend.  Where it is known
that the value is unlimited the value will also be omitted.

Use the --full flag to see the numbers written out in full, eg

    Total:   18253611008
    Used:    7993453766
    Free:    1411001220
    Trashed: 104857602
    Other:   8849156022

Use the --json flag for a computer readable output, eg

    {
        "total": 18253611008,
        "used": 7993453766,
        "trashed": 104857602,
        "other": 8849156022,
        "free": 1411001220
    }


```
rclone about remote: [flags]
```

### Options

```
      --full   Full numbers instead of SI units
  -h, --help   help for about
      --json   Format output as JSON
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


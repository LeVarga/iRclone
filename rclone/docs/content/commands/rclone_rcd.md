---
date: 2019-11-19T16:02:36Z
title: "rclone rcd"
slug: rclone_rcd
url: /commands/rclone_rcd/
---
## rclone rcd

Run rclone listening to remote control commands only.

### Synopsis


This runs rclone so that it only listens to remote control commands.

This is useful if you are controlling rclone via the rc API.

If you pass in a path to a directory, rclone will serve that directory
for GET requests on the URL passed in.  It will also open the URL in
the browser when rclone is run.

See the [rc documentation](/rc/) for more info on the rc flags.


```
rclone rcd <path to files to serve>* [flags]
```

### Options

```
  -h, --help   help for rcd
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


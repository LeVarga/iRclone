---
date: 2019-11-19T16:02:36Z
title: "rclone cryptcheck"
slug: rclone_cryptcheck
url: /commands/rclone_cryptcheck/
---
## rclone cryptcheck

Cryptcheck checks the integrity of a crypted remote.

### Synopsis


rclone cryptcheck checks a remote against a crypted remote.  This is
the equivalent of running rclone check, but able to check the
checksums of the crypted remote.

For it to work the underlying remote of the cryptedremote must support
some kind of checksum.

It works by reading the nonce from each file on the cryptedremote: and
using that to encrypt each file on the remote:.  It then checks the
checksum of the underlying file on the cryptedremote: against the
checksum of the file it has just encrypted.

Use it like this

    rclone cryptcheck /path/to/files encryptedremote:path

You can use it like this also, but that will involve downloading all
the files in remote:path.

    rclone cryptcheck remote:path encryptedremote:path

After it has run it will log the status of the encryptedremote:.

If you supply the --one-way flag, it will only check that files in source
match the files in destination, not the other way around. Meaning extra files in
destination that are not in the source will not trigger an error.


```
rclone cryptcheck remote:path cryptedremote:path [flags]
```

### Options

```
  -h, --help      help for cryptcheck
      --one-way   Check one way only, source files must exist on destination
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


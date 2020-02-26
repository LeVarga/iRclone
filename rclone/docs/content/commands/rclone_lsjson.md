---
date: 2019-11-19T16:02:36Z
title: "rclone lsjson"
slug: rclone_lsjson
url: /commands/rclone_lsjson/
---
## rclone lsjson

List directories and objects in the path in JSON format.

### Synopsis

List directories and objects in the path in JSON format.

The output is an array of Items, where each Item looks like this

   {
      "Hashes" : {
         "SHA-1" : "f572d396fae9206628714fb2ce00f72e94f2258f",
         "MD5" : "b1946ac92492d2347c6235b4d2611184",
         "DropboxHash" : "ecb65bb98f9d905b70458986c39fcbad7715e5f2fcc3b1f07767d7c83e2438cc"
      },
      "ID": "y2djkhiujf83u33",
      "OrigID": "UYOJVTUW00Q1RzTDA",
      "IsBucket" : false,
      "IsDir" : false,
      "MimeType" : "application/octet-stream",
      "ModTime" : "2017-05-31T16:15:57.034468261+01:00",
      "Name" : "file.txt",
      "Encrypted" : "v0qpsdq8anpci8n929v3uu9338",
      "EncryptedPath" : "kja9098349023498/v0qpsdq8anpci8n929v3uu9338",
      "Path" : "full/path/goes/here/file.txt",
      "Size" : 6,
      "Tier" : "hot",
   }

If --hash is not specified the Hashes property won't be emitted.

If --no-modtime is specified then ModTime will be blank.

If --encrypted is not specified the Encrypted won't be emitted.

If --dirs-only is not specified files in addition to directories are returned

If --files-only is not specified directories in addition to the files will be returned.

The Path field will only show folders below the remote path being listed.
If "remote:path" contains the file "subfolder/file.txt", the Path for "file.txt"
will be "subfolder/file.txt", not "remote:path/subfolder/file.txt".
When used without --recursive the Path will always be the same as Name.

If the directory is a bucket in a bucket based backend, then
"IsBucket" will be set to true. This key won't be present unless it is
"true".

The time is in RFC3339 format with up to nanosecond precision.  The
number of decimal digits in the seconds will depend on the precision
that the remote can hold the times, so if times are accurate to the
nearest millisecond (eg Google Drive) then 3 digits will always be
shown ("2017-05-31T16:15:57.034+01:00") whereas if the times are
accurate to the nearest second (Dropbox, Box, WebDav etc) no digits
will be shown ("2017-05-31T16:15:57+01:00").

The whole output can be processed as a JSON blob, or alternatively it
can be processed line by line as each item is written one to a line.

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
rclone lsjson remote:path [flags]
```

### Options

```
      --dirs-only    Show only directories in the listing.
  -M, --encrypted    Show the encrypted names.
      --files-only   Show only files in the listing.
      --hash         Include hashes in the output (may take longer).
  -h, --help         help for lsjson
      --no-modtime   Don't read the modification time (can speed things up).
      --original     Show the ID of the underlying Object.
  -R, --recursive    Recurse into the listing.
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone](/commands/rclone/)	 - Show help for rclone commands, flags and backends.


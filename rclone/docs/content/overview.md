---
title: "Overview of cloud storage systems"
description: "Overview of cloud storage systems"
type: page
date: "2019-02-25"
---

# Overview of cloud storage systems #

Each cloud storage system is slightly different.  Rclone attempts to
provide a unified interface to them, but some underlying differences
show through.

## Features ##

Here is an overview of the major features of each cloud storage system.

| Name                         | Hash        | ModTime | Case Insensitive | Duplicate Files | MIME Type |
| ---------------------------- |:-----------:|:-------:|:----------------:|:---------------:|:---------:|
| 1Fichier                     | Whirlpool   | No      | No               | Yes             | R         |
| Amazon Drive                 | MD5         | No      | Yes              | No              | R         |
| Amazon S3                    | MD5         | Yes     | No               | No              | R/W       |
| Backblaze B2                 | SHA1        | Yes     | No               | No              | R/W       |
| Box                          | SHA1        | Yes     | Yes              | No              | -         |
| Citrix ShareFile             | MD5         | Yes     | Yes              | No              | -         |
| Dropbox                      | DBHASH †    | Yes     | Yes              | No              | -         |
| FTP                          | -           | No      | No               | No              | -         |
| Google Cloud Storage         | MD5         | Yes     | No               | No              | R/W       |
| Google Drive                 | MD5         | Yes     | No               | Yes             | R/W       |
| Google Photos                | -           | No      | No               | Yes             | R         |
| HTTP                         | -           | No      | No               | No              | R         |
| Hubic                        | MD5         | Yes     | No               | No              | R/W       |
| Jottacloud                   | MD5         | Yes     | Yes              | No              | R/W       |
| Koofr                        | MD5         | No      | Yes              | No              | -         |
| Mail.ru Cloud                | Mailru ‡‡‡  | Yes     | Yes              | No              | -         |
| Mega                         | -           | No      | No               | Yes             | -         |
| Microsoft Azure Blob Storage | MD5         | Yes     | No               | No              | R/W       |
| Microsoft OneDrive           | SHA1 ‡‡     | Yes     | Yes              | No              | R         |
| OpenDrive                    | MD5         | Yes     | Yes              | No              | -         |
| Openstack Swift              | MD5         | Yes     | No               | No              | R/W       |
| pCloud                       | MD5, SHA1   | Yes     | No               | No              | W         |
| premiumize.me                | -           | No      | Yes              | No              | R         |
| put.io                       | CRC-32      | Yes     | No               | Yes             | R         |
| QingStor                     | MD5         | No      | No               | No              | R/W       |
| SFTP                         | MD5, SHA1 ‡ | Yes     | Depends          | No              | -         |
| WebDAV                       | MD5, SHA1 ††| Yes ††† | Depends          | No              | -         |
| Yandex Disk                  | MD5         | Yes     | No               | No              | R/W       |
| The local filesystem         | All         | Yes     | Depends          | No              | -         |

### Hash ###

The cloud storage system supports various hash types of the objects.
The hashes are used when transferring data as an integrity check and
can be specifically used with the `--checksum` flag in syncs and in
the `check` command.

To use the verify checksums when transferring between cloud storage
systems they must support a common hash type.

† Note that Dropbox supports [its own custom
hash](https://www.dropbox.com/developers/reference/content-hash).
This is an SHA256 sum of all the 4MB block SHA256s.

‡ SFTP supports checksums if the same login has shell access and `md5sum`
or `sha1sum` as well as `echo` are in the remote's PATH.

†† WebDAV supports hashes when used with Owncloud and Nextcloud only.

††† WebDAV supports modtimes when used with Owncloud and Nextcloud only.

‡‡ Microsoft OneDrive Personal supports SHA1 hashes, whereas OneDrive
for business and SharePoint server support Microsoft's own
[QuickXorHash](https://docs.microsoft.com/en-us/onedrive/developer/code-snippets/quickxorhash).

‡‡‡ Mail.ru uses its own modified SHA1 hash

### ModTime ###

The cloud storage system supports setting modification times on
objects.  If it does then this enables a using the modification times
as part of the sync.  If not then only the size will be checked by
default, though the MD5SUM can be checked with the `--checksum` flag.

All cloud storage systems support some kind of date on the object and
these will be set when transferring from the cloud storage system.

### Case Insensitive ###

If a cloud storage systems is case sensitive then it is possible to
have two files which differ only in case, eg `file.txt` and
`FILE.txt`.  If a cloud storage system is case insensitive then that
isn't possible.

This can cause problems when syncing between a case insensitive
system and a case sensitive system.  The symptom of this is that no
matter how many times you run the sync it never completes fully.

The local filesystem and SFTP may or may not be case sensitive
depending on OS.

  * Windows - usually case insensitive, though case is preserved
  * OSX - usually case insensitive, though it is possible to format case sensitive
  * Linux - usually case sensitive, but there are case insensitive file systems (eg FAT formatted USB keys)

Most of the time this doesn't cause any problems as people tend to
avoid files whose name differs only by case even on case sensitive
systems.

### Duplicate files ###

If a cloud storage system allows duplicate files then it can have two
objects with the same name.

This confuses rclone greatly when syncing - use the `rclone dedupe`
command to rename or remove duplicates.

### Restricted filenames ###

Some cloud storage systems might have restrictions on the characters
that are usable in file or directory names.
When `rclone` detects such a name during a file upload, it will
transparently replace the restricted characters with similar looking
Unicode characters.

This process is designed to avoid ambiguous file names as much as
possible and allow to move files between many cloud storage systems
transparently.

The name shown by `rclone` to the user or during log output will only
contain a minimal set of [replaced characters](#restricted-characters)
to ensure correct formatting and not necessarily the actual name used
on the cloud storage.

This transformation is reversed when downloading a file or parsing
`rclone` arguments.
For example, when uploading a file named `my file?.txt` to Onedrive
will be displayed as `my file?.txt` on the console, but stored as
`my file？.txt` (the `?` gets replaced by the similar looking `？`
character) to Onedrive.
The reverse transformation allows to read a file`unusual/name.txt`
from Google Drive, by passing the name `unusual／name.txt` (the `/` needs
to be replaced by the similar looking `／` character) on the command line.

#### Default restricted characters {#restricted-characters}

The table below shows the characters that are replaced by default.

When a replacement character is found in a filename, this character
will be escaped with the `‛` character to avoid ambiguous file names.
(e.g. a file named `␀.txt` would shown as `‛␀.txt`)

Each cloud storage backend can use a different set of characters,
which will be specified in the documentation for each backend.

| Character | Value | Replacement |
| --------- |:-----:|:-----------:|
| NUL       | 0x00  | ␀           |
| SOH       | 0x01  | ␁           |
| STX       | 0x02  | ␂           |
| ETX       | 0x03  | ␃           |
| EOT       | 0x04  | ␄           |
| ENQ       | 0x05  | ␅           |
| ACK       | 0x06  | ␆           |
| BEL       | 0x07  | ␇           |
| BS        | 0x08  | ␈           |
| HT        | 0x09  | ␉           |
| LF        | 0x0A  | ␊           |
| VT        | 0x0B  | ␋           |
| FF        | 0x0C  | ␌           |
| CR        | 0x0D  | ␍           |
| SO        | 0x0E  | ␎           |
| SI        | 0x0F  | ␏           |
| DLE       | 0x10  | ␐           |
| DC1       | 0x11  | ␑           |
| DC2       | 0x12  | ␒           |
| DC3       | 0x13  | ␓           |
| DC4       | 0x14  | ␔           |
| NAK       | 0x15  | ␕           |
| SYN       | 0x16  | ␖           |
| ETB       | 0x17  | ␗           |
| CAN       | 0x18  | ␘           |
| EM        | 0x19  | ␙           |
| SUB       | 0x1A  | ␚           |
| ESC       | 0x1B  | ␛           |
| FS        | 0x1C  | ␜           |
| GS        | 0x1D  | ␝           |
| RS        | 0x1E  | ␞           |
| US        | 0x1F  | ␟           |
| /         | 0x2F  | ／           |
| DEL       | 0x7F  | ␡           |

The default encoding will also encode these file names as they are
problematic with many cloud storage systems.

| File name | Replacement |
| --------- |:-----------:|
| .         | ．          |
| ..        | ．．         |

#### Invalid UTF-8 bytes {#invalid-utf8}

Some backends only support a sequence of well formed UTF-8 bytes
as file or directory names.

In this case all invalid UTF-8 bytes will be replaced with a quoted
representation of the byte value to allow uploading a file to such a
backend. For example, the invalid byte `0xFE` will be encoded as `‛FE`.

A common source of invalid UTF-8 bytes are local filesystems, that store
names in a different encoding than UTF-8 or UTF-16, like latin1. See the
[local filenames](/local/#filenames) section for details.

### MIME Type ###

MIME types (also known as media types) classify types of documents
using a simple text classification, eg `text/html` or
`application/pdf`.

Some cloud storage systems support reading (`R`) the MIME type of
objects and some support writing (`W`) the MIME type of objects.

The MIME type can be important if you are serving files directly to
HTTP from the storage system.

If you are copying from a remote which supports reading (`R`) to a
remote which supports writing (`W`) then rclone will preserve the MIME
types.  Otherwise they will be guessed from the extension, or the
remote itself may assign the MIME type.

## Optional Features ##

All the remotes support a basic set of features, but there are some
optional features supported by some remotes used to make some
operations more efficient.

| Name                         | Purge | Copy | Move | DirMove | CleanUp | ListR | StreamUpload | LinkSharing | About | EmptyDir |
| ---------------------------- |:-----:|:----:|:----:|:-------:|:-------:|:-----:|:------------:|:------------:|:-----:| :------: |
| 1Fichier                     | No    | No   | No   | No      | No      | No    | No           | No           |   No  |  Yes |
| Amazon Drive                 | Yes   | No   | Yes  | Yes     | No [#575](https://github.com/rclone/rclone/issues/575) | No  | No  | No [#2178](https://github.com/rclone/rclone/issues/2178) | No  | Yes |
| Amazon S3                    | No    | Yes  | No   | No      | No      | Yes   | Yes          | No [#2178](https://github.com/rclone/rclone/issues/2178) | No  | No |
| Backblaze B2                 | No    | Yes  | No   | No      | Yes     | Yes   | Yes          | Yes | No  | No |
| Box                          | Yes   | Yes  | Yes  | Yes     | No [#575](https://github.com/rclone/rclone/issues/575) | No  | Yes | Yes | No  | Yes |
| Citrix ShareFile             | Yes   | Yes  | Yes  | Yes     | No      | No    | Yes          | No          | No  | Yes |
| Dropbox                      | Yes   | Yes  | Yes  | Yes     | No [#575](https://github.com/rclone/rclone/issues/575) | No  | Yes | Yes | Yes | Yes |
| FTP                          | No    | No   | Yes  | Yes     | No      | No    | Yes          | No [#2178](https://github.com/rclone/rclone/issues/2178) | No  | Yes |
| Google Cloud Storage         | Yes   | Yes  | No   | No      | No      | Yes   | Yes          | No [#2178](https://github.com/rclone/rclone/issues/2178) | No  | No |
| Google Drive                 | Yes   | Yes  | Yes  | Yes     | Yes     | Yes   | Yes          | Yes         | Yes | Yes |
| Google Photos                | No    | No   | No   | No      | No      | No    | No           | No          | No | No |
| HTTP                         | No    | No   | No   | No      | No      | No    | No           | No [#2178](https://github.com/rclone/rclone/issues/2178) | No  | Yes |
| Hubic                        | Yes † | Yes  | No   | No      | No      | Yes   | Yes          | No [#2178](https://github.com/rclone/rclone/issues/2178) | Yes | No |
| Jottacloud                   | Yes   | Yes  | Yes  | Yes     | No      | Yes   | No           | Yes                                                   | Yes | Yes |
| Mail.ru Cloud                | Yes   | Yes  | Yes  | Yes     | Yes     | No    | No           | Yes                                                   | Yes | Yes |
| Mega                         | Yes   | No   | Yes  | Yes     | Yes     | No    | No           | No [#2178](https://github.com/rclone/rclone/issues/2178) | Yes | Yes |
| Microsoft Azure Blob Storage | Yes   | Yes  | No   | No      | No      | Yes   | No           | No [#2178](https://github.com/rclone/rclone/issues/2178) | No  | No |
| Microsoft OneDrive           | Yes   | Yes  | Yes  | Yes     | No [#575](https://github.com/rclone/rclone/issues/575) | No | No | Yes | Yes | Yes |
| OpenDrive                    | Yes   | Yes  | Yes  | Yes     | No      | No    | No           | No                                                    | No  | Yes |
| Openstack Swift              | Yes † | Yes  | No   | No      | No      | Yes   | Yes          | No [#2178](https://github.com/rclone/rclone/issues/2178) | Yes | No |
| pCloud                       | Yes   | Yes  | Yes  | Yes     | Yes     | No    | No           | No [#2178](https://github.com/rclone/rclone/issues/2178) | Yes | Yes |
| premiumize.me                | Yes   | No   | Yes  | Yes     | No      | No    | No           | Yes         | Yes | Yes |
| put.io                       | Yes   | No   | Yes  | Yes     | Yes     | No    | Yes          | No [#2178](https://github.com/rclone/rclone/issues/2178) | Yes | Yes |
| QingStor                     | No    | Yes  | No   | No      | No      | Yes   | No           | No [#2178](https://github.com/rclone/rclone/issues/2178) | No  | No |
| SFTP                         | No    | No   | Yes  | Yes     | No      | No    | Yes          | No [#2178](https://github.com/rclone/rclone/issues/2178) | Yes  | Yes |
| WebDAV                       | Yes   | Yes  | Yes  | Yes     | No      | No    | Yes ‡        | No [#2178](https://github.com/rclone/rclone/issues/2178) | Yes  | Yes |
| Yandex Disk                  | Yes   | Yes  | Yes  | Yes     | Yes     | No    | Yes          | Yes         | Yes | Yes |
| The local filesystem         | Yes   | No   | Yes  | Yes     | No      | No    | Yes          | No          | Yes | Yes |

### Purge ###

This deletes a directory quicker than just deleting all the files in
the directory.

† Note Swift and Hubic implement this in order to delete directory
markers but they don't actually have a quicker way of deleting files
other than deleting them individually.

‡ StreamUpload is not supported with Nextcloud

### Copy ###

Used when copying an object to and from the same remote.  This known
as a server side copy so you can copy a file without downloading it
and uploading it again.  It is used if you use `rclone copy` or
`rclone move` if the remote doesn't support `Move` directly.

If the server doesn't support `Copy` directly then for copy operations
the file is downloaded then re-uploaded.

### Move ###

Used when moving/renaming an object on the same remote.  This is known
as a server side move of a file.  This is used in `rclone move` if the
server doesn't support `DirMove`.

If the server isn't capable of `Move` then rclone simulates it with
`Copy` then delete.  If the server doesn't support `Copy` then rclone
will download the file and re-upload it.

### DirMove ###

This is used to implement `rclone move` to move a directory if
possible.  If it isn't then it will use `Move` on each file (which
falls back to `Copy` then download and upload - see `Move` section).

### CleanUp ###

This is used for emptying the trash for a remote by `rclone cleanup`.

If the server can't do `CleanUp` then `rclone cleanup` will return an
error.

### ListR ###

The remote supports a recursive list to list all the contents beneath
a directory quickly.  This enables the `--fast-list` flag to work.
See the [rclone docs](/docs/#fast-list) for more details.

### StreamUpload ###

Some remotes allow files to be uploaded without knowing the file size
in advance. This allows certain operations to work without spooling the
file to local disk first, e.g. `rclone rcat`.

### LinkSharing ###

Sets the necessary permissions on a file or folder and prints a link
that allows others to access them, even if they don't have an account
on the particular cloud provider.

### About ###

This is used to fetch quota information from the remote, like bytes
used/free/quota and bytes used in the trash.

This is also used to return the space used, available for `rclone mount`.

If the server can't do `About` then `rclone about` will return an
error.

### EmptyDir ###

The remote supports empty directories. See [Limitations](/bugs/#limitations)
 for details. Most Object/Bucket based remotes do not support this.

---
title: "rclone completion bash"
description: "Output bash completion script for rclone."
slug: rclone_completion_bash
url: /commands/rclone_completion_bash/
# autogenerated - DO NOT EDIT, instead edit the source code in cmd/completion/bash/ and as part of making a release run "make commanddocs"
---
# rclone completion bash

Output bash completion script for rclone.

## Synopsis


Generates a bash shell autocompletion script for rclone.

This writes to /etc/bash_completion.d/rclone by default so will
probably need to be run with sudo or as root, e.g.

    sudo rclone genautocomplete bash

Logout and login again to use the autocompletion scripts, or source
them directly

    . /etc/bash_completion

If you supply a command line argument the script will be written
there.

If output_file is "-", then the output will be written to stdout.


```
rclone completion bash [output_file] [flags]
```

## Options

```
  -h, --help   help for bash
```


See the [global flags page](/flags/) for global options not listed here.

# SEE ALSO

* [rclone completion](/commands/rclone_completion/)	 - Output completion script for a given shell.

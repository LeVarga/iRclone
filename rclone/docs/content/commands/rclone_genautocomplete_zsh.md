---
date: 2019-11-19T16:02:36Z
title: "rclone genautocomplete zsh"
slug: rclone_genautocomplete_zsh
url: /commands/rclone_genautocomplete_zsh/
---
## rclone genautocomplete zsh

Output zsh completion script for rclone.

### Synopsis


Generates a zsh autocompletion script for rclone.

This writes to /usr/share/zsh/vendor-completions/_rclone by default so will
probably need to be run with sudo or as root, eg

    sudo rclone genautocomplete zsh

Logout and login again to use the autocompletion scripts, or source
them directly

    autoload -U compinit && compinit

If you supply a command line argument the script will be written
there.


```
rclone genautocomplete zsh [output_file] [flags]
```

### Options

```
  -h, --help   help for zsh
```

See the [global flags page](/flags/) for global options not listed here.

### SEE ALSO

* [rclone genautocomplete](/commands/rclone_genautocomplete/)	 - Output completion script for a given shell.


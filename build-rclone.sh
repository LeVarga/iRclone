#!/bin/sh
set -x
GOPATH="${GOPATH:-"$HOME/go"}"
GOBINPATH="$GOPATH/bin"
case ":$PATH:" in
  *:GOBINPATH:*);;
  *) export PATH="$PATH:$GOBINPATH" ;;
esac

sim_arch="arm64"
case $(uname -m) in
    x86_64) sim_arch="amd64" ;;
esac

if [ -d "$GOPATH" ]; then
    if [ ! -f "$GOBINPATH/gomobile" ] ; then
        go install golang.org/x/mobile/cmd/gomobile@latest
        gomobile init
    fi
    gomobile bind -target=ios/arm64,iossimulator/$sim_arch -o ./rclone.xcframework
else
    echo "Could not find go path. Make sure go is installed and you are NOT running as root."
fi

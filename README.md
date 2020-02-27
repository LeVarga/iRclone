# iRclone
Rclone for iOS devices

Work in progress...

iOS 13+ Required

## Features
- Configure remotes (create / delete)
- Browse remote directories
- Copy / move files and directories (including to and from local fs)
- Stream video content from remote
- View supported photos, documents and audio files (downloads to tmp directory)
- Files app support for local files
## Build
### Requirements:
- cocoapods
- go
- gomobile
- Xcode 11
### Steps:
- git clone https://github.com/lvarga/iRclone
- cd iRclone
- mkdir $GOPATH/src/github.com/rclone/
- ln -s rclone $GOPATH/src/github.com/rclone/
- cd rclone
- gomobile bind -target=ios/amd64,ios/arm64 -o ../rclone.framework
- cd ..
- pod install
- open iRclone.xcworkspace
- configure code signing, build

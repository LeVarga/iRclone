# iRclone
Rclone for iOS devices

Work in progress...

## Working features
- Configure remotes (create / delete)
- Browse remote directories
- Copy / move files and directories (including to and from local fs)
- Stream video content from remote
- View supported photos, documents and audio files (downloads to tmp directory)
- Files app support for local files and rclone.conf
## Build
### Requirements:
- cocoapods (brew install cocoapods)
- go (brew install go)
- Xcode (with command-line tools installed and configured)
### Steps:
- git clone https://github.com/levarga/iRclone && cd iRclone
- go get
- ./build-rclone.sh
- pod install
- open iRclone.xcworkspace
- configure code signing, build

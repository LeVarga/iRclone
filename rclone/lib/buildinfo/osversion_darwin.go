//go:build darwin
// +build darwin

package buildinfo

import (
	"github.com/matishsiao/goInfo"
)

// GetOSVersion returns OS version, kernel and bitness
func GetOSVersion() (osVersion, osKernel string) {
	if info, err := goInfo.GetInfo(); err == nil {
		osVersion = " " + info.OS + " " + info.Core
		osKernel = info.Kernel + " " + info.Platform
	}
	return
}

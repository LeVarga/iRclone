//go:build !windows && !darwin
// +build !windows,!darwin

package buildinfo

import (
	"github.com/shirou/gopsutil/v3/host"
	"strings"
)

// GetOSVersion returns OS version, kernel and bitness
func GetOSVersion() (osVersion, osKernel string) {
	if platform, _, version, err := host.PlatformInformation(); err == nil && platform != "" {
		osVersion = platform
		if version != "" {
			osVersion += " " + version
		}
	}

	if version, err := host.KernelVersion(); err == nil && version != "" {
		osKernel = version
	}

	if arch, err := host.KernelArch(); err == nil && arch != "" {
		if strings.HasSuffix(arch, "64") && osVersion != "" {
			osVersion += " (64 bit)"
		}
		if osKernel != "" {
			osKernel += " (" + arch + ")"
		}
	}

	return
}

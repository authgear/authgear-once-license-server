package uname

import (
	"strings"
)

const (
	KernelDarwin = "darwin"
	KernelLinux  = "linux"
)

const (
	ArchArm64 = "arm64"
	ArchAmd64 = "amd64"
)

// NormalizeUnameS normalizes the output of `uname -s`.
// See https://pubs.opengroup.org/onlinepubs/9799919799/utilities/uname.html
func NormalizeUnameS(uname_s string) string {
	uname_s = strings.TrimSpace(uname_s)
	uname_s = strings.ToLower(uname_s)

	switch {
	case strings.Contains(uname_s, "darwin"):
		return KernelDarwin
	case strings.Contains(uname_s, "linux"):
		return KernelLinux
	default:
		return KernelLinux
	}
}

// NormalizeUnameM normalizes the output of `uname -m`.
// See https://pubs.opengroup.org/onlinepubs/9799919799/utilities/uname.html
func NormalizeUnameM(uname_m string) string {
	uname_m = strings.TrimSpace(uname_m)
	uname_m = strings.ToLower(uname_m)

	switch {
	case strings.Contains(uname_m, "arm64") || strings.Contains(uname_m, "aarch64") || strings.Contains(uname_m, "arm"):
		return ArchArm64
	case strings.Contains(uname_m, "amd64") || strings.Contains(uname_m, "x86_64") || strings.Contains(uname_m, "x64"):
		return ArchAmd64
	default:
		return ArchAmd64
	}
}

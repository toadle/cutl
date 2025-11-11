package version

import (
	"fmt"
	"runtime/debug"
)

// Build information. These variables are set via ldflags during build.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// GetVersion returns the version string with build information
func GetVersion() string {
	if Version == "dev" {
		// In development, try to get version from debug info
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" && len(setting.Value) >= 7 {
					return fmt.Sprintf("dev-%s", setting.Value[:7])
				}
			}
		}
		return "dev"
	}
	return Version
}

// GetFullVersion returns detailed version information
func GetFullVersion() string {
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" && len(setting.Value) >= 7 {
					return fmt.Sprintf("dev-%s (development build)", setting.Value[:7])
				}
			}
		}
		return "dev (development build)"
	}
	
	result := fmt.Sprintf("v%s", Version)
	if Commit != "unknown" && len(Commit) >= 7 {
		result += fmt.Sprintf(" (%s)", Commit[:7])
	}
	if Date != "unknown" {
		result += fmt.Sprintf(" built on %s", Date)
	}
	return result
}
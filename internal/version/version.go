package version

// version is set at build time via:
//
//	-ldflags "-X codeberg.org/dbus/shushingface/internal/version.version=v0.1.0"
//
// In dev builds, it defaults to "dev".
var version = "dev"

// Version returns the build version string.
func Version() string { return version }

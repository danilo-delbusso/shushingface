package platform

// DisplayServer identifies the graphics stack.
type DisplayServer string

const (
	Wayland DisplayServer = "wayland"
	X11     DisplayServer = "x11"
	Native  DisplayServer = "native" // macOS, Windows
	Unknown DisplayServer = "unknown"
)

// PackageManager identifies the system package manager.
type PackageManager string

const (
	Apt       PackageManager = "apt"
	Dnf       PackageManager = "dnf"
	Pacman    PackageManager = "pacman"
	Zypper    PackageManager = "zypper"
	Apk       PackageManager = "apk"
	Nix       PackageManager = "nix"
	UnknownPM PackageManager = "unknown"
)

// Info holds all detected runtime platform details.
type Info struct {
	OS             string         `json:"os"`
	DisplayServer  DisplayServer  `json:"displayServer"`
	Desktop        string         `json:"desktop"`
	PackageManager PackageManager `json:"packageManager"`
}

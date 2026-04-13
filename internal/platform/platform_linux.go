//go:build linux

package platform

import (
	"os"
	"os/exec"
	"sync"
)

var (
	info     Info
	infoOnce sync.Once
)

// Detect returns the runtime platform info (cached after first call).
func Detect() Info {
	infoOnce.Do(func() {
		info = Info{
			OS:             "linux",
			DisplayServer:  detectDisplayServer(),
			Desktop:        os.Getenv("XDG_CURRENT_DESKTOP"),
			PackageManager: detectPackageManager(),
		}
	})
	return info
}

func detectDisplayServer() DisplayServer {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return Wayland
	}
	if os.Getenv("DISPLAY") != "" {
		return X11
	}
	return Unknown
}

func detectPackageManager() PackageManager {
	managers := []struct {
		cmd string
		pm  PackageManager
	}{
		{"apt", Apt},
		{"dnf", Dnf},
		{"pacman", Pacman},
		{"zypper", Zypper},
		{"apk", Apk},
		{"nix-env", Nix},
	}
	for _, m := range managers {
		if _, err := exec.LookPath(m.cmd); err == nil {
			return m.pm
		}
	}
	return UnknownPM
}

// InstallCmd returns the command to install a package on the detected distro.
func InstallCmd(pkg string) string {
	pm := Detect().PackageManager
	switch pm {
	case Apt:
		return "sudo apt install " + pkg
	case Dnf:
		return "sudo dnf install " + pkg
	case Pacman:
		return "sudo pacman -S " + pkg
	case Zypper:
		return "sudo zypper install " + pkg
	case Apk:
		return "sudo apk add " + pkg
	case Nix:
		return "nix-env -iA nixpkgs." + pkg
	default:
		return "install " + pkg + " using your package manager"
	}
}

// HasCommand checks if a CLI tool is available on PATH.
func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

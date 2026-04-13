package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// appName is the per-platform directory name we live under inside the
// OS-standard config / state / cache roots.
const appName = "shushingface"

// Config returns the directory where user-modifiable settings live and
// ensures it exists.
//
// Linux:   $XDG_CONFIG_HOME/shushingface, fallback ~/.config/shushingface
// Windows: %APPDATA%\shushingface (roaming — survives roaming-profile sync)
//
// Use this for files the user might want to back up, sync between machines,
// or hand-edit. NOT for logs, caches, or generated databases.
func Config() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, appName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// State returns the directory where logs and persistent generated data
// (history database, etc.) live and ensures it exists.
//
// Linux:   $XDG_STATE_HOME/shushingface, fallback ~/.local/state/shushingface
// Windows: %LOCALAPPDATA%\shushingface (machine-local — does not roam)
//
// State is the right home for things the app produces during use: logs,
// SQLite databases, on-disk caches that should survive restarts but
// don't need to follow the user across machines.
func State() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	full := filepath.Join(dir, appName)
	if err := os.MkdirAll(full, 0o700); err != nil {
		return "", err
	}
	return full, nil
}

func stateDir() (string, error) {
	if runtime.GOOS == "windows" {
		// %LOCALAPPDATA% — set on every supported Windows version.
		if v := os.Getenv("LOCALAPPDATA"); v != "" {
			return v, nil
		}
		// Fallback if LOCALAPPDATA is unset for some reason: derive from home.
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "AppData", "Local"), nil
	}

	// Linux / other unix: XDG_STATE_HOME with ~/.local/state fallback.
	if v := os.Getenv("XDG_STATE_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state"), nil
}

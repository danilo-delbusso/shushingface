//go:build linux

package indicator

import (
	_ "embed"
	"log/slog"
	"sync"

	"codeberg.org/dbus/shushingface/internal/platform"
	"codeberg.org/dbus/shushingface/internal/sni"
)

func capability() platform.Capability { return platform.Supported() }

//go:embed icons/idle.png
var idleIcon []byte

//go:embed icons/recording.png
var recordingIcon []byte

var (
	instance *sni.Item
	once     sync.Once
)

// Start registers a StatusNotifierItem with the panel. onActivate is
// called when the user clicks the icon. All D-Bus + SNI plumbing lives
// in internal/sni.
func Start(onActivate func()) {
	once.Do(func() {
		item, err := sni.New(idleIcon, recordingIcon, onActivate)
		if err != nil {
			slog.Debug("indicator: panel registration unavailable", "error", err)
			return
		}
		instance = item
		slog.Info("panel indicator registered")
	})
}

// SetRecording flips the displayed icon between idle and recording.
func SetRecording(recording bool) {
	if instance == nil {
		return
	}
	instance.SetRecording(recording)
}

// Stop releases the bus name; the panel detects it vanishing and pulls
// the icon. Resets the once-guard so a future Start can re-register
// (used when the user toggles the indicator setting back on).
func Stop() {
	if instance == nil {
		return
	}
	if err := instance.Close(); err != nil {
		slog.Warn("indicator: close failed", "error", err)
	}
	instance = nil
	once = sync.Once{}
}

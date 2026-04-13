//go:build windows

package notify

import (
	"log/slog"

	toast "git.sr.ht/~jackmordaunt/go-toast/v2"

	"codeberg.org/dbus/shushingface/internal/platform"
)

const appID = "shushingface"

func capability() platform.Capability { return platform.Supported() }

// RecordingStarted shows a toast indicating recording is active.
func RecordingStarted() { push("Recording...", "") }

// RecordingProcessing updates the toast to show processing state.
func RecordingProcessing() { push("Processing with AI...", "") }

// Error shows an error toast.
func Error(title, body string) { push(title, body) }

// RecordingDone is a no-op on Windows — toasts auto-dismiss and there is no
// cheap way to retract a specific one without the full Action Center API.
func RecordingDone() {}

func push(title, body string) {
	n := toast.Notification{AppID: appID, Title: title, Body: body}
	if err := n.Push(); err != nil {
		slog.Debug("toast notification failed", "title", title, "error", err)
	}
}

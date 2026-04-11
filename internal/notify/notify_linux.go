//go:build linux

package notify

import (
	"log/slog"

	"github.com/godbus/dbus/v5"
)

const (
	dest  = "org.freedesktop.Notifications"
	npath = "/org/freedesktop/Notifications"
	iface = "org.freedesktop.Notifications"
)

var lastID uint32

// RecordingStarted shows a persistent notification indicating recording is active.
func RecordingStarted() {
	send("sussurro", "Recording...", "audio-input-microphone", 0)
}

// RecordingProcessing updates the notification to show processing state.
func RecordingProcessing() {
	send("sussurro", "Processing with AI...", "audio-input-microphone", 0)
}

// RecordingDone dismisses the recording notification.
func RecordingDone() {
	if lastID == 0 {
		return
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		return
	}
	conn.Object(dest, dbus.ObjectPath(npath)).Call(iface+".CloseNotification", 0, lastID)
	lastID = 0
}

func send(summary, body, icon string, timeout int32) {
	conn, err := dbus.SessionBus()
	if err != nil {
		slog.Debug("dbus not available for notification", "error", err)
		return
	}

	call := conn.Object(dest, dbus.ObjectPath(npath)).Call(
		iface+".Notify",
		0,
		"sussurro",
		lastID,
		icon,
		summary,
		body,
		[]string{},
		map[string]dbus.Variant{
			"urgency":  dbus.MakeVariant(byte(0)),
			"resident": dbus.MakeVariant(true),
		},
		timeout,
	)
	if call.Err != nil {
		slog.Debug("failed to send notification", "error", call.Err)
		return
	}
	call.Store(&lastID)
}

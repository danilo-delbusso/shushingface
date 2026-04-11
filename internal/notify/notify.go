package notify

import (
	"log/slog"

	"github.com/godbus/dbus/v5"
)

const (
	dest = "org.freedesktop.Notifications"
	path = "/org/freedesktop/Notifications"
	iface = "org.freedesktop.Notifications"
)

var lastID uint32

// RecordingStarted shows a persistent notification indicating recording is active.
func RecordingStarted() {
	send("Sussurro", "Recording...", "audio-input-microphone", 0)
}

// RecordingProcessing updates the notification to show processing state.
func RecordingProcessing() {
	send("Sussurro", "Processing with AI...", "audio-input-microphone", 0)
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
	conn.Object(dest, dbus.ObjectPath(path)).Call(iface+".CloseNotification", 0, lastID)
	lastID = 0
}

func send(summary, body, icon string, timeout int32) {
	conn, err := dbus.SessionBus()
	if err != nil {
		slog.Debug("dbus not available for notification", "error", err)
		return
	}

	call := conn.Object(dest, dbus.ObjectPath(path)).Call(
		iface+".Notify",
		0,
		"Sussurro",     // app_name
		lastID,         // replaces_id (0 = new, >0 = replace)
		icon,           // app_icon
		summary,        // summary
		body,           // body
		[]string{},     // actions
		map[string]dbus.Variant{
			"urgency":  dbus.MakeVariant(byte(0)), // low urgency
			"resident": dbus.MakeVariant(true),     // keep after action
		},
		timeout, // expire_timeout (0 = persistent)
	)
	if call.Err != nil {
		slog.Debug("failed to send notification", "error", call.Err)
		return
	}
	call.Store(&lastID)
}

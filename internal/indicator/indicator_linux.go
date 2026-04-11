//go:build linux

package indicator

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	sniIface     = "org.kde.StatusNotifierItem"
	watcherDest  = "org.kde.StatusNotifierWatcher"
	watcherPath  = "/StatusNotifierWatcher"
	watcherIface = "org.kde.StatusNotifierWatcher"
	itemPath     = "/StatusNotifierItem"
)

// sniItem implements the StatusNotifierItem D-Bus interface.
type sniItem struct {
	mu        sync.RWMutex
	conn      *dbus.Conn
	busName   string
	recording bool
}

// D-Bus property getters (called by the panel)
func (s *sniItem) Get(iface, prop string) (dbus.Variant, *dbus.Error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if iface != sniIface {
		return dbus.Variant{}, nil
	}
	switch prop {
	case "Category":
		return dbus.MakeVariant("ApplicationStatus"), nil
	case "Id":
		return dbus.MakeVariant("sussurro"), nil
	case "Title":
		return dbus.MakeVariant("Sussurro"), nil
	case "Status":
		if s.recording {
			return dbus.MakeVariant("NeedsAttention"), nil
		}
		return dbus.MakeVariant("Active"), nil
	case "IconName":
		return dbus.MakeVariant("audio-input-microphone"), nil
	case "AttentionIconName":
		return dbus.MakeVariant("media-record"), nil
	case "ToolTip":
		title := "Sussurro — Ready"
		if s.recording {
			title = "Sussurro — Recording"
		}
		// ToolTip is (sa(iiay)ss) — icon_name, icon_data, title, description
		return dbus.MakeVariant(struct {
			IconName string
			IconData []struct {
				Width, Height int32
				Data          []byte
			}
			Title string
			Desc  string
		}{
			IconName: "audio-input-microphone",
			Title:    title,
		}), nil
	case "IconThemePath":
		return dbus.MakeVariant(""), nil
	case "Menu":
		return dbus.MakeVariant(dbus.ObjectPath("/NO_DBUSMENU")), nil
	case "ItemIsMenu":
		return dbus.MakeVariant(false), nil
	case "WindowId":
		return dbus.MakeVariant(int32(0)), nil
	}
	return dbus.Variant{}, nil
}

func (s *sniItem) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	props := []string{"Category", "Id", "Title", "Status", "IconName",
		"AttentionIconName", "ToolTip", "IconThemePath", "Menu", "ItemIsMenu", "WindowId"}
	result := make(map[string]dbus.Variant)
	for _, p := range props {
		v, _ := s.Get(iface, p)
		result[p] = v
	}
	return result, nil
}

func (s *sniItem) Set(string, string, dbus.Variant) *dbus.Error {
	return nil
}

// Activate is called when the user clicks the indicator
func (s *sniItem) Activate(x, y int32) *dbus.Error {
	return nil
}

func (s *sniItem) SecondaryActivate(x, y int32) *dbus.Error {
	return nil
}

func (s *sniItem) Scroll(delta int32, orientation string) *dbus.Error {
	return nil
}

var (
	instance *sniItem
	once     sync.Once
)

// Start registers a StatusNotifierItem with the panel. Call once at startup.
func Start() {
	once.Do(func() {
		conn, err := dbus.SessionBus()
		if err != nil {
			slog.Debug("indicator: no session bus", "error", err)
			return
		}

		busName := fmt.Sprintf("org.kde.StatusNotifierItem-%d-1", os.Getpid())
		reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue)
		if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
			slog.Debug("indicator: could not claim bus name", "name", busName, "error", err)
			return
		}

		item := &sniItem{conn: conn, busName: busName}
		instance = item

		// Export the properties interface
		conn.ExportMethodTable(map[string]interface{}{
			"Get":    item.Get,
			"GetAll": item.GetAll,
			"Set":    item.Set,
		}, itemPath, "org.freedesktop.DBus.Properties")

		// Export SNI methods
		conn.ExportMethodTable(map[string]interface{}{
			"Activate":          item.Activate,
			"SecondaryActivate": item.SecondaryActivate,
			"Scroll":            item.Scroll,
		}, itemPath, sniIface)

		// Export introspection
		node := &introspect.Node{
			Name: itemPath,
			Interfaces: []introspect.Interface{
				introspect.IntrospectData,
				{
					Name: sniIface,
					Methods: []introspect.Method{
						{Name: "Activate", Args: []introspect.Arg{
							{Name: "x", Type: "i", Direction: "in"},
							{Name: "y", Type: "i", Direction: "in"},
						}},
						{Name: "SecondaryActivate", Args: []introspect.Arg{
							{Name: "x", Type: "i", Direction: "in"},
							{Name: "y", Type: "i", Direction: "in"},
						}},
						{Name: "Scroll", Args: []introspect.Arg{
							{Name: "delta", Type: "i", Direction: "in"},
							{Name: "orientation", Type: "s", Direction: "in"},
						}},
					},
					Properties: []introspect.Property{
						{Name: "Category", Type: "s", Access: "read"},
						{Name: "Id", Type: "s", Access: "read"},
						{Name: "Title", Type: "s", Access: "read"},
						{Name: "Status", Type: "s", Access: "read"},
						{Name: "IconName", Type: "s", Access: "read"},
						{Name: "AttentionIconName", Type: "s", Access: "read"},
						{Name: "ToolTip", Type: "(sa(iiay)ss)", Access: "read"},
						{Name: "Menu", Type: "o", Access: "read"},
						{Name: "ItemIsMenu", Type: "b", Access: "read"},
					},
					Signals: []introspect.Signal{
						{Name: "NewStatus", Args: []introspect.Arg{{Name: "status", Type: "s"}}},
						{Name: "NewToolTip"},
					},
				},
			},
		}
		conn.Export(introspect.NewIntrospectable(node), itemPath, "org.freedesktop.DBus.Introspectable")

		// Register with the watcher
		call := conn.Object(watcherDest, dbus.ObjectPath(watcherPath)).Call(
			watcherIface+".RegisterStatusNotifierItem", 0, busName)
		if call.Err != nil {
			slog.Debug("indicator: failed to register with watcher", "error", call.Err)
			return
		}

		slog.Info("panel indicator registered")
	})
}

// SetRecording updates the indicator to reflect recording state.
func SetRecording(recording bool) {
	if instance == nil {
		return
	}
	instance.mu.Lock()
	instance.recording = recording
	instance.mu.Unlock()

	// Signal the panel to re-read our status
	instance.conn.Emit(itemPath, sniIface+".NewStatus", func() string {
		if recording {
			return "NeedsAttention"
		}
		return "Active"
	}())
	instance.conn.Emit(itemPath, sniIface+".NewToolTip")
}

// Stop cleans up the indicator.
func Stop() {
	if instance == nil {
		return
	}
	instance.conn.ReleaseName(instance.busName)
}

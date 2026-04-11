//go:build linux

package indicator

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/png"
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

//go:embed icons/idle.png
var idleIcon []byte

//go:embed icons/recording.png
var recordingIcon []byte

// sniItem implements the StatusNotifierItem D-Bus interface.
type sniItem struct {
	mu        sync.RWMutex
	conn      *dbus.Conn
	busName   string
	recording bool
}

func (s *sniItem) icon() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.recording {
		return recordingIcon
	}
	return idleIcon
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
		return dbus.MakeVariant("sussurro"), nil
	case "Status":
		if s.recording {
			return dbus.MakeVariant("NeedsAttention"), nil
		}
		return dbus.MakeVariant("Active"), nil
	case "IconName":
		return dbus.MakeVariant(""), nil
	case "IconPixmap":
		if s.recording {
			return dbus.MakeVariant(iconPixmaps(recordingIcon)), nil
		}
		return dbus.MakeVariant(iconPixmaps(idleIcon)), nil
	case "AttentionIconName":
		return dbus.MakeVariant(""), nil
	case "AttentionIconPixmap":
		if s.recording {
			return dbus.MakeVariant(iconPixmaps(recordingIcon)), nil
		}
		return dbus.MakeVariant(iconPixmaps(idleIcon)), nil
	case "ToolTip":
		title := "sussurro — Ready"
		if s.recording {
			title = "sussurro — Recording"
		}
		return dbus.MakeVariant(struct {
			IconName string
			IconData []struct {
				Width, Height int32
				Data          []byte
			}
			Title string
			Desc  string
		}{
			Title: title,
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
		"IconPixmap", "AttentionIconName", "AttentionIconPixmap",
		"ToolTip", "IconThemePath", "Menu", "ItemIsMenu", "WindowId"}
	result := make(map[string]dbus.Variant)
	for _, p := range props {
		v, _ := s.Get(iface, p)
		result[p] = v
	}
	return result, nil
}

func (s *sniItem) Set(string, string, dbus.Variant) *dbus.Error { return nil }

func (s *sniItem) Activate(x, y int32) *dbus.Error   { return nil }
func (s *sniItem) SecondaryActivate(x, y int32) *dbus.Error { return nil }
func (s *sniItem) Scroll(delta int32, orientation string) *dbus.Error { return nil }

type iconPixmap struct {
	Width, Height int32
	Data          []byte
}

func iconPixmaps(pngData []byte) []iconPixmap {
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	argb := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			off := (y*w + x) * 4
			argb[off+0] = byte(a >> 8)
			argb[off+1] = byte(r >> 8)
			argb[off+2] = byte(g >> 8)
			argb[off+3] = byte(b >> 8)
		}
	}
	return []iconPixmap{{int32(w), int32(h), argb}}
}


var (
	instance *sniItem
	once     sync.Once
)

// Start registers a StatusNotifierItem with the panel.
func Start() {
	once.Do(func() {
		conn, err := dbus.SessionBusPrivate()
		if err != nil {
			slog.Debug("indicator: no session bus", "error", err)
			return
		}
		if err = conn.Auth(nil); err != nil {
			conn.Close()
			slog.Debug("indicator: auth failed", "error", err)
			return
		}
		if err = conn.Hello(); err != nil {
			conn.Close()
			slog.Debug("indicator: hello failed", "error", err)
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

		conn.ExportMethodTable(map[string]interface{}{
			"Get":    item.Get,
			"GetAll": item.GetAll,
			"Set":    item.Set,
		}, itemPath, "org.freedesktop.DBus.Properties")

		conn.ExportMethodTable(map[string]interface{}{
			"Activate":          item.Activate,
			"SecondaryActivate": item.SecondaryActivate,
			"Scroll":            item.Scroll,
		}, itemPath, sniIface)

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
						{Name: "IconPixmap", Type: "a(iiay)", Access: "read"},
						{Name: "AttentionIconName", Type: "s", Access: "read"},
						{Name: "AttentionIconPixmap", Type: "a(iiay)", Access: "read"},
						{Name: "ToolTip", Type: "(sa(iiay)ss)", Access: "read"},
						{Name: "Menu", Type: "o", Access: "read"},
						{Name: "ItemIsMenu", Type: "b", Access: "read"},
					},
					Signals: []introspect.Signal{
						{Name: "NewIcon"},
						{Name: "NewAttentionIcon"},
						{Name: "NewStatus", Args: []introspect.Arg{{Name: "status", Type: "s"}}},
						{Name: "NewToolTip"},
					},
				},
			},
		}
		conn.Export(introspect.NewIntrospectable(node), itemPath, "org.freedesktop.DBus.Introspectable")

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

	status := "Active"
	if recording {
		status = "NeedsAttention"
	}
	instance.conn.Emit(itemPath, sniIface+".NewStatus", status)
	instance.conn.Emit(itemPath, sniIface+".NewToolTip")
	instance.conn.Emit(itemPath, sniIface+".NewIcon")
	instance.conn.Emit(itemPath, sniIface+".NewAttentionIcon")
}

// Stop removes the indicator from the panel by closing the private D-Bus connection.
// The watcher detects the name vanishing and removes the icon.
func Stop() {
	if instance == nil {
		return
	}
	instance.conn.Close()
	instance = nil
	once = sync.Once{}
}

// Package sni implements the StatusNotifierItem D-Bus protocol so the
// app can show a panel/tray icon on Linux desktops (KDE, COSMIC, GNOME
// with the appindicator extension, etc.). It is the only consumer of
// godbus today.
//
// The parent internal/indicator package wraps this in a thin build-tagged
// adapter that satisfies the indicator API and provides the icon bytes.
package sni

import (
	"bytes"
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

// Item is a registered StatusNotifierItem. The icon swap on SetRecording
// is announced via the standard NewIcon / NewStatus / NewToolTip signals,
// which is what panels listen for.
type Item struct {
	mu            sync.RWMutex
	conn          *dbus.Conn
	busName       string
	idleIcon      []byte
	recordingIcon []byte
	recording     bool
	onActivate    func()
}

// New connects to the session bus, claims a unique name, exports the
// item interface and registers with the StatusNotifierWatcher. Returns
// nil + a debug-level reason on systems where any of those steps fail
// (no session bus, no watcher, name conflict) — the caller treats that
// as "panel indicators are unavailable here".
//
// idleIcon and recordingIcon are PNG-encoded byte slices the caller
// owns; the package decodes them on demand for D-Bus property reads.
func New(idleIcon, recordingIcon []byte, onActivate func()) (*Item, error) {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, fmt.Errorf("session bus: %w", err)
	}
	if err = conn.Auth(nil); err != nil {
		closeWarn(conn, "after auth failure")
		return nil, fmt.Errorf("auth: %w", err)
	}
	if err = conn.Hello(); err != nil {
		closeWarn(conn, "after hello failure")
		return nil, fmt.Errorf("hello: %w", err)
	}
	busName := fmt.Sprintf("org.kde.StatusNotifierItem-%d-1", os.Getpid())
	reply, err := conn.RequestName(busName, dbus.NameFlagDoNotQueue)
	if err != nil {
		closeWarn(conn, "after request-name failure")
		return nil, fmt.Errorf("request name %q: %w", busName, err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		closeWarn(conn, "after non-primary owner")
		return nil, fmt.Errorf("not primary owner of %q", busName)
	}

	item := &Item{
		conn:          conn,
		busName:       busName,
		idleIcon:      idleIcon,
		recordingIcon: recordingIcon,
		onActivate:    onActivate,
	}

	if err := conn.ExportMethodTable(map[string]interface{}{
		"Get":    item.Get,
		"GetAll": item.GetAll,
		"Set":    item.Set,
	}, itemPath, "org.freedesktop.DBus.Properties"); err != nil {
		closeWarn(conn, "after export properties")
		return nil, fmt.Errorf("export properties: %w", err)
	}

	if err := conn.ExportMethodTable(map[string]interface{}{
		"Activate":          item.Activate,
		"SecondaryActivate": item.SecondaryActivate,
		"Scroll":            item.Scroll,
	}, itemPath, sniIface); err != nil {
		closeWarn(conn, "after export item")
		return nil, fmt.Errorf("export item: %w", err)
	}

	if err := conn.Export(introspect.NewIntrospectable(introspectionNode()), itemPath, "org.freedesktop.DBus.Introspectable"); err != nil {
		closeWarn(conn, "after export introspectable")
		return nil, fmt.Errorf("export introspectable: %w", err)
	}

	call := conn.Object(watcherDest, dbus.ObjectPath(watcherPath)).Call(
		watcherIface+".RegisterStatusNotifierItem", 0, busName)
	if call.Err != nil {
		closeWarn(conn, "after watcher register")
		return nil, fmt.Errorf("register with watcher: %w", call.Err)
	}

	return item, nil
}

// SetRecording flips the indicator state and emits the standard signals
// so panels refresh the displayed icon.
func (s *Item) SetRecording(recording bool) {
	s.mu.Lock()
	s.recording = recording
	s.mu.Unlock()

	status := "Active"
	if recording {
		status = "NeedsAttention"
	}
	for sig, args := range map[string][]interface{}{
		sniIface + ".NewStatus":        {status},
		sniIface + ".NewToolTip":       nil,
		sniIface + ".NewIcon":          nil,
		sniIface + ".NewAttentionIcon": nil,
	} {
		if err := s.conn.Emit(itemPath, sig, args...); err != nil {
			slog.Warn("sni: emit failed", "signal", sig, "error", err)
		}
	}
}

// Close releases the bus name. The watcher detects the name vanishing
// and pulls the icon. Idempotent.
func (s *Item) Close() error {
	if s == nil || s.conn == nil {
		return nil
	}
	err := s.conn.Close()
	s.conn = nil
	return err
}

// --- D-Bus property handlers -----------------------------------------------

func (s *Item) Get(iface, prop string) (dbus.Variant, *dbus.Error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if iface != sniIface {
		return dbus.Variant{}, nil
	}
	switch prop {
	case "Category":
		return dbus.MakeVariant("ApplicationStatus"), nil
	case "Id", "Title":
		return dbus.MakeVariant("shushingface"), nil
	case "Status":
		if s.recording {
			return dbus.MakeVariant("NeedsAttention"), nil
		}
		return dbus.MakeVariant("Active"), nil
	case "IconName", "AttentionIconName", "IconThemePath":
		return dbus.MakeVariant(""), nil
	case "IconPixmap", "AttentionIconPixmap":
		if s.recording {
			return dbus.MakeVariant(iconPixmaps(s.recordingIcon)), nil
		}
		return dbus.MakeVariant(iconPixmaps(s.idleIcon)), nil
	case "ToolTip":
		title := "shushingface — Ready"
		if s.recording {
			title = "shushingface — Recording"
		}
		return dbus.MakeVariant(struct {
			IconName string
			IconData []struct {
				Width, Height int32
				Data          []byte
			}
			Title string
			Desc  string
		}{Title: title}), nil
	case "Menu":
		return dbus.MakeVariant(dbus.ObjectPath("/NO_DBUSMENU")), nil
	case "ItemIsMenu":
		return dbus.MakeVariant(false), nil
	case "WindowId":
		return dbus.MakeVariant(int32(0)), nil
	}
	return dbus.Variant{}, nil
}

func (s *Item) GetAll(iface string) (map[string]dbus.Variant, *dbus.Error) {
	props := []string{"Category", "Id", "Title", "Status", "IconName",
		"IconPixmap", "AttentionIconName", "AttentionIconPixmap",
		"ToolTip", "IconThemePath", "Menu", "ItemIsMenu", "WindowId"}
	result := make(map[string]dbus.Variant, len(props))
	for _, p := range props {
		v, _ := s.Get(iface, p)
		result[p] = v
	}
	return result, nil
}

func (s *Item) Set(string, string, dbus.Variant) *dbus.Error { return nil }

func (s *Item) Activate(_, _ int32) *dbus.Error {
	if s.onActivate != nil {
		s.onActivate()
	}
	return nil
}

func (s *Item) SecondaryActivate(_, _ int32) *dbus.Error { return nil }
func (s *Item) Scroll(_ int32, _ string) *dbus.Error     { return nil }

// --- helpers ---------------------------------------------------------------

type iconPixmap struct {
	Width, Height int32
	Data          []byte
}

// iconPixmaps decodes a PNG into the ARGB byte format StatusNotifierItem
// requires. Returns nil if decoding fails — the panel falls back to the
// blank IconName.
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

func introspectionNode() *introspect.Node {
	return &introspect.Node{
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
}

func closeWarn(conn *dbus.Conn, what string) {
	if err := conn.Close(); err != nil {
		slog.Warn("sni: failed to close D-Bus connection", "what", what, "error", err)
	}
}

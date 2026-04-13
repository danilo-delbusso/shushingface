// Package winhotkey is the Win32 implementation of global hotkey
// registration. Toggle shortcuts are routed through golang.design/x/hotkey
// (which owns its own HWND + message pump, debounced for OS auto-repeat);
// push-to-talk shortcuts use a low-level keyboard hook installed on a
// dedicated message-pumping OS thread.
//
// The parent internal/hotkey package wraps this in a thin build-tagged
// adapter that satisfies hotkey.Manager — see hotkey/hotkey_windows.go.
package winhotkey

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"syscall"
	"time"

	gdhotkey "golang.design/x/hotkey"

	"codeberg.org/dbus/shushingface/internal/win32"
)

// ErrConflict is returned when another process already owns the hotkey.
var ErrConflict = errors.New("hotkey is already registered by another application")

// Modifier bitmask matching the Win32 RegisterHotKey constants.
const (
	ModCtrl  uint32 = 0x0002
	ModAlt   uint32 = 0x0001
	ModShift uint32 = 0x0004
	ModWin   uint32 = 0x0008
)

// Mode picks which delivery model a registration uses.
type Mode int

const (
	// ModeToggle delivers one Trigger event per discrete press.
	ModeToggle Mode = iota
	// ModePushToTalk delivers Press on keydown and Release on keyup.
	ModePushToTalk
)

// EventType describes which kind of hotkey event was delivered.
type EventType int

const (
	Trigger EventType = iota
	Press
	Release
)

// Event is one occurrence of a registered hotkey.
type Event struct {
	Name string
	Type EventType
}

// Spec is an already-parsed shortcut.
type Spec struct {
	Mods uint32 // bitmask of ModCtrl/ModAlt/ModShift/ModWin
	VK   uint32 // Win32 virtual-key code
}

// Minimum gap between consecutive toggle-hotkey triggers we'll emit.
// golang.design/x/hotkey does not set MOD_NOREPEAT, so holding the key
// for even a few ms fires a flood of WM_HOTKEY events. We swallow
// anything within this window of the previous trigger.
const toggleDebounce = 350 * time.Millisecond

// Windows error returned by RegisterHotKey when the combination is
// already claimed by another process.
const errHotkeyAlreadyRegistered syscall.Errno = 1409

// Thread-message code used to tell the hook pump goroutine to exit.
const wmHookShutdown = win32.WMApp + 1

type registration struct {
	mode Mode
	// Toggle-mode: library handle + cancel signal for the watcher goroutine.
	hk       *gdhotkey.Hotkey
	cancelCh chan struct{}
	// PTT-mode: match criteria + current keydown latch (debounces auto-repeat).
	mods uint32
	vk   uint32
	down bool
}

// Manager owns a set of named hotkey registrations and one Events channel
// across both modes. Safe for use from any goroutine.
type Manager struct {
	mu           sync.Mutex
	regs         map[string]*registration
	events       chan Event
	hookHwnd     uintptr
	hookThreadID uint32
	hookStopped  chan struct{}
}

// New returns an empty Manager. No OS resources are claimed until the
// first registration.
func New() *Manager {
	return &Manager{
		regs:   map[string]*registration{},
		events: make(chan Event, 16),
	}
}

// activeManager is read by the C-style hook callback; set while a PTT
// registration is active. Only one Manager can hold the hook at a time —
// a single app instance is the only realistic caller.
var activeManager *Manager

// Register installs a hotkey under the given name. Re-registering an
// existing name replaces the previous binding atomically. Returns
// ErrConflict when another process owns the combination.
func (m *Manager) Register(name string, spec Spec, mode Mode) error {
	if spec.Mods == 0 {
		return errors.New("at least one modifier required")
	}
	if spec.VK == 0 {
		return errors.New("virtual key code required")
	}

	if err := m.Unregister(name); err != nil {
		slog.Warn("winhotkey: prior unregister failed during re-register", "name", name, "error", err)
	}

	slog.Debug("winhotkey: registering",
		"name", name, "mode", mode, "vk", spec.VK, "mods", spec.Mods)

	switch mode {
	case ModeToggle:
		hk := gdhotkey.New(libMods(spec.Mods), gdhotkey.Key(spec.VK))
		if err := hk.Register(); err != nil {
			if isConflictErr(err) {
				slog.Warn("winhotkey: conflict", "name", name)
				return ErrConflict
			}
			return fmt.Errorf("RegisterHotKey: %w", err)
		}
		reg := &registration{mode: mode, hk: hk, cancelCh: make(chan struct{})}
		m.mu.Lock()
		m.regs[name] = reg
		m.mu.Unlock()
		go m.toggleLoop(name, reg)
		return nil

	case ModePushToTalk:
		if err := m.ensureHook(); err != nil {
			return err
		}
		reg := &registration{mode: mode, mods: spec.Mods, vk: spec.VK}
		m.mu.Lock()
		m.regs[name] = reg
		activeManager = m
		m.mu.Unlock()
		return nil
	}
	return fmt.Errorf("unknown mode %d", mode)
}

func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	reg, ok := m.regs[name]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	delete(m.regs, name)
	hasPTT := false
	for _, r := range m.regs {
		if r.mode == ModePushToTalk {
			hasPTT = true
			break
		}
	}
	m.mu.Unlock()

	switch reg.mode {
	case ModeToggle:
		if reg.hk != nil {
			if err := reg.hk.Unregister(); err != nil {
				slog.Warn("winhotkey: lib Unregister failed", "name", name, "error", err)
			}
		}
		close(reg.cancelCh)
	case ModePushToTalk:
		if !hasPTT {
			m.tearDownHook()
		}
	}
	return nil
}

func (m *Manager) Events() <-chan Event { return m.events }

func (m *Manager) Close() error {
	m.mu.Lock()
	names := make([]string, 0, len(m.regs))
	for name := range m.regs {
		names = append(names, name)
	}
	m.mu.Unlock()
	for _, n := range names {
		if err := m.Unregister(n); err != nil {
			slog.Warn("winhotkey: unregister on close failed", "name", n, "error", err)
		}
	}
	return nil
}

func (m *Manager) toggleLoop(name string, reg *registration) {
	var lastFired time.Time
	for {
		select {
		case _, ok := <-reg.hk.Keydown():
			if !ok {
				return
			}
			if since := time.Since(lastFired); since < toggleDebounce {
				slog.Debug("winhotkey: trigger suppressed (auto-repeat)",
					"name", name, "since_last", since)
				continue
			}
			lastFired = time.Now()
			select {
			case m.events <- Event{Name: name, Type: Trigger}:
			default:
				slog.Warn("winhotkey: event channel full, dropping", "name", name)
			}
		case <-reg.cancelCh:
			return
		}
	}
}

// ensureHook installs the WH_KEYBOARD_LL hook on a dedicated OS thread
// running a Win32 message loop. Without an active GetMessageW pump the
// system silently drops hook callbacks.
func (m *Manager) ensureHook() error {
	m.mu.Lock()
	have := m.hookHwnd != 0
	m.mu.Unlock()
	if have {
		return nil
	}

	ready := make(chan error, 1)
	stopped := make(chan struct{})
	go m.hookPump(ready, stopped)
	if err := <-ready; err != nil {
		return err
	}
	m.mu.Lock()
	m.hookStopped = stopped
	m.mu.Unlock()
	return nil
}

func (m *Manager) hookPump(ready chan<- error, stopped chan<- struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(stopped)

	tid := win32.GetCurrentThreadID()
	hook, err := win32.SetWindowsHookExW(
		win32.WHKeyboardLL,
		win32.NewCallback(keyboardProc),
		win32.GetModuleHandleW(),
		0,
	)
	if err != nil {
		ready <- fmt.Errorf("SetWindowsHookEx: %w", err)
		return
	}

	m.mu.Lock()
	m.hookHwnd = hook
	m.hookThreadID = tid
	m.mu.Unlock()
	ready <- nil

	var msg win32.MSG
	for {
		r := win32.GetMessageW(&msg)
		if r <= 0 {
			break
		}
		if msg.Message == wmHookShutdown {
			break
		}
		win32.TranslateMessage(&msg)
		win32.DispatchMessageW(&msg)
	}

	win32.UnhookWindowsHookEx(hook)
	m.mu.Lock()
	m.hookHwnd = 0
	m.hookThreadID = 0
	m.mu.Unlock()
}

func (m *Manager) tearDownHook() {
	m.mu.Lock()
	tid := m.hookThreadID
	stopped := m.hookStopped
	m.hookStopped = nil
	if activeManager == m {
		activeManager = nil
	}
	m.mu.Unlock()
	if tid == 0 {
		return
	}
	win32.PostThreadMessageW(tid, wmHookShutdown, 0, 0)
	if stopped != nil {
		<-stopped
	}
}

func keyboardProc(nCode int32, wParam uintptr, lParam uintptr) uintptr {
	if nCode < 0 || activeManager == nil {
		return win32.CallNextHookEx(nCode, wParam, lParam)
	}
	kbd := win32.KBDFromLParam(lParam)
	vk := kbd.VkCode
	switch wParam {
	case win32.WMKeyDown, win32.WMSysKeyDown:
		activeManager.handlePTT(vk, true)
	case win32.WMKeyUp, win32.WMSysKeyUp:
		activeManager.handlePTT(vk, false)
	}
	return win32.CallNextHookEx(nCode, wParam, lParam)
}

func (m *Manager) handlePTT(vk uint32, down bool) {
	mods := currentMods()
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, reg := range m.regs {
		if reg.mode != ModePushToTalk || reg.vk != vk {
			continue
		}
		if down {
			if mods != reg.mods || reg.down {
				return
			}
			reg.down = true
			select {
			case m.events <- Event{Name: name, Type: Press}:
			default:
				slog.Warn("winhotkey: PTT press dropped", "name", name)
			}
			return
		}
		if reg.down {
			reg.down = false
			select {
			case m.events <- Event{Name: name, Type: Release}:
			default:
				slog.Warn("winhotkey: PTT release dropped", "name", name)
			}
			return
		}
	}
}

func currentMods() uint32 {
	var mods uint32
	if win32.IsKeyDown(win32.VKLControl) || win32.IsKeyDown(win32.VKRControl) {
		mods |= ModCtrl
	}
	if win32.IsKeyDown(win32.VKLAlt) || win32.IsKeyDown(win32.VKRAlt) {
		mods |= ModAlt
	}
	if win32.IsKeyDown(win32.VKLShift) || win32.IsKeyDown(win32.VKRShift) {
		mods |= ModShift
	}
	if win32.IsKeyDown(win32.VKLWin) || win32.IsKeyDown(win32.VKRWin) {
		mods |= ModWin
	}
	return mods
}

func isConflictErr(err error) bool {
	if err == nil {
		return false
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == errHotkeyAlreadyRegistered
	}
	return false
}

func libMods(m uint32) []gdhotkey.Modifier {
	var out []gdhotkey.Modifier
	if m&ModCtrl != 0 {
		out = append(out, gdhotkey.ModCtrl)
	}
	if m&ModAlt != 0 {
		out = append(out, gdhotkey.ModAlt)
	}
	if m&ModShift != 0 {
		out = append(out, gdhotkey.ModShift)
	}
	if m&ModWin != 0 {
		out = append(out, gdhotkey.ModWin)
	}
	return out
}

// VKFromKey maps a canonical key string (as produced by hotkey.canonKey)
// to its Win32 virtual-key code. Returns false for keys this package
// doesn't know about — callers should treat that as ErrInvalidSpec.
func VKFromKey(key string) (uint32, bool) {
	if len(key) == 1 {
		c := key[0]
		switch {
		case c >= 'A' && c <= 'Z':
			return uint32(c), true
		case c >= '0' && c <= '9':
			return uint32(c), true
		}
	}
	switch key {
	case "Space":
		return 0x20, true
	case "Enter", "Return":
		return 0x0D, true
	case "Tab":
		return 0x09, true
	case "Escape":
		return 0x1B, true
	case "Backspace":
		return 0x08, true
	case "Delete":
		return 0x2E, true
	case "Insert":
		return 0x2D, true
	case "Home":
		return 0x24, true
	case "End":
		return 0x23, true
	case "Pageup":
		return 0x21, true
	case "Pagedown":
		return 0x22, true
	case "Left", "Arrowleft":
		return 0x25, true
	case "Up", "Arrowup":
		return 0x26, true
	case "Right", "Arrowright":
		return 0x27, true
	case "Down", "Arrowdown":
		return 0x28, true
	}
	if len(key) >= 2 && (key[0] == 'F' || key[0] == 'f') {
		var n int
		if _, err := fmt.Sscanf(key[1:], "%d", &n); err == nil && n >= 1 && n <= 24 {
			return uint32(0x70 + n - 1), true
		}
	}
	return 0, false
}

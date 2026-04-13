//go:build windows

package hotkey

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	gdhotkey "golang.design/x/hotkey"

	"codeberg.org/dbus/shushingface/internal/platform"
	"codeberg.org/dbus/shushingface/internal/win32"
)

func capability() platform.Capability { return platform.Supported() }

// Minimum gap between consecutive toggle-hotkey triggers we'll emit.
// golang.design/x/hotkey does not set MOD_NOREPEAT, so holding the key for
// even a few ms fires a flood of WM_HOTKEY events. We swallow anything
// within this window of the previous trigger.
const toggleDebounce = 350 * time.Millisecond

// Windows error returned by RegisterHotKey when the combination is already
// claimed by another process.
const errHotkeyAlreadyRegistered syscall.Errno = 1409

// Local RegisterHotKey modifier bitmask (distinct from the hook-path
// modifier tracking done via win32.VK* key-state checks).
const (
	modControl = 0x0002
	modAlt     = 0x0001
	modShift   = 0x0004
	modWin     = 0x0008
)

type registration struct {
	mode Mode
	// Toggle mode: library handle + signal channel to stop the watcher goroutine.
	hk       *gdhotkey.Hotkey
	cancelCh chan struct{}
	// PTT mode: match criteria + current keydown state (debounces OS auto-repeat).
	mods uint32
	vk   uint32
	down bool
}

type manager struct {
	mu           sync.Mutex
	regs         map[string]*registration
	events       chan Event
	hookHwnd     uintptr
	hookThreadID uint32        // OS thread hosting the WH_KEYBOARD_LL pump
	hookStopped  chan struct{} // closed when hook goroutine exits
}

// Thread-message code used to tell the hook pump goroutine to exit.
const wmHookShutdown = win32.WMApp + 1

// New creates a Windows hotkey manager. Toggle shortcuts use
// golang.design/x/hotkey (HWND-backed, battle-tested message loop);
// push-to-talk shortcuts use a shared low-level keyboard hook.
func New() Manager {
	return &manager{
		regs:   map[string]*registration{},
		events: make(chan Event, 16),
	}
}

// activeManager is read by the C-style hook callback; set while a PTT
// registration is active.
var activeManager *manager

func (m *manager) Register(name string, spec Spec, mode Mode) error {
	vk, ok := vkFromKey(spec.Key)
	if !ok {
		return fmt.Errorf("%w: unsupported key %q", ErrInvalidSpec, spec.Key)
	}
	mods := winMods(spec.Mods)
	if mods == 0 {
		return fmt.Errorf("%w: at least one modifier required", ErrInvalidSpec)
	}

	// Replace any prior registration with the same name.
	if err := m.Unregister(name); err != nil {
		slog.Warn("hotkey: prior unregister failed during re-register", "name", name, "error", err)
	}

	slog.Debug("hotkey: registering",
		"name", name, "spec", FormatSpec(spec), "mode", mode, "vk", vk, "mods", mods)

	switch mode {
	case ModeToggle:
		hk := gdhotkey.New(libMods(spec.Mods), gdhotkey.Key(vk))
		if err := hk.Register(); err != nil {
			if isConflictErr(err) {
				slog.Warn("hotkey: conflict", "name", name, "spec", FormatSpec(spec))
				return ErrConflict
			}
			slog.Warn("hotkey: register failed", "name", name, "spec", FormatSpec(spec), "error", err)
			return fmt.Errorf("RegisterHotKey: %w", err)
		}

		reg := &registration{mode: mode, hk: hk, cancelCh: make(chan struct{})}
		m.mu.Lock()
		m.regs[name] = reg
		m.mu.Unlock()

		go m.toggleLoop(name, reg)
		slog.Debug("hotkey: toggle registered", "name", name)
		return nil

	case ModePushToTalk:
		if err := m.ensureHook(); err != nil {
			return err
		}
		reg := &registration{mode: mode, mods: mods, vk: vk}
		m.mu.Lock()
		m.regs[name] = reg
		activeManager = m
		m.mu.Unlock()
		slog.Debug("hotkey: ptt registered", "name", name)
		return nil
	}
	return fmt.Errorf("unknown mode %d", mode)
}

func (m *manager) toggleLoop(name string, reg *registration) {
	slog.Debug("hotkey: toggleLoop starting", "name", name)
	defer slog.Debug("hotkey: toggleLoop exiting", "name", name)
	var lastFired time.Time
	for {
		select {
		case _, ok := <-reg.hk.Keydown():
			if !ok {
				return
			}
			if since := time.Since(lastFired); since < toggleDebounce {
				slog.Debug("hotkey: trigger suppressed (auto-repeat)",
					"name", name, "since_last", since)
				continue
			}
			lastFired = time.Now()
			slog.Debug("hotkey: trigger fired", "name", name)
			select {
			case m.events <- Event{Name: name, Type: Trigger}:
			default:
				slog.Warn("hotkey: event channel full, dropping", "name", name)
			}
		case <-reg.cancelCh:
			return
		}
	}
}

func (m *manager) Unregister(name string) error {
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

	slog.Debug("hotkey: unregistering", "name", name, "mode", reg.mode)

	switch reg.mode {
	case ModeToggle:
		if reg.hk != nil {
			if err := reg.hk.Unregister(); err != nil {
				slog.Warn("hotkey: lib Unregister failed", "name", name, "error", err)
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

func (m *manager) Events() <-chan Event { return m.events }

func (m *manager) Close() error {
	m.mu.Lock()
	names := make([]string, 0, len(m.regs))
	for name := range m.regs {
		names = append(names, name)
	}
	m.mu.Unlock()
	for _, n := range names {
		if err := m.Unregister(n); err != nil {
			slog.Warn("hotkey: unregister on close failed", "name", n, "error", err)
		}
	}
	return nil
}

// ensureHook installs the low-level keyboard hook on a dedicated OS thread
// that runs a Win32 message loop. WH_KEYBOARD_LL callbacks only fire on
// threads with an active GetMessageW pump — without this, no PTT event
// would ever dispatch.
func (m *manager) ensureHook() error {
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
	slog.Debug("hotkey: keyboard hook installed")
	return nil
}

// hookPump is pinned to an OS thread; it installs the hook on that thread
// and then services its Windows message queue until told to shut down.
func (m *manager) hookPump(ready chan<- error, stopped chan<- struct{}) {
	runtimeLockOSThread()
	defer runtimeUnlockOSThread()
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
			break // WM_QUIT or error
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
	slog.Debug("hotkey: keyboard hook pump exited")
}

func (m *manager) tearDownHook() {
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
	slog.Debug("hotkey: keyboard hook removed")
}

func runtimeLockOSThread()   { runtime.LockOSThread() }
func runtimeUnlockOSThread() { runtime.UnlockOSThread() }

// keyboardProc is the low-level keyboard hook callback.
func keyboardProc(nCode int32, wParam uintptr, lParam uintptr) uintptr {
	if nCode < 0 || activeManager == nil {
		return win32.CallNextHookEx(nCode, wParam, lParam)
	}
	kbd := (*win32.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
	vk := kbd.VkCode

	switch wParam {
	case win32.WMKeyDown, win32.WMSysKeyDown:
		activeManager.handlePTT(vk, true)
	case win32.WMKeyUp, win32.WMSysKeyUp:
		activeManager.handlePTT(vk, false)
	}

	return win32.CallNextHookEx(nCode, wParam, lParam)
}

func (m *manager) handlePTT(vk uint32, down bool) {
	mods := currentMods()
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, reg := range m.regs {
		if reg.mode != ModePushToTalk {
			continue
		}
		if reg.vk != vk {
			continue
		}
		if down {
			if mods != reg.mods {
				continue
			}
			if reg.down {
				return // ignore auto-repeat
			}
			reg.down = true
			select {
			case m.events <- Event{Name: name, Type: Press}:
			default:
				slog.Warn("hotkey: PTT press dropped", "name", name)
			}
			return
		}
		if reg.down {
			reg.down = false
			select {
			case m.events <- Event{Name: name, Type: Release}:
			default:
				slog.Warn("hotkey: PTT release dropped", "name", name)
			}
			return
		}
	}
}

func currentMods() uint32 {
	var mods uint32
	if win32.IsKeyDown(win32.VKLControl) || win32.IsKeyDown(win32.VKRControl) {
		mods |= modControl
	}
	if win32.IsKeyDown(win32.VKLAlt) || win32.IsKeyDown(win32.VKRAlt) {
		mods |= modAlt
	}
	if win32.IsKeyDown(win32.VKLShift) || win32.IsKeyDown(win32.VKRShift) {
		mods |= modShift
	}
	if win32.IsKeyDown(win32.VKLWin) || win32.IsKeyDown(win32.VKRWin) {
		mods |= modWin
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

func winMods(m Modifier) uint32 {
	var out uint32
	if m&ModCtrl != 0 {
		out |= modControl
	}
	if m&ModAlt != 0 {
		out |= modAlt
	}
	if m&ModShift != 0 {
		out |= modShift
	}
	if m&ModSuper != 0 {
		out |= modWin
	}
	return out
}

func libMods(m Modifier) []gdhotkey.Modifier {
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
	if m&ModSuper != 0 {
		out = append(out, gdhotkey.ModWin)
	}
	return out
}

func vkFromKey(key string) (uint32, bool) {
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

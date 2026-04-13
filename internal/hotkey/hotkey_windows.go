//go:build windows

package hotkey

import (
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func runtimeLockOSThread()   { runtime.LockOSThread() }
func runtimeUnlockOSThread() { runtime.UnlockOSThread() }

const (
	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008
	modNoRepeat = 0x4000

	wmHotkey     = 0x0312
	wmKeyDown    = 0x0100
	wmKeyUp      = 0x0101
	wmSysKeyDown = 0x0104
	wmSysKeyUp   = 0x0105

	wmClose      = 0x0010
	wmApp        = 0x8000
	wmRegister   = wmApp + 1
	wmUnregister = wmApp + 2

	whKeyboardLL = 13

	vkLControl = 0xA2
	vkRControl = 0xA3
	vkLAlt     = 0xA4
	vkRAlt     = 0xA5
	vkLShift   = 0xA0
	vkRShift   = 0xA1
	vkLWin     = 0x5B
	vkRWin     = 0x5C
)

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procPostThreadMessage  = user32.NewProc("PostThreadMessageW")
	procSetWindowsHookExW  = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx     = user32.NewProc("CallNextHookEx")
	procGetKeyState        = user32.NewProc("GetKeyState")

	kernel32               = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")
	procGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
)

type msg struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       [2]int32
	LPrivate uint32
}

type kbdHookStruct struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type registration struct {
	id   uint16
	mode Mode
	mods uint32
	vk   uint32
	down bool // for PTT: tracks current keydown state (debounces auto-repeat)
}

type regReq struct {
	id   uint16
	mods uint32
	vk   uint32
	resp chan error
}

type unregReq struct {
	id   uint16
	resp chan error
}

type manager struct {
	mu       sync.Mutex
	regs     map[string]*registration
	next     uint16
	threadID uint32
	events   chan Event
	regCh    chan regReq
	unregCh  chan unregReq
	hookHwnd uintptr
	done     chan struct{}
}

// New creates a Windows hotkey manager. The returned manager runs its own
// OS-thread-pinned message loop; Close to tear it down.
func New() Manager {
	m := &manager{
		regs:    map[string]*registration{},
		next:    1,
		events:  make(chan Event, 16),
		regCh:   make(chan regReq, 16),
		unregCh: make(chan unregReq, 16),
		done:    make(chan struct{}),
	}
	ready := make(chan struct{})
	go m.loop(ready)
	<-ready
	return m
}

// Detect returns platform capabilities for hotkey registration.
func Detect() Capabilities {
	return Capabilities{Supported: true, ConflictCheck: true}
}

// activeManager is set while the manager goroutine runs so the C-style
// hook callback can find the channel/registrations.
var activeManager *manager

func (m *manager) loop(ready chan<- struct{}) {
	runtimeLockOSThread()
	defer runtimeUnlockOSThread()

	tid, _, _ := procGetCurrentThreadID.Call()
	m.threadID = uint32(tid)

	activeManager = m
	defer func() { activeManager = nil }()

	close(ready)

	var message msg
	for {
		select {
		case req := <-m.regCh:
			req.resp <- m.doRegister(req.id, req.mods, req.vk)
			continue
		case req := <-m.unregCh:
			req.resp <- m.doUnregister(req.id)
			continue
		case <-m.done:
			m.tearDownHook()
			return
		default:
		}

		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		if message.Message == wmHotkey {
			m.dispatchHotkey(uint16(message.WParam))
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}
}

func (m *manager) dispatchHotkey(id uint16) {
	m.mu.Lock()
	var name string
	for n, r := range m.regs {
		if r.id == id {
			name = n
			break
		}
	}
	m.mu.Unlock()
	if name == "" {
		return
	}
	select {
	case m.events <- Event{Name: name, Type: Trigger}:
	default:
		slog.Warn("hotkey: event channel full, dropping", "name", name)
	}
}

func (m *manager) doRegister(id uint16, mods, vk uint32) error {
	r, _, err := procRegisterHotKey.Call(0, uintptr(id), uintptr(mods|modNoRepeat), uintptr(vk))
	if r == 0 {
		if errno, ok := err.(syscall.Errno); ok && errno == 1409 {
			return ErrConflict
		}
		return fmt.Errorf("RegisterHotKey: %w", err)
	}
	return nil
}

func (m *manager) doUnregister(id uint16) error {
	r, _, err := procUnregisterHotKey.Call(0, uintptr(id))
	if r == 0 {
		return fmt.Errorf("UnregisterHotKey: %w", err)
	}
	return nil
}

func (m *manager) wake(code uint32) {
	procPostThreadMessage.Call(uintptr(m.threadID), uintptr(code), 0, 0)
}

func (m *manager) Register(name string, spec Spec, mode Mode) error {
	vk, ok := vkFromKey(spec.Key)
	if !ok {
		return fmt.Errorf("%w: unsupported key %q", ErrInvalidSpec, spec.Key)
	}
	mods := winMods(spec.Mods)
	if mods == 0 {
		return fmt.Errorf("%w: at least one modifier required", ErrInvalidSpec)
	}

	m.mu.Lock()
	if existing, ok := m.regs[name]; ok {
		existingID := existing.id
		existingMode := existing.mode
		delete(m.regs, name)
		m.mu.Unlock()
		if existingMode == ModeToggle {
			resp := make(chan error, 1)
			m.unregCh <- unregReq{id: existingID, resp: resp}
			m.wake(wmUnregister)
			if err := <-resp; err != nil {
				slog.Warn("hotkey: stale unregister failed", "error", err)
			}
		}
		m.mu.Lock()
	}
	id := m.next
	m.next++
	reg := &registration{id: id, mode: mode, mods: mods, vk: vk}
	m.regs[name] = reg
	m.mu.Unlock()

	switch mode {
	case ModeToggle:
		resp := make(chan error, 1)
		m.regCh <- regReq{id: id, mods: mods, vk: vk, resp: resp}
		m.wake(wmRegister)
		if err := <-resp; err != nil {
			m.mu.Lock()
			delete(m.regs, name)
			m.mu.Unlock()
			return err
		}
	case ModePushToTalk:
		if err := m.ensureHook(); err != nil {
			m.mu.Lock()
			delete(m.regs, name)
			m.mu.Unlock()
			return err
		}
	}
	return nil
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
	if reg.mode == ModeToggle {
		resp := make(chan error, 1)
		m.unregCh <- unregReq{id: reg.id, resp: resp}
		m.wake(wmUnregister)
		return <-resp
	}
	if !hasPTT {
		m.tearDownHook()
	}
	return nil
}

func (m *manager) Events() <-chan Event { return m.events }

func (m *manager) Close() error {
	m.mu.Lock()
	toggleIDs := make([]uint16, 0, len(m.regs))
	for _, r := range m.regs {
		if r.mode == ModeToggle {
			toggleIDs = append(toggleIDs, r.id)
		}
	}
	m.regs = map[string]*registration{}
	m.mu.Unlock()

	for _, id := range toggleIDs {
		resp := make(chan error, 1)
		m.unregCh <- unregReq{id: id, resp: resp}
		m.wake(wmUnregister)
		if err := <-resp; err != nil {
			slog.Warn("hotkey: unregister on close failed", "id", id, "error", err)
		}
	}
	close(m.done)
	m.wake(wmClose)
	return nil
}

func (m *manager) ensureHook() error {
	m.mu.Lock()
	have := m.hookHwnd != 0
	m.mu.Unlock()
	if have {
		return nil
	}

	hInst, _, _ := procGetModuleHandleW.Call(0)
	hook, _, err := procSetWindowsHookExW.Call(
		uintptr(whKeyboardLL),
		syscall.NewCallback(keyboardProc),
		hInst,
		0,
	)
	if hook == 0 {
		return fmt.Errorf("SetWindowsHookEx: %w", err)
	}
	m.mu.Lock()
	m.hookHwnd = hook
	m.mu.Unlock()
	return nil
}

func (m *manager) tearDownHook() {
	m.mu.Lock()
	hook := m.hookHwnd
	m.hookHwnd = 0
	m.mu.Unlock()
	if hook != 0 {
		procUnhookWindowsHookEx.Call(hook)
	}
}

// keyboardProc is the low-level keyboard hook callback. Runs on the manager's
// message-loop thread (where the hook was installed).
func keyboardProc(nCode int32, wParam uintptr, lParam uintptr) uintptr {
	if nCode < 0 || activeManager == nil {
		r, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, uintptr(lParam))
		return r
	}
	kbd := (*kbdHookStruct)(unsafe.Pointer(lParam))
	vk := kbd.VkCode

	switch wParam {
	case wmKeyDown, wmSysKeyDown:
		activeManager.handlePTT(vk, true)
	case wmKeyUp, wmSysKeyUp:
		activeManager.handlePTT(vk, false)
	}

	r, _, _ := procCallNextHookEx.Call(0, uintptr(nCode), wParam, uintptr(lParam))
	return r
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
		// On keydown require the configured modifiers to be held.
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
		// keyup: only emit Release if we previously emitted Press.
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

// currentMods reads the current modifier-key state via GetKeyState.
func currentMods() uint32 {
	var mods uint32
	if isDown(vkLControl) || isDown(vkRControl) {
		mods |= modControl
	}
	if isDown(vkLAlt) || isDown(vkRAlt) {
		mods |= modAlt
	}
	if isDown(vkLShift) || isDown(vkRShift) {
		mods |= modShift
	}
	if isDown(vkLWin) || isDown(vkRWin) {
		mods |= modWin
	}
	return mods
}

func isDown(vk uint32) bool {
	r, _, _ := procGetKeyState.Call(uintptr(vk))
	return int16(r) < 0 // high bit = pressed
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

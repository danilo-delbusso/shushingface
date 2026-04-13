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
	modAlt     = 0x0001
	modControl = 0x0002
	modShift   = 0x0004
	modWin     = 0x0008
	modNoRepeat = 0x4000

	wmHotkey   = 0x0312
	wmClose    = 0x0010
	wmApp      = 0x8000
	wmRegister   = wmApp + 1
	wmUnregister = wmApp + 2
)

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey    = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey  = user32.NewProc("UnregisterHotKey")
	procGetMessageW       = user32.NewProc("GetMessageW")
	procDispatchMessageW  = user32.NewProc("DispatchMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procPostThreadMessage = user32.NewProc("PostThreadMessageW")
	kernel32              = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")
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
	ids      map[string]uint16
	next     uint16
	threadID uint32
	events   chan string
	regCh    chan regReq
	unregCh  chan unregReq
	done     chan struct{}
}

// New creates a Windows hotkey manager. The returned manager runs its own
// OS-thread-pinned message loop; Close to tear it down.
func New() Manager {
	m := &manager{
		ids:     map[string]uint16{},
		next:    1,
		events:  make(chan string, 16),
		regCh:   make(chan regReq),
		unregCh: make(chan unregReq),
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

func (m *manager) loop(ready chan<- struct{}) {
	// OS thread lock is required: RegisterHotKey is thread-affine.
	runtimeLockOSThread()
	defer runtimeUnlockOSThread()

	tid, _, _ := procGetCurrentThreadID.Call()
	m.threadID = uint32(tid)
	close(ready)

	// Pumping loop: drain requests and dispatch WM_HOTKEY messages.
	var message msg
	pending := map[uint16]string{}

	for {
		// Handle any channel requests by posting a wakeup before blocking on GetMessage.
		select {
		case req := <-m.regCh:
			req.resp <- m.doRegister(req.id, req.mods, req.vk)
			continue
		case req := <-m.unregCh:
			req.resp <- m.doUnregister(req.id)
			continue
		case <-m.done:
			return
		default:
		}

		r, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&message)),
			0, 0, 0,
		)
		if int32(r) <= 0 {
			return
		}
		switch message.Message {
		case wmHotkey:
			id := uint16(message.WParam)
			m.mu.Lock()
			name := pending[id]
			if name == "" {
				for k, v := range m.ids {
					if v == id {
						name = k
						pending[id] = k
						break
					}
				}
			}
			m.mu.Unlock()
			if name != "" {
				select {
				case m.events <- name:
				default:
					slog.Warn("hotkey: event channel full, dropping", "id", name)
				}
			}
		case wmRegister, wmUnregister:
			// No-op: handled through the select above after PostThreadMessage wakes us.
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
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

// wake posts a thread message so GetMessage returns and the select fires.
func (m *manager) wake(code uint32) {
	procPostThreadMessage.Call(uintptr(m.threadID), uintptr(code), 0, 0)
}

func (m *manager) Register(name string, spec Spec) error {
	vk, ok := vkFromKey(spec.Key)
	if !ok {
		return fmt.Errorf("%w: unsupported key %q", ErrInvalidSpec, spec.Key)
	}
	mods := winMods(spec.Mods)
	if mods == 0 {
		return fmt.Errorf("%w: at least one modifier required", ErrInvalidSpec)
	}

	m.mu.Lock()
	if existing, ok := m.ids[name]; ok {
		m.mu.Unlock()
		resp := make(chan error, 1)
		m.unregCh <- unregReq{id: existing, resp: resp}
		m.wake(wmUnregister)
		if err := <-resp; err != nil {
			slog.Warn("hotkey: stale unregister failed", "error", err)
		}
		m.mu.Lock()
		delete(m.ids, name)
	}
	id := m.next
	m.next++
	m.mu.Unlock()

	resp := make(chan error, 1)
	m.regCh <- regReq{id: id, mods: mods, vk: vk, resp: resp}
	m.wake(wmRegister)
	if err := <-resp; err != nil {
		return err
	}
	m.mu.Lock()
	m.ids[name] = id
	m.mu.Unlock()
	return nil
}

func (m *manager) Unregister(name string) error {
	m.mu.Lock()
	id, ok := m.ids[name]
	if !ok {
		m.mu.Unlock()
		return nil
	}
	delete(m.ids, name)
	m.mu.Unlock()
	resp := make(chan error, 1)
	m.unregCh <- unregReq{id: id, resp: resp}
	m.wake(wmUnregister)
	return <-resp
}

func (m *manager) Events() <-chan string { return m.events }

func (m *manager) Close() error {
	m.mu.Lock()
	ids := make([]uint16, 0, len(m.ids))
	for _, id := range m.ids {
		ids = append(ids, id)
	}
	m.ids = map[string]uint16{}
	m.mu.Unlock()

	for _, id := range ids {
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

// vkFromKey returns the Win32 virtual key code for a canonical key name.
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
	// F1–F24
	if len(key) >= 2 && (key[0] == 'F' || key[0] == 'f') {
		var n int
		if _, err := fmt.Sscanf(key[1:], "%d", &n); err == nil && n >= 1 && n <= 24 {
			return uint32(0x70 + n - 1), true
		}
	}
	return 0, false
}


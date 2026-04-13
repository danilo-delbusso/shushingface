//go:build windows

// Package win32 wraps the raw Win32 syscalls used across the app into one
// place so no other package needs to declare DLL handles or proc pointers.
// Feature packages import wrappers from here rather than re-loading user32
// etc. themselves.
package win32

import "golang.org/x/sys/windows"

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	gdi32    = windows.NewLazySystemDLL("gdi32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
)

// User32 returns the lazy-loaded user32.dll handle for callers that need
// to declare procs this package does not yet wrap (e.g. overlay's niche
// paint calls). Prefer adding a typed wrapper here over exposing raw procs.
func User32() *windows.LazyDLL { return user32 }

// GDI32 returns the lazy-loaded gdi32.dll handle.
func GDI32() *windows.LazyDLL { return gdi32 }

// Kernel32 returns the lazy-loaded kernel32.dll handle.
func Kernel32() *windows.LazyDLL { return kernel32 }

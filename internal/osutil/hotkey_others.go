//go:build !linux

package osutil

import "golang.design/x/hotkey"

var modMap = map[string]hotkey.Modifier{
	"CTRL":  hotkey.ModCtrl,
	"SHIFT": hotkey.ModShift,
	"ALT":   hotkey.ModOption,
	"CMD":   hotkey.ModCmd,
	"WIN":   hotkey.ModCmd,
	"SUPER": hotkey.ModCmd,
}

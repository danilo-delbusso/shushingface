//go:build linux

package osutil

import "golang.design/x/hotkey"

var modMap = map[string]hotkey.Modifier{
	"CTRL":  hotkey.ModCtrl,
	"SHIFT": hotkey.ModShift,
	"ALT":   hotkey.Mod1,
	"CMD":   hotkey.Mod4,
	"WIN":   hotkey.Mod4,
	"SUPER": hotkey.Mod4,
}

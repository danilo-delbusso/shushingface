package osutil

import (
	"fmt"
	"strings"

	"golang.design/x/hotkey"
)

// RegisterHotkey parses a string like "Ctrl+Shift+R" and registers it globally.
// It returns a channel that emits whenever the hotkey is pressed.
func RegisterHotkey(hotkeyStr string) (<-chan bool, error) {
	modifiers, key, err := parseHotkey(hotkeyStr)
	if err != nil {
		return nil, err
	}

	hk := hotkey.New(modifiers, key)
	if err := hk.Register(); err != nil {
		return nil, fmt.Errorf("failed to register hotkey: %v", err)
	}

	triggerCh := make(chan bool)
	go func() {
		for {
			<-hk.Keydown()
			triggerCh <- true
		}
	}()

	return triggerCh, nil
}

// parseHotkey converts a string like "Ctrl+Shift+R" into hotkey.Modifier and hotkey.Key types.
func parseHotkey(s string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(strings.ToUpper(s), "+")
	if len(parts) == 0 {
		return nil, 0, fmt.Errorf("invalid hotkey format")
	}

	var modifiers []hotkey.Modifier
	var key hotkey.Key

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if i == len(parts)-1 {
			// The last part should be the main key
			k, ok := keyMap[part]
			if !ok {
				return nil, 0, fmt.Errorf("unsupported key: %s", part)
			}
			key = k
		} else {
			// Previous parts should be modifiers
			m, ok := modMap[part]
			if !ok {
				return nil, 0, fmt.Errorf("unsupported modifier: %s", part)
			}
			modifiers = append(modifiers, m)
		}
	}

	return modifiers, key, nil
}

// keyMap provides a cross-platform mapping for common keys.
var keyMap = map[string]hotkey.Key{
	"A": hotkey.KeyA, "B": hotkey.KeyB, "C": hotkey.KeyC, "D": hotkey.KeyD,
	"E": hotkey.KeyE, "F": hotkey.KeyF, "G": hotkey.KeyG, "H": hotkey.KeyH,
	"I": hotkey.KeyI, "J": hotkey.KeyJ, "K": hotkey.KeyK, "L": hotkey.KeyL,
	"M": hotkey.KeyM, "N": hotkey.KeyN, "O": hotkey.KeyO, "P": hotkey.KeyP,
	"Q": hotkey.KeyQ, "R": hotkey.KeyR, "S": hotkey.KeyS, "T": hotkey.KeyT,
	"U": hotkey.KeyU, "V": hotkey.KeyV, "W": hotkey.KeyW, "X": hotkey.KeyX,
	"Y": hotkey.KeyY, "Z": hotkey.KeyZ,
	"0": hotkey.Key0, "1": hotkey.Key1, "2": hotkey.Key2, "3": hotkey.Key3,
	"4": hotkey.Key4, "5": hotkey.Key5, "6": hotkey.Key6, "7": hotkey.Key7,
	"8": hotkey.Key8, "9": hotkey.Key9,
	"SPACE": hotkey.KeySpace, "ENTER": hotkey.KeyReturn, "ESC": hotkey.KeyEscape,
}

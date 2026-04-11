//go:build !linux

package paste

// Type is not yet implemented on this platform.
// TODO: macOS — NSPasteboard + CGEventCreateKeyboardEvent
// TODO: Windows — clipboard API + SendInput
func Type(text string) error {
	return nil
}

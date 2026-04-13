package platform

// Capability is how platform-specific features report whether they work
// on the current OS. Keep it tiny — a bool and a user-facing reason string
// is enough to (1) short-circuit code paths and (2) let the UI grey out
// toggles for features the current platform can't run.
//
// Convention: when Supported is false, Reason is a short human-readable
// sentence that can be shown verbatim in tooltips or warning banners.
// When Supported is true, Reason is usually empty.
type Capability struct {
	Supported bool   `json:"supported"`
	Reason    string `json:"reason,omitempty"`
}

// Supported returns a Capability marking the feature as available on this
// platform. Prefer this constructor over literal struct construction so the
// intent reads clearly at the call site.
func Supported() Capability { return Capability{Supported: true} }

// Unsupported returns a Capability marking the feature as unavailable with
// the given user-facing reason. The reason should finish a sentence like
// "This feature is unavailable because …".
func Unsupported(reason string) Capability {
	return Capability{Supported: false, Reason: reason}
}

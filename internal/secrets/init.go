package secrets

import "log/slog"

// New creates the best available secret store.
// Tries the OS keyring first, falls back to in-memory/config storage.
func New(opts ...FallbackOption) Store {
	if s := NewKeyringStore(); s != nil {
		slog.Info("using OS keyring for secret storage")
		return s
	}
	slog.Info("using config file for secret storage (OS keyring not available)")
	return NewFallbackStore(opts...)
}

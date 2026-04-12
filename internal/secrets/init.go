package secrets

import "log/slog"

func New(opts ...FallbackOption) Store {
	if s := NewKeyringStore(); s != nil {
		slog.Info("using OS keyring for secret storage")
		return s
	}
	slog.Info("using config file for secret storage (OS keyring not available)")
	return NewFallbackStore(opts...)
}

package secrets

import "sync"

// fallbackStore stores secrets in memory, backed by the config file.
// Used when the OS keyring is not available. Not encrypted, but still
// stored in the user's config directory with 0600 permissions.
type fallbackStore struct {
	mu     sync.RWMutex
	values map[string]string
	onSave func(key, value string) // called on Set to persist to config
}

// FallbackOption configures the fallback store.
type FallbackOption func(*fallbackStore)

// WithValues pre-loads secrets (e.g. from config file on startup).
func WithValues(values map[string]string) FallbackOption {
	return func(f *fallbackStore) {
		for k, v := range values {
			f.values[k] = v
		}
	}
}

// WithSaveFunc sets a callback that persists secrets on Set/Delete.
func WithSaveFunc(fn func(key, value string)) FallbackOption {
	return func(f *fallbackStore) {
		f.onSave = fn
	}
}

// NewFallbackStore creates an in-memory secret store for when the keyring
// is unavailable. Pre-load values from config with WithValues.
func NewFallbackStore(opts ...FallbackOption) Store {
	f := &fallbackStore{values: make(map[string]string)}
	for _, o := range opts {
		o(f)
	}
	return f
}

func (f *fallbackStore) Get(key string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	v, ok := f.values[key]
	if !ok || v == "" {
		return "", ErrNotFound
	}
	return v, nil
}

func (f *fallbackStore) Set(key, value string) error {
	f.mu.Lock()
	f.values[key] = value
	f.mu.Unlock()
	if f.onSave != nil {
		f.onSave(key, value)
	}
	return nil
}

func (f *fallbackStore) Delete(key string) error {
	f.mu.Lock()
	delete(f.values, key)
	f.mu.Unlock()
	if f.onSave != nil {
		f.onSave(key, "")
	}
	return nil
}

func (f *fallbackStore) IsSecure() bool { return false }

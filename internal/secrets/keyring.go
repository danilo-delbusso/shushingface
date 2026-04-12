package secrets

import (
	"errors"
	"log/slog"

	"github.com/zalando/go-keyring"
)

const serviceName = "shushingface"

// keyringStore uses the OS keyring (GNOME Keyring, macOS Keychain, Windows Credential Manager).
type keyringStore struct{}

func (k *keyringStore) Get(key string) (string, error) {
	val, err := keyring.Get(serviceName, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return val, err
}

func (k *keyringStore) Set(key, value string) error {
	return keyring.Set(serviceName, key, value)
}

func (k *keyringStore) Delete(key string) error {
	err := keyring.Delete(serviceName, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func (k *keyringStore) IsSecure() bool { return true }

// NewKeyringStore creates a Store backed by the OS keyring.
// Returns nil if the keyring is not available.
func NewKeyringStore() Store {
	// Test if the keyring is usable by writing and deleting a probe key
	if err := keyring.Set(serviceName, "__probe__", "test"); err != nil {
		slog.Debug("OS keyring not available, will use fallback", "error", err)
		return nil
	}
	keyring.Delete(serviceName, "__probe__")
	return &keyringStore{}
}

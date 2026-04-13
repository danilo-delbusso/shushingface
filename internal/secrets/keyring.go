package secrets

import (
	"errors"
	"log/slog"

	"github.com/zalando/go-keyring"
)

const serviceName = "shushingface"

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

func NewKeyringStore() Store {
	// Test if the keyring is usable by writing and deleting a probe key
	if err := keyring.Set(serviceName, "__probe__", "test"); err != nil {
		slog.Debug("OS keyring not available, will use fallback", "error", err)
		return nil
	}
	if err := keyring.Delete(serviceName, "__probe__"); err != nil {
		slog.Warn("failed to clean up keyring probe key", "error", err)
	}
	return &keyringStore{}
}

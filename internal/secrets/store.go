// Package secrets provides a cross-platform interface for storing sensitive
// values (API keys, tokens) securely. It uses the OS keyring when available
// and falls back to plaintext config storage.
package secrets

import "errors"

// ErrNotFound is returned when a secret doesn't exist in the store.
var ErrNotFound = errors.New("secret not found")

// Store is the interface for reading and writing secrets.
// Implementations exist for OS keyrings and plaintext fallback.
type Store interface {
	// Get retrieves a secret by key. Returns ErrNotFound if it doesn't exist.
	Get(key string) (string, error)
	// Set stores a secret.
	Set(key, value string) error
	// Delete removes a secret. No error if it doesn't exist.
	Delete(key string) error
	// IsSecure returns true if the store uses OS-level encryption (keyring),
	// false if it falls back to plaintext.
	IsSecure() bool
}

package secrets

import "errors"

var ErrNotFound = errors.New("secret not found")

type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
	IsSecure() bool
}

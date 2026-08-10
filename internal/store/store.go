package store

import "errors"

// ErrNotFound is returned when a key does not exist.
var ErrNotFound = errors.New("store: not found")

// Store defines the key-value operations required by the application.
type Store interface {
	Get(key []byte) ([]byte, error)
	Put(key, value []byte) error
	Delete(key []byte) error
	DeleteMany(keys [][]byte) error
	Iterate(prefix []byte, fn func(key, value []byte) error) error
	Close() error
}

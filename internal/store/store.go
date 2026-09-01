package store

import (
	"github.com/shurco/mycart/pkg/storage"
)

var store storage.Storage

// New initializes the storage backend and returns an error if any occurs.
func New(s storage.Storage) error {
	store = s
	return nil
}

// Store returns the storage instance. If storage is not initialized, returns nil.
// Use New() to initialize storage before calling Store().
func Store() storage.Storage {
	return store
}

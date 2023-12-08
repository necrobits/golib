package main

import (
	"github.com/necrobits/x/kvstore"
)

var _ kvstore.KvStore = (*MemStore)(nil)

type MemStore struct {
	data    map[string]kvstore.Data
	txId    int
	changes map[string]kvstore.Data
}

// GetAll implements kvstore.KvStore.
func (store *MemStore) GetAll() (map[string]kvstore.Data, error) {
	return store.data, nil
}

// DeleteMany implements kvstore.KvStore.
func (store *MemStore) DeleteMany(keys []string) error {
	for _, k := range keys {
		delete(store.data, k)
	}
	return nil
}

// SetMany implements kvstore.KvStore.
func (store *MemStore) SetMany(data map[string]kvstore.Data) error {
	for k, v := range data {
		store.data[k] = v
	}
	return nil
}

// Delete implements kvstore.KvStore.
func (store *MemStore) Delete(key string) error {
	delete(store.data, key)
	return nil
}

// Get implements kvstore.KvStore.
func (store *MemStore) Get(key string) (kvstore.Data, error) {
	return store.data[key], nil
}

// Set implements kvstore.KvStore.
func (store *MemStore) Set(key string, value kvstore.Data) error {
	if store.txId > 0 {
		// old value is stored in changes
		store.changes[key] = store.data[key]
	}
	store.data[key] = value
	return nil
}

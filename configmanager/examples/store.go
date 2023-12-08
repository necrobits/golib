package main

import (
	"sync"

	"github.com/necrobits/x/kvstore"
)

var _ kvstore.KvStore = (*MemStore)(nil)

type MemStore struct {
	mu   sync.RWMutex
	data map[string]kvstore.Data
}

func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[string]kvstore.Data),
	}
}

// GetAll implements kvstore.KvStore.
func (store *MemStore) GetAll() (map[string]kvstore.Data, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.data, nil
}

// DeleteMany implements kvstore.KvStore.
func (store *MemStore) DeleteMany(keys []string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	for _, k := range keys {
		delete(store.data, k)
	}
	return nil
}

// SetMany implements kvstore.KvStore.
func (store *MemStore) SetMany(data map[string]kvstore.Data) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	for k, v := range data {
		store.data[k] = v
	}
	return nil
}

// Delete implements kvstore.KvStore.
func (store *MemStore) Delete(key string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.data, key)
	return nil
}

// Get implements kvstore.KvStore.
func (store *MemStore) Get(key string) (kvstore.Data, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.data[key], nil
}

// Set implements kvstore.KvStore.
func (store *MemStore) Set(key string, value kvstore.Data) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.data[key] = value
	return nil
}

package main

import (
	"fmt"

	"github.com/necrobits/x/kvstore"
)

var _ kvstore.KvStore = (*MemStore)(nil)

type MemStore struct {
	Data    map[string]kvstore.Data
	txId    int
	changes map[string]kvstore.Data
}

// Delete implements kvstore.KvStore.
func (store *MemStore) Delete(key string) error {
	delete(store.Data, key)
	return nil
}

// Get implements kvstore.KvStore.
func (store *MemStore) Get(key string) (kvstore.Data, error) {
	return store.Data[key], nil
}

// Set implements kvstore.KvStore.
func (store *MemStore) Set(key string, value kvstore.Data) error {
	if store.txId > 0 {
		// old value is stored in changes
		store.changes[key] = store.Data[key]
	}
	store.Data[key] = value
	return nil
}

// Transaction implements kvstore.KvStore.
func (store *MemStore) Transaction(fn func(kvstore.KvStore) error) error {
	tx := &MemStore{
		Data:    store.Data,
		txId:    store.txId + 1,
		changes: make(map[string]kvstore.Data),
	}
	err := fn(tx)
	if err != nil {
		fmt.Printf("Transaction %d failed: %s, rolling back\n", tx.txId, err)
		for key, value := range tx.changes {
			store.Data[key] = value
			if value == nil {
				delete(store.Data, key)
			}
		}
	}
	return nil
}

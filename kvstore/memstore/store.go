package memstore

import (
	"context"
	"sync"

	"github.com/necrobits/x/errors"
	"github.com/necrobits/x/kvstore"
)

var _ kvstore.KvStore = &Store{}

var ErrKeyNotFound = "key_not_found"

type Store struct {
	mu   sync.RWMutex
	data map[string]kvstore.Data
}

func New() *Store {
	return &Store{
		data: make(map[string]kvstore.Data),
	}
}

// Delete implements kvstore.KvStore.
func (s *Store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// Get implements kvstore.KvStore.
func (s *Store) Get(ctx context.Context, key string) (kvstore.Data, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return nil, errors.B().
			Code(ErrKeyNotFound).
			Op("memstore.Get").
			Msgf("key %s not found", key).Build()
	}

	return data, nil
}

// Set implements kvstore.KvStore.
func (s *Store) Set(ctx context.Context, key string, value kvstore.Data) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Transaction implements kvstore.KvStore.
func (s *Store) Transaction(ctx context.Context, fn func(tx kvstore.KvStore) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clonedData := make(map[string]kvstore.Data)
	for k, v := range s.data {
		clonedData[k] = v
	}

	tx := &Store{
		data: clonedData,
	}

	err := fn(tx)
	// Rollback
	if err != nil {
		return err
	}
	// Commit
	s.data = tx.data

	return nil
}

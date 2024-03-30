package memstore

import (
	"context"
	"reflect"
	"sync"

	"github.com/necrobits/x/errors"
	"github.com/necrobits/x/kvstore"
)

var _ kvstore.KvStore = &store{}

var (
	ErrKeyNotFound = "key_not_found"
)

type store struct {
	mu   sync.RWMutex
	data map[string]any
}

func New() *store {
	return &store{
		data: make(map[string]any),
	}
}

// Has implements kvstore.KvStore.
func (s *store) Has(ctx context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok, nil
}

// DeleteAll implements kvstore.KvStore.
func (s *store) DeleteAll(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
	return nil
}

// DeleteMany implements kvstore.KvStore.
func (s *store) DeleteMany(ctx context.Context, keys []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, key := range keys {
		delete(s.data, key)
	}
	return nil
}

// SetMany implements kvstore.KvStore.
func (s *store) SetMany(ctx context.Context, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range data {
		s.data[k] = v
	}
	return nil
}

// GetAll implements kvstore.KvStore.
func (s *store) GetAll(ctx context.Context, values map[string]any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k, v := range s.data {
		values[k] = v
	}

	return nil
}

// GetMany implements kvstore.KvStore.
func (s *store) GetMany(ctx context.Context, keys []string, values map[string]any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, key := range keys {
		v, ok := s.data[key]
		if ok {
			values[key] = v
		}
	}

	return nil
}

// Delete implements kvstore.KvStore.
func (s *store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// Get implements kvstore.KvStore.
func (s *store) Get(ctx context.Context, key string, value any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return errors.B().
			Code(ErrKeyNotFound).
			Op("memstore.Get").
			Msgf("key %s not found", key).Build()
	}

	// copy data to value using reflection
	v := reflect.ValueOf(value)
	v.Elem().Set(reflect.ValueOf(data))
	return nil
}

// Set implements kvstore.KvStore.
func (s *store) Set(ctx context.Context, key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return nil
}

// Transaction implements kvstore.KvStore.
func (s *store) Transaction(ctx context.Context, fn func(tx kvstore.KvStore) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clonedData := make(map[string]any)
	for k, v := range s.data {
		clonedData[k] = v
	}

	tx := &store{
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

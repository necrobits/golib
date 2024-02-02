package memstore

import (
	"context"
	"testing"

	"github.com/necrobits/x/errors"
	"github.com/necrobits/x/kvstore"
)

func TestGet(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	t.Run("found", func(t *testing.T) {
		data, err := store.Get(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data != "value1" {
			t.Errorf("unexpected data: %v", data)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := store.Get(context.Background(), "key3")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, kvstore.ErrKeyNotFound) {
			t.Errorf("unexpected error %v", err)
		}
	})
}

func TestGetAll(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	data, err := store.GetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data["key1"] != "value1" {
		t.Errorf("unexpected data: %v", data["key1"])
	}
	if data["key2"] != "value2" {
		t.Errorf("unexpected data: %v", data["key2"])
	}
}

func TestGetMany(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	data, err := store.GetMany(context.Background(), []string{"key1", "key2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data["key1"] != "value1" {
		t.Errorf("unexpected data: %v", data["key1"])
	}
	if data["key2"] != "value2" {
		t.Errorf("unexpected data: %v", data["key2"])
	}
}

func TestSet(t *testing.T) {
	store := &store{
		data: make(map[string]any),
	}

	err := store.Set(context.Background(), "key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.data["key1"] != "value1" {
		t.Errorf("unexpected data: %v", store.data["key1"])
	}
}

func TestSetMany(t *testing.T) {
	store := &store{
		data: make(map[string]any),
	}

	err := store.SetMany(context.Background(), map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.data["key1"] != "value1" {
		t.Errorf("unexpected data: %v", store.data["key1"])
	}
	if store.data["key2"] != "value2" {
		t.Errorf("unexpected data: %v", store.data["key2"])
	}
}

func TestDelete(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err := store.Delete(context.Background(), "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := store.data["key1"]; ok {
		t.Errorf("unexpected data: %v", store.data["key1"])
	}
}

func TestDeleteMany(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err := store.DeleteMany(context.Background(), []string{"key1", "key2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.data) != 0 {
		t.Errorf("unexpected data: %v", store.data)
	}
}

func TestDeleteAll(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err := store.DeleteAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.data) != 0 {
		t.Errorf("unexpected data: %v", store.data)
	}
}

func TestTransaction(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err := store.Transaction(context.Background(), func(tx kvstore.KvStore) error {
		err := tx.Set(context.Background(), "key3", "value3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err = tx.Delete(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := store.data["key1"]; ok {
		t.Errorf("unexpected data: %v", store.data["key1"])
	}
	if store.data["key3"] != "value3" {
		t.Errorf("unexpected data: %v", store.data["key3"])
	}
}

func TestTransactionFailed(t *testing.T) {
	store := &store{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	err := store.Transaction(context.Background(), func(tx kvstore.KvStore) error {
		err := tx.Set(context.Background(), "key3", "value3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		err = tx.Delete(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return errors.B().Msg("test error").Build()
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if _, ok := store.data["key1"]; !ok {
		t.Errorf("unexpected data: %v", store.data["key1"])
	}
	if _, ok := store.data["key3"]; ok {
		t.Errorf("unexpected data: %v", store.data["key3"])
	}
}

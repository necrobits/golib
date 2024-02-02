package configmanager

import (
	"context"
	"testing"

	"github.com/necrobits/x/errors"
	"github.com/necrobits/x/kvstore"
	"github.com/necrobits/x/kvstore/memstore"
)

func simpleManager() *Manager {
	store := memstore.New()
	ctx := context.Background()
	store.Set(ctx, "key1", "value1")
	store.Set(ctx, "key2", "value2")

	return NewManager(store)
}

func TestGet(t *testing.T) {
	m := simpleManager()

	t.Run("found", func(t *testing.T) {
		cfg, err := m.Get(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != "value1" {
			t.Errorf("unexpected data: %v", cfg)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := m.Get(context.Background(), "key3")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, kvstore.ErrKeyNotFound) {
			t.Errorf("unexpected error %v", err)
		}
	})
}

func TestSubscribe(t *testing.T) {
	m := simpleManager()

	ch1 := m.Subscribe("key1")
	ch2 := m.Subscribe("key1")

	m.eb.Publish("key1", "value1")

	cfg := <-ch1
	if cfg != "value1" {
		t.Errorf("unexpected data: %v", cfg)
	}

	cfg = <-ch2
	if cfg != "value1" {
		t.Errorf("unexpected data: %v", cfg)
	}
}

func TestUpdateOne(t *testing.T) {
	m := simpleManager()

	err := m.UpdateOne(context.Background(), "key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := m.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != "value1" {
		t.Errorf("unexpected data: %v", cfg)
	}
}

func TestUpdateOneWithValidator(t *testing.T) {
	m := simpleManager()
	m.RegisterValidator("key1", func(cfg Config) error {
		if cfg != "value1" {
			return errors.B().Msg("invalid config").Build()
		}
		return nil
	})

	t.Run("valid", func(t *testing.T) {
		err := m.UpdateOne(context.Background(), "key1", "value1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cfg, err := m.Get(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != "value1" {
			t.Errorf("unexpected data: %v", cfg)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		err := m.UpdateOne(context.Background(), "key1", "value2")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		cfg, err := m.Get(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != "value1" {
			t.Errorf("unexpected data: %v", cfg)
		}
	})
}

func TestUpdateMany(t *testing.T) {
	m := simpleManager()

	err := m.UpdateMany(context.Background(), map[string]Config{
		"key1": "value1",
		"key2": "value2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := m.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != "value1" {
		t.Errorf("unexpected data: %v", cfg)
	}

	cfg, err = m.Get(context.Background(), "key2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != "value2" {
		t.Errorf("unexpected data: %v", cfg)
	}
}

func TestUpdateManyWithValidator(t *testing.T) {
	m := simpleManager()
	m.RegisterValidator("key1", func(cfg Config) error {
		if cfg != "value1" {
			return errors.B().Msg("invalid config").Build()
		}
		return nil
	})
	m.RegisterValidator("key2", func(cfg Config) error {
		if cfg != "value2" {
			return errors.B().Msg("invalid config").Build()
		}
		return nil
	})

	t.Run("valid", func(t *testing.T) {
		err := m.UpdateMany(context.Background(), map[string]Config{
			"key1": "value1",
			"key2": "value2",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cfg, err := m.Get(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != "value1" {
			t.Errorf("unexpected data: %v", cfg)
		}

		cfg, err = m.Get(context.Background(), "key2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != "value2" {
			t.Errorf("unexpected data: %v", cfg)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		err := m.UpdateMany(context.Background(), map[string]Config{
			"key1": "value1",
			"key2": "value1",
		})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}

		cfg, err := m.Get(context.Background(), "key1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != "value1" {
			t.Errorf("unexpected data: %v", cfg)
		}

		cfg, err = m.Get(context.Background(), "key2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != "value2" {
			t.Errorf("unexpected data: %v", cfg)
		}
	})
}

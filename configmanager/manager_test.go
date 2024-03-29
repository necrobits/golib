package configmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/necrobits/x/errors"
	"github.com/necrobits/x/kvstore/memstore"
	"github.com/stretchr/testify/require"
)

func simpleManager(initData map[string]json.RawMessage) *Manager {
	ctx := context.Background()
	store := memstore.New()
	for k, v := range initData {
		err := store.Set(ctx, k, v)
		if err != nil {
			panic(err)
		}
	}
	return NewManager(store)
}

func TestManagerGet(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		m := &Manager{
			cfgs: make(map[string]Config),
		}
		_, err := m.Get(context.Background(), "test")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrKeyNotRegistered) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("Found", func(t *testing.T) {
		cfg := &struct{}{}
		m := &Manager{
			cfgs: map[string]Config{
				"test": cfg,
			},
		}
		actualCfg, err := m.Get(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		require.Equal(t, cfg, actualCfg)
	})
}

func TestManagerGetAll(t *testing.T) {
	m := &Manager{
		cfgs: map[string]Config{
			"test1": &struct{}{},
			"test2": &struct{}{},
		},
	}
	cfgs, err := m.GetAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	require.Equal(t, m.cfgs, cfgs)
}

func TestManagerValidator(t *testing.T) {
	type TestConfig struct {
		StrField string `json:"str_field"`
	}
	validator := func(cfg Config) error {
		testCfg, ok := cfg.(*TestConfig)
		if !ok {
			return fmt.Errorf("unexpected config type: %T", cfg)
		}
		if testCfg.StrField != "test" {
			return fmt.Errorf("unexpected str field: %s", testCfg.StrField)
		}
		return nil
	}

	t.Run("RegisterValidator_WithoutConfigRegistration", func(t *testing.T) {
		m := NewManager(memstore.New())
		err := m.RegisterValidator("test", validator)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrKeyNotRegistered) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("RegisterValidator_WithConfigRegistration", func(t *testing.T) {
		m := NewManager(memstore.New())

		err := m.RegisterConfig("test", &TestConfig{})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.RegisterValidator("test", validator)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("ValidateAll_Valid", func(t *testing.T) {
		m := NewManager(memstore.New())
		err := m.RegisterConfig("test", &TestConfig{
			StrField: "test",
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.RegisterValidator("test", validator)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.ValidateAll()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("ValidateAll_Invalid", func(t *testing.T) {
		m := NewManager(memstore.New())
		err := m.RegisterConfig("test", &TestConfig{
			StrField: "invalid",
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.RegisterValidator("test", validator)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.ValidateAll()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestManagerRegisterConfig(t *testing.T) {
	type TestConfig struct {
		StrField  string          `json:"str_field"`
		IntField  int             `json:"int_field"`
		JsonField json.RawMessage `json:"json_field"`
	}
	type JsonField struct {
		Test string `json:"test"`
	}

	t.Run("WithoutStoreData_WithoutDefault", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.RegisterConfig("test", &TestConfig{})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		cfg, err := m.Get(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if cfg == nil {
			t.Fatalf("config is nil")
		}
	})

	t.Run("WithoutStoreData_WithDefault", func(t *testing.T) {
		defaultCfg := &TestConfig{
			StrField: "test",
			IntField: 1,
		}
		m := simpleManager(nil)
		err := m.RegisterConfig("test", defaultCfg)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		cfg, err := m.Get(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		castedCfg, ok := cfg.(*TestConfig)
		if !ok {
			t.Fatalf("unexpected config type: %T", cfg)
		}
		require.Equal(t, defaultCfg, castedCfg)
	})

	t.Run("WithStoreData", func(t *testing.T) {
		jsonField := &JsonField{
			Test: "test",
		}
		jsonFieldBytes, err := json.Marshal(jsonField)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		expectedCfg := &TestConfig{
			StrField:  "test",
			IntField:  1,
			JsonField: jsonFieldBytes,
		}
		expectedCfgBytes, err := json.Marshal(expectedCfg)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		m := simpleManager(map[string]json.RawMessage{
			"test": expectedCfgBytes,
		})
		err = m.RegisterConfig("test", &TestConfig{})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		actualCfg, err := m.Get(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		castedActualCfg, ok := actualCfg.(*TestConfig)
		if !ok {
			t.Fatalf("unexpected config type: %T", actualCfg)
		}
		require.Equal(t, expectedCfg, castedActualCfg)
	})

	t.Run("WithStoreData_InvalidDefault", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.RegisterConfig("test", nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, errors.EUnexpectedDataType) {
			t.Fatalf("unexpected error: %s", err)
		}
	})
}

func TestManagerSubscribeConfig(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		m := simpleManager(nil)
		ch, err := m.SubscribeConfig("test")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrKeyNotRegistered) {
			t.Fatalf("unexpected error: %s", err)
		}
		if ch != nil {
			t.Fatalf("expected nil, got %v", ch)
		}
	})

	t.Run("Found", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.RegisterConfig("test", &struct{}{})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		ch, err := m.SubscribeConfig("test")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if ch == nil {
			t.Fatalf("expected channel, got nil")
		}
	})
}

func TestManagerSubscribeAndUpdateOne(t *testing.T) {
	type TestConfig struct {
		StrField string `json:"str_field"`
	}

	t.Run("InvalidKey", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.UpdateOne(context.Background(), "test", nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrKeyNotRegistered) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("InvalidData_WithoutValidator", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.RegisterConfig("test", &TestConfig{})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.UpdateOne(context.Background(), "test", nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("InvalidData_WithValidator", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.RegisterConfig("test", &TestConfig{})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.RegisterValidator("test", func(cfg Config) error {
			return fmt.Errorf("test error")
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.UpdateOne(context.Background(), "test", nil)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("ValidData_WithoutValidator", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.RegisterConfig("test", &TestConfig{
			StrField: "test",
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		cfgCh, err := m.SubscribeConfig("test")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.UpdateOne(context.Background(), "test", json.RawMessage(`{"str_field":"newTest"}`))
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		cfg := <-cfgCh
		castedCfg, ok := cfg.(*TestConfig)
		if !ok {
			t.Fatalf("unexpected config type: %T", cfg)
		}
		require.Equal(t, "newTest", castedCfg.StrField)
		require.Equal(t, "newTest", m.cfgs["test"].(*TestConfig).StrField)
	})

	t.Run("ValidData_WithValidator", func(t *testing.T) {
		m := simpleManager(nil)
		err := m.RegisterConfig("test", &TestConfig{
			StrField: "test",
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		cfgCh, err := m.SubscribeConfig("test")
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.RegisterValidator("test", func(cfg Config) error {
			testCfg, ok := cfg.(*TestConfig)
			if !ok {
				return fmt.Errorf("unexpected config type: %T", cfg)
			}
			if testCfg.StrField != "newTest" {
				return fmt.Errorf("unexpected str field: %s", testCfg.StrField)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		err = m.UpdateOne(context.Background(), "test", json.RawMessage(`{"str_field":"newTest"}`))
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		cfg := <-cfgCh
		castedCfg, ok := cfg.(*TestConfig)
		if !ok {
			t.Fatalf("unexpected config type: %T", cfg)
		}
		require.Equal(t, "newTest", castedCfg.StrField)
		require.Equal(t, "newTest", m.cfgs["test"].(*TestConfig).StrField)
	})
}

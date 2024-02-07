package configmanager

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/necrobits/x/errors"
	"github.com/necrobits/x/event"
	"github.com/necrobits/x/kvstore"
)

var (
	ErrKeyNotRegistered = "key_not_registered"
)

type Manager struct {
	store      kvstore.KvStore
	eb         *event.EventBus
	validators map[string]ValidateFunc
	cfgs       map[string]Config
}

func NewManager(store kvstore.KvStore) *Manager {
	return &Manager{
		store:      store,
		eb:         event.NewEventBus(),
		validators: make(map[string]ValidateFunc),
		cfgs:       make(map[string]Config),
	}
}

// ValidateAll validates all the registered configs in the manager.
// When a validator is registered for a key, it will be used to validate the config.
// Otherwise, the config will be considered valid.
func (m *Manager) ValidateAll() error {
	for key, cfg := range m.cfgs {
		validator, ok := m.validators[key]
		if ok {
			err := validator(cfg)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) GetAll(ctx context.Context) (map[string]Config, error) {
	return m.cfgs, nil
}

func (m *Manager) Get(ctx context.Context, key string) (Config, error) {
	return m.getConfig(key)
}

func (m *Manager) UpdateOne(ctx context.Context, key string, configData json.RawMessage) error {
	newCfg, err := m.updateConfig(ctx, key, configData, nil)
	if err != nil {
		return err
	}

	m.cfgs[key] = newCfg
	m.eb.Publish(event.Topic(key), newCfg)

	return nil
}

func (m *Manager) UpdateMany(ctx context.Context, configDatas map[string]json.RawMessage) error {
	var newCfgs = make(map[string]Config)

	err := m.store.Transaction(ctx, func(tx kvstore.KvStore) error {
		for key, configData := range configDatas {
			newCfg, err := m.updateConfig(ctx, key, configData, tx)
			if err != nil {
				return err
			}

			newCfgs[key] = newCfg
		}
		return nil
	})
	if err != nil {
		return err
	}

	for key, cfg := range newCfgs {
		m.cfgs[key] = cfg
		m.eb.Publish(event.Topic(key), cfg)
	}

	return nil
}

// RegisterConfig returns a channel that will receive updates for the given key
//
// It also takes a defaultConfig, which will be used when no value is found in the store
// The defaultConfig is also used to determine the type of the Config, so it should not be nil
// or just an empty interface, otherwise an error will be returned
//
// Keep in mind that the defaultConfig is not validated and stored,
// so it's possible to have an invalid config in the manager
// Instead, the config will be validated when it's updated, in case a validator for the key is registered
func (m *Manager) RegisterConfig(key string, defaultConfig Config) (chan Config, error) {
	ch := event.NewEventChannel()
	m.eb.Subscribe(event.Topic(key), ch)

	// check if the defaultConfig is valid type
	rt := reflect.TypeOf(defaultConfig)
	if rt == nil {
		return nil, errors.B().
			Code(errors.EUnexpectedDataType).
			Msg("default config should not be nil").Build()
	}

	ctx := context.Background()
	ok, err := m.store.Has(ctx, key)
	if err != nil {
		return nil, err
	}
	if !ok {
		cfgData, err := json.Marshal(defaultConfig)
		if err != nil {
			return nil, err
		}
		err = m.store.Set(ctx, key, cfgData)
		if err != nil {
			return nil, err
		}
		m.cfgs[key] = defaultConfig
	} else {
		storedCfg, err := m.store.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		castedStoredCfg, ok := storedCfg.(json.RawMessage)
		if !ok {
			return nil, errors.B().
				Code(errors.EMalformedData).
				Msg("stored config is not a valid json").Build()
		}
		cfg := reflect.New(rt).Interface()
		err = json.Unmarshal(castedStoredCfg, cfg)
		if err != nil {
			return nil, err
		}
		cfg = reflect.ValueOf(cfg).Elem().Interface().(Config)

		m.cfgs[key] = cfg
	}

	cfgCh := make(chan Config)
	go func() {
		for e := range ch {
			cfgCh <- e.Data().(Config)
		}
	}()

	return cfgCh, nil
}

func (m *Manager) RegisterValidator(key string, validator ValidateFunc) error {
	_, ok := m.cfgs[key]
	if !ok {
		return errors.B().
			Code(ErrKeyNotRegistered).
			Msgf("key %s not registered", key).Build()
	}
	m.validators[key] = validator
	return nil
}

func (m *Manager) updateConfig(ctx context.Context, key string, newConfigData json.RawMessage, tx kvstore.KvStore) (Config, error) {
	currentCfg, err := m.getConfig(key)
	if err != nil {
		return nil, err
	}

	// Get type of the config and create a new instance of it
	cfgType := reflect.TypeOf(currentCfg)
	newCfg := reflect.New(cfgType).Interface()

	err = json.Unmarshal(newConfigData, newCfg)
	if err != nil {
		return nil, err
	}
	newCfg = reflect.ValueOf(newCfg).Elem().Interface()

	validator, ok := m.validators[key]
	if ok {
		err := validator(newCfg)
		if err != nil {
			return nil, err
		}
	}

	if tx == nil {
		tx = m.store
	}
	err = tx.Set(ctx, key, newConfigData)
	if err != nil {
		return nil, err
	}

	return newCfg.(Config), nil
}

func (m *Manager) getConfig(key string) (Config, error) {
	cfg, ok := m.cfgs[key]
	if !ok {
		return nil, errors.B().
			Code(ErrKeyNotRegistered).
			Msgf("key %s not registered", key).Build()
	}
	return cfg, nil
}

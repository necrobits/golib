package configmanager

import (
	"context"

	"github.com/necrobits/x/event"
	"github.com/necrobits/x/kvstore"
)

type Manager struct {
	store      kvstore.KvStore
	eb         *event.EventBus
	validators map[string]ValidateFunc
}

func NewManager(store kvstore.KvStore) *Manager {
	return &Manager{
		store:      store,
		eb:         event.NewEventBus(),
		validators: make(map[string]ValidateFunc),
	}
}

func (m *Manager) ValidateConfig(key string, cfg Config) error {
	validator, ok := m.validators[key]
	if !ok {
		return nil
	}
	return validator(cfg)
}

func (m *Manager) Get(ctx context.Context, key string) (Config, error) {
	return m.store.Get(ctx, key)
}

func (m *Manager) UpdateOne(ctx context.Context, key string, cfg Config) error {
	validator, ok := m.validators[key]
	if ok {
		err := validator(cfg)
		if err != nil {
			return err
		}
	}

	err := m.store.Set(ctx, key, cfg)
	if err != nil {
		return err
	}

	m.eb.Publish(event.Topic(key), cfg)

	return nil
}

func (m *Manager) UpdateMany(ctx context.Context, cfgs map[string]Config) error {
	var dataMap = make(map[string]kvstore.Data)

	err := m.store.Transaction(ctx, func(tx kvstore.KvStore) error {
		for key, cfg := range cfgs {
			validator, ok := m.validators[key]
			if ok {
				err := validator(cfg)
				if err != nil {
					return err
				}
			}
			dataMap[key] = cfg
			err := tx.Set(ctx, key, cfg)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for key, cfg := range cfgs {
		m.eb.Publish(event.Topic(key), cfg)
	}

	return nil
}

func (m *Manager) Subscribe(key string) chan Config {
	ch := event.NewEventChannel()
	m.eb.Subscribe(event.Topic(key), ch)
	cfgCh := make(chan Config)

	go func() {
		for e := range ch {
			cfgCh <- e.Data().(Config)
		}
	}()

	return cfgCh
}

func (m *Manager) RegisterValidator(key string, validator ValidateFunc) {
	m.validators[key] = validator
}

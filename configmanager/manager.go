package configmanager

import (
	"reflect"
	"strconv"

	"github.com/necrobits/x/event"
)

const defaultTagKey = "cfg"

type Manager struct {
	tagKey  string
	rootCfg Config
	eb      *event.EventBus
}

type ManagerOpts struct {
	RootCfg Config
	TagKey  string
}

func NewManager(opts *ManagerOpts) *Manager {
	tagKey := opts.TagKey
	if tagKey == "" {
		tagKey = defaultTagKey
	}
	return &Manager{
		rootCfg: opts.RootCfg,
		tagKey:  tagKey,
		eb:      event.NewEventBus(),
	}
}

func (m *Manager) RootConfig() Config {
	return m.rootCfg
}

func (m *Manager) Register(cfg RegistrableConfig) event.EventChannel {
	cfgCh := event.NewEventChannel()
	m.eb.Subscribe(cfg.Topic(), cfgCh)
	return cfgCh
}

func (m *Manager) Update(data map[string]interface{}) error {
	var canAddr bool

	dataValue := reflect.ValueOf(convertDotNotationToMap(data))
	eventQueue := make(EventQueue, 0)
	rollbacks := make(RollbackList, 0)

	config := reflect.ValueOf(m.rootCfg)
	if config.Kind() == reflect.Ptr {
		canAddr = true
		config = config.Elem()
	} else {
		canAddr = false
		configPtr := reflect.New(config.Type())
		configPtr.Elem().Set(config)
		config = configPtr.Elem()
	}

	if err := m.updateConfig(&eventQueue, &rollbacks, config, dataValue); err != nil {
		rollbacks.rollback()
		return err
	}

	if !canAddr {
		m.rootCfg = config.Interface()
	}
	for _, event := range eventQueue {
		m.eb.Publish(event.Topic(), event.Data())
	}

	return nil
}

func (m *Manager) updateConfig(eventQueue *EventQueue, rollbacks *RollbackList, cfg reflect.Value, data reflect.Value) error {
	if data.Kind() != reflect.Map {
		oldCfg := reflect.New(cfg.Type()).Elem()
		oldCfg.Set(cfg)
		*rollbacks = append(*rollbacks, Rollback{
			value:    cfg,
			oldValue: oldCfg,
		})
		cfg.Set(data)
		return publish(eventQueue, cfg)
	}

	switch cfg.Kind() {
	case reflect.Struct:
		for i := 0; i < cfg.NumField(); i++ {
			field := cfg.Field(i)
			tag := cfg.Type().Field(i).Tag.Get(m.tagKey)

			_data := data.MapIndex(reflect.ValueOf(tag))
			if _data.IsValid() {
				_data = reflect.ValueOf(_data.Interface())
				if err := m.updateConfig(eventQueue, rollbacks, field, _data); err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		for _, key := range data.MapKeys() {
			_cfg := cfg.MapIndex(key)
			if !_cfg.IsValid() {
				*rollbacks = append(*rollbacks, Rollback{
					value:    cfg,
					key:      key,
					oldValue: reflect.Value{},
				})
				cfg.SetMapIndex(key, reflect.New(cfg.Type().Elem()).Elem())
			}

			canAddr := _cfg.CanAddr()
			if !canAddr {
				_cfg = reflect.New(_cfg.Type()).Elem()
				_cfg.Set(cfg.MapIndex(key))
			}
			cfg.SetMapIndex(key, clone(_cfg))

			_data := reflect.ValueOf(data.MapIndex(key).Interface())
			if err := m.updateConfig(eventQueue, rollbacks, _cfg, _data); err != nil {
				return err
			}
			if !canAddr {
				*rollbacks = append(*rollbacks, Rollback{
					value:    cfg,
					key:      key,
					oldValue: _cfg,
				})
				cfg.SetMapIndex(key, _cfg)
			}

		}
	case reflect.Ptr:
		_cfg := cfg.Elem()
		_cfg.Set(clone(_cfg))
		if err := m.updateConfig(eventQueue, rollbacks, _cfg, data); err != nil {
			return err
		}
	case reflect.Slice:
		for _, key := range data.MapKeys() {
			idx, err := strconv.Atoi(key.String())
			if err != nil {
				return err
			}
			len := cfg.Len()
			if idx == len {
				*rollbacks = append(*rollbacks, Rollback{
					value:    cfg,
					key:      key,
					oldValue: reflect.Value{},
				})
				cfg.Set(reflect.Append(cfg, reflect.New(cfg.Type().Elem()).Elem()))
			}
			_cfg := cfg.Index(idx)
			_cfg.Set(clone(_cfg))

			_data := reflect.ValueOf(data.MapIndex(key).Interface())
			if err := m.updateConfig(eventQueue, rollbacks, _cfg, _data); err != nil {
				return err
			}
		}
	}

	return publish(eventQueue, cfg)
}

func publish(eventQueue *EventQueue, config reflect.Value) error {
	if cfg, ok := config.Interface().(ValidatableConfig); ok {
		if err := cfg.Validate(); err != nil {
			return err
		}
	}
	if cfg, ok := config.Interface().(RegistrableConfig); ok {
		if config.Kind() == reflect.Ptr {
			eventQueue.add(cfg.Topic(), config.Elem().Interface())
		} else {
			eventQueue.add(cfg.Topic(), config.Interface())
		}
	}
	return nil
}

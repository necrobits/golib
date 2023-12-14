package configmanager

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/necrobits/x/event"
	"github.com/necrobits/x/kvstore"
)

const defaultTagKey = "cfg"

type Manager struct {
	store   kvstore.KvStore
	tagKey  string
	rootCfg Config
	eb      *event.EventBus
}

type ManagerOpts struct {
	RootCfg Config
	TagKey  string
	Store   kvstore.KvStore
}

func NewManager(opts *ManagerOpts) (*Manager, error) {
	tagKey := opts.TagKey
	if tagKey == "" {
		tagKey = defaultTagKey
	}
	flattedCfg := dotFlatConfig(opts.RootCfg, tagKey)

	persistedCfg, err := opts.Store.GetAll()
	if err != nil {
		return nil, err
	}
	for k, v := range persistedCfg {
		flattedCfg[k] = v
	}
	manager := &Manager{
		store:   opts.Store,
		rootCfg: opts.RootCfg,
		tagKey:  tagKey,
		eb:      event.NewEventBus(),
	}
	if err := manager.UpdateWithoutValidation(flattedCfg); err != nil {
		return nil, err
	}
	return manager, nil
}

func (m *Manager) RootConfig() Config {
	return m.rootCfg
}

func (m *Manager) DotFlatRootConfig() map[string]interface{} {
	return dotFlatConfig(m.rootCfg, m.tagKey)
}

func (m *Manager) Register(cfg RegistrableConfig) event.EventChannel {
	cfgCh := event.NewEventChannel()
	m.eb.Subscribe(cfg.Topic(), cfgCh)
	return cfgCh
}

func (m *Manager) ValidateConfig() error {
	return m.validate(reflect.ValueOf(m.rootCfg))
}

func (m *Manager) validate(cfg reflect.Value) error {
	if validatableConfig, ok := cfg.Interface().(ValidatableConfig); ok {
		if err := validatableConfig.Validate(); err != nil {
			return err
		}
	}

	switch cfg.Kind() {
	case reflect.Struct:
		for i := 0; i < cfg.NumField(); i++ {
			field := cfg.Field(i)
			if err := m.validate(field); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range cfg.MapKeys() {
			if err := m.validate(cfg.MapIndex(key)); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		return m.validate(cfg.Elem())
	case reflect.Slice:
		for i := 0; i < cfg.Len(); i++ {
			if err := m.validate(cfg.Index(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) UpdateWithoutValidation(data map[string]interface{}) error {
	return m.setConfig(false, data)
}

func (m *Manager) Update(data map[string]interface{}) error {
	return m.setConfig(true, data)
}

func (m *Manager) setConfig(withValidation bool, data map[string]interface{}) error {
	var canAddr bool

	dataValue := reflect.ValueOf(convertDotNotationToMap(data, m.tagKey))
	eventQueue := make(EventQueue, 0)
	rollbacks := make(RollbackList, 0)
	changes := make(map[string]kvstore.Data)
	_data := dataValue.MapIndex(reflect.ValueOf(m.rootCfg.Name()))
	if !_data.IsValid() {
		return fmt.Errorf("invalid config data")
	} else {
		_data = reflect.ValueOf(_data.Interface())
	}

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

	params := &updateConfigParams{
		withValidation: withValidation,
		cfg:            config,
		data:           _data,
		dottedKey:      m.rootCfg.Name(),
		changes:        changes,
		eventQueue:     &eventQueue,
		rollbacks:      &rollbacks,
	}

	if err := m.updateConfig(params); err != nil {
		rollbacks.rollback()
		return err
	}

	err := m.store.SetMany(changes)
	if err != nil {
		rollbacks.rollback()
		return err
	}

	if !canAddr {
		m.rootCfg = config.Interface().(Config)
	}
	for _, event := range eventQueue {
		m.eb.Publish(event.Topic(), event.Data())
	}

	return nil
}

type updateConfigParams struct {
	changes        map[string]kvstore.Data
	withValidation bool
	cfg            reflect.Value
	data           reflect.Value
	dottedKey      string
	eventQueue     *EventQueue
	rollbacks      *RollbackList
}

func (m *Manager) updateConfig(params *updateConfigParams) error {
	data := params.data
	cfg := params.cfg
	eventQueue := params.eventQueue
	rollbacks := params.rollbacks
	dottedKey := params.dottedKey

	if data.Kind() != reflect.Map {
		oldCfg := reflect.New(cfg.Type()).Elem()
		oldCfg.Set(cfg)
		*rollbacks = append(*rollbacks, Rollback{
			value:    cfg,
			oldValue: oldCfg,
		})
		cfgType := cfg.Type()
		castedData := data.Convert(cfgType)
		cfg.Set(castedData)
		params.changes[dottedKey] = kvstore.Data(castedData.Interface())
		return publish(params.withValidation, eventQueue, cfg)
	}

	switch cfg.Kind() {
	case reflect.Struct:
		for i := 0; i < cfg.NumField(); i++ {
			field := cfg.Field(i)
			tag := cfg.Type().Field(i).Tag.Get(m.tagKey)

			_data := data.MapIndex(reflect.ValueOf(tag))
			if _data.IsValid() {
				_data = reflect.ValueOf(_data.Interface())
				params.cfg = field
				params.data = _data
				params.dottedKey = dottedKey + "." + tag
				if err := m.updateConfig(params); err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		for _, key := range data.MapKeys() {
			if cfg.IsNil() {
				*rollbacks = append(*rollbacks, Rollback{
					value:    cfg,
					oldValue: reflect.Value{},
				})
				cfg.Set(reflect.MakeMap(cfg.Type()))
			}
			keyType := cfg.Type().Key()
			castedKey := key.Convert(keyType)
			_cfg := cfg.MapIndex(castedKey)
			if !_cfg.IsValid() {
				*rollbacks = append(*rollbacks, Rollback{
					value:    cfg,
					key:      castedKey,
					oldValue: reflect.Value{},
				})
				cfg.SetMapIndex(castedKey, reflect.New(cfg.Type().Elem()).Elem())
				_cfg = cfg.MapIndex(castedKey)
			}

			canAddr := _cfg.CanAddr()
			if !canAddr {
				_cfg = reflect.New(_cfg.Type()).Elem()
				_cfg.Set(cfg.MapIndex(castedKey))
			}
			cfg.SetMapIndex(castedKey, clone(_cfg))

			_data := reflect.ValueOf(data.MapIndex(key).Interface())
			params.cfg = _cfg
			params.data = _data
			params.dottedKey = dottedKey + "." + castedKey.String()
			if err := m.updateConfig(params); err != nil {
				return err
			}
			if !canAddr {
				*rollbacks = append(*rollbacks, Rollback{
					value:    cfg,
					key:      castedKey,
					oldValue: _cfg,
				})
				cfg.SetMapIndex(castedKey, _cfg)
			}

		}
	case reflect.Ptr:
		if cfg.IsNil() {
			*rollbacks = append(*rollbacks, Rollback{
				value:    cfg,
				oldValue: reflect.Value{},
			})
			cfg.Set(reflect.New(cfg.Type().Elem()))
		}
		_cfg := cfg.Elem()
		_cfg.Set(clone(_cfg))
		params.rollbacks = rollbacks
		params.cfg = _cfg
		if err := m.updateConfig(params); err != nil {
			return err
		}
	case reflect.Slice:
		for _, key := range data.MapKeys() {
			idx, err := strconv.Atoi(key.String())
			if err != nil {
				return err
			}
			if cfg.IsNil() {
				*rollbacks = append(*rollbacks, Rollback{
					value:         cfg,
					sliceAppended: true,
					oldValue:      reflect.Value{},
				})
				cfg.Set(reflect.MakeSlice(cfg.Type(), 0, 0))
			}
			len := cfg.Len()
			if idx == len {
				*rollbacks = append(*rollbacks, Rollback{
					value:         cfg,
					sliceAppended: true,
					oldValue:      reflect.Value{},
				})
				cfg.Set(reflect.Append(cfg, reflect.New(cfg.Type().Elem()).Elem()))
			}
			_cfg := cfg.Index(idx)
			_cfg.Set(clone(_cfg))

			_data := reflect.ValueOf(data.MapIndex(key).Interface())
			params.cfg = _cfg
			params.data = _data
			params.dottedKey = dottedKey + "." + key.String()
			if err := m.updateConfig(params); err != nil {
				return err
			}
		}
	}

	return publish(params.withValidation, eventQueue, cfg)
}

func publish(withValidation bool, eventQueue *EventQueue, config reflect.Value) error {
	if cfg, ok := config.Interface().(ValidatableConfig); withValidation && ok {
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

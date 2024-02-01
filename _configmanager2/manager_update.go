package old_configmanager

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/necrobits/x/errors"
)

var (
	EInvalidConfigData = "invalid config data"
	EInvalidConfigKey  = "invalid config key"
)

type UpdateConfigOpts struct {
	Data           map[string]interface{}
	SortedKeyPaths []string // If this is nil, keys will be sorted by the config manager
}

func (m *Manager) UpdateConfig(opts *UpdateConfigOpts) error {
	if opts.SortedKeyPaths == nil {
		opts.SortedKeyPaths = sortedMapKeys(opts.Data)
	}

	path := m.cfgKey
	rflCfgVal := reflect.ValueOf(m.cfg)

	cloneRflCfgVal := clone(rflCfgVal)

	err := m.updateConfig(&updateConfigOpts{
		data:           opts.Data,
		currentPath:    path,
		rflCfgVal:      cloneRflCfgVal,
		sortedKeyPaths: opts.SortedKeyPaths,
	})
	if err != nil {
		return err
	}

	if rflCfgVal.Kind() == reflect.Ptr {
		m.cfg = cloneRflCfgVal.Addr().Interface().(Config)
	} else {
		m.cfg = cloneRflCfgVal.Interface().(Config)
	}

	return nil
}

type updateConfigOpts struct {
	data           map[string]interface{}
	sortedKeyPaths []string
	currentPath    string
	rflCfgVal      reflect.Value
}

func (m *Manager) updateConfig(opts *updateConfigOpts) error {
	cfg := opts.rflCfgVal

	nodeMng, ok := m.nodes[opts.currentPath]
	if ok {
		nodeUpdateData := make(map[string]interface{})
		nodeSortedKeys := make([]string, 0)

		for _, dataKeyPath := range opts.sortedKeyPaths {
			var newPath string

			currentKey := nodeMng.ConfigKey()
			if dataKeyPath == opts.currentPath {
				newPath = currentKey
			} else {
				newPath = joinPath(currentKey, strings.TrimPrefix(dataKeyPath, opts.currentPath+"."))
			}
			nodeSortedKeys = append(nodeSortedKeys, newPath)
			nodeUpdateData[newPath] = opts.data[dataKeyPath]
		}

		err := nodeMng.UpdateConfig(&UpdateConfigOpts{
			Data:           nodeUpdateData,
			SortedKeyPaths: nodeSortedKeys,
		})
		if err != nil {
			return err
		}
		cfg.Set(reflect.ValueOf(nodeMng.Config()))

		return nil
	}

	// If key is empty, this indicates that the prefix is the last key of key path,
	// that means the current config field can be updated
	key := getNextKey(opts.sortedKeyPaths[0], opts.currentPath)
	if key == "" {
		data := opts.data[opts.currentPath]
		rflData := reflect.ValueOf(data)
		cfgType := cfg.Type()
		if !rflData.CanConvert(cfgType) {
			return errors.B().
				Code(EInvalidConfigData).
				Msgf("expected type %s for key %s, got %s", cfgType, opts.currentPath, rflData.Type()).
				Build()
		}
		rflCastedData := rflData.Convert(cfgType)
		cfg.Set(rflCastedData)

		return nil
	}

	currentPath := joinPath(opts.currentPath, key)
	sortedKeyPaths := make([]string, 0)

	for i, keyPath := range opts.sortedKeyPaths {
		if currentPath == key && strings.HasPrefix(opts.sortedKeyPaths[i+1], currentPath) {
			return errors.B().
				Code(EInvalidConfigKey).
				Msgf("conflicted keys %s and %s", currentPath, opts.sortedKeyPaths[i+1]).
				Build()
		}
		if strings.HasPrefix(keyPath, currentPath) {
			sortedKeyPaths = append(sortedKeyPaths, keyPath)
		}
		if i < len(opts.sortedKeyPaths)-1 && strings.HasPrefix(opts.sortedKeyPaths[i+1], currentPath) {
			continue
		}

		err := m.enterField(key, &updateConfigOpts{
			data:           opts.data,
			currentPath:    currentPath,
			rflCfgVal:      cfg,
			sortedKeyPaths: sortedKeyPaths,
		})
		if err != nil {
			return err
		}

		if i < len(opts.sortedKeyPaths)-1 {
			sortedKeyPaths = make([]string, 0)
			key = getNextKey(opts.sortedKeyPaths[i+1], opts.currentPath)
			currentPath = joinPath(opts.currentPath, key)
		}
	}

	return nil
}

func (m *Manager) enterField(key string, updateOpts *updateConfigOpts) error {
	cfg := updateOpts.rflCfgVal
	data := updateOpts.data
	path := updateOpts.currentPath
	sortedKeys := updateOpts.sortedKeyPaths

	switch cfg.Kind() {
	case reflect.Struct:
		for i := 0; i < cfg.NumField(); i++ {
			field := cfg.Field(i)
			tag := cfg.Type().Field(i).Tag.Get(TagKey)
			if tag == key {
				err := m.updateConfig(&updateConfigOpts{
					data:           data,
					currentPath:    path,
					rflCfgVal:      field,
					sortedKeyPaths: sortedKeys,
				})
				if err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		if cfg.IsNil() {
			cfg.Set(reflect.MakeMap(cfg.Type()))
		}

		cloneCfg := clone(cfg)

		rflKey := reflect.ValueOf(key)
		keyType := cloneCfg.Type().Key()
		rflCastedKey := rflKey.Convert(keyType)
		field := cloneCfg.MapIndex(rflCastedKey)
		if !field.IsValid() {
			cloneCfg.SetMapIndex(rflCastedKey, reflect.New(cloneCfg.Type().Elem()).Elem())
			field = cloneCfg.MapIndex(rflCastedKey)
		}

		canAddr := field.CanAddr()
		if !canAddr {
			field = reflect.New(field.Type()).Elem()
			field.Set(cloneCfg.MapIndex(rflCastedKey))
		}
		cloneCfg.SetMapIndex(rflCastedKey, field)

		err := m.updateConfig(&updateConfigOpts{
			data:           data,
			currentPath:    path,
			rflCfgVal:      field,
			sortedKeyPaths: sortedKeys,
		})
		if err != nil {
			return err
		}

		if !canAddr {
			cloneCfg.SetMapIndex(rflCastedKey, field)
		}

		cfg.Set(cloneCfg)
	case reflect.Slice:
		if cfg.IsNil() {
			cfg.Set(reflect.MakeSlice(cfg.Type(), 0, 0))
		}
		idx, err := strconv.Atoi(key)
		if err != nil {
			return err
		}
		if idx >= cfg.Len() {
			cfg.Set(reflect.Append(cfg, reflect.New(cfg.Type().Elem()).Elem()))
		}

		cloneCfg := clone(cfg)

		field := cloneCfg.Index(idx)

		err = m.updateConfig(&updateConfigOpts{
			data:           data,
			currentPath:    path,
			rflCfgVal:      field,
			sortedKeyPaths: sortedKeys,
		})
		if err != nil {
			return err
		}

		cfg.Set(cloneCfg)
	case reflect.Ptr:
		if cfg.IsNil() {
			cfg.Set(reflect.New(cfg.Type().Elem()))
		}
		elemCfg := clone(cfg).Elem()
		err := m.enterField(key, &updateConfigOpts{
			data:           data,
			currentPath:    path,
			rflCfgVal:      elemCfg,
			sortedKeyPaths: sortedKeys,
		})
		if err != nil {
			return err
		}
		cfg.Set(elemCfg.Addr())
	default:
		return errors.B().
			Code(EInvalidConfigKey).
			Msgf("config field with key %s does not exist", path).
			Build()
	}

	return nil
}

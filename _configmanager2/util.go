package old_configmanager

import (
	"reflect"
	"slices"
	"strings"
)

func sortedMapKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	return keys
}

func getNextKey(keyPath string, prefix string) string {
	rest := strings.TrimPrefix(keyPath, prefix)
	if rest == "" {
		return ""
	}
	return strings.Split(rest[1:], ".")[0]
}

func joinPath(prefix string, key string) string {
	return prefix + "." + key
}

func clone(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			newVal := reflect.New(val.Type().Elem())
			val.Set(newVal)
		} else {
			newVal := reflect.New(val.Type().Elem())
			newVal.Elem().Set(val.Elem())
			return newVal
		}
	} else if val.Kind() == reflect.Map {
		newVal := reflect.MakeMap(val.Type())
		for _, key := range val.MapKeys() {
			newVal.SetMapIndex(key, val.MapIndex(key))
		}
		return newVal
	} else if val.Kind() == reflect.Slice {
		newVal := reflect.MakeSlice(val.Type(), val.Len(), val.Len())
		for i := 0; i < val.Len(); i++ {
			newVal.Index(i).Set(val.Index(i))
		}
		return newVal
	} else if val.Kind() == reflect.Struct {
		newVal := reflect.New(val.Type())
		newVal.Elem().Set(val)
		return newVal.Elem()
	}
	return val
}

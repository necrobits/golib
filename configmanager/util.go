package configmanager

import (
	"reflect"
	"strings"
)

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
	}
	return val
}

// convert from map of dot notation to map of map/value
// e.g. {"a.b.c": 1} => {"a": {"b": {"c": 1}}}
func convertDotNotationToMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		keys := strings.Split(k, ".")
		if len(keys) == 1 {
			result[k] = v
		} else {
			currentMap := result
			var ok bool
			for i := 0; i < len(keys)-1; i++ {
				if _, ok = currentMap[keys[i]]; !ok {
					currentMap[keys[i]] = make(map[string]interface{})
				}
				currentMap = currentMap[keys[i]].(map[string]interface{})
			}
			currentMap[keys[len(keys)-1]] = v
		}

	}
	return result
}

package configmanager

import (
	"reflect"
	"strconv"
	"strings"
)

func dotFlatConfig(cfg Config, cfgTag string) map[string]interface{} {
	result := make(map[string]interface{})
	toDottedNotationHelper(cfg, result, cfg.Name()+".", cfgTag)
	return result
}

func toDottedNotationHelper(cfg interface{}, result map[string]interface{}, prefix string, cfgTag string) {
	cfgValue := reflect.ValueOf(cfg)
	if cfgValue.Kind() == reflect.Ptr {
		cfgValue = cfgValue.Elem()
	}

	if cfgValue.Type().String() == "json.RawMessage" {
		result[prefix[:len(prefix)-1]] = cfg
	} else if cfgValue.Kind() == reflect.Map {
		for _, key := range cfgValue.MapKeys() {
			toDottedNotationHelper(cfgValue.MapIndex(key).Interface(), result, prefix+key.String()+".", cfgTag)
		}
	} else if cfgValue.Kind() == reflect.Slice {
		for i := 0; i < cfgValue.Len(); i++ {
			toDottedNotationHelper(cfgValue.Index(i).Interface(), result, prefix+strconv.Itoa(i)+".", cfgTag)
		}
	} else if cfgValue.Kind() == reflect.Struct {
		for i := 0; i < cfgValue.NumField(); i++ {
			tag := cfgValue.Type().Field(i).Tag.Get(cfgTag)
			if tag == "" {
				tag = cfgValue.Type().Field(i).Name
			} else if tag == "-" {
				continue
			}
			toDottedNotationHelper(cfgValue.Field(i).Interface(), result, prefix+tag+".", cfgTag)
		}
	} else {
		result[prefix[:len(prefix)-1]] = cfg
	}
}

// convert from map of dot notation to map of map/value
// e.g. {"a.b.c": 1} => {"a": {"b": {"c": 1}}}
func convertDotNotationToMap(data map[string]interface{}, cfgTag string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		keys := strings.Split(k, ".")
		if len(keys) == 1 {
			result[k] = toPrimitiveMap(v, cfgTag)
		} else {
			currentMap := result
			var ok bool
			for i := 0; i < len(keys)-1; i++ {
				if _, ok = currentMap[keys[i]]; !ok {
					currentMap[keys[i]] = make(map[string]interface{})
				}
				currentMap = currentMap[keys[i]].(map[string]interface{})
			}
			currentMap[keys[len(keys)-1]] = toPrimitiveMap(v, cfgTag)
		}

	}
	return result
}

// convert value to map of only primitive types
func toPrimitiveMap(val interface{}, cfgTag string) interface{} {
	if val == nil {
		return nil
	}

	valValue := reflect.ValueOf(val)
	if valValue.Kind() == reflect.Ptr {
		valValue = valValue.Elem()
	}

	if valValue.Type().String() == "json.RawMessage" {
		return val
	} else if valValue.Kind() == reflect.Map {
		result := make(map[string]interface{})
		for _, key := range valValue.MapKeys() {
			result[key.String()] = toPrimitiveMap(valValue.MapIndex(key).Interface(), cfgTag)
		}
		return result
	} else if valValue.Kind() == reflect.Slice {
		result := make(map[string]interface{})
		for i := 0; i < valValue.Len(); i++ {
			key := strconv.Itoa(i)
			result[key] = toPrimitiveMap(valValue.Index(i).Interface(), cfgTag)
		}
		return result
	} else if valValue.Kind() == reflect.Struct {
		result := make(map[string]interface{})
		for i := 0; i < valValue.NumField(); i++ {
			field := valValue.Field(i)
			tag := valValue.Type().Field(i).Tag.Get(cfgTag)
			if tag == "" {
				tag = valValue.Type().Field(i).Name
			} else if tag == "-" {
				continue
			}
			result[tag] = toPrimitiveMap(field.Interface(), cfgTag)
		}
		return result
	} else {
		return val
	}
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
	}
	return val
}

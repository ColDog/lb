package tools

import (
	"encoding/json"
	"fmt"
)

// a general purpose map data structure that is polymorphic. Used for configuration and in http body requests.
type Map struct {
	json 	[]byte
	data 	map[string] interface{}
}


func (obj *Map) before() {
	fmt.Println("checking json")
	if obj.json != nil {
		fmt.Println("parsing json")
		if err := json.Unmarshal(obj.json, &obj.data); err != nil {
			panic(err)
		}
		obj.json = nil
	}

	if obj.data == nil {
		obj.data = make(map[string] interface{})
	}
}

func (obj *Map) Str(key string) string {
	obj.before()
	if item, ok := obj.data[key]; ok {
		if asserted, ok := item.(string); ok {
			return asserted
		}
	}

	return ""
}

func (obj *Map) StrArray(key string) []string {
	obj.before()
	if item, ok := obj.data[key]; ok {
		if asserted, ok := item.([]string); ok {
			return asserted
		}
	}

	return make([]string, 0)
}

func (obj *Map) MapArray(key string) []Map {
	obj.before()
	results := make([]Map, 0)
	if item, ok := obj.data[key]; ok {
		if asserted, ok := item.([]map[string] interface{}); ok {
			for _, newobj := range asserted {
				results = append(results, Map{data: newobj})
			}
		}
	}

	return results
}

func (obj *Map) Int32(key string) int32 {
	return int32(obj.Int64(key))
}

func (obj *Map) Int8(key string) int8 {
	return int8(obj.Int64(key))
}

func (obj *Map) Int16(key string) int16 {
	return int16(obj.Int64(key))
}

func (obj *Map) Int(key string) int {
	return int(obj.Int64(key))
}

func (obj *Map) Float32(key string) float32 {
	return float32(obj.Float64(key))
}

func (obj *Map) Float64(key string) float64 {
	obj.before()
	if item, ok := obj.data[key]; ok {
		switch v := item.(type) {
		case int:
			return float64(v)
		case int8:
			return float64(v)
		case int16:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case uint8:
			return float64(v)
		case uint16:
			return float64(v)
		case uint32:
			return float64(v)
		case uint64:
			return float64(v)
		case float64:
			return v
		case float32:
			return float64(v)
		}
	}

	return float64(0)
}

func (obj *Map) Int64(key string) int64 {
	obj.before()
	if item, ok := obj.data[key]; ok {
		switch v := item.(type) {
		case int:
			return int64(v)
		case int8:
			return int64(v)
		case int16:
			return int64(v)
		case int32:
			return int64(v)
		case int64:
			return v
		case uint8:
			return int64(v)
		case uint16:
			return int64(v)
		case uint32:
			return int64(v)
		case uint64:
			return int64(v)
		case float64:
			return int64(v)
		case float32:
			return int64(v)
		}
	}

	return int64(0)
}


func (obj *Map) Bool(key string) bool {
	obj.before()
	if item, ok := obj.data[key]; ok {
		if asserted, ok := item.(bool); ok {
			return asserted
		}
	}

	return false
}

func (obj *Map) Exists(key string) bool {
	obj.before()
	_, ok := obj.data[key]
	return ok
}

func (obj *Map) Put(key string, val interface{}) {
	obj.before()
	obj.data[key] = val
}

func (obj *Map) Extend(newMap map[string] interface{}) {
	obj.before()
	for key, val := range newMap {
		obj.data[key] = val
	}
}

func (obj *Map) ToJson(key string, val interface{}) {
	obj.before()
}

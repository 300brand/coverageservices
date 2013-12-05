package StorageReader

import (
	"labix.org/v2/mgo/bson"
	"reflect"
)

// Recursively scans a query to convert any ObjectId string values back into
// ObjectIds because JSON loses the type during transport
func objectIdify(in interface{}) {
	v := reflect.ValueOf(in)
	switch v.Kind() {
	case reflect.Ptr:
		v = v.Elem()
	default:
		return
	}

	reflectInto(v)
}

func reflectInto(v reflect.Value) {
	orig := v
	// Spin out the map/slice stuff
	for {
		switch v.Kind() {
		case reflect.Map:
			// Maps keep coming in as map[...]interface{}, should be able to
			// keep it this way until it fails
			for _, key := range v.MapKeys() {
				valInt := v.MapIndex(key).Elem().Interface()
				valPtr := reflect.ValueOf(&valInt)
				reflectInto(valPtr)
				v.SetMapIndex(key, valPtr.Elem())
			}
			return
		case reflect.Slice:
			newSlice := make([]interface{}, v.Len())
			for i := 0; i < v.Len(); i++ {
				newSlice[i] = v.Index(i).Interface()
				reflectInto(reflect.ValueOf(&newSlice[i]))
			}
			// Need the original pointer value to reset the slice to []interface{}
			orig.Elem().Set(reflect.ValueOf(newSlice))
			return
		case reflect.Ptr:
			// If it's a pointer, drop to the element pointed to
			v = v.Elem()
			// Test if the pointed element is a map or slice
			newV := reflect.ValueOf(v.Interface())
			if k := newV.Kind(); k == reflect.Slice || k == reflect.Map {
				v = newV
				continue
			}
		}
		break
	}

	if s, ok := v.Interface().(string); ok && bson.IsObjectIdHex(s) {
		v.Set(reflect.ValueOf(bson.ObjectIdHex(s)))
	}
}

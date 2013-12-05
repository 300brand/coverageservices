package StorageReader

import (
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestSingleDimension(t *testing.T) {
	id := bson.NewObjectId()
	m := bson.M{"_id": id.Hex()}
	objectIdify(&m)
	switch v := m["_id"].(type) {
	case bson.ObjectId:
		if id.String() != v.String() {
			t.Errorf("IDs do not match %s != %s", id, v)
		}
	default:
		t.Errorf("Invalid _id type [%T]: %+v", v, v)
	}
}

func TestMultiDimension(t *testing.T) {
	id := bson.NewObjectId()
	m := bson.M{"_id": bson.M{"$not": id.Hex()}}
	objectIdify(&m)
	switch v := m["_id"].(type) {
	case bson.M:
		switch w := v["$not"].(type) {
		case bson.ObjectId:
			if id.String() != w.String() {
				t.Errorf("IDs do not match %s != %s", id, w)
			}
		default:
			t.Errorf("Invalid $not type [%T]: %+v", w, w)
		}
	default:
		t.Errorf("Invalid _id type [%T]: %+v", v, v)
	}
}

func TestIdArr(t *testing.T) {
	ids := []bson.ObjectId{
		bson.NewObjectId(),
		bson.NewObjectId(),
	}
	m := bson.M{"_id": bson.M{"$in": []string{ids[0].Hex(), ids[1].Hex()}}}
	objectIdify(&m)
	switch v := m["_id"].(type) {
	case bson.M:
		switch w := v["$in"].(type) {
		case []interface{}:
			for i, id := range w {
				bId, ok := id.(bson.ObjectId)
				if !ok {
					t.Errorf("Invalid id type [%T]: %v", id, id)
					continue
				}
				if ids[i].String() != bId.String() {
					t.Errorf("IDs do not match %s != %s", ids[i], id)
				}
			}
		default:
			t.Errorf("Invalid $in type [%T]: %+v", w, w)
		}
	default:
		t.Errorf("Invalid _id type [%T]: %+v", v, v)
	}
}

func TestMapInterface(t *testing.T) {
	id := bson.NewObjectId()
	m := bson.M{"_id": map[string]interface{}{"$not": id.Hex()}}
	objectIdify(&m)
	switch v := m["_id"].(type) {
	case map[string]interface{}:
		switch w := v["$not"].(type) {
		case bson.ObjectId:
			if id.String() != w.String() {
				t.Errorf("IDs do not match %s != %s", id, w)
			}
		default:
			t.Errorf("Invalid $not type [%T]: %+v", w, w)
		}
	default:
		t.Errorf("Invalid _id type [%T]: %+v", v, v)
	}
}

func TestMapInterfaceArr(t *testing.T) {
	ids := []bson.ObjectId{
		bson.NewObjectId(),
		bson.NewObjectId(),
	}
	m := bson.M{"_id": map[string]interface{}{"$in": []interface{}{ids[0].Hex(), ids[1].Hex()}}}
	objectIdify(&m)
	switch v := m["_id"].(type) {
	case map[string]interface{}:
		switch w := v["$in"].(type) {
		case []interface{}:
			for i, id := range w {
				bId, ok := id.(bson.ObjectId)
				if !ok {
					t.Errorf("Invalid id type [%T]: %v", id, id)
					continue
				}
				if ids[i].String() != bId.String() {
					t.Errorf("IDs do not match %s != %s", ids[i], id)
				}
			}
		default:
			t.Errorf("Invalid $in type [%T]: %+v", w, w)
		}
	default:
		t.Errorf("Invalid _id type [%T]: %+v", v, v)
	}
}

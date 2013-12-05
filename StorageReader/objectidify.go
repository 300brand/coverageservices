package StorageReader

import (
	"labix.org/v2/mgo/bson"
)

// Recursively scans a bson.M to convert any string values back into ObjectIds
// because JSON loses the type during transport
func objectIdify(m *bson.M) {
	for k, v := range *m {
		switch i := v.(type) {
		case string:
			if bson.IsObjectIdHex(i) {
				(*m)[k] = bson.ObjectIdHex(i)
			}
		case []string:
			ids := make([]bson.ObjectId, len(i))
			for idx := range i {
				if bson.IsObjectIdHex(i[idx]) {
					ids[idx] = bson.ObjectIdHex(i[idx])
				}
			}
			(*m)[k] = ids
		case bson.M:
			objectIdify(&i)
		}
	}
}

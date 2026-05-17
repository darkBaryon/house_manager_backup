package adm

import "go.mongodb.org/mongo-driver/v2/bson"

func compactObjectIDs(ids []bson.ObjectID) []bson.ObjectID {
	if len(ids) == 0 {
		return nil
	}
	result := make([]bson.ObjectID, 0, len(ids))
	seen := make(map[bson.ObjectID]struct{}, len(ids))
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

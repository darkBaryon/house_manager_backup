package common

import "go.mongodb.org/mongo-driver/v2/bson"

// UpdateDoc is a typed wrapper around Mongo update operators.
type UpdateDoc struct {
	set         bson.M
	unset       bson.M
	inc         bson.M
	setOnInsert bson.M
}

// NewUpdateDoc creates an empty update document.
func NewUpdateDoc() UpdateDoc {
	return UpdateDoc{}
}

// Set adds a $set field.
func (u UpdateDoc) Set(field Field, value any) UpdateDoc {
	if u.set == nil {
		u.set = bson.M{}
	}
	u.set[string(field)] = value
	return u
}

// Unset adds a $unset field.
func (u UpdateDoc) Unset(field Field) UpdateDoc {
	if u.unset == nil {
		u.unset = bson.M{}
	}
	u.unset[string(field)] = ""
	return u
}

// Inc adds a $inc field.
func (u UpdateDoc) Inc(field Field, value any) UpdateDoc {
	if u.inc == nil {
		u.inc = bson.M{}
	}
	u.inc[string(field)] = value
	return u
}

// SetOnInsert adds a $setOnInsert field.
func (u UpdateDoc) SetOnInsert(field Field, value any) UpdateDoc {
	if u.setOnInsert == nil {
		u.setOnInsert = bson.M{}
	}
	u.setOnInsert[string(field)] = value
	return u
}

// IsEmpty reports whether the update has no operators.
func (u UpdateDoc) IsEmpty() bool {
	return len(u.set) == 0 && len(u.unset) == 0 && len(u.inc) == 0 && len(u.setOnInsert) == 0
}

// BSON returns a Mongo update document.
func (u UpdateDoc) BSON() bson.M {
	doc := bson.M{}
	if len(u.set) > 0 {
		doc["$set"] = cloneBsonM(u.set)
	}
	if len(u.unset) > 0 {
		doc["$unset"] = cloneBsonM(u.unset)
	}
	if len(u.inc) > 0 {
		doc["$inc"] = cloneBsonM(u.inc)
	}
	if len(u.setOnInsert) > 0 {
		doc["$setOnInsert"] = cloneBsonM(u.setOnInsert)
	}
	return doc
}

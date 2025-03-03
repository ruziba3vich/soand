package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	Background struct {
		ID       primitive.ObjectID `bson:"_id,omitempty"`
		Filename string             `bson:"filename"`
	}
)

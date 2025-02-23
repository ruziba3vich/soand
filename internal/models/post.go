package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`                // MongoDB ObjectID
	CreatorId   primitive.ObjectID `bson:"creator_id,omitempty" json:"creator_id"` // Creator id
	Pictures    []string           `bson:"pictures" json:"picture"`                // Image URLs or file path
	Tags        []string           `bson:"tags" json:"tags"`                       // List of tags
	Description string             `bson:"description" json:"description"`         // Post description
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`           // Timestamp
	DeleteAt    time.Time          `bson:"delete_at" json:"delete_at"`             // Field for automatic deletion
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SenderID    primitive.ObjectID `bson:"sender_id" json:"sender_id"`
	RecipientID primitive.ObjectID `bson:"recipient_id" json:"recipient_id"`
	Content     string             `bson:"content" json:"content"`
	Pictures    []string           `bson:"pictures,omitempty" json:"pictures,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

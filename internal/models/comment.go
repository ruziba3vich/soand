package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	PostID       primitive.ObjectID `json:"post_id" bson:"post_id"`
	Text         string             `json:"text,omitempty" bson:"text,omitempty"`
	VoiceMessage string             `json:"voice_message,omitempty" bson:"voice_message,omitempty"`
	Pictures     []string           `json:"pictures,omitempty" bson:"pictures,omitempty"`
	ReplyTo      primitive.ObjectID `json:"reply_to" bson:"reply_to"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
}

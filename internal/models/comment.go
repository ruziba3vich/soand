package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID          primitive.ObjectID `json:"user_id" bson:"user_id"`
	PostID          primitive.ObjectID `json:"post_id" bson:"post_id"`
	Text            string             `json:"text,omitempty" bson:"text,omitempty"`
	VoiceMessage    string             `json:"voice_message,omitempty" bson:"voice_message,omitempty"`
	Pictures        []string           `json:"pictures,omitempty" bson:"pictures,omitempty"`
	ReplyTo         primitive.ObjectID `json:"reply_to,omitempty" bson:"reply_to"`
	OwnerFullname   string             `bson:"owner_full_name" json:"owner_full_name"`
	OwnerProfilePic string             `bson:"owner_profile_pic" json:"owner_profile_pic"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	Reactions       map[string]int     `json:"reactions" bson:"reactions"`
}

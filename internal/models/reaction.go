package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Reaction struct {
	CommentId primitive.ObjectID `bson:"comment_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Reaction  string             `json:"reaction"`
	Incr      bool               `json:"incr"`
}

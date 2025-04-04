package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Reaction struct {
	PostID primitive.ObjectID `bson:"post_id"`
	UserID primitive.ObjectID `bson:"user_id"`
}

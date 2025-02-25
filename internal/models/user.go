package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Fullname      string             `json:"full_name" bson:"full_name"`
	Phone         string             `json:"phone" bson:"phone" binding:"required"`
	Username      string             `json:"username" bson:"username" binding:"required"`
	Password      string             `json:"password" bson:"password"`
	Status        string             `json:"status" bson:"status"`
	ProfilePics   []string           `json:"profile_pics" bson:"profile_pics"`
	HiddenProfile bool               `json:"profile_hidden" bson:"profile_hidden"`
}

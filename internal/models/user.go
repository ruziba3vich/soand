package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID            primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Fullname      string             `json:"full_name" bson:"full_name" binding:"required"`
	Phone         string             `json:"phone" bson:"phone"`
	Username      string             `json:"username" bson:"username"`
	Password      string             `json:"password" bson:"password" binding:"required"`
	Bio           string             `json:"bio" bson:"bio"`
	Status        string             `json:"status" bson:"status"`
	ProfilePics   []string           `json:"profile_pics" bson:"profile_pics"`
	BackgroundPic string             `json:"background_pic" bson:"background_pic"`
	HiddenProfile bool               `json:"profile_hidden" bson:"profile_hidden"`
}

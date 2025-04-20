package storage

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LikesStorage struct {
	db *mongo.Collection
}

func NewLikesStorage(db *mongo.Collection) *LikesStorage {
	return &LikesStorage{
		db: db,
	}
}

// LikePost adds a like for a post if the user hasn't liked it already.
func (s *LikesStorage) LikePost(ctx context.Context, userID, postID primitive.ObjectID) error {
	// Insert the like
	like := bson.M{
		"user_id": userID,
		"post_id": postID,
	}
	_, err := s.db.InsertOne(ctx, like)
	return err
}

// DislikePost removes a like from a post if the user has liked it before.
func (s *LikesStorage) DislikePost(ctx context.Context, userID, postID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID, "post_id": postID}
	_, err := s.db.DeleteOne(ctx, filter)
	return err
}

// HasUserLiked checks if a user has already liked a given post.
func (s *LikesStorage) HasUserLiked(ctx context.Context, userID, postID primitive.ObjectID) (bool, error) {
	filter := bson.M{"user_id": userID, "post_id": postID}
	result := s.db.FindOne(ctx, filter)

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return false, nil // user has not liked the post
	}
	if result.Err() != nil {
		return false, result.Err() // some unexpected error
	}
	return true, nil // user has liked the post
}

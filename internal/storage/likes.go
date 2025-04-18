package storage

import (
	"context"
	"fmt"

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
	// Check if the like already exists
	filter := bson.M{"user_id": userID, "post_id": postID}
	existing := s.db.FindOne(ctx, filter)
	if existing.Err() == nil {
		return fmt.Errorf("you have already liked this post")
	} else if existing.Err() != mongo.ErrNoDocuments {
		// Some other error occurred
		return existing.Err()
	}

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

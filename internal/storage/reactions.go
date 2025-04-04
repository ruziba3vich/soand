package storage

import (
	"context"
	"fmt"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReactionsStorage struct {
	db *mongo.Collection
}

func NewReactionsStorage(db *mongo.Collection) *ReactionsStorage {
	return &ReactionsStorage{
		db: db,
	}
}

// AddReaction adds a reaction (either "like" or "dislike")
// Returns 1 if added, 0 if already exists
func (r *ReactionsStorage) AddReaction(ctx context.Context, postID, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{"post_id": postID, "user_id": userID}
	var existing models.Reaction

	err := r.db.FindOne(ctx, filter).Decode(&existing)
	if err == nil {
		// User already reacted
		return false, nil
	} else if err != mongo.ErrNoDocuments {
		return false, err // unexpected error
	}

	// Add new reaction
	newReaction := models.Reaction{
		PostID: postID,
		UserID: userID,
	}

	_, err = r.db.InsertOne(ctx, newReaction)
	if err != nil {
		return false, err
	}

	return true, nil
}

// RemoveReaction removes a user's reaction from a post
// Returns 1 if deleted, 0 if it didn't exist
func (r *ReactionsStorage) RemoveReaction(ctx context.Context, postID, userID string) (bool, error) {
	filter := bson.M{"post_id": postID, "user_id": userID}

	res, err := r.db.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, fmt.Errorf("you have not reacted to this post")
		}
		return false, err
	}

	if res.DeletedCount == 0 {
		return false, nil
	}

	return true, nil
}

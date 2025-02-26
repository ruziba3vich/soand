package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommentStorage struct {
	db *mongo.Collection
}

// NewCommentStorage initializes the comment storage
func NewCommentStorage(db *mongo.Collection) *CommentStorage {
	return &CommentStorage{
		db: db,
	}
}

// CreateComment inserts a new comment into the database
func (s *CommentStorage) CreateComment(ctx context.Context, comment *models.Comment) error {
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()

	// If it's a reply, ensure the parent comment exists within the same post
	if !comment.ReplyTo.IsZero() {
		var parentComment models.Comment
		err := s.db.FindOne(ctx, bson.M{"_id": comment.ReplyTo, "post_id": comment.PostID}).Decode(&parentComment)
		if err != nil {
			return fmt.Errorf("parent comment not found within the same post")
		}
	}

	_, err := s.db.InsertOne(ctx, comment)
	return err
}

// DeleteComment removes a comment by ID
func (s *CommentStorage) DeleteComment(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID) error {
	res, err := s.db.DeleteOne(ctx, bson.M{"_id": commentID, "user_id": userID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("comment not found or unauthorized")
	}
	return nil
}

// GetCommentsByPostID retrieves paginated comments for a specific post
func (s *CommentStorage) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, pageSize int64) ([]models.Comment, error) {
	var comments []models.Comment

	// Ensure page and pageSize have valid values
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	skip := (page - 1) * pageSize

	cursor, err := s.db.Find(ctx, bson.M{"post_id": postID}, &options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
		Sort:  bson.M{"created_at": -1}, // Sort by newest comments first
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

// UpdateCommentText updates the text of a comment by its ID
func (s *CommentStorage) UpdateCommentText(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID, newText string) error {
	if newText == "" {
		return fmt.Errorf("comment text cannot be empty")
	}

	// Define the update filter (only allow the owner of the comment to edit)
	filter := bson.M{"_id": commentID, "user_id": userID}
	update := bson.M{"$set": bson.M{"text": newText, "updated_at": time.Now()}}

	result, err := s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// If no document was modified, return an error (comment not found or not owned by user)
	if result.ModifiedCount == 0 {
		return fmt.Errorf("comment not found or user not authorized to edit")
	}

	return nil
}

// GetCommentsByUserID retrieves all comments made by a user
// func (s *CommentStorage) GetCommentsByUserID(ctx context.Context, userID primitive.ObjectID) ([]models.Comment, error) {
// 	var comments []models.Comment

// 	cursor, err := s.db.Find(ctx, bson.M{"user_id": userID})
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)

// 	if err := cursor.All(ctx, &comments); err != nil {
// 		return nil, err
// 	}
// 	return comments, nil
// }

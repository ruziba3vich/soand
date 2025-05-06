package storage

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	dto "github.com/ruziba3vich/soand/internal/dtos"
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

// AddReactionToComment adds a user's reaction to a comment
func (s *CommentStorage) AddReactionToComment(ctx context.Context, reaction *models.Reaction) error {

	comment, err := s.GetCommentByID(ctx, reaction.CommentId)
	if err != nil {
		return err
	}
	comment.Reactions[reaction.Reaction] = append(comment.Reactions[reaction.Reaction], reaction.UserID)

	_, err = s.db.ReplaceOne(
		ctx,
		bson.M{"_id": comment.ID},
		comment,
	)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return nil
}

// RemoveReactionFromComment removes a user's reaction from a comment
func (s *CommentStorage) RemoveReactionFromComment(ctx context.Context, reaction *models.Reaction) error {
	comment, err := s.GetCommentByID(ctx, reaction.CommentId)
	if err != nil {
		return err
	}

	var found bool

	for r, users := range comment.Reactions {
		ind := slices.Index(users, reaction.UserID)
		if ind != -1 {
			newUsers := slices.Delete(users, ind, ind+1)
			if len(newUsers) == 0 {
				delete(comment.Reactions, r)
			} else {
				comment.Reactions[r] = newUsers
			}
			found = true
			break
		}
	}

	if !found {
		return dto.ErrNotReacted
	}

	_, err = s.db.ReplaceOne(
		ctx,
		bson.M{"_id": comment.ID},
		comment,
	)
	if err != nil {
		return fmt.Errorf("failed to update comment: %s", err.Error())
	}

	return nil
}

func (s *CommentStorage) GetParentComment(ctx context.Context, comment *models.Comment) error {
	var parentComment models.Comment
	return s.db.FindOne(ctx, bson.M{"_id": comment.ReplyTo, "post_id": comment.PostID}).Decode(&parentComment)
}

func (s *CommentStorage) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, pageSize int64) ([]*models.Comment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	skip := (page - 1) * pageSize

	opts := options.Find().
		SetLimit(pageSize).
		SetSkip(skip).
		SetSort(bson.M{"created_at": -1})

	cursor, err := s.db.Find(ctx, bson.M{"post_id": postID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*models.Comment
	for cursor.Next(ctx) {
		var comment models.Comment
		if err := cursor.Decode(&comment); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err := cursor.Err(); err != nil {
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

// GetCommentByID retrieves a comment by its ID, hiding UserID if the user's profile is private
func (s *CommentStorage) GetCommentByID(ctx context.Context, commentID primitive.ObjectID) (*models.Comment, error) {
	var comment models.Comment

	// Find the comment by ID
	err := s.db.FindOne(ctx, bson.M{"_id": commentID}).Decode(&comment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}

	return &comment, nil
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

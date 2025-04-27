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
	db           *mongo.Collection
	user_storage *UserStorage
}

// NewCommentStorage initializes the comment storage
func NewCommentStorage(db *mongo.Collection, user_storage *UserStorage) *CommentStorage {
	return &CommentStorage{
		db:           db,
		user_storage: user_storage,
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
func (s *CommentStorage) AddReactionToComment(ctx context.Context, reaction *models.Reaction) (*models.Comment, error) {

	comment, err := s.GetCommentByID(ctx, reaction.CommentId)
	if err != nil {
		return nil, err
	}
	comment.Reactions[reaction.Reaction] = append(comment.Reactions[reaction.Reaction], reaction.UserID)

	_, err = s.db.ReplaceOne(
		ctx,
		bson.M{"_id": comment.ID},
		comment,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

// RemoveReactionFromComment removes a user's reaction from a comment
func (s *CommentStorage) RemoveReactionFromComment(ctx context.Context, reaction *models.Reaction) (*models.Comment, error) {
	comment, err := s.GetCommentByID(ctx, reaction.CommentId)
	if err != nil {
		return nil, err
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
		return nil, dto.ErrNotReacted
	}

	_, err = s.db.ReplaceOne(
		ctx,
		bson.M{"_id": comment.ID},
		comment,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update comment: %s", err.Error())
	}

	return comment, nil
}

func (s *CommentStorage) GetParentComment(ctx context.Context, comment *models.Comment) error {
	var parentComment models.Comment
	return s.db.FindOne(ctx, bson.M{"_id": comment.ReplyTo, "post_id": comment.PostID}).Decode(&parentComment)
}

// GetCommentsByPostID retrieves paginated comments for a specific post
func (s *CommentStorage) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page, pageSize int64) ([]models.Comment, error) {
	// Ensure page and pageSize have valid values
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	skip := (page - 1) * pageSize

	// Find comments with pagination and sorting
	cursor, err := s.db.Find(ctx, bson.M{"post_id": postID}, &options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
		Sort:  bson.M{"created_at": -1}, // Sort by newest comments first
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Initialize an empty slice to store comments
	var comments []models.Comment

	// Iterate over the cursor and process each comment
	for cursor.Next(ctx) {
		var comment models.Comment
		if err := cursor.Decode(&comment); err != nil {
			return nil, err
		}
		owner, err := s.user_storage.GetUserByID(ctx, comment.UserID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				comment.UserID = primitive.NilObjectID
				comment.OwnerFullname = "Deleted Account"
			} else {
				return nil, err
			}
		}
		if owner.HiddenProfile {
			comment.UserID = primitive.NilObjectID
			comment.OwnerFullname = "Anonim user"
		} else {
			comment.OwnerFullname = owner.Fullname
			if len(owner.ProfilePics) > 0 {
				comment.OwnerProfilePic = owner.ProfilePics[0].Url
			}
		}
		comments = append(comments, comment)
	}

	// Check for any errors that occurred during iteration
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// UpdateCommentText updates the text of a comment by its ID
func (s *CommentStorage) UpdateCommentText(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID, newText string) (*models.Comment, error) {
	if newText == "" {
		return nil, fmt.Errorf("comment text cannot be empty")
	}

	// Define the update filter (only allow the owner of the comment to edit)
	filter := bson.M{"_id": commentID, "user_id": userID}
	update := bson.M{"$set": bson.M{"text": newText, "updated_at": time.Now()}}

	result, err := s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// If no document was modified, return an error (comment not found or not owned by user)
	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("comment not found or user not authorized to edit")
	}

	return s.GetCommentByID(ctx, commentID)
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

	// Check if the user's profile is hidden
	user, err := s.user_storage.GetUserByID(ctx, comment.UserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			comment.UserID = primitive.NilObjectID
			comment.OwnerFullname = "Deleted Account"
			return &comment, bson.ErrDecodeToNil
		}
		return nil, err
	}

	// If the user's profile is private, clear the UserID
	if user.HiddenProfile {
		comment.OwnerFullname = "Anonim user"
		comment.UserID = primitive.NilObjectID // Set to zero value to hide it
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

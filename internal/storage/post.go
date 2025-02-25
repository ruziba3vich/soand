package storage

import (
	"context"
	"errors"
	"time"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	db            *mongo.Collection
	users_storage *UserStorage
}

// CreatePost inserts a new post into the database
func (s *Storage) CreatePost(ctx context.Context, post *models.Post, deleteAfter int) (primitive.ObjectID, error) {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.DeleteAt = post.CreatedAt.Add(time.Duration(deleteAfter) * time.Hour)
	_, err := s.db.InsertOne(ctx, post)
	return post.ID, err
}

// GetPost retrieves a post by ID
func (s *Storage) GetPost(ctx context.Context, id primitive.ObjectID) (*models.Post, error) {
	var post models.Post
	err := s.db.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("post not found")
	}
	return &post, err
}

func (s *Storage) UpdatePost(ctx context.Context, id, updaterID primitive.ObjectID, update bson.M) error {
	// Find the post and check ownership
	post, err := s.GetPost(ctx, id)
	if err != nil {
		return err
	}
	// Check if the updater is the creator
	if post.CreatorId != updaterID {
		return errors.New("only the creator can update this post")
	}

	// Perform the update
	result, err := s.db.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("failed to update post")
	}

	return nil
}

// DeletePost permanently removes a post from the database
func (s *Storage) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.db.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("post not found")
	}
	return nil
}

func (s *Storage) GetAllPosts(ctx context.Context, page, pageSize int64) ([]models.Post, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	skip := int64((page - 1) * pageSize) // Calculate how many documents to skip

	cursor, err := s.db.Find(ctx, bson.M{}, &options.FindOptions{
		Sort:  bson.M{"created_at": -1}, // Sort by newest first
		Skip:  &skip,
		Limit: &pageSize,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	for cursor.Next(ctx) {
		var post models.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, err
		}

		// Fetch owner details
		owner, err := s.users_storage.GetUserByID(ctx, post.CreatorId)
		if err != nil {
			return nil, err
		}

		// Check if the owner's profile is hidden
		if owner.HiddenProfile {
			post.CreatorId = primitive.NilObjectID // Set to "00000" equivalent
		}

		posts = append(posts, post)
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

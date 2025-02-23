package service

import (
	"context"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostService struct to handle post-related operations
type PostService struct {
	storage *storage.Storage
	logger  *logrus.Logger
}

// NewPostService initializes a new PostService with storage and logger
func NewPostService(storage *storage.Storage, logger *logrus.Logger) *PostService {
	// Create a logger
	return &PostService{
		storage: storage,
		logger:  logger,
	}
}

// CreatePost inserts a new post into the database
func (s *PostService) CreatePost(ctx context.Context, post *models.Post, deleteAfter int) (primitive.ObjectID, error) {
	id, err := s.storage.CreatePost(ctx, post, deleteAfter)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"content": post.Description,
			"error":   err.Error(),
		}).Error("Failed to create post")
		return primitive.NilObjectID, err
	}

	s.logger.WithFields(logrus.Fields{
		"id":      id.Hex(),
		"content": post.Description,
	}).Info("Post created successfully")
	return id, nil
}

// DeletePost removes a post by ID
func (s *PostService) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	err := s.storage.DeletePost(ctx, id)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"id":    id.Hex(),
			"error": err.Error(),
		}).Error("Failed to delete post")
		return err
	}

	s.logger.WithField("id", id.Hex()).Info("Post deleted successfully")
	return nil
}

// EnsureTTLIndex ensures TTL index is set on the collection
func (s *PostService) EnsureTTLIndex(ctx context.Context) error {
	err := s.storage.EnsureTTLIndex(ctx)
	if err != nil {
		s.logger.WithField("error", err.Error()).Error("Failed to ensure TTL index")
		return err
	}

	s.logger.Info("TTL index ensured successfully")
	return nil
}

// GetAllPosts retrieves paginated posts
func (s *PostService) GetAllPosts(ctx context.Context, page int64, pageSize int64) ([]models.Post, error) {
	posts, err := s.storage.GetAllPosts(ctx, page, pageSize)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"page":     page,
			"pageSize": pageSize,
			"error":    err.Error(),
		}).Error("Failed to retrieve posts")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"page":     page,
		"pageSize": pageSize,
		"count":    len(posts),
	}).Info("Retrieved posts successfully")
	return posts, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, id primitive.ObjectID) (*models.Post, error) {
	post, err := s.storage.GetPost(ctx, id)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"id":    id.Hex(),
			"error": err.Error(),
		}).Error("Failed to retrieve post")
		return nil, err
	}

	s.logger.WithField("id", post.ID.Hex()).Info("Retrieved post successfully")
	return post, nil
}

// UpdatePost updates a post by ID
func (s *PostService) UpdatePost(ctx context.Context, id primitive.ObjectID, updaterID primitive.ObjectID, update bson.M) error {
	err := s.storage.UpdatePost(ctx, id, updaterID, update)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"id":        id.Hex(),
			"updaterID": updaterID.Hex(),
			"update":    update,
			"error":     err.Error(),
		}).Error("Failed to update post")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"id":        id.Hex(),
		"updaterID": updaterID.Hex(),
		"update":    update,
	}).Info("Post updated successfully")
	return nil
}

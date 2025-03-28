package service

import (
	"context"
	"log"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostService struct to handle post-related operations
type PostService struct {
	storage *storage.Storage
	logger  *log.Logger
}

// NewPostService initializes a new PostService with storage and logger
func NewPostService(storage *storage.Storage, logger *log.Logger) *PostService {
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
		s.logger.Println(logrus.Fields{
			"content": post.Description,
			"error":   err.Error(),
		})
		return primitive.NilObjectID, err
	}

	s.logger.Println(logrus.Fields{
		"id":      id.Hex(),
		"content": post.Description,
	})
	return id, nil
}

// DeletePost removes a post by ID
func (s *PostService) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	err := s.storage.DeletePost(ctx, id)
	if err != nil {
		s.logger.Println(logrus.Fields{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return err
	}

	s.logger.Println("id", id.Hex())
	return nil
}

// EnsureTTLIndex ensures TTL index is set on the collection
func (s *PostService) EnsureTTLIndex(ctx context.Context) error {
	err := s.storage.EnsureTTLIndex(ctx)
	if err != nil {
		s.logger.Println("error", err.Error())
		return err
	}

	s.logger.Println("TTL index ensured successfully")
	return nil
}

// GetAllPosts retrieves paginated posts
func (s *PostService) GetAllPosts(ctx context.Context, page int64, pageSize int64) ([]models.Post, error) {
	posts, err := s.storage.GetAllPosts(ctx, page, pageSize)
	if err != nil {
		s.logger.Println(logrus.Fields{
			"page":     page,
			"pageSize": pageSize,
			"error":    err.Error(),
		})
		return nil, err
	}

	s.logger.Println(logrus.Fields{
		"page":     page,
		"pageSize": pageSize,
		"count":    len(posts),
	})
	return posts, nil
}

// GetPost retrieves a post by ID
func (s *PostService) GetPost(ctx context.Context, id primitive.ObjectID) (*models.Post, error) {
	post, err := s.storage.GetPost(ctx, id)
	if err != nil {
		s.logger.Println(logrus.Fields{
			"id":    id.Hex(),
			"error": err.Error(),
		})
		return nil, err
	}

	s.logger.Println("id", post.ID.Hex())
	return post, nil
}

// UpdatePost updates a post by ID
func (s *PostService) UpdatePost(ctx context.Context, id primitive.ObjectID, updaterID primitive.ObjectID, update bson.M) error {
	err := s.storage.UpdatePost(ctx, id, updaterID, update)
	if err != nil {
		s.logger.Println(logrus.Fields{
			"id":        id.Hex(),
			"updaterID": updaterID.Hex(),
			"update":    update,
			"error":     err.Error(),
		})
		return err
	}

	s.logger.Println(logrus.Fields{
		"id":        id.Hex(),
		"updaterID": updaterID.Hex(),
		"update":    update,
	})
	return nil
}

func (s *PostService) SearchPostsByTitle(ctx context.Context, query string, page, pageSize int64) ([]models.Post, error) {
	posts, err := s.storage.SearchPostsByTitle(ctx, query, page, pageSize)
	if err != nil {
		s.logger.Println(err.Error())
	}
	return posts, nil
}

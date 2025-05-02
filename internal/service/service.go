package service

import (
	"context"
	"errors"
	"log"
	"mime/multipart"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostService struct to handle post-related operations
type PostService struct {
	storage       *storage.Storage
	likes_storage *storage.LikesStorage
	logger        *log.Logger
	file_service  repos.IFIleStoreService
}

// NewPostService initializes a new PostService with storage and logger
func NewPostService(storage *storage.Storage, likes_storage *storage.LikesStorage, file_service repos.IFIleStoreService, logger *log.Logger) repos.IPostService {
	// Create a logger
	return &PostService{
		storage:       storage,
		likes_storage: likes_storage,
		logger:        logger,
		file_service:  file_service,
	}
}

// CreatePost inserts a new post into the database
func (s *PostService) CreatePost(ctx context.Context, post *models.Post, files []*multipart.FileHeader, deleteAfter int) error {
	for _, file := range files {
		filename, err := s.file_service.UploadFile(file)
		if err != nil {
			s.logger.Println(logrus.Fields{
				"error": err.Error(),
			})
			return err
		}
		post.Pictures = append(post.Pictures, filename)
	}
	err := s.storage.CreatePost(ctx, post, deleteAfter)
	if err != nil {
		s.logger.Println(logrus.Fields{
			"content": post.Description,
			"error":   err.Error(),
		})
		return err
	}

	s.logger.Println(logrus.Fields{
		"id":      post.ID.Hex(),
		"content": post.Description,
	})
	return nil
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

func (s *PostService) LikeOrDislikePost(ctx context.Context, userId primitive.ObjectID, postId primitive.ObjectID, count int) error {
	liked, err := s.likes_storage.HasUserLiked(ctx, userId, postId)
	if err != nil {
		s.logger.Printf("failed to check if user liked %s\n", err.Error())
		return err
	}
	if count == 1 {
		if liked {
			return errors.New("user already liked this post")
		}
		if err := s.likes_storage.LikePost(ctx, userId, postId); err != nil {
			s.logger.Printf("failed to like the post [%s] by user [%s]\n", postId.Hex(), userId.Hex())
			return err
		}
	} else {
		if !liked {
			return errors.New("user has not liked this post")
		}
		if err := s.likes_storage.DislikePost(ctx, userId, postId); err != nil {
			s.logger.Printf("failed to dislike the post [%s] by user [%s]\n", postId.Hex(), userId.Hex())
			return err
		}
	}

	if err := s.storage.LikeOrDislikePost(ctx, userId, postId, count); err != nil {
		s.logger.Println(err.Error())
	}
	return nil
}

func (s *PostService) changeFiles(post *models.Post) error {
	for i := range post.Pictures {
		fileUrl, err := s.file_service.GetFile(post.Pictures[i])
		if err != nil {
			return err
		}
		post.Pictures[i] = fileUrl
	}

	return nil
}

package service

import (
	"context"
	"log"
	"time"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommentService struct {
	storage storage.CommentStorage
	logger  *log.Logger
}

func NewCommentService(storage storage.CommentStorage, logger *log.Logger) *CommentService {
	return &CommentService{
		storage: storage,
		logger:  logger,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, comment *models.Comment) error {
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()

	if err := s.storage.CreateComment(ctx, comment); err != nil {
		s.logger.Printf("Failed to create comment: %v\n", err)
		return err
	}
	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, userID, commentID primitive.ObjectID) error {
	if err := s.storage.DeleteComment(ctx, userID, commentID); err != nil {
		s.logger.Printf("Failed to delete comment (ID: %s) by user (ID: %s): %v\n", commentID.Hex(), userID.Hex(), err)
		return err
	}
	return nil
}

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, limit, offset int64) ([]models.Comment, error) {
	comments, err := s.storage.GetCommentsByPostID(ctx, postID, limit, offset)
	if err != nil {
		s.logger.Printf("Failed to get comments for post (ID: %s): %v\n", postID.Hex(), err)
		return nil, err
	}
	return comments, nil
}

func (s *CommentService) UpdateCommentText(ctx context.Context, userID, commentID primitive.ObjectID, newText string) error {
	if err := s.storage.UpdateCommentText(ctx, userID, commentID, newText); err != nil {
		s.logger.Printf("Failed to update comment (ID: %s) by user (ID: %s): %v\n", commentID.Hex(), userID.Hex(), err)
		return err
	}
	return nil
}

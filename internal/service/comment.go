package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
)

type CommentService struct {
	storage *storage.CommentStorage
	redis   *redis.Client
	logger  *log.Logger
}

func NewCommentService(storage *storage.CommentStorage, redis *redis.Client, logger *log.Logger) *CommentService {
	return &CommentService{
		storage: storage,
		redis:   redis,
		logger:  logger,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, comment *models.Comment) error {
	// Store the comment in MongoDB
	if err := s.storage.CreateComment(ctx, comment); err != nil {
		s.logger.Println("Error storing comment:", err)
		return err
	}

	// Publish to Redis Pub/Sub
	commentData, err := json.Marshal(comment)
	if err != nil {
		s.logger.Println("Error marshalling comment for Redis:", err)
		return err
	}

	channel := "comments:" + comment.PostID.Hex()
	if err := s.redis.Publish(ctx, channel, commentData).Err(); err != nil {
		s.logger.Println("Error publishing comment to Redis:", err)
		return err
	}

	s.logger.Println("Comment published to Redis on channel:", channel)
	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID) error {
	err := s.storage.DeleteComment(ctx, commentID, userID)
	if err != nil {
		s.logger.Println("Error deleting comment:", err)
		return err
	}

	s.logger.Println("Comment deleted successfully:", commentID.Hex())
	return nil
}

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page int64, pageSize int64) ([]models.Comment, error) {
	comments, err := s.storage.GetCommentsByPostID(ctx, postID, page, pageSize)
	if err != nil {
		s.logger.Println("Error fetching comments:", err)
		return nil, err
	}

	s.logger.Printf("Fetched %d comments for post %s\n", len(comments), postID.Hex())
	return comments, nil
}

func (s *CommentService) UpdateCommentText(ctx context.Context, commentID primitive.ObjectID, userID primitive.ObjectID, newText string) error {
	err := s.storage.UpdateCommentText(ctx, commentID, userID, newText)
	if err != nil {
		s.logger.Println("Error updating comment text:", err)
		return err
	}

	s.logger.Println("Comment updated successfully:", commentID.Hex())
	return nil
}

func (s *CommentService) SubscribeToComments(ctx context.Context, postID primitive.ObjectID, handleMessage func(comment *models.Comment)) {
	channel := "comments:" + postID.Hex()
	pubsub := s.redis.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var comment models.Comment
		if err := json.Unmarshal([]byte(msg.Payload), &comment); err != nil {
			s.logger.Println("Error unmarshalling comment:", err)
			continue
		}
		handleMessage(&comment)
	}
}

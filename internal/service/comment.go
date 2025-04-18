package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
)

type CommentService struct {
	storage      *storage.CommentStorage
	redis        *redis.Client
	logger       *log.Logger
	user_storage *storage.UserStorage
}

func NewCommentService(storage *storage.CommentStorage, user_storage *storage.UserStorage, redis *redis.Client, logger *log.Logger) *CommentService {
	return &CommentService{
		storage:      storage,
		redis:        redis,
		logger:       logger,
		user_storage: user_storage,
	}
}

func (s *CommentService) CreateComment(ctx context.Context, comment *models.Comment) error {
	comment.ID = primitive.NewObjectID()
	comment.CreatedAt = time.Now()
	comment.Reactions = make(map[string][]primitive.ObjectID)

	// If it's a reply, ensure the parent comment exists within the same post
	if !comment.ReplyTo.IsZero() {
		err := s.storage.GetParentComment(ctx, comment)
		if err != nil {
			return fmt.Errorf("parent comment not found within the same post")
		}
	}

	// Store the comment in MongoDB
	if err := s.storage.CreateComment(ctx, comment); err != nil {
		s.logger.Println("Error storing comment:", err)
		return err
	}

	user, err := s.user_storage.GetUserByID(ctx, comment.UserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			comment.OwnerFullname = "Deleted Account"
			return nil
		}
		return err
	}
	comment.OwnerFullname = user.Fullname
	if len(user.ProfilePics) > 0 {
		comment.OwnerProfilePic = user.ProfilePics[0].Url
	}
	return err

}

func (s *CommentService) ReactToComment(ctx context.Context, reaction *models.Reaction) error {
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

func (s *CommentService) GetCommentByID(ctx context.Context, commentID primitive.ObjectID) (*models.Comment, error) {
	comment, err := s.storage.GetCommentByID(ctx, commentID)
	if err != nil {
		s.logger.Println("Error getting comment by id:", err)
		return nil, err
	}

	s.logger.Println("Comment updated successfully:", comment.ID.Hex())
	return comment, nil
}

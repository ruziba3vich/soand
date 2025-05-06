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

	dto "github.com/ruziba3vich/soand/internal/dtos"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/ruziba3vich/soand/internal/storage"
)

type CommentService struct {
	storage      *storage.CommentStorage
	redis        *redis.Client
	logger       *log.Logger
	user_storage *storage.UserStorage
	file_storage repos.IFIleStoreService
}

func NewCommentService(
	storage *storage.CommentStorage,
	user_storage *storage.UserStorage,
	file_storage repos.IFIleStoreService,
	redis *redis.Client, logger *log.Logger) repos.ICommentService {
	return &CommentService{
		storage:      storage,
		redis:        redis,
		file_storage: file_storage,
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

	if comment.Pictures == nil {
		comment.Pictures = make([]string, 0)
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
	if len(comment.VoiceMessage) > 0 {
		if err := s.fetchVoiceMessage(comment); err != nil {
			return err
		}
	}
	if len(comment.Pictures) > 0 {
		if err := s.fetchPictures(comment); err != nil {
			return err
		}
	}
	return err

}

func (s *CommentService) ReactToComment(ctx context.Context, reaction *models.Reaction) error {
	if err := s.storage.RemoveReactionFromComment(ctx, reaction); err != nil {
		if !errors.Is(err, dto.ErrNotReacted) {
			s.logger.Println(err.Error())
			return err
		}
	}
	if reaction.Incr {
		if err := s.storage.AddReactionToComment(ctx, reaction); err != nil {
			s.logger.Printf("could not react to comment %s by user %s : %s", reaction.CommentId.Hex(), reaction.UserID.Hex(), err.Error())
			return err
		}
	}
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

func (s *CommentService) GetCommentsByPostID(ctx context.Context, postID primitive.ObjectID, page int64, pageSize int64) ([]*models.Comment, error) {
	comments, err := s.storage.GetCommentsByPostID(ctx, postID, page, pageSize)
	if err != nil {
		s.logger.Println("Error fetching comments:", err)
		return nil, err
	}

	// Iterate over the cursor and process each comment
	for _, comment := range comments {
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
		if len(comment.VoiceMessage) > 0 {
			if err := s.fetchVoiceMessage(comment); err != nil {
				return nil, err
			}
		}
		if len(comment.Pictures) > 0 {
			if err := s.fetchPictures(comment); err != nil {
				return nil, err
			}
		}
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
	if len(comment.VoiceMessage) > 0 {
		if err := s.fetchVoiceMessage(comment); err != nil {
			return nil, err
		}
	}

	if len(comment.Pictures) > 0 {
		if err := s.fetchPictures(comment); err != nil {
			return nil, err
		}
	}

	// Check if the user's profile is hidden
	user, err := s.user_storage.GetUserByID(ctx, comment.UserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			comment.UserID = primitive.NilObjectID
			comment.OwnerFullname = "Deleted Account"
			return comment, nil
		}
		return nil, err
	}

	// If the user's profile is private, clear the UserID
	if user.HiddenProfile {
		comment.OwnerFullname = "Anonim user"
		comment.UserID = primitive.NilObjectID // Set to zero value to hide it
	}
	s.logger.Println("Comment updated successfully:", comment.ID.Hex())
	return comment, nil
}

func (s *CommentService) fetchVoiceMessage(comment *models.Comment) (err error) {
	comment.VoiceMessage, err = s.file_storage.GetFile(comment.VoiceMessage)
	if err != nil {
		return err
	}
	return nil
}

func (s *CommentService) fetchPictures(comment *models.Comment) (err error) {
	for i := range comment.Pictures {
		comment.Pictures[i], err = s.file_storage.GetFile(comment.Pictures[i])
		if err != nil {
			return err
		}
	}
	return nil
}

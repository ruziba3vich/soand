package service

import (
	"context"
	"log"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/ruziba3vich/soand/internal/storage"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatService struct {
	storage *storage.ChatStorage
	logger  *log.Logger
}

func NewChatService(storage *storage.ChatStorage, logger *log.Logger) repos.IChatService {
	return &ChatService{
		storage: storage,
		logger:  logger,
	}
}

// CreateMessage creates a new message and publishes it to Redis
func (s *ChatService) CreateMessage(ctx context.Context, message *models.Message) error {
	s.logger.Printf("Creating message from %s to %s", message.SenderID.Hex(), message.RecipientID.Hex())
	err := s.storage.CreateMessage(ctx, message)
	if err != nil {
		s.logger.Printf("Failed to create message: %v", err)
		return err
	}
	s.logger.Printf("Message created and published: %s", message.ID.Hex())
	return nil
}

// GetMessagesBetweenUsers retrieves paginated messages between two users
func (s *ChatService) GetMessagesBetweenUsers(ctx context.Context, userID, otherUserID primitive.ObjectID, page, pageSize int64) ([]*models.Message, error) {
	s.logger.Printf("Fetching messages between %s and %s (page %d, size %d)", userID.Hex(), otherUserID.Hex(), page, pageSize)
	messages, err := s.storage.GetMessagesBetweenUsers(ctx, userID, otherUserID, page, pageSize)
	if err != nil {
		s.logger.Printf("Failed to fetch messages: %v", err)
		return nil, err
	}
	s.logger.Printf("Retrieved %d messages", len(messages))
	return messages, nil
}

// UpdateMessageText updates the text of a message and notifies via Redis
func (s *ChatService) UpdateMessageText(ctx context.Context, messageID primitive.ObjectID, newText string) error {
	s.logger.Printf("Updating message %s with new text", messageID.Hex())
	err := s.storage.UpdateMessageText(ctx, messageID, newText)
	if err != nil {
		s.logger.Printf("Failed to update message %s: %v", messageID.Hex(), err)
		return err
	}
	s.logger.Printf("Message %s updated successfully", messageID.Hex())
	return nil
}

// DeleteMessage deletes a message and notifies via Redis
func (s *ChatService) DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error {
	s.logger.Printf("Deleting message %s", messageID.Hex())
	err := s.storage.DeleteMessage(ctx, messageID)
	if err != nil {
		s.logger.Printf("Failed to delete message %s: %v", messageID.Hex(), err)
		return err
	}
	s.logger.Printf("Message %s deleted successfully", messageID.Hex())
	return nil
}

func (s *ChatService) GetMessageByID(ctx context.Context, messageID primitive.ObjectID) (*models.Message, error) {
	s.logger.Printf("Fetching message %s", messageID.Hex())
	message, err := s.storage.GetMessageByID(ctx, messageID)
	if err != nil {
		s.logger.Printf("Failed to fetch message %s: %v", messageID.Hex(), err)
		return nil, err
	}
	s.logger.Printf("Message %s fetched successfully", messageID.Hex())
	return message, nil
}

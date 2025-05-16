package service

import (
	"context"
	"log"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PinnedChatsService struct {
	storage repos.IPinnedChatsService
	logger  *log.Logger
}

func NewPinnedChatService(storage repos.IPinnedChatsService, logger *log.Logger) *PinnedChatsService {
	return &PinnedChatsService{
		storage: storage,
		logger:  logger,
	}
}

func (s *PinnedChatsService) SetPinned(ctx context.Context, userID primitive.ObjectID, chatID primitive.ObjectID, pinned bool) error {
	if err := s.storage.SetPinned(ctx, userID, chatID, pinned); err != nil {
		s.logger.Println(logrus.Fields{
			"user_id": userID,
			"chat_id": chatID,
			"error":   err.Error(),
		})
		return err
	}
	return nil
}

func (s *PinnedChatsService) GetPinnedChatsByUser(ctx context.Context, userID primitive.ObjectID, page int64, pageSize int64) ([]*models.PinnedChat, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	skip := (page - 1) * pageSize

	rawChats, err := s.storage.GetPinnedChatsByUser(ctx, userID, page, skip, pageSize)
	if err != nil {
		s.logger.Println(logrus.Fields{
			"user_id":   userID,
			"page":      page,
			"page_size": pageSize,
			"error":     err.Error(),
		})
		return nil, err
	}

	var chats []*models.PinnedChat
	for _, raw := range rawChats {
		chat := &models.PinnedChat{}

		if chatID, ok := raw["chat_id"].(primitive.ObjectID); ok {
			chat.ChatId = chatID.Hex()
		}
		if pinned, ok := raw["pinned"].(bool); ok {
			if pinned {
				chat.Pinned = pinned
				chats = append(chats, chat)
			}
		}
	}

	return chats, nil
}

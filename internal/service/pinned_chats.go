package service

import (
	"context"
	"log"

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

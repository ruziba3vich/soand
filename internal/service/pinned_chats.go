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
	storage     repos.IPinnedChatsService
	logger      *log.Logger
	postService *PostService
}

func NewPinnedChatService(storage repos.IPinnedChatsService, postService *PostService, logger *log.Logger) *PinnedChatsService {
	return &PinnedChatsService{
		storage:     storage,
		logger:      logger,
		postService: postService,
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

func (s *PinnedChatsService) GetPinnedChatsByUser(ctx context.Context, userID primitive.ObjectID, page int64, pageSize int64) ([]*models.Post, error) {
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

	response := []*models.Post{}

	for _, raw := range rawChats {

		chatID, ok := raw["chat_id"].(primitive.ObjectID)
		if ok {
			if pinned, ok := raw["pinned"].(bool); ok {
				if pinned {
					chat, err := s.postService.GetPost(ctx, chatID)
					if err == nil {
						response = append(response, chat)
					}
				}
			}
		}
	}

	return response, nil
}

package repos

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IPinnedChatsService interface {
	GetPinnedChatsByUser(ctx context.Context, userID primitive.ObjectID, page int64, skip int64, pageSize int64) ([]bson.M, error)
	SetPinned(ctx context.Context, userID primitive.ObjectID, chatID primitive.ObjectID, pinned bool) error
}

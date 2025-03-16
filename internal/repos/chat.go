package repos

import (
	"context"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IChatService interface {
	CreateMessage(ctx context.Context, message *models.Message) error
	GetMessagesBetweenUsers(ctx context.Context, userID, otherUserID primitive.ObjectID, page, pageSize int64) ([]*models.Message, error)
	UpdateMessageText(ctx context.Context, messageID primitive.ObjectID, newText string) error
	DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error
}

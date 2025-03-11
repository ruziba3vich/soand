package repos

import (
	"context"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	ICommentService interface {
		CreateComment(context.Context, *models.Comment) error
		DeleteComment(context.Context, primitive.ObjectID, primitive.ObjectID) error
		GetCommentsByPostID(context.Context, primitive.ObjectID, int64, int64) ([]models.Comment, error)
		UpdateCommentText(context.Context, primitive.ObjectID, primitive.ObjectID, string) error
		GetCommentByID(context.Context, primitive.ObjectID) (*models.Comment, error)
	}
)

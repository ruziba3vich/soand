package repos

import (
	"context"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	UserRepo interface {
		CreateUser(context.Context, *models.User) (string, error)
		DeleteUser(context.Context, primitive.ObjectID) error
		GetUserByID(context.Context, primitive.ObjectID) (*models.User, error)
		GetUserByUsername(context.Context, string) (*models.User, error)
		UpdateFullname(context.Context, primitive.ObjectID, string) error
		UpdatePassword(context.Context, primitive.ObjectID, string, string) error
		UpdateUsername(context.Context, primitive.ObjectID, string) error
		ValidateJWT(string) (string, error)
		LoginUser(context.Context, string, string) (string, error)
		ChangeProfileVisibility(context.Context, primitive.ObjectID, bool) error
		SetBio(context.Context, primitive.ObjectID, string) error
	}
)

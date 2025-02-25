package repos

import (
	"context"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	UserRepo interface {
		CreateUser(ctx context.Context, user *models.User) (string, error)
		DeleteUser(ctx context.Context, userID primitive.ObjectID) error
		GetUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
		GetUserByUsername(ctx context.Context, username string) (*models.User, error)
		UpdateFullname(ctx context.Context, userID primitive.ObjectID, newFullname string) error
		UpdatePassword(ctx context.Context, userID primitive.ObjectID, oldPassword string, newPassword string) error
		UpdateUsername(ctx context.Context, userID primitive.ObjectID, newUsername string) error
		ValidateJWT(tokenString string) (string, error)
	}
)

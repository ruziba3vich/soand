package repos

import (
	"context"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostServiceInterface defines all methods for the PostService
type IPostService interface {
	CreatePost(ctx context.Context, post *models.Post, deleteAfter int) (primitive.ObjectID, error)
	DeletePost(ctx context.Context, id primitive.ObjectID) error
	EnsureTTLIndex(ctx context.Context) error
	GetAllPosts(ctx context.Context, page int64, pageSize int64) ([]models.Post, error)
	GetPost(ctx context.Context, id primitive.ObjectID) (*models.Post, error)
	UpdatePost(ctx context.Context, id primitive.ObjectID, updaterID primitive.ObjectID, update bson.M) error
	SearchPostsByTitle(ctx context.Context, query string, page, pageSize int64) ([]models.Post, error)
	LikeOrDislikePost(ctx context.Context, userId primitive.ObjectID, postId primitive.ObjectID, count int) error
	// ReactToPost(ctx context.Context, postId primitive.ObjectID, userId primitive.ObjectID, reaction string, add bool) error
}

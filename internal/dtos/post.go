package dto

import (
	"errors"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostRequest represents the request payload for creating a post
type PostRequest struct {
	Description string   `json:"description" binding:"required"`
	CreatorId   string   `json:"creator_id"`
	DeleteAfter int      `json:"delete_after" binding:"required"`
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
}

// ToPost converts PostRequest to models.Post
func (p *PostRequest) ToPost() *models.Post {
	creatorId, _ := primitive.ObjectIDFromHex(p.CreatorId)
	return &models.Post{
		Description: p.Description,
		CreatorId:   creatorId,
		Tags:        p.Tags,
		Title:       p.Title,
	}
}

var ErrNotReacted = errors.New("user has not reacted")

/*
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`                // MongoDB ObjectID
    CreatorId   primitive.ObjectID `bson:"creator_id,omitempty" json:"creator_id"` // Creator id
    Pictures    []string           `bson:"pictures" json:"picture"`                // Image URLs or file path
    Tags        []string           `bson:"tags" json:"tags"`                       // List of tags
    Description string             `bson:"description" json:"description"`         // Post description
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`           // Timestamp
    DeleteAt    time.Time          `bson:"delete_at" json:"delete_at"`             // Field for automatic deletion
*/

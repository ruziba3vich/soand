package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	db            *mongo.Collection
	users_storage *UserStorage
	// reactions_storage *ReactionsStorage
}

// NewStorage initializes storage with a MongoDB collection
func NewStorage(
	collection *mongo.Collection,
	users_storage *UserStorage) *Storage {
	return &Storage{
		db:            collection,
		users_storage: users_storage,
	}
}

// CreatePost inserts a new post into the database
func (s *Storage) CreatePost(ctx context.Context, post *models.Post, deleteAfter int) error {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.DeleteAt = post.CreatedAt.Add(time.Duration(deleteAfter) * time.Hour)
	if post.Pictures == nil {
		post.Pictures = []string{}
	}
	if post.Tags == nil {
		post.Tags = []string{}
	}
	if post.Reactions == nil {
		post.Reactions = make(map[string]int)
	}
	_, err := s.db.InsertOne(ctx, post)
	return err
}

// GetPost retrieves a post by ID
func (s *Storage) GetPost(ctx context.Context, id primitive.ObjectID) (*models.Post, error) {
	var post models.Post
	err := s.db.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("post not found")
	}
	owner, err := s.users_storage.GetUserByID(ctx, post.CreatorId)
	if err != nil {
		return nil, err
	}
	post.OwnerFullname = owner.Fullname
	if len(owner.ProfilePics) > 0 {
		post.OwnerProfilePic = owner.ProfilePics[0].Url
	}
	return &post, err
}

func (s *Storage) UpdatePost(ctx context.Context, id, updaterID primitive.ObjectID, update bson.M) error {
	// Find the post and check ownership
	post, err := s.GetPost(ctx, id)
	if err != nil {
		return err
	}
	// Check if the updater is the creator
	if post.CreatorId != updaterID {
		return errors.New("only the creator can update this post")
	}

	// Perform the update
	result, err := s.db.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("failed to update post")
	}

	return nil
}

// DeletePost permanently removes a post from the database
func (s *Storage) DeletePost(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.db.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("post not found")
	}
	return nil
}

func (s *Storage) GetAllPosts(ctx context.Context, page, pageSize int64) ([]models.Post, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	skip := int64((page - 1) * pageSize) // Calculate how many documents to skip

	cursor, err := s.db.Find(ctx, bson.M{}, &options.FindOptions{
		Sort:  bson.M{"created_at": -1}, // Sort by newest first
		Skip:  &skip,
		Limit: &pageSize,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	for cursor.Next(ctx) {
		var post models.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, err
		}

		// Fetch owner details
		owner, err := s.users_storage.GetUserByID(ctx, post.CreatorId)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				username := "deleted"
				owner = &models.User{
					Fullname:      "Deleted Account",
					Username:      &username,
					ProfilePics:   []models.ProfilePic{},
					HiddenProfile: false,
				}
				post.OwnerFullname = owner.Fullname

			} else {
				return nil, err
			}
		}

		// Check if the owner's profile is hidden
		if owner.HiddenProfile {
			post.OwnerFullname = "Anonim user"
			post.CreatorId = primitive.NilObjectID // Set to "00000" equivalent
		} else {
			post.OwnerFullname = owner.Fullname
			if len(owner.ProfilePics) > 0 {
				post.OwnerProfilePic = owner.ProfilePics[0].Url
			}
		}
		posts = append(posts, post)
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	if posts != nil {
		return posts, nil
	}

	return []models.Post{}, nil
}

func (s *Storage) SearchPostsByTitle(ctx context.Context, query string, page, pageSize int64) ([]models.Post, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	skip := (page - 1) * pageSize

	// Use $text to perform a full-text search on the title field
	filter := bson.M{
		"$text": bson.M{
			"$search":        query,
			"$caseSensitive": false, // Case-insensitive search
		},
	}

	// Project the relevance score and the post fields
	projection := bson.M{
		"score": bson.M{"$meta": "textScore"},
		// Include other fields as needed
	}

	// Find options: sort by relevance score, paginate
	findOptions := options.Find().
		SetProjection(projection).
		SetSort(bson.M{"score": bson.M{"$meta": "textScore"}}). // Sort by relevance
		SetSkip(skip).
		SetLimit(pageSize)

	// Execute the search
	cursor, err := s.db.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %v", err)
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %v", err)
	}

	return posts, nil
}

func (s *Storage) LikeOrDislikePost(ctx context.Context, userId primitive.ObjectID, postId primitive.ObjectID, count int) error {
	filter := bson.M{"_id": postId}
	update := bson.M{"$inc": bson.M{"likes": count}}

	_, err := s.db.UpdateOne(ctx, filter, update)
	return err
}

// func (s *Storage) ReactToPost(ctx context.Context, postId primitive.ObjectID, userId primitive.ObjectID, reaction string, add bool) error {
// 	var updated bool
// 	var err error
// 	var val int

// 	filter := bson.M{"_id": postId}

// 	if add {
// 		updated, err = s.reactions_storage.AddReaction(ctx, postId, userId)
// 		if err != nil {
// 			return err
// 		}
// 		if updated {
// 			val = 1
// 		} else {
// 			return nil
// 		}
// 	} else {
// 		updated, err = s.reactions_storage.RemoveReaction(ctx, postId, userId)
// 		if err != nil {
// 			return err
// 		}
// 		if updated {
// 			val = -1
// 		} else {
// 			return nil
// 		}
// 	}

// 	if val == 0 {
// 		return nil
// 	}

// 	// Update the reaction count
// 	update := bson.M{
// 		"$inc": bson.M{"reactions." + reaction: val},
// 	}
// 	_, err = s.db.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		return err
// 	}

// 	// Check if the reaction count is now 0 — if so, remove it
// 	var post struct {
// 		Reactions map[string]int `bson:"reactions"`
// 	}
// 	err = s.db.FindOne(ctx, filter).Decode(&post)
// 	if err != nil {
// 		return err
// 	}

// 	if count, ok := post.Reactions[reaction]; ok && count < 1 {
// 		_, err = s.db.UpdateOne(ctx, filter, bson.M{
// 			"$unset": bson.M{"reactions." + reaction: ""},
// 		})
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

package storage

import (
	"context"

	"github.com/ruziba3vich/soand/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongoDB initializes a MongoDB connection and returns a *mongo.Collection
func ConnectMongoDB(ctx context.Context, cfg *config.Config, collectionName string) *mongo.Collection {
	// Set MongoDB client options
	return nil
}

func (s *Storage) EnsureTTLIndex(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"delete_at": 1},                   // Index on delete_at field
		Options: options.Index().SetExpireAfterSeconds(0), // TTL index (MongoDB auto-deletes expired docs)
	}
	_, err := s.db.Indexes().CreateOne(ctx, indexModel)
	return err
}

// NewStorage initializes storage with a MongoDB collection
func NewStorage(collection *mongo.Collection) *Storage {
	return &Storage{db: collection}
}

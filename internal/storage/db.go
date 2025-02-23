package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Storage) EnsureTTLIndex(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"delete_at": 1},                   // Index on delete_at field
		Options: options.Index().SetExpireAfterSeconds(0), // TTL index (MongoDB auto-deletes expired docs)
	}
	_, err := s.DB.Indexes().CreateOne(ctx, indexModel)
	return err
}

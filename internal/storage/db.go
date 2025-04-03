package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ruziba3vich/soand/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongoDB initializes a MongoDB connection and returns a *mongo.Collection
func ConnectMongoDB(ctx context.Context, cfg *config.Config, collectionName string) (*mongo.Collection, error) {
	// Print connection details for debugging (don't include password in production logs)
	fmt.Printf("Attempting to connect to MongoDB at %s with user %s\n", cfg.MongoDB.URI, cfg.MongoDB.User)

	// cfg.MongoDB.URI = "mongodb://mongo:27017/soand"
	// cfg.MongoDB.User = "mongo_user"
	// cfg.MongoDB.Password = "Dost0n1k"

	// Then in your connection code
	credential := options.Credential{
		Username:   cfg.MongoDB.User,
		Password:   cfg.MongoDB.Password,
		AuthSource: "admin",
	}
	clientOptions := options.Client().ApplyURI(cfg.MongoDB.URI).SetAuth(credential)
	clientOptions.SetAuth(credential)

	// Set a timeout for the connection
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database to ensure the connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to ping MongoDB: %v", err)
	}

	fmt.Println("✅ Connected to MongoDB!")

	// Return the specified collection
	return client.Database(cfg.MongoDB.Database).Collection(collectionName), nil
}

func (s *Storage) EnsureTTLIndex(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys: bson.M{"delete_at": 1}, // Create an index on delete_at field
		Options: options.Index().
			SetExpireAfterSeconds(0), // TTL index, MongoDB auto-deletes expired documents
	}

	// Create the index
	_, err := s.db.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	return nil
}

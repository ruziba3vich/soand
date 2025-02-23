package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ruziba3vich/soand/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongoDB initializes a MongoDB connection and returns a *mongo.Collection
func ConnectMongoDB(cfg *config.Config, collectionName string) *mongo.Collection {
	// Set MongoDB client options
	clientOptions := options.Client().ApplyURI(cfg.MongoDB.URI)

	// Create a MongoDB client
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	// Establish a connection with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Verify connection with a ping
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	fmt.Println("Connected to MongoDB successfully!")

	// Return the specified collection
	return client.Database(cfg.MongoDB.Database).Collection(collectionName)
}

func (s *Storage) EnsureTTLIndex(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"delete_at": 1},                   // Index on delete_at field
		Options: options.Index().SetExpireAfterSeconds(0), // TTL index (MongoDB auto-deletes expired docs)
	}
	_, err := s.DB.Indexes().CreateOne(ctx, indexModel)
	return err
}

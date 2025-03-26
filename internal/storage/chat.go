package storage

import (
	"context"
	"fmt"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatStorage struct {
	db *mongo.Collection
}

func NewChatStorage(db *mongo.Collection) *ChatStorage {
	return &ChatStorage{
		db: db,
	}
}

func (s *ChatStorage) CreateMessage(ctx context.Context, message *models.Message) error {
	_, err := s.db.InsertOne(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to save message to MongoDB: %v", err)
	}

	return nil
}

func (s *ChatStorage) GetMessagesBetweenUsers(ctx context.Context, userID, otherUserID primitive.ObjectID, page, pageSize int64) ([]*models.Message, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	skip := (page - 1) * pageSize

	filter := bson.M{
		"$or": []bson.M{
			{"sender_id": userID, "recipient_id": otherUserID},
			{"sender_id": otherUserID, "recipient_id": userID},
		},
	}
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(pageSize).
		SetSkip(skip)

	cursor, err := s.db.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *ChatStorage) UpdateMessageText(ctx context.Context, messageID primitive.ObjectID, newText string) error {
	var message models.Message
	err := s.db.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		return fmt.Errorf("failed to fetch message for update: %v", err)
	}

	filter := bson.M{"_id": messageID}
	update := bson.M{"$set": bson.M{"content": newText}}
	_, err = s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update message in MongoDB: %v", err)
	}

	return nil
}

func (s *ChatStorage) DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error {
	var message models.Message
	err := s.db.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		return fmt.Errorf("failed to fetch message for deletion: %v", err)
	}

	filter := bson.M{"_id": messageID}
	_, err = s.db.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete message from MongoDB: %v", err)
	}

	return nil
}

func (s *ChatStorage) GetMessageByID(ctx context.Context, messageID primitive.ObjectID) (*models.Message, error) {
	var message models.Message
	err := s.db.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("message not found")
		}
		return nil, fmt.Errorf("failed to fetch message: %v", err)
	}
	return &message, nil
}

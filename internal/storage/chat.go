package storage

import (
	"context"

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
	return &ChatStorage{db: db}
}

func (s *ChatStorage) CreateMessage(ctx context.Context, message *models.Message) error {
	_, err := s.db.InsertOne(ctx, message)
	return err
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
	filter := bson.M{"_id": messageID}
	update := bson.M{"$set": bson.M{"content": newText}}
	_, err := s.db.UpdateOne(ctx, filter, update)
	return err
}

func (s *ChatStorage) DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error {
	filter := bson.M{"_id": messageID}
	_, err := s.db.DeleteOne(ctx, filter)
	return err
}

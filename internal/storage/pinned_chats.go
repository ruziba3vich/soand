package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PinnedChat struct {
	db *mongo.Collection
}

func NewPinnedChat(db *mongo.Collection) *PinnedChat {
	return &PinnedChat{
		db: db,
	}
}

func (p *PinnedChat) SetPinned(ctx context.Context, userID, chatID primitive.ObjectID, pinned bool) error {
	filter := bson.M{
		"chat_id": chatID,
		"user_id": userID,
	}
	update := bson.M{"$set": bson.M{"pinned": pinned}}

	opts := options.Update().SetUpsert(true)
	_, err := p.db.UpdateOne(ctx, filter, update, opts)
	return err
}

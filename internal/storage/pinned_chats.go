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

func (p *PinnedChat) GetPinnedChatsByUser(ctx context.Context, userID primitive.ObjectID, page, pageSize int64) ([]bson.M, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	skip := (page - 1) * pageSize

	filter := bson.M{
		"user_id": userID,
		"pinned":  true,
	}

	opts := options.Find().SetSkip(skip).SetLimit(pageSize)

	cursor, err := p.db.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

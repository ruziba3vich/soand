package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/ruziba3vich/soand/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	BackgroundStorage struct {
		storage *FileStorage
		db      *mongo.Collection
	}
)

func NewBackgroundStorage(storage *FileStorage, db *mongo.Collection) *BackgroundStorage {
	return &BackgroundStorage{
		storage: storage,
		db:      db,
	}
}

// GetAllBackgrounds retrieves all background photos with pagination
func (bs *BackgroundStorage) GetAllBackgrounds(page, pageSize int64) ([]models.Background, error) {
	offset := (page - 1) * pageSize
	cursor, err := bs.db.Find(context.TODO(), bson.M{}, &options.FindOptions{
		Skip:  &offset,
		Limit: &pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve backgrounds: %w", err)
	}
	defer cursor.Close(context.TODO())

	var backgrounds []models.Background
	if err := cursor.All(context.TODO(), &backgrounds); err != nil {
		return nil, fmt.Errorf("failed to decode backgrounds: %w", err)
	}
	return backgrounds, nil
}

// GetBackgroundByID retrieves a background photo by ID
func (bs *BackgroundStorage) GetBackgroundByID(id string) (*models.Background, error) {
	var background models.Background
	err := bs.db.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&background)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("background not found")
		}
		return nil, fmt.Errorf("failed to retrieve background: %w", err)
	}
	return &background, nil
}

// CreateBackground uploads a new background photo
func (bs *BackgroundStorage) CreateBackground(file *multipart.FileHeader) (string, error) {
	filename, err := bs.storage.UploadFile(file)
	if err != nil {
		return "", err
	}

	background := models.Background{Filename: filename}
	_, err = bs.db.InsertOne(context.TODO(), background)
	if err != nil {
		bs.storage.DeleteFile(filename)
		return "", fmt.Errorf("failed to save background record: %w", err)
	}

	return filename, nil
}

// DeleteBackground permanently deletes a background photo
func (bs *BackgroundStorage) DeleteBackground(id string) error {
	var background models.Background
	err := bs.db.FindOneAndDelete(context.TODO(), bson.M{"_id": id}).Decode(&background)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("background not found")
		}
		return fmt.Errorf("failed to delete background record: %s", err.Error())
	}

	// Delete file from MinIO
	return bs.storage.DeleteFile(background.Filename)
}

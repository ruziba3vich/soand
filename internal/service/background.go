package service

import (
	"log"
	"mime/multipart"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
)

type BackgroundService struct {
	logger  *log.Logger
	storage *storage.BackgroundStorage
}

func NewBackgroundService(logger *log.Logger, storage *storage.BackgroundStorage) *BackgroundService {
	return &BackgroundService{
		logger:  logger,
		storage: storage,
	}
}

func (s *BackgroundService) CreateBackground(file *multipart.FileHeader) (string, error) {
	s.logger.Println("Creating new background...")

	filename, err := s.storage.CreateBackground(file)
	if err != nil {
		s.logger.Println("Failed to upload background file: ", err)
		return "", err
	}

	s.logger.Println("Background created successfully")
	return filename, nil
}

func (s *BackgroundService) DeleteBackground(id string) error {
	s.logger.Printf("Deleting background with ID: %s\n", id)

	err := s.storage.DeleteBackground(id)
	if err != nil {
		s.logger.Println("Failed to delete background: ", err)
		return err
	}

	s.logger.Println("Background deleted successfully")
	return nil
}

func (s *BackgroundService) GetAllBackgrounds(page int64, pageSize int64) ([]models.Background, error) {
	s.logger.Println("Fetching backgrounds - Page:", page, "PageSize:", pageSize)

	backgrounds, err := s.storage.GetAllBackgrounds(page, pageSize)
	if err != nil {
		s.logger.Println("Failed to fetch backgrounds: ", err)
		return nil, err
	}

	return backgrounds, nil
}

func (s *BackgroundService) GetBackgroundByID(id string) (string, error) {
	s.logger.Printf("Fetching background with ID: %s\n", id)

	background, err := s.storage.GetBackgroundByID(id)
	if err != nil {
		s.logger.Println("Failed to fetch background: ", err)
		return "", err
	}

	return background, nil
}

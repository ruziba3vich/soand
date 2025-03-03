package service

import (
	"mime/multipart"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/sirupsen/logrus"
)

type BackgroundService struct {
	logger  *logrus.Logger
	storage *storage.BackgroundStorage
}

func NewBackgroundService(logger *logrus.Logger, storage *storage.BackgroundStorage) *BackgroundService {
	return &BackgroundService{
		logger:  logger,
		storage: storage,
	}
}

func (s *BackgroundService) CreateBackground(file *multipart.FileHeader) (string, error) {
	s.logger.Info("Creating new background...")

	filename, err := s.storage.CreateBackground(file)
	if err != nil {
		s.logger.Error("Failed to upload background file: ", err)
		return "", err
	}

	s.logger.Info("Background created successfully")
	return filename, nil
}

func (s *BackgroundService) DeleteBackground(id string) error {
	s.logger.Infof("Deleting background with ID: %s", id)

	err := s.storage.DeleteBackground(id)
	if err != nil {
		s.logger.Error("Failed to delete background: ", err)
		return err
	}

	s.logger.Info("Background deleted successfully")
	return nil
}

func (s *BackgroundService) GetAllBackgrounds(page int64, pageSize int64) ([]models.Background, error) {
	s.logger.Infof("Fetching backgrounds - Page: %d, PageSize: %d", page, pageSize)

	backgrounds, err := s.storage.GetAllBackgrounds(page, pageSize)
	if err != nil {
		s.logger.Error("Failed to fetch backgrounds: ", err)
		return nil, err
	}

	return backgrounds, nil
}

func (s *BackgroundService) GetBackgroundByID(id string) (*models.Background, error) {
	s.logger.Infof("Fetching background with ID: %s", id)

	background, err := s.storage.GetBackgroundByID(id)
	if err != nil {
		s.logger.Error("Failed to fetch background: ", err)
		return nil, err
	}

	return background, nil
}

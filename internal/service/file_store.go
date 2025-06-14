package service

import (
	"log"
	"mime/multipart"

	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/ruziba3vich/soand/internal/storage"
)

type (
	FileStoreService struct {
		storage *storage.FileStorage
		logger  *log.Logger
	}
)

func NewFileStoreService(storage *storage.FileStorage, logger *log.Logger) repos.IFIleStoreService {
	return &FileStoreService{
		storage: storage,
		logger:  logger,
	}
}

func (s *FileStoreService) UploadFile(file *multipart.FileHeader) (string, error) {
	path, err := s.storage.UploadFile(file)
	if err != nil {
		s.logger.Println("Error uploading file:", err)
		return "", err
	}
	s.logger.Println("File uploaded successfully to:", path)
	return s.storage.GetFile(path)
}

func (s *FileStoreService) UploadFileFromBytes(data []byte, contentType string) (string, error) {
	response, err := s.storage.UploadFileFromBytes(data, contentType)
	if err != nil {
		s.logger.Println("Error uploading file:", err)
		return "", err
	}
	s.logger.Println("File uploaded successfully to:", response)
	return response, nil
}

func (s *FileStoreService) GetFile(fileID string) (string, error) {
	path, err := s.storage.GetFile(fileID)
	if err != nil {
		s.logger.Println("Error retrieving file:", err)
		return "", err
	}
	s.logger.Println("File retrieved successfully:", path)
	return path, nil
}

func (s *FileStoreService) DeleteFile(fileID string) error {
	err := s.storage.DeleteFile(fileID)
	if err != nil {
		s.logger.Println("Error deleting file:", err)
		return err
	}
	s.logger.Println("File deleted successfully:", fileID)
	return nil
}

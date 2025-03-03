package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/ruziba3vich/soand/pkg/config"
)

type (
	FileStorage struct {
		cfg          *config.Config
		minio_client *minio.Client
	}
)

func NewFileStorage(cfg *config.Config, minio_client *minio.Client) *FileStorage {
	return &FileStorage{
		cfg:          cfg,
		minio_client: minio_client,
	}
}

func (s *FileStorage) UploadFile(file *multipart.FileHeader) (string, error) {

	// Open file
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("Cannot open file: " + err.Error())
	}
	defer f.Close()

	// Generate unique filename
	filename := fmt.Sprintf("%d", time.Now().UnixMilli())
	
	// Upload file to MinIO
	_, err = s.minio_client.PutObject(
		context.Background(),
		s.cfg.MinIO.Bucket,
		filename,
		f,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")},
	)
	if err != nil {
		return "", fmt.Errorf("Failed to upload file: " + err.Error())
	}

	return filename, nil
}

func (s *FileStorage) GetFile(filename string) (string, error) {
	// Check if the file exists
	_, err := s.minio_client.StatObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("File not found: " + err.Error())
	}

	// Generate pre-signed URL
	expiry := time.Hour * 24
	url, err := s.minio_client.PresignedGetObject(context.Background(), s.cfg.MinIO.Bucket, filename, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to get file: " + err.Error())
	}

	return url.String(), nil
}

func (s *FileStorage) DeleteFile(filename string) error {
	// Check if the file exists before attempting deletion
	_, err := s.minio_client.StatObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("File not found: " + err.Error())
	}

	// Delete file from MinIO
	err = s.minio_client.RemoveObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete file: " + err.Error())
	}

	return nil
}

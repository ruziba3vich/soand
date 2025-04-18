package storage

import (
	"bytes"
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
		return "", fmt.Errorf("cannot open file: %s", err.Error())
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
		return "", fmt.Errorf("failed to upload file: %s", err.Error())
	}

	return s.GetFile(filename)
}

func (s *FileStorage) UploadFileFromBytes(data []byte, contentType string) (string, error) {
	// Create a reader from the raw bytes
	reader := bytes.NewReader(data)
	filename := fmt.Sprintf("%d", time.Now().UnixMilli())

	// Upload file to MinIO
	_, err := s.minio_client.PutObject(
		context.Background(),
		s.cfg.MinIO.Bucket,
		filename,
		reader,
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file from bytes: %s", err.Error())
	}

	return s.GetFile(filename)
}

func (s *FileStorage) GetFile(filename string) (string, error) {
	// Check if the file exists
	_, err := s.minio_client.StatObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("file not found: %s", err.Error())
	}

	// Generate pre-signed URL
	expiry := time.Hour * 24
	url, err := s.minio_client.PresignedGetObject(context.Background(), s.cfg.MinIO.Bucket, filename, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get file: %s", err.Error())
	}

	return url.String(), nil
}

func (s *FileStorage) DeleteFile(filename string) error {
	// Check if the file exists before attempting deletion
	_, err := s.minio_client.StatObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("file not found: %s", err.Error())
	}

	// Delete file from MinIO
	err = s.minio_client.RemoveObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %s", err.Error())
	}

	return nil
}

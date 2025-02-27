package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

func (s *FileStorage) UploadFile(file *multipart.FileHeader) error {

	// Open file
	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("Cannot open file: " + err.Error())
	}
	defer f.Close()

	// Generate unique filename
	filename := fmt.Sprintf("%d", time.Now().UnixMilli())

	// Initialize MinIO client
	minioClient, err := minio.New(s.cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.cfg.MinIO.AccessKey, s.cfg.MinIO.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("Failed to connect to MinIO: " + err.Error())
	}

	// Upload file to MinIO
	_, err = minioClient.PutObject(
		context.Background(),
		s.cfg.MinIO.Bucket,
		filename,
		f,
		file.Size,
		minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")},
	)
	if err != nil {
		return fmt.Errorf("Failed to upload file: " + err.Error())
	}

	return nil
}

func (s *FileStorage) GetFile(filename string) (string, error) {

	// Initialize MinIO client
	minioClient, err := minio.New(s.cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.cfg.MinIO.AccessKey, s.cfg.MinIO.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return "", fmt.Errorf("Failed to connect to MinIO: " + err.Error())
	}

	// Check if the file exists
	_, err = minioClient.StatObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.StatObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("File not found: " + err.Error())
	}

	// Generate pre-signed URL
	expiry := time.Hour * 24
	url, err := minioClient.PresignedGetObject(context.Background(), s.cfg.MinIO.Bucket, filename, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to get file: " + err.Error())
	}

	return url.String(), nil
}

func (s *FileStorage) DeleteFile(filename string) error {

	// Initialize MinIO client
	minioClient, err := minio.New(s.cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.cfg.MinIO.AccessKey, s.cfg.MinIO.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("Failed to connect to MinIO: " + err.Error())
	}

	// Check if the file exists before attempting deletion
	_, err = minioClient.StatObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("File not found: " + err.Error())
	}

	// Delete file from MinIO
	err = minioClient.RemoveObject(context.Background(), s.cfg.MinIO.Bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete file: " + err.Error())
	}

	return nil
}

/*
// Ensure bucket exists
	err = createBucket(minioClient, bucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create/check bucket"})
		return
	}

    func createBucket(client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}

	if !exists {
		return client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
	}
	return nil
}
*/

package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/ruziba3vich/soand/internal/middleware"
	limiter "github.com/ruziba3vich/soand/internal/rate_limiter"
	"github.com/ruziba3vich/soand/internal/registerar"
	"github.com/ruziba3vich/soand/internal/service"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/ruziba3vich/soand/pkg/config"
)

func Run(ctx context.Context, logger *log.Logger) error {
	cfg := config.LoadConfig()

	router := gin.Default()

	// Initialize MinIO client
	minio_client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("Failed to connect to MinIO: " + err.Error())
	}

	if err := createBucket(minio_client, cfg.MinIO.Bucket); err != nil {
		return err
	}

	file_storage := storage.NewFileStorage(cfg, minio_client)
	file_store_service := service.NewFileStoreService(file_storage, logger)

	// Background
	background_collection, err := storage.ConnectMongoDB(ctx, cfg, "background_collection")
	background_storage := storage.NewBackgroundStorage(file_storage, background_collection)

	background_service := service.NewBackgroundService(logger, background_storage)

	// Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// users

	user_collection, err := storage.ConnectMongoDB(ctx, cfg, "users_collection")
	if err != nil {
		logger.Println(err)
		return err
	}

	rate_limiter := limiter.NewTokenBucketLimiter(redisClient, 15, 0.25, 1*time.Minute)

	user_storage := storage.NewUserStorage(user_collection, cfg, background_storage)
	user_service := service.NewUserService(user_storage, logger)

	authMiddleware := middleware.NewAuthHandler(user_service, logger, rate_limiter)

	registerar.RegisterUserRoutes(router, user_service, logger, authMiddleware.AuthMiddleware())

	// posts

	posts_collection, err := storage.ConnectMongoDB(ctx, cfg, "posts_collection")
	if err != nil {
		return err
	}

	posts_storage := storage.NewStorage(posts_collection, user_storage)
	posts_service := service.NewPostService(posts_storage, logger)

	registerar.RegisterPostRoutes(
		router,
		posts_service,
		logger,
		file_store_service,
		authMiddleware.AuthMiddleware(),
	)

	if err := posts_storage.EnsureTTLIndex(ctx); err != nil {
		return err
	}

	// Comments
	comments_collection, err := storage.ConnectMongoDB(ctx, cfg, "comments_collection")
	if err != nil {
		logger.Println("Error connecting to comments collection:", err)
		return err
	}

	comments_storage := storage.NewCommentStorage(comments_collection)
	comments_service := service.NewCommentService(comments_storage, redisClient, logger)

	registerar.RegisterCommentRoutes(
		router,
		comments_service,
		file_store_service,
		logger,
		redisClient,
		authMiddleware.AuthMiddleware(),
		authMiddleware.WebSocketAuthMiddleware(),
	)

	registerar.RegisterBackgroundHandler(router, background_service, logger)

	// direct messages

	chat_collection, err := storage.ConnectMongoDB(ctx, cfg, "chat_collection")
	chat_storage := storage.NewChatStorage(chat_collection, redisClient)
	chat_service := service.NewChatService(chat_storage, logger)
	registerar.RegisterChatHandler(
		router,
		chat_service,
		file_store_service,
		logger,
		redisClient,
		authMiddleware.AuthMiddleware(),
		authMiddleware.WebSocketAuthMiddleware(),
	)

	return router.Run(":7777")
}

// Ensure bucket exists
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

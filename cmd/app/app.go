package app

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ruziba3vich/soand/internal/middleware"
	"github.com/ruziba3vich/soand/internal/registerar"
	"github.com/ruziba3vich/soand/internal/service"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/ruziba3vich/soand/pkg/config"
)

func Run(ctx context.Context, logger *log.Logger) error {
	cfg := config.LoadConfig()

	router := gin.Default()

	// users

	user_collection, err := storage.ConnectMongoDB(ctx, cfg, "users_collection")
	if err != nil {
		logger.Println(err)
		return err
	}

	user_storage := storage.NewUserStorage(user_collection, cfg)
	user_service := service.NewUserService(user_storage, logger)

	authMiddleware := middleware.NewAuthHandler(user_service, logger)

	registerar.RegisterUserRoutes(router, user_service, logger, authMiddleware.AuthMiddleware())

	// posts

	posts_collection, err := storage.ConnectMongoDB(ctx, cfg, "posts_collection")
	if err != nil {
		return err
	}

	posts_storage := storage.NewStorage(posts_collection, user_storage)
	posts_service := service.NewPostService(posts_storage, logger)

	registerar.RegisterPostRoutes(router, posts_service, logger, authMiddleware.AuthMiddleware())

	if err := posts_storage.EnsureTTLIndex(ctx); err != nil {
		return err
	}

	// Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Comments
	comments_collection, err := storage.ConnectMongoDB(ctx, cfg, "comments_collection")
	if err != nil {
		logger.Println("Error connecting to comments collection:", err)
		return err
	}

	comments_storage := storage.NewCommentStorage(comments_collection)
	comments_service := service.NewCommentService(comments_storage, redisClient, logger)

	registerar.RegisterCommentRoutes(router, comments_service, logger, redisClient, authMiddleware.AuthMiddleware(), authMiddleware.WebSocketAuthMiddleware())

	return router.Run(":7777")
}

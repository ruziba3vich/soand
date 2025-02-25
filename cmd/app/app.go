package app

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
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

	authMiddleware := middleware.AuthMiddleware(user_service, logger)

	registerar.RegisterUserRoutes(router, user_service, logger, authMiddleware)

	// posts

	posts_collection, err := storage.ConnectMongoDB(ctx, cfg, "posts_collection")
	if err != nil {
		return err
	}

	posts_storage := storage.NewStorage(posts_collection, user_storage)
	posts_service := service.NewPostService(posts_storage, logger)

	registerar.RegisterPostRoutes(router, posts_service, logger, authMiddleware)

	if err := posts_storage.EnsureTTLIndex(ctx); err != nil {
		return err
	}

	return router.Run(":7777")
}

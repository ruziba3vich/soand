package app

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	handler "github.com/ruziba3vich/soand/internal/http"
	"github.com/ruziba3vich/soand/internal/service"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/ruziba3vich/soand/pkg/config"
)

func Run(ctx context.Context, logger *log.Logger) error {
	cfg := config.LoadConfig()
	posts_collection, err := storage.ConnectMongoDB(ctx, cfg, "posts_collection")
	if err != nil {
		return err
	}

	posts_storage := storage.NewStorage(posts_collection)
	posts_service := service.NewPostService(posts_storage, logger)

	posts_handler := handler.NewPostHandler(posts_service, logger)

	if err := posts_storage.EnsureTTLIndex(ctx); err != nil {
		return err
	}

	router := gin.Default()
	posts_handler.RegisterRoutes(router)

	return router.Run(":7777")
}

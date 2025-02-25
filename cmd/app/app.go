package app

import (
	"context"

	"github.com/gin-gonic/gin"
	handler "github.com/ruziba3vich/soand/internal/http"
	"github.com/ruziba3vich/soand/internal/service"
	"github.com/ruziba3vich/soand/internal/storage"
	"github.com/ruziba3vich/soand/pkg/config"
	"github.com/sirupsen/logrus"
)

func Run(ctx context.Context, logger *logrus.Logger) error {
	cfg := config.LoadConfig()
	collection, err := storage.ConnectMongoDB(ctx, cfg, "posts_collection")
	if err != nil {
		return err
	}

	storage := storage.NewStorage(collection)
	service := service.NewPostService(storage, logger)

	handler := handler.NewPostHandler(service, logger)

	if err := storage.EnsureTTLIndex(ctx); err != nil {
		return err
	}

	router := gin.Default()
	handler.RegisterRoutes(router)

	return router.Run(":7777")
}

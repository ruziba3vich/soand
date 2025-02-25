package registerar

import (
	"log"

	"github.com/gin-gonic/gin"
	handler "github.com/ruziba3vich/soand/internal/http"
	"github.com/ruziba3vich/soand/internal/middleware"
	"github.com/ruziba3vich/soand/internal/repos"
)

func RegisterUserRoutes(r *gin.Engine, userRepo repos.UserRepo, logger *log.Logger) {
	userHandler := handler.NewUserHandler(userRepo, logger)

	authMiddleware := middleware.AuthMiddleware(userRepo, logger)

	userRoutes := r.Group("/users")
	{
		userRoutes.POST("/", userHandler.CreateUser)
		userRoutes.POST("/validate", userHandler.ValidateJWT)

		// Protected routes (require authentication)
		userRoutes.Use(authMiddleware)
		userRoutes.DELETE("/:id", userHandler.DeleteUser)
		userRoutes.GET("/:id", userHandler.GetUserByID)
		userRoutes.GET("/username/:username", userHandler.GetUserByUsername)
		userRoutes.PATCH("/fullname", userHandler.UpdateFullname)
		userRoutes.PATCH("/password", userHandler.UpdatePassword)
		userRoutes.PATCH("/username", userHandler.UpdateUsername)
	}
}

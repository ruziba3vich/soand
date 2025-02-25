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
		userRoutes.POST("/login", userHandler.LoginUser)
		userRoutes.DELETE("/:id", authMiddleware(userHandler.DeleteUser))
		userRoutes.GET("/:id", authMiddleware(userHandler.GetUserByID))
		userRoutes.GET("/username/:username", authMiddleware(userHandler.GetUserByUsername))
		userRoutes.PATCH("/fullname", authMiddleware(userHandler.UpdateFullname))
		userRoutes.PATCH("/password", authMiddleware(userHandler.UpdatePassword))
		userRoutes.PATCH("/username", authMiddleware(userHandler.UpdateUsername))
	}
}

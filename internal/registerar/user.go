package registerar

import (
	"log"

	"github.com/gin-gonic/gin"
	handler "github.com/ruziba3vich/soand/internal/http"
	"github.com/ruziba3vich/soand/internal/repos"
)

func RegisterUserRoutes(r *gin.Engine, userRepo repos.UserRepo, logger *log.Logger, authMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {
	userHandler := handler.NewUserHandler(userRepo, logger)

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
		userRoutes.PATCH("/visibility", authMiddleware(userHandler.ChangeProfileVisibility))

	}
}

func RegisterPostRoutes(r *gin.Engine, postRepo repos.IPostService, logger *log.Logger, authMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {
	h := handler.NewPostHandler(postRepo, logger)

	posts := r.Group("/posts")
	{
		posts.POST("", authMiddleware(h.CreatePost))
		posts.GET("", h.GetPost)                           // Get post by query param "id"
		posts.GET("/all", h.GetAllPosts)                   // Get all posts with pagination
		posts.PUT("/:id", authMiddleware(h.UpdatePost))    // Update post by ID
		posts.DELETE("/:id", authMiddleware(h.DeletePost)) // Delete post by ID
	}
}

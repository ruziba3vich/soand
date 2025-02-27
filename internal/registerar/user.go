package registerar

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	handler "github.com/ruziba3vich/soand/internal/http"
	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/ruziba3vich/soand/internal/service"
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

func RegisterPostRoutes(
	r *gin.Engine,
	postRepo repos.IPostService,
	logger *log.Logger,
	file_service repos.IFIleStoreService,
	authMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {
	h := handler.NewPostHandler(postRepo, logger, file_service)

	posts := r.Group("/posts")
	{
		posts.POST("", authMiddleware(h.CreatePost))
		posts.GET("", h.GetPost)                           // Get post by query param "id"
		posts.GET("/all", h.GetAllPosts)                   // Get all posts with pagination
		posts.PUT("/:id", authMiddleware(h.UpdatePost))    // Update post by ID
		posts.DELETE("/:id", authMiddleware(h.DeletePost)) // Delete post by ID
	}
}

func RegisterCommentRoutes(
	r *gin.Engine,
	commentService *service.CommentService,
	file_service repos.IFIleStoreService,
	logger *log.Logger,
	redis *redis.Client,
	authMiddleware func(gin.HandlerFunc) gin.HandlerFunc,
	wsMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {
	commentHandler := handler.NewCommentHandler(commentService, file_service, logger, redis)

	commentRoutes := r.Group("/comments")
	{
		commentRoutes.GET("/ws", wsMiddleware(commentHandler.HandleWebSocket))             // WebSocket endpoint
		commentRoutes.GET("/:post_id", commentHandler.GetCommentsByPostID)                 // Fetch comments with pagination
		commentRoutes.PATCH("/:comment_id", authMiddleware(commentHandler.UpdateComment))  // Update comment text
		commentRoutes.DELETE("/:comment_id", authMiddleware(commentHandler.DeleteComment)) // Delete comment
	}
}

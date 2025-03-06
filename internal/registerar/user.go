package registerar

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	_ "github.com/ruziba3vich/soand/docs"
	handler "github.com/ruziba3vich/soand/internal/http"
	"github.com/ruziba3vich/soand/internal/repos"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterUserRoutes(r *gin.Engine, userRepo repos.UserRepo, logger *log.Logger, authMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {
	r.Use(CORSMiddleware())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	userHandler := handler.NewUserHandler(userRepo, logger)
	r.GET("/", userHandler.Home)

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
		userRoutes.PATCH("/bio", authMiddleware(userHandler.SetBio))
		userRoutes.PATCH("/background", authMiddleware(userHandler.SetBackgroundPic))
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
	commentService repos.ICommentService,
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

func RegisterBackgroundHandler(
	r *gin.Engine,
	background_service repos.IBackgroundService,
	logger *log.Logger,
) {

	background_handler := handler.NewBackgroundHandler(background_service, logger)

	backgroundRoutes := r.Group("/backgrounds")

	backgroundRoutes.POST("/", background_handler.CreateBackground)
	backgroundRoutes.GET("/", background_handler.GetAllBackgrounds)
	backgroundRoutes.GET("/:id", background_handler.GetBackgroundByID)
	backgroundRoutes.DELETE("/:id", background_handler.DeleteBackground)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://45.130.164.130:7777") // Replace with your origin
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // Cache for 24 hours

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

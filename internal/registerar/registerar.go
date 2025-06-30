package registerar

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	_ "github.com/ruziba3vich/soand/docs"
	handler "github.com/ruziba3vich/soand/internal/http"
	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/ruziba3vich/soand/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterUserRoutes(r *gin.Engine, userRepo repos.UserRepo, file_store repos.IFIleStoreService, logger *log.Logger, authMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {
	r.Use(CORSMiddleware())
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	userHandler := handler.NewUserHandler(userRepo, file_store, logger)
	r.GET("/", userHandler.Home)

	userRoutes := r.Group("/users")
	{
		userRoutes.POST("/", userHandler.CreateUser)
		userRoutes.POST("/login", userHandler.LoginUser)
		userRoutes.POST("profile/pic", authMiddleware(userHandler.AddProfilePicture))
		userRoutes.DELETE("profile/pic", authMiddleware(userHandler.DeleteProfilePicture))
		userRoutes.DELETE("/:id", authMiddleware(userHandler.DeleteUser))
		userRoutes.GET("/me", authMiddleware(userHandler.GetUserMe))
		userRoutes.GET("/:id", userHandler.GetUserByID)
		userRoutes.GET("/username/:username", userHandler.GetUserByUsername)
		userRoutes.GET("profile/pic", userHandler.GetProfilePictures)
		// userRoutes.PATCH("/fullname", authMiddleware(userHandler.UpdateFullname))
		userRoutes.PUT("/update", authMiddleware(userHandler.UpdateUser))
		userRoutes.PATCH("/password", authMiddleware(userHandler.UpdatePassword))
		userRoutes.PATCH("/username", authMiddleware(userHandler.UpdateUsername))
		// userRoutes.PATCH("/visibility", authMiddleware(userHandler.ChangeProfileVisibility))
		// userRoutes.PATCH("/bio", authMiddleware(userHandler.SetBio))
		userRoutes.PATCH("/background", authMiddleware(userHandler.SetBackgroundPic))
	}
}

func RegisterPostRoutes(
	r *gin.Engine,
	postRepo repos.IPostService,
	logger *log.Logger,
	authMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {
	h := handler.NewPostHandler(postRepo, logger)

	posts := r.Group("/posts")
	{
		posts.POST("", authMiddleware(h.CreatePost))
		posts.POST("search/title", h.SearchPostsByTitle)
		posts.POST("/like", authMiddleware(h.LikePostHandler))
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
	wsMiddleware func(gin.HandlerFunc) gin.HandlerFunc,
	commentMiddleware func(gin.HandlerFunc) gin.HandlerFunc) {

	commentHandler := handler.NewCommentHandler(commentService, file_service, logger, redis)

	commentRoutes := r.Group("/comments")
	{
		commentRoutes.POST("/react", authMiddleware(commentHandler.ReactToComment))
		commentRoutes.GET("/ws", wsMiddleware(commentHandler.HandleWebSocket))                // WebSocket endpoint
		commentRoutes.GET("/:post_id", commentMiddleware(commentHandler.GetCommentsByPostID)) // Fetch comments with pagination
		commentRoutes.PATCH("/:comment_id", authMiddleware(commentHandler.UpdateComment))     // Update comment text
		commentRoutes.DELETE("/:comment_id", authMiddleware(commentHandler.DeleteComment))    // Delete comment
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

func RegisterChatHandler(
	r *gin.Engine,
	service repos.IChatService,
	fileService repos.IFIleStoreService,
	logger *log.Logger,
	redis *redis.Client,
	authMiddleware func(gin.HandlerFunc) gin.HandlerFunc,
	wsMiddleware func(gin.HandlerFunc) gin.HandlerFunc,
) {
	chat_handler := handler.NewChatHandler(service, fileService, logger, redis)

	chat_handler_routes := r.Group("/chat")

	chat_handler_routes.GET("direct", wsMiddleware(chat_handler.HandleChatWebSocket))
	chat_handler_routes.GET("direct/messages", authMiddleware(chat_handler.GetMessages))
	chat_handler_routes.PATCH("update", wsMiddleware(chat_handler.UpdateMessage))
	chat_handler_routes.DELETE("dlete", wsMiddleware(chat_handler.DeleteMessage))
}

func RegisterFileStorageHandler(r *gin.Engine, file_service repos.IFIleStoreService, logger *log.Logger) {
	file_getter_handler := handler.NewFIleGetterHandler(file_service, logger)

	r.POST("/upload/file/soand/secure", file_getter_handler.UploadFile)
	r.GET("get/file/by/query", file_getter_handler.GetFileById)
}

func RegisterPinnedChatsHandler(r *gin.Engine, pinnedChatService *service.PinnedChatsService, authMiddleware func(gin.HandlerFunc) gin.HandlerFunc, logger *log.Logger) {
	pinnedChatHandler := handler.NewPinnedChatsHandler(pinnedChatService, logger)

	r.POST("/chats/pin", authMiddleware(pinnedChatHandler.SetChatPinned))
	r.GET("/chats/pinned", authMiddleware(pinnedChatHandler.GetPinnedChats))
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://soand.prodonik.uz") // Allow only this origin
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

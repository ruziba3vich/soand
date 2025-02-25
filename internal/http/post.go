package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dto "github.com/ruziba3vich/soand/internal/dtos"
	"github.com/ruziba3vich/soand/internal/repos"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostHandler struct that handles HTTP requests
type PostHandler struct {
	service repos.IPostService
	logger  *logrus.Logger
}

// NewPostHandler initializes a new PostHandler with a service and logger
func NewPostHandler(service repos.IPostService, logger *logrus.Logger) *PostHandler {
	return &PostHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers the post-related endpoints to the given router
func (h *PostHandler) RegisterRoutes(router *gin.RouterGroup) {
	posts := router.Group("/posts")
	{
		posts.POST("", h.CreatePost)
		posts.GET("", h.GetPost)           // Get post by query param "id"
		posts.GET("/all", h.GetAllPosts)   // Get all posts with pagination
		posts.PUT("/:id", h.UpdatePost)    // Update post by ID
		posts.DELETE("/:id", h.DeletePost) // Delete post by ID
	}
}

// CreatePost handles creating a new post
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req dto.PostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err.Error()).Error("Invalid request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Convert PostRequest to models.Post
	post := req.ToPost()

	// Call service to create post
	id, err := h.service.CreatePost(c.Request.Context(), post, req.DeleteAfter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Post created successfully", "id": id.Hex()})
}

// GetPost handles retrieving a single post by ID
func (h *PostHandler) GetPost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Invalid post ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	post, err := h.service.GetPost(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetAllPosts handles retrieving all posts with pagination
func (h *PostHandler) GetAllPosts(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	posts, err := h.service.GetAllPosts(c.Request.Context(), stringToInt64(page), stringToInt64(pageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// UpdatePost handles updating an existing post
func (h *PostHandler) UpdatePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Invalid post ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.WithField("error", err.Error()).Error("Invalid update payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update payload"})
		return
	}

	updaterID := primitive.NewObjectID() // Assume it's coming from auth context
	if err := h.service.UpdatePost(c.Request.Context(), id, updaterID, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully"})
}

// DeletePost handles deleting a post by ID
func (h *PostHandler) DeletePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Invalid post ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	if err := h.service.DeletePost(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// Helper function to convert string to int64
func stringToInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 1
	}
	return val
}

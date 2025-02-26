package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type CommentHandler struct {
	service repos.ICommentService
	logger  *log.Logger
	redis   *redis.Client
}

func NewCommentHandler(service repos.ICommentService, logger *log.Logger, redis *redis.Client) *CommentHandler {
	return &CommentHandler{
		service: service,
		logger:  logger,
		redis:   redis,
	}
}

// Handle WebSocket connections for real-time comments
func (h *CommentHandler) HandleWebSocket(c *gin.Context) {
	// Extract post ID from query parameters
	postID := c.Query("post_id")
	if postID == "" {
		h.logger.Println("Missing post ID in WebSocket request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id is required"})
		return
	}

	// Upgrade HTTP to WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	h.logger.Println("New WebSocket client connected for post:", postID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to Redis channel for this specific post
	pubsub := h.redis.Subscribe(ctx, "comments:"+postID)
	defer pubsub.Close()

	// Goroutine to listen for messages from Redis and send to WebSocket client
	go func() {
		for msg := range pubsub.Channel() {
			h.logger.Println("Received message from Redis for post", postID, ":", msg.Payload)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				h.logger.Println("Error sending message to WebSocket client:", err)
				cancel() // Cancel context to stop subscription
				return
			}
		}
	}()

	// Listen for new comments from WebSocket client
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.logger.Println("WebSocket connection closed:", err)
			break
		}

		var comment models.Comment
		if err := json.Unmarshal(msg, &comment); err != nil {
			h.logger.Println("Invalid comment format:", err)
			continue
		}

		// Ensure comment belongs to the correct post
		if comment.PostID.Hex() != postID {
			h.logger.Println("Comment post ID mismatch:", comment.PostID.Hex(), "Expected:", postID)
			continue
		}

		comment.ID = primitive.NewObjectID()
		comment.CreatedAt = time.Now()

		// Save comment to DB
		if err := h.service.CreateComment(ctx, &comment); err != nil {
			h.logger.Println("Error saving comment:", err)
			continue
		}

		// Publish the comment to Redis channel for this post
		commentJSON, _ := json.Marshal(comment)
		h.redis.Publish(ctx, "comments:"+postID, string(commentJSON))

		h.logger.Println("New comment published to post", postID, "ID:", comment.ID.Hex())
	}
}

// GetCommentsByPostID retrieves all comments for a post with pagination
func (h *CommentHandler) GetCommentsByPostID(c *gin.Context) {
	postIDStr := c.Param("post_id")
	postID, err := primitive.ObjectIDFromHex(postIDStr)
	if err != nil {
		h.logger.Println("Invalid post ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	// Get pagination params
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)

	comments, err := h.service.GetCommentsByPostID(c.Request.Context(), postID, page, pageSize)
	if err != nil {
		h.logger.Println("Failed to fetch comments:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch comments"})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// UpdateComment updates the text of a comment
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		h.logger.Println("Invalid comment ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	// Extract user ID from context (Assuming AuthMiddleware sets user_id)
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Println("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userObjectID, ok := userID.(primitive.ObjectID)
	if !ok {
		h.logger.Println("Invalid user ID format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		NewText string `json:"new_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Println("Invalid request body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err = h.service.UpdateCommentText(c.Request.Context(), commentID, userObjectID, req.NewText)
	if err != nil {
		h.logger.Println("Failed to update comment:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment updated successfully"})
}

// DeleteComment removes a comment
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		h.logger.Println("Invalid comment ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	// Extract user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Println("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userObjectID, ok := userID.(primitive.ObjectID)
	if !ok {
		h.logger.Println("Invalid user ID format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = h.service.DeleteComment(c.Request.Context(), commentID, userObjectID)
	if err != nil {
		h.logger.Println("Failed to delete comment:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted successfully"})
}

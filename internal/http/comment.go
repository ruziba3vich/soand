package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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

var supportedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// Supported MIME types
var supportedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

/*
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return r.Header.Get("Origin") == "https://trusted-domain.com" },
}
*/

type CommentHandler struct {
	service      repos.ICommentService
	file_service repos.IFIleStoreService
	logger       *log.Logger
	redis        *redis.Client
}

func NewCommentHandler(
	service repos.ICommentService,
	file_service repos.IFIleStoreService,
	logger *log.Logger,
	redis *redis.Client) *CommentHandler {
	return &CommentHandler{
		service:      service,
		file_service: file_service,
		logger:       logger,
		redis:        redis,
	}
}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
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
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					h.logger.Println("Redis subscription error:", err)
					time.Sleep(5 * time.Second) // Retry after delay
					continue
				}
				h.logger.Println("Received message from Redis for post", postID, ":", msg.Payload)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					h.logger.Println("Error sending message to WebSocket client:", err)
					cancel() // Cancel context to stop subscription
					return
				}
			}
		}
	}()

	// Listen for new comments from WebSocket client
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			h.logger.Println("WebSocket connection closed:", err)
			break
		}

		var comment models.Comment

		// Handle JSON comment message
		if messageType == websocket.TextMessage {
			if err := json.Unmarshal(msg, &comment); err != nil {
				h.logger.Println("Invalid comment format:", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid comment format"}`))
				continue
			}
		} else if messageType == websocket.BinaryMessage {
			// Handle binary messages (e.g., images or voice messages)
			fileHeader := &multipart.FileHeader{
				Filename: fmt.Sprintf("%d", time.Now().UnixMilli()), // Unique filename
				Size:     int64(len(msg)),                           // File size
			}

			fileURL, err := h.file_service.UploadFile(fileHeader)
			if err != nil {
				h.logger.Println("Error uploading file:", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "file upload failed"}`))
				continue
			}

			// Validate file extension
			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			if !supportedImageExtensions[ext] {
				h.logger.Println("Unsupported image type:", ext)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unsupported image type"}`))
				continue
			}

			// Validate MIME type
			mimeType := http.DetectContentType(msg)
			if !supportedMimeTypes[mimeType] {
				h.logger.Println("Unsupported MIME type:", mimeType)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unsupported MIME type"}`))
				continue
			}

			// Check file type and store accordingly
			if strings.HasSuffix(fileHeader.Filename, ".jpg") || strings.HasSuffix(fileHeader.Filename, ".png") {
				// Append multiple images to the list
				comment.Pictures = append(comment.Pictures, fileURL)
			} else if strings.HasSuffix(fileHeader.Filename, ".mp3") || strings.HasSuffix(fileHeader.Filename, ".wav") {
				// Only allow **one** voice message
				if comment.VoiceMessage == "" {
					comment.VoiceMessage = fileURL
				} else {
					h.logger.Println("Multiple voice messages detected. Only one is allowed.")
					conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "only one voice message allowed"}`))
					continue
				}
			} else {
				h.logger.Println("Unsupported file type:", fileHeader.Filename)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unsupported file type"}`))
				continue
			}
		}

		// Validate post ID
		if comment.PostID.IsZero() || comment.PostID.Hex() != postID {
			h.logger.Println("Comment post ID mismatch or missing:", comment.PostID.Hex(), "Expected:", postID)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid post ID"}`))
			continue
		}

		// Extract user ID
		userID, err := getUserIdFromRequest(c)
		if err != nil {
			h.logger.Println("Failed to extract user ID:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unauthorized"}`))
			continue
		}

		comment.ID = primitive.NewObjectID()
		comment.CreatedAt = time.Now()
		comment.UserID = userID

		// Save to DB
		if err := h.service.CreateComment(ctx, &comment); err != nil {
			h.logger.Println("Error saving comment:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "could not save comment"}`))
			continue
		}

		// Publish comment to Redis
		commentJSON, err := json.Marshal(comment)
		if err != nil {
			h.logger.Println("Error marshaling comment:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "internal server error"}`))
			continue
		}
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

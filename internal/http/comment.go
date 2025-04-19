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

// PendingComment tracks a comment being built across multiple messages
type (
	CommentResponse struct {
		Data map[string]any `json:"data"`
	}
	pendingComment struct {
		Comment models.Comment
	}
)

// HandleWebSocket handles WebSocket connections for real-time comments
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

	// Extract user ID from request
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		h.logger.Println("Failed to extract user ID:", err)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unauthorized"}`))
		return
	}

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
				var comment models.Comment
				if err := json.Unmarshal([]byte(msg.Payload), &comment); err != nil {
					h.logger.Println("error converting comment: ", msg.Payload)
					return
				}
				response := CommentResponse{
					Data: map[string]any{
						"comment": comment,
						"user_id": userID,
					},
				}
				jsonBytes, err := json.Marshal(response)
				if err != nil {
					h.logger.Println("error marshaling response: ", response)
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
					h.logger.Println("Error sending message to WebSocket client:", err)
					cancel() // Cancel context to stop subscription
					return
				}
			}
		}
	}()

	// Store pending comments per connection
	pending := make(map[*websocket.Conn]pendingComment)
	defer delete(pending, conn) // Cleanup on disconnect

	// Listen for messages from WebSocket client
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			h.logger.Println("WebSocket connection closed:", err)
			break
		}

		// Only handle text messages (JSON)
		if messageType != websocket.TextMessage {
			h.logger.Println("Unsupported message type:", messageType)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "only JSON messages are supported"}`))
			continue
		}

		// Get or initialize the pending comment for this connection
		current, exists := pending[conn]
		if !exists {
			current = pendingComment{Comment: models.Comment{}}
		}

		// Parse the JSON comment message
		if err := json.Unmarshal(msg, &current.Comment); err != nil {
			h.logger.Println("Invalid comment format:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid comment format"}`))
			continue
		}

		// Validate and set required fields
		postObjectID, err := primitive.ObjectIDFromHex(postID)
		if err != nil || current.Comment.PostID.IsZero() {
			current.Comment.PostID = postObjectID
		} else if current.Comment.PostID.Hex() != postID {
			h.logger.Println("Comment post ID mismatch:", current.Comment.PostID.Hex(), "Expected:", postID)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid post ID"}`))
			continue
		}

		// Ensure the comment has at least some content
		if current.Comment.Text == "" {
			h.logger.Println("Empty comment received")
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "comment text is required"}`))
			continue
		}

		// Set metadata
		current.Comment.ID = primitive.NewObjectID()
		current.Comment.UserID = userID
		current.Comment.CreatedAt = time.Now()

		// Save to database
		if err := h.service.CreateComment(ctx, &current.Comment); err != nil {
			h.logger.Println("Error saving comment:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "could not save comment"}`))
			continue
		}

		// Publish to Redis
		commentJSON, err := json.Marshal(current.Comment)
		if err != nil {
			h.logger.Println("Error marshaling comment:", err)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "internal server error"}`))
			continue
		}
		h.redis.Publish(ctx, "comments:"+postID, string(commentJSON))
		h.logger.Println("New comment published to post", postID, "ID:", current.Comment.ID.Hex())

		// Reset pending comment after saving
		delete(pending, conn)
	}
}

// GetCommentsByPostID retrieves all comments for a post with pagination
// @Summary Get comments by post ID
// @Description Retrieves a paginated list of comments for a specific post
// @Tags comments
// @Produce json
// @Param post_id path string true "Post ID (MongoDB ObjectID)"
// @Param page query string false "Page number (default: 1)"
// @Param pageSize query string false "Number of comments per page (default: 10)"
// @Success 200 {array} interface{} "List of comments"
// @Failure 400 {object} map[string]string "Invalid post ID"
// @Failure 500 {object} map[string]string "Could not fetch comments"
// @Router /posts/{post_id}/comments [get]
func (h *CommentHandler) GetCommentsByPostID(c *gin.Context) {
	userId, _ := getUserIdFromRequest(c)
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

	c.JSON(http.StatusOK, gin.H{"data": map[string]any{
		"comments": comments,
		"user_id":  userId,
	}})
}

func (h *CommentHandler) ReactToComment(c *gin.Context) {
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var req models.Reaction
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: " + err.Error()})
		return
	}

	req.UserID = userId

	if err := h.service.ReactToComment(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "reacted successfully"})
}

// UpdateComment updates the text of a comment
// @Summary Update a comment
// @Description Updates the text of a specific comment for the authenticated user
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param comment_id path string true "Comment ID (MongoDB ObjectID)"
// @Param comment body object{new_text=string} true "New comment text"
// @Success 200 {object} map[string]string "Comment updated successfully"
// @Failure 400 {object} map[string]string "Invalid comment ID or request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Could not update comment"
// @Router /comments/{comment_id} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		h.logger.Println("Invalid comment ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}
	postIdStr := c.Param("post_id")
	// Extract user ID from context (Assuming AuthMiddleware sets user_id)
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		h.logger.Println("User ID not found in context")
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

	// Update the comment and retrieve the associated post_id
	err = h.service.UpdateCommentText(c.Request.Context(), commentID, userID, req.NewText)
	if err != nil {
		h.logger.Println("Failed to update comment:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update comment"})
		return
	}

	// Publish the update to Redis
	updateMessage := map[string]string{
		"action":     "update",
		"comment_id": commentID.Hex(),
		"new_text":   req.NewText,
		"updated_at": time.Now().String(),
	}
	messageJSON, err := json.Marshal(updateMessage)
	if err != nil {
		h.logger.Println("Error marshaling update message:", err)
		// Don't fail the request, just log the error
	} else {
		h.redis.Publish(c.Request.Context(), "comments:"+postIdStr, string(messageJSON))
		h.logger.Println("Published comment update for comment", commentID.Hex(), "to post", postIdStr)
	}

	c.JSON(http.StatusOK, gin.H{"data": "comment updated successfully"})
}

// DeleteComment removes a comment
// @Summary Delete a comment
// @Description Deletes a specific comment for the authenticated user
// @Tags comments
// @Produce json
// @Security BearerAuth
// @Param comment_id path string true "Comment ID (MongoDB ObjectID)"
// @Success 200 {object} map[string]string "Comment deleted successfully"
// @Failure 400 {object} map[string]string "Invalid comment ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Could not delete comment"
// @Router /comments/{comment_id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err := primitive.ObjectIDFromHex(commentIDStr)
	if err != nil {
		h.logger.Println("Invalid comment ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment ID"})
		return
	}

	// Extract user ID from context
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Fetch the comment to get the post_id before deletion (if needed)
	comment, err := h.service.GetCommentByID(c.Request.Context(), commentID) // New service method
	if err != nil {
		h.logger.Println("Failed to fetch comment for deletion:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete comment"})
		return
	}

	// Delete the comment
	err = h.service.DeleteComment(c.Request.Context(), commentID, userID)
	if err != nil {
		h.logger.Println("Failed to delete comment:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete comment"})
		return
	}

	// Publish the deletion to Redis
	deleteMessage := map[string]interface{}{
		"action":     "delete",
		"comment_id": commentID.Hex(),
		"deleted_at": time.Now(),
	}
	messageJSON, err := json.Marshal(deleteMessage)
	if err != nil {
		h.logger.Println("Error marshaling delete message:", err)
		// Don't fail the request, just log the error
	} else {
		postID := comment.PostID.Hex() // From fetched comment
		h.redis.Publish(c.Request.Context(), "comments:"+postID, string(messageJSON))
		h.logger.Println("Published comment deletion for comment", commentID.Hex(), "to post", postID)
	}

	c.JSON(http.StatusOK, gin.H{"data": "comment deleted successfully"})
}

// extFromMime maps MIME types to file extensions
func extFromMime(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav":
		return ".wav"
	default:
		return ".bin" // Fallback for unknown types
	}
}

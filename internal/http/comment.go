package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Connection struct to track users per post
type Connection struct {
	PostID primitive.ObjectID
	Conn   *websocket.Conn
}

// CommentHandler struct
type CommentHandler struct {
	service repos.ICommentService
	logger  *log.Logger
	mu      sync.Mutex
	clients map[*Connection]bool
}

// NewCommentHandler initializes a new comment handler
func NewCommentHandler(service repos.ICommentService, logger *log.Logger) *CommentHandler {
	return &CommentHandler{
		service: service,
		logger:  logger,
		clients: make(map[*Connection]bool),
	}
}

// WebSocket connection handler
func (h *CommentHandler) HandleWebSocket(c *gin.Context) {
	postID, err := primitive.ObjectIDFromHex(c.Param("post_id"))
	if err != nil {
		h.logger.Println("Invalid post ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	client := &Connection{PostID: postID, Conn: conn}

	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	h.logger.Println("New WebSocket connection for post:", postID.Hex())

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.logger.Println("WebSocket read error:", err)
			h.mu.Lock()
			delete(h.clients, client)
			h.mu.Unlock()
			return
		}

		var comment models.Comment
		if err := json.Unmarshal(message, &comment); err != nil {
			h.logger.Println("Invalid comment format:", err)
			continue
		}

		comment.PostID = postID
		comment.CreatedAt = time.Now()

		if err := h.service.CreateComment(context.Background(), &comment); err != nil {
			h.logger.Println("Failed to store comment:", err)
			continue
		}

		h.broadcastComment(comment)
	}
}

// Broadcasts comments to all WebSocket clients
func (h *CommentHandler) broadcastComment(comment models.Comment) {
	h.mu.Lock()
	defer h.mu.Unlock()

	message, err := json.Marshal(comment)
	if err != nil {
		h.logger.Println("Failed to serialize comment:", err)
		return
	}

	for client := range h.clients {
		if client.PostID == comment.PostID {
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				h.logger.Println("Failed to send message:", err)
				client.Conn.Close()
				delete(h.clients, client)
			}
		}
	}
}

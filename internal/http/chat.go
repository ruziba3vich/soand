package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatHandler struct {
	service     repos.IChatService
	fileService repos.IFIleStoreService
	logger      *log.Logger
	redis       *redis.Client
}

func NewChatHandler(service repos.IChatService, fileService repos.IFIleStoreService, logger *log.Logger, redis *redis.Client) *ChatHandler {
	return &ChatHandler{
		service:     service,
		fileService: fileService,
		logger:      logger,
		redis:       redis,
	}
}

type pendingMessage struct {
	Message models.Message
}

func (h *ChatHandler) HandleChatWebSocket(c *gin.Context) {
	recipientIDStr := c.Query("recipient_id")
	if recipientIDStr == "" {
		h.logger.Println("Missing recipient ID in WebSocket request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipient_id is required"})
		return
	}
	recipientID, err := primitive.ObjectIDFromHex(recipientIDStr)
	if err != nil {
		h.logger.Println("Invalid recipient ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipient_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Println("WebSocket upgrade failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
		return
	}
	defer conn.Close()

	senderID, err := getUserIdFromRequest(c)
	if err != nil {
		h.logger.Println("Failed to extract sender ID:", err)
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unauthorized"}`))
		return
	}

	h.logger.Println("New chat WebSocket client connected:", senderID.Hex(), "to", recipientID.Hex())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chatChannel := fmt.Sprintf("chat:%s:%s", min(senderID.Hex(), recipientID.Hex()), max(senderID.Hex(), recipientID.Hex()))
	pubsub := h.redis.Subscribe(ctx, chatChannel)
	defer pubsub.Close()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					h.logger.Println("Redis subscription error:", err)
					time.Sleep(5 * time.Second)
					continue
				}
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
					h.logger.Println("Error sending message to WebSocket client:", err)
					cancel()
					return
				}
			}
		}
	}()

	pending := make(map[*websocket.Conn]pendingMessage)
	defer delete(pending, conn)
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			h.logger.Println("WebSocket connection closed:", err)
			break
		}

		current, exists := pending[conn]
		if !exists {
			current = pendingMessage{
				Message: models.Message{
					ID:          primitive.NewObjectID(),
					SenderID:    senderID,
					RecipientID: recipientID,
					CreatedAt:   time.Now(),
				},
			}
		}

		if messageType == websocket.TextMessage {
			var incoming struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal(msg, &incoming); err != nil {
				h.logger.Println("Invalid message format:", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid message format"}`))
				continue
			}
			current.Message.Content = incoming.Content
			pending[conn] = current
		} else if messageType == websocket.BinaryMessage {
			mimeType := http.DetectContentType(msg)
			fileURL, err := h.fileService.UploadFileFromBytes(msg, mimeType)
			if err != nil {
				h.logger.Println("Error uploading file:", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "file upload failed"}`))
				continue
			}

			if !supportedMimeTypes[mimeType] {
				h.logger.Println("Unsupported file type:", mimeType)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "unsupported file type"}`))
				continue
			}

			current.Message.Pictures = append(current.Message.Pictures, fileURL)
			pending[conn] = current
		}

		if current.Message.Content != "" || len(current.Message.Pictures) > 0 {
			if err := h.service.CreateMessage(ctx, &current.Message); err != nil {
				h.logger.Println("Error creating message:", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "could not send message"}`))
				continue
			}

			h.logger.Println("Message sent from", senderID.Hex(), "to", recipientID.Hex(), "with", len(current.Message.Pictures), "files")

			delete(pending, conn)
		}
	}
}

func (h *ChatHandler) UpdateMessage(c *gin.Context) {
	messageIDStr := c.Param("message_id")
	messageID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		h.logger.Println("Invalid message ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID"})
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

	err = h.service.UpdateMessageText(c.Request.Context(), messageID, req.NewText)
	if err != nil {
		h.logger.Println("Error updating message:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message updated successfully"})
}

func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	messageIDStr := c.Param("message_id")
	messageID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		h.logger.Println("Invalid message ID:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message ID"})
		return
	}

	err = h.service.DeleteMessage(c.Request.Context(), messageID)
	if err != nil {
		h.logger.Println("Error deleting message:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "message deleted successfully"})
}

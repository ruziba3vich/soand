package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PinnedChatsHandler struct {
	service *service.PinnedChatsService
	logger  *log.Logger
}

func NewPinnedChatsHandler(service *service.PinnedChatsService, logger *log.Logger) *PinnedChatsHandler {
	return &PinnedChatsHandler{
		service: service,
		logger:  logger,
	}
}

func (h *PinnedChatsHandler) SetChatPinned(c *gin.Context) {
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.PinChatRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	primitiveChatID, err := primitive.ObjectIDFromHex(req.ChatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	if err := h.service.SetPinned(c.Request.Context(), userID, primitiveChatID, req.Pin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

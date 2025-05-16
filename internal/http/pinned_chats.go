package handler

import (
	"log"
	"net/http"
	"strconv"

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

func (h *PinnedChatsHandler) GetPinnedChats(c *gin.Context) {
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)

	response, err := h.service.GetPinnedChatsByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

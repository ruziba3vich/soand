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

// SetChatPinned pins or unpins a chat for the authenticated user
// @Summary Pin or unpin a chat
// @Description Pins or unpins a chat for the authenticated user
// @Tags chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param pinChatRequest body models.PinChatRequest true "Pin chat request"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /chats/pin [post]
func (h *PinnedChatsHandler) SetChatPinned(c *gin.Context) {
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.PinChatRequest
	if err := c.BindJSON(&req); err != nil {
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

// GetPinnedChats gets pinned chats for the authenticated user
// @Summary Get pinned chats
// @Description Retrieves a paginated list of pinned chats for the authenticated user
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} map[string]interface{} "List of pinned chats"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /chats/pinned [get]
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

	if response == nil {
		response = []*models.Post{}
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

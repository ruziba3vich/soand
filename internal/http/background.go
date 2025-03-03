package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
)

// BackgroundHandler handles background-related requests.
type BackgroundHandler struct {
	service repos.IBackgroundService
	logger  *log.Logger
}

// NewBackgroundHandler creates a new BackgroundHandler instance.
func NewBackgroundHandler(service repos.IBackgroundService, logger *log.Logger) *BackgroundHandler {
	return &BackgroundHandler{service: service, logger: logger}
}

// CreateBackground godoc
// @Summary Upload a new background image
// @Description Uploads a new background image and stores it
// @Tags backgrounds
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Background image file"
// @Success 201 {string} string "File uploaded successfully"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /backgrounds [post]
func (h *BackgroundHandler) CreateBackground(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Println("Failed to get file:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}

	id, err := h.service.CreateBackground(file)
	if err != nil {
		h.logger.Println("Failed to create background:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// DeleteBackground godoc
// @Summary Delete a background image
// @Description Deletes a background image by ID
// @Tags backgrounds
// @Param id path string true "Background ID"
// @Success 200 {string} string "Background deleted successfully"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Background not found"
// @Failure 500 {string} string "Internal server error"
// @Router /backgrounds/{id} [delete]
func (h *BackgroundHandler) DeleteBackground(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Println("Invalid ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err := h.service.DeleteBackground(id)
	if err != nil {
		h.logger.Println("Failed to delete background:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete background"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Background deleted successfully"})
}

// GetAllBackgrounds godoc
// @Summary Get all background images
// @Description Retrieves a list of background images with pagination
// @Tags backgrounds
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {array} models.Background
// @Failure 400 {string} string "Invalid query parameters"
// @Failure 500 {string} string "Internal server error"
// @Router /backgrounds [get]
func (h *BackgroundHandler) GetAllBackgrounds(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)

	backgrounds, err := h.service.GetAllBackgrounds(page, pageSize)
	if err != nil {
		h.logger.Println("Failed to get backgrounds:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve backgrounds"})
		return
	}

	c.JSON(http.StatusOK, backgrounds)
}

// GetBackgroundByID godoc
// @Summary Get a background image by ID
// @Description Retrieves a specific background image by its ID
// @Tags backgrounds
// @Param id path string true "Background ID"
// @Success 200 {object} models.Background
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Background not found"
// @Failure 500 {string} string "Internal server error"
// @Router /backgrounds/{id} [get]
func (h *BackgroundHandler) GetBackgroundByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Println("Invalid ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	background, err := h.service.GetBackgroundByID(id)
	if err != nil {
		h.logger.Println("Failed to get background:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve background"})
		return
	}

	c.JSON(http.StatusOK, background)
}

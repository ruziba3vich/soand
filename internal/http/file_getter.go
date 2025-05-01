package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/soand/internal/repos"
)

type (
	FIleStorageHandler struct {
		file_service repos.IFIleStoreService
		logger       *log.Logger
	}
)

func NewFIleGetterHandler(file_service repos.IFIleStoreService, logger *log.Logger) *FIleStorageHandler {
	return &FIleStorageHandler{
		file_service: file_service,
		logger:       logger,
	}
}

// GetFileById retrieves a file by its ID
// @Summary      Get file by ID
// @Description  Retrieves file information using the provided file ID
// @Tags         Files
// @Accept       json
// @Produce      json
// @Param        file_id  query  string  true  "File ID to retrieve"
// @Success      200  {object}  map[string]interface{}  "Returns the file information"
// @Failure      400  {object}  map[string]interface{}  "Missing or invalid file ID"
// @Failure      404  {object}  map[string]interface{}  "File not found"
// @Router       /files [get]
func (h *FIleStorageHandler) GetFileById(c *gin.Context) {
	id := c.Query("file_id")
	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file id has not been specified"})
		return
	}

	file, err := h.file_service.GetFile(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no file found with the provided id " + id})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": file})
}

// UploadFile handles file uploads via form data
// @Summary      Upload a file
// @Description  Uploads a single file to the storage service (MinIO). The file is sent as form data and stored, returning the file URL on success.
// @Tags         Files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "File to upload (Supported formats: any file type supported by MinIO, e.g., images, PDFs, audio. Max size: 10MB recommended)"
// @Success      200  {object}  map[string]interface{}  "Returns the uploaded file URL"
// @Failure      400  {object}  map[string]interface{}  "Invalid file upload or request format"
// @Failure      500  {object}  map[string]interface{}  "Server error during file upload"
// @Router       /files [post]
// @Note        For frontend devs: Send the file in a multipart/form-data request with the key 'file'. Example in JS: `formData.append('file', fileInput.files[0])`. Keep files under 10MB to avoid timeouts. No authentication required for this endpoint (add if needed).
func (h *FIleStorageHandler) UploadFile(c *gin.Context) {
	// Get the file from the form data
	file, err := c.FormFile("file")
	if err != nil {
		h.logger.Println("Invalid file upload:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file upload"})
		return
	}

	// Upload the file to MinIO using file_service
	fileURL, err := h.file_service.UploadFile(file)
	if err != nil {
		h.logger.Println("Failed to upload file to storage:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	// Return the file URL on success
	c.JSON(http.StatusOK, gin.H{
		"data": map[string]string{
			"message": "file uploaded successfully",
			"url":     fileURL,
		},
	})
}

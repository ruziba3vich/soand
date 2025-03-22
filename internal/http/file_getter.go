package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/soand/internal/repos"
)

type (
	FIleGetterHandler struct {
		file_service repos.IFIleStoreService
		logger       *log.Logger
	}
)

func NewFIleGetterHandler(file_service repos.IFIleStoreService, logger *log.Logger) *FIleGetterHandler {
	return &FIleGetterHandler{
		file_service: file_service,
		logger:       logger,
	}
}

func (h *FIleGetterHandler) GetFileById(c *gin.Context) {
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

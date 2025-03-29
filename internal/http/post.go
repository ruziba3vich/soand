package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	dto "github.com/ruziba3vich/soand/internal/dtos"
	"github.com/ruziba3vich/soand/internal/repos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostHandler struct that handles HTTP requests
type PostHandler struct {
	service      repos.IPostService
	user_service repos.UserRepo
	logger       *log.Logger
	file_service repos.IFIleStoreService
}

// NewPostHandler initializes a new PostHandler with a service and logger
func NewPostHandler(service repos.IPostService, logger *log.Logger, file_service repos.IFIleStoreService) *PostHandler {
	return &PostHandler{
		service:      service,
		logger:       logger,
		file_service: file_service,
	}
}

// CreatePost creates a new post with optional file uploads
// @Summary Create a new post
// @Description Creates a post with description, tags, optional delete_after time, and file attachments
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param description formData string true "Post description"
// @Param delete_after formData string false "Time in minutes after which the post will be deleted"
// @Param tags formData string false "Comma-separated list of tags or JSON array"
// @Param tags_json formData string false "JSON stringified array of tags (alternative to tags)"
// @Param files formData file false "Files to upload (multiple allowed)"
// @Success 201 {object} map[string]string "Post created successfully with ID"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	req := dto.PostRequest{
		Description: c.PostForm("description"),
		DeleteAfter: stringToInt(c.PostForm("delete_after")),
		Tags:        c.PostFormArray("tags"),
		Title:       c.PostForm("title"),
	}

	tagsStr := c.PostForm("tags_json")
	if tagsStr != "" {
		if err := json.Unmarshal([]byte(tagsStr), &req.Tags); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tags format"})
			return
		}
	}

	post := req.ToPost()
	post.CreatorId = userId

	form, err := c.MultipartForm()
	if err == nil {
		uploadedFiles := form.File["files"]
		for _, file := range uploadedFiles {
			fileURL, err := h.file_service.UploadFile(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "File upload failed: " + err.Error()})
				return
			}
			post.Pictures = append(post.Pictures, fileURL)
		}
	}

	id, err := h.service.CreatePost(c.Request.Context(), post, req.DeleteAfter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Post created successfully", "id": id.Hex()})
}

// GetPost retrieves a post by its ID
// @Summary Get a post by ID
// @Description Retrieves a single post using its MongoDB ObjectID
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID (MongoDB ObjectID)"
// @Success 200 {object} interface{} "Post details"
// @Failure 400 {object} map[string]string "Invalid post ID format"
// @Failure 404 {object} map[string]string "Post not found"
// @Router /posts [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	idParam := c.Query("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		h.logger.Println("error", err.Error()+idParam+" "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format " + idParam + " " + err.Error()})
		return
	}

	post, err := h.service.GetPost(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetAllPosts retrieves all posts with pagination
// @Summary Get all posts
// @Description Retrieves a paginated list of all posts
// @Tags posts
// @Accept json
// @Produce json
// @Param page query string false "Page number (default: 1)"
// @Param pageSize query string false "Number of posts per page (default: 10)"
// @Success 200 {array} interface{} "List of posts"
// @Failure 500 {object} map[string]string "Failed to retrieve posts"
// @Router /posts/all [get]
func (h *PostHandler) GetAllPosts(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	posts, err := h.service.GetAllPosts(c.Request.Context(), stringToInt64(page), stringToInt64(pageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// UpdatePost updates an existing post
// @Summary Update a post
// @Description Updates a post by ID with new data
// @Tags posts
// @Accept json
// @Security BearerAuth
// @Produce json
// @Param id path string true "Post ID (MongoDB ObjectID)"
// @Param updateData body map[string]interface{} true "Fields to update (e.g., description, tags)"
// @Success 200 {object} map[string]string "Post updated successfully"
// @Failure 400 {object} map[string]string "Invalid post ID or payload"
// @Failure 500 {object} map[string]string "Failed to update post"
// @Router /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		h.logger.Println("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		return
	}

	var updateData map[string]any
	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.logger.Println("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update payload " + err.Error()})
		return
	}

	updaterId, _ := primitive.ObjectIDFromHex(updateData["creator_id"].(string))

	if err := h.service.UpdatePost(c.Request.Context(), id, updaterId, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Post updated successfully"})
}

// DeletePost deletes a post by its ID
// @Summary Delete a post
// @Description Deletes a post using its MongoDB ObjectID
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Post ID (MongoDB ObjectID)"
// @Success 200 {object} map[string]string "Post deleted successfully"
// @Failure 400 {object} map[string]string "Invalid post ID format"
// @Failure 500 {object} map[string]string "Failed to delete post"
// @Router /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		h.logger.Println("error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format "})
		return
	}

	if err := h.service.DeletePost(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Post deleted successfully"})
}

func (h *PostHandler) SearchPostsByTitle(c *gin.Context) {
	var reuest struct {
		Query string `json:"query"`
	}
	if err := c.BindJSON(&reuest); err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: " + err.Error()})
		return
	}
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: " + err.Error()})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: " + err.Error()})
		return
	}

	posts, err := h.service.SearchPostsByTitle(c.Request.Context(), reuest.Query, int64(page), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// Helper function to convert string to int64
func stringToInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 1
	}
	return val
}

func stringToInt(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return num
}

package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin" // Assuming your model is here
	dto "github.com/ruziba3vich/soand/internal/dtos"
	_ "github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos" // Assuming a package for common swagger DTOs
	_ "github.com/ruziba3vich/soand/pkg/swagger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostHandler struct that handles HTTP requests
type PostHandler struct {
	service repos.IPostService
	logger  *log.Logger
}

// NewPostHandler initializes a new PostHandler with a service and logger
func NewPostHandler(service repos.IPostService, logger *log.Logger) *PostHandler {
	return &PostHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePost creates a new post from a JSON payload
// @Summary Create a new post
// @Description Creates a post with description and tags from a JSON body. Note: This version does not support file uploads.
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param postRequest body dto.PostRequest true "Post creation payload"
// @Success 201 {object} swagger.Response{data=models.Post} "Post created successfully"
// @Failure 400 {object} swagger.ErrorResponse "Invalid request payload"
// @Failure 401 {object} swagger.ErrorResponse "Unauthorized"
// @Failure 500 {object} swagger.ErrorResponse "Internal server error"
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.PostRequest

	// This correctly binds a JSON request body. The annotations now match this.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request : " + err.Error()})
		return
	}

	// This part of the code is now unreachable if you are using ShouldBindJSON,
	// because c.PostForm reads from form data, not a JSON body.
	// tagsStr := c.PostForm("tags_json")
	// if tagsStr != "" {
	// 	if err := json.Unmarshal([]byte(tagsStr), &req.Tags); err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tags format"})
	// 		return
	// 	}
	// }

	post := req.ToPost()
	post.CreatorId = userId

	// This code for multipart form handling is incompatible with ShouldBindJSON
	// form, err := c.MultipartForm()
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// files := form.File["files"]

	err = h.service.CreatePost(c.Request.Context(), post, req.DeleteAfter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": post})
}

// GetPost retrieves a post by its ID
// @Summary Get a post by ID
// @Description Retrieves a single post using its MongoDB ObjectID from a query parameter.
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID (MongoDB ObjectID)" Format(hex)
// @Success 200 {object} models.Post "Post details"
// @Failure 400 {object} swagger.ErrorResponse "Invalid post ID format"
// @Failure 404 {object} swagger.ErrorResponse "Post not found"
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
// @Description Retrieves a paginated list of all posts using query parameters for pagination.
// @Tags posts
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param pageSize query integer false "Number of posts per page" default(10)
// @Success 200 {object} swagger.PaginatedPostsResponse "List of posts"
// @Failure 500 {object} swagger.ErrorResponse "Failed to retrieve posts"
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
// @Description Updates a post by its ID (from the URL path) with data from a JSON body.
// @Description **Security Note:** The current implementation is insecure as it takes `creator_id` from the body and does not check for post ownership.
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Post ID (MongoDB ObjectID)" Format(hex)
// @Param updateRequest body map[string]interface{} true "Fields to update in JSON format"
// @Success 200 {object} swagger.SuccessResponse "Post updated successfully"
// @Failure 400 {object} swagger.ErrorResponse "Invalid post ID or payload"
// @Failure 401 {object} swagger.ErrorResponse "Unauthorized"
// @Failure 500 {object} swagger.ErrorResponse "Failed to update post"
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

	// FIXME: This is a security vulnerability. The updater's ID should come from the request context (token), not the payload.
	updaterId, _ := primitive.ObjectIDFromHex(updateData["creator_id"].(string))

	if err := h.service.UpdatePost(c.Request.Context(), id, updaterId, updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "Post updated successfully"})
}

// DeletePost deletes a post by its ID
// @Summary Delete a post
// @Description Deletes a post using its MongoDB ObjectID from the URL path.
// @Description **Security Note:** This does not currently check if the requester is the owner of the post.
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Post ID (MongoDB ObjectID)" Format(hex)
// @Success 200 {object} swagger.SuccessResponse "Post deleted successfully"
// @Failure 400 {object} swagger.ErrorResponse "Invalid post ID format"
// @Failure 401 {object} swagger.ErrorResponse "Unauthorized"
// @Failure 500 {object} swagger.ErrorResponse "Failed to delete post"
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

// SearchPostsByTitle searches for posts by title
// @Summary Search for posts
// @Description Searches for posts by title (from a JSON body) with pagination (from query parameters).
// @Tags posts
// @Accept json
// @Produce json
// @Param searchRequest body swagger.SearchRequest true "Search query payload"
// @Param page query integer false "Page number" default(1)
// @Param limit query integer false "Number of results per page" default(10)
// @Success 200 {object} swagger.PaginatedPostsResponse "A list of matching posts"
// @Failure 400 {object} swagger.ErrorResponse "Invalid request payload or query parameters"
// @Failure 500 {object} swagger.ErrorResponse "Internal server error"
// @Router /posts/search/title [post]
func (h *PostHandler) SearchPostsByTitle(c *gin.Context) {
	var request struct {
		Query string `json:"query"`
	}
	if err := c.BindJSON(&request); err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: " + err.Error()})
		return
	}
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: invalid page number"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.logger.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request: invalid limit number"})
		return
	}

	posts, err := h.service.SearchPostsByTitle(c.Request.Context(), request.Query, int64(page), int64(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// LikePostHandler handles liking and unliking a post
// @Summary Like or unlike a post
// @Description Submits a like (or removes a like) for a specific post. The post ID is a query param, and the like status is in the JSON body.
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param post_id query string true "ID of the post to like/unlike" Format(hex)
// @Param likeRequest body swagger.LikeRequest true "Like action"
// @Success 200 {object} swagger.SuccessResponse "Action completed successfully"
// @Failure 400 {object} swagger.ErrorResponse "Invalid post ID or request body"
// @Failure 401 {object} swagger.ErrorResponse "Unauthorized"
// @Failure 500 {object} swagger.ErrorResponse "Internal server error"
// @Router /posts/like [post]
func (h *PostHandler) LikePostHandler(c *gin.Context) {
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	postIdStr := c.Query("post_id")
	if len(postIdStr) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no post id provided"})
		return
	}

	postId, err := primitive.ObjectIDFromHex(postIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id provided: " + err.Error()})
		return
	}
	var req struct {
		Like bool `json:"like"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	var count int
	if req.Like {
		count = 1
	} else {
		count = -1
	}

	err = h.service.LikeOrDislikePost(c.Request.Context(), userId, postId, count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "post is liked"})
}

// Helper function to convert string to int64
func stringToInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Default to 1 for page, but for pageSize, 0 might be better to indicate error
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

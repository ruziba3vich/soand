// @title Soand API
// @version 1.0
// @description Soand API Documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "JWT Authorization header using the Bearer scheme. Example: 'Bearer {token}'"
package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler handles user-related API requests
type UserHandler struct {
	repo       repos.UserRepo
	file_store repos.IFIleStoreService
	logger     *log.Logger
}

// NewUserHandler initializes a new UserHandler
func NewUserHandler(repo repos.UserRepo, file_store repos.IFIleStoreService, logger *log.Logger) *UserHandler {
	return &UserHandler{repo: repo, logger: logger, file_store: file_store}
}

// CreateUser handles user creation requests
// @Summary Create a new user
// @Description Creates a new user and returns an authentication token
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User data"
// @Success 200 {object} map[string]string "Token for the created user"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Failed to create user"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		h.logger.Printf("Error parsing user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	user.Status = "basic"
	h.logger.Println(user.Password, len(user.Password))

	token, err := h.repo.CreateUser(c.Request.Context(), &user)
	if err != nil {
		h.logger.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": token})
}

// LoginUser handles user login requests
// @Summary Login a user
// @Description Authenticates a user and returns an authentication token
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body object{username=string,password=string} true "Login credentials"
// @Success 200 {object} map[string]string "Authentication token"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 500 {object} map[string]string "Failed to login user"
// @Router /users/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("Error parsing user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := h.repo.LoginUser(c.Request.Context(), request.Username, request.Password)
	if err != nil {
		h.logger.Printf("Error logging in user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login user " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": token})
}

// DeleteUser handles user deletion requests
// @Summary Delete a user
// @Description Deletes the authenticated user's account
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 204 "User deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to delete user"
// @Router /users [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.repo.DeleteUser(c.Request.Context(), userId); err != nil {
		h.logger.Printf("Error deleting user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "user has successfully been deleted"})
}

// GetUserByID handles retrieving a user by ID
// @Summary Get user by ID
// @Description Retrieves the authenticated user's details by their ID
// @Tags users
// @Produce json
// @Success 200 {object} models.User "User details"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/me [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userId := c.Param("id")
	if len(userId) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid user id provided"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid user id provided"})
		return
	}

	user, err := h.repo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Printf("Error fetching user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// GetUserByUsername handles retrieving a user by username
// @Summary Get user by username
// @Description Retrieves a user's details by their username
// @Tags users
// @Produce json
// @Param username path string true "Username of the user"
// @Success 200 {object} models.User "User details"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/{username} [get]
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := h.repo.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		h.logger.Printf("Error fetching user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// UpdateFullname handles updating a user's full name
// @Summary Update user full name
// @Description Updates the authenticated user's full name
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param fullname body object{new_fullname=string} true "New full name"
// @Success 200 "Full name updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to update fullname"
// @Router /users/fullname [put]
func (h *UserHandler) UpdateFullname(c *gin.Context) {
	var request struct {
		NewFullname string `json:"new_fullname"`
	}

	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("Error parsing fullname update request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.repo.UpdateFullname(c.Request.Context(), userId, request.NewFullname); err != nil {
		h.logger.Printf("Error updating fullname: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update fullname"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

// UpdatePassword handles updating a user's password
// @Summary Update user password
// @Description Updates the authenticated user's password
// @Tags users
// @Accept json
// @Security BearerAuth
// @Produce json
// @Param passwords body object{old_password=string,new_password=string} true "Old and new passwords"
// @Success 200 "Password updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to update password"
// @Router /users/password [put]
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var request struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("Error parsing password update request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.repo.UpdatePassword(c.Request.Context(), userId, request.OldPassword, request.NewPassword); err != nil {
		h.logger.Printf("Error updating password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

// UpdateUsername handles updating a user's username
// @Summary Update user username
// @Description Updates the authenticated user's username
// @Tags users
// @Accept json
// @Security BearerAuth
// @Produce json
// @Param username body object{new_username=string} true "New username"
// @Success 200 "Username updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to update username"
// @Router /users/username [put]
func (h *UserHandler) UpdateUsername(c *gin.Context) {
	var request struct {
		NewUsername string `json:"new_username"`
	}

	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("Error parsing username update request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.repo.UpdateUsername(c.Request.Context(), userId, request.NewUsername); err != nil {
		h.logger.Printf("Error updating username: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update username"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "username updated successfully"})
}

// ChangeProfileVisibility handles changing a user's profile visibility
// @Summary Change profile visibility
// @Description Updates the authenticated user's profile visibility (hidden or visible)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param visibility body object{hidden=boolean} true "Profile visibility status"
// @Success 200 {object} map[string]string "Profile visibility updated"
// @Failure 400 {object} map[string]string "Invalid request payload"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to update profile visibility"
// @Router /users/visibility [put]
func (h *UserHandler) ChangeProfileVisibility(c *gin.Context) {
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Hidden bool `json:"hidden"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Printf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := h.repo.ChangeProfileVisibility(c.Request.Context(), userId, req.Hidden); err != nil {
		h.logger.Printf("Failed to change profile visibility for user %s: %v", userId.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile visibility"})
		return
	}

	h.logger.Printf("Successfully changed profile visibility for user %s", userId.Hex())
	c.JSON(http.StatusOK, gin.H{"data": "profile visibility updated"})
}

// SetBio handles updating a user's bio
// @Summary Set user bio
// @Description Updates the authenticated user's bio
// @Tags users
// @Accept json
// @Security BearerAuth
// @Produce json
// @Param bio body object{bio=string} true "New bio"
// @Success 200 "Bio updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to update bio"
// @Router /users/bio [put]
func (h *UserHandler) SetBio(c *gin.Context) {
	var request struct {
		Bio string `json:"bio"`
	}

	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("Error parsing bio request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.repo.SetBio(c.Request.Context(), userId, request.Bio); err != nil {
		h.logger.Printf("Error updating bio: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func getUserIdFromRequest(c *gin.Context) (primitive.ObjectID, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, fmt.Errorf(" ")
	}

	oid, err := primitive.ObjectIDFromHex(userID.(string))
	return oid, err
}

// SetBackgroundPic handles setting a user's chat background picture
// @Summary Set user background picture
// @Description Sets the authenticated user's background picture
// @Tags users
// @Accept json
// @Security BearerAuth
// @Produce json
// @Param background_pic body object{pic_id=string} true "New background picture ID (e.g., UUID or MinIO object key)"
// @Success 200 {object} map[string]string "Background picture is set successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Failed to update background picture"
// @Router /users/background [put]
func (h *UserHandler) SetBackgroundPic(c *gin.Context) {
	var request struct {
		BackgroundPic string `json:"pic_id"`
	}

	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("Error parsing background picture update request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.repo.SetBackgroundPic(c.Request.Context(), userId, request.BackgroundPic); err != nil {
		h.logger.Printf("Error updating background picture: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update background picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

// Home handles the home endpoint requiring a valid JWT token
// @Summary Get home endpoint
// @Description Returns a successful response if the user is authenticated with a valid JWT token
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "User authenticated successfully"
// @Failure 401 {object} map[string]string "Unauthorized or invalid token"
// @Header 200 {string} Authorization "Bearer <token>" "Required JWT token for authentication"
// @Router / [get]
func (h *UserHandler) Home(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		h.logger.Println("Missing authorization token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	parts := strings.Split(tokenString, " ")

	_, err := h.repo.ValidateJWT(parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	c.Status(http.StatusOK)
}

// AddProfilePicture godoc
// @Summary      Add a new profile picture
// @Description  Uploads a single profile picture for the authenticated user. The file is stored in MinIO, then added to the user's profile in MongoDB if successful.
// @Tags         Profile
// @Accept       multipart/form-data
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token for authentication"
// @Param        picture        formData  file    true  "Profile picture file (Supported formats: JPEG, PNG, GIF, WEBP. Max size: 5MB recommended)"
// @Success      200  {object}  map[string]interface{}  "Returns the uploaded file URL"
// @Failure      400  {object}  map[string]interface{}  "Invalid file upload or request format"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - missing or invalid token"
// @Failure      500  {object}  map[string]interface{}  "Server error during file upload or database update"
// @Router       /profile/picture [post]
// @Note        For frontend devs: Send the file in a multipart/form-data request with the key 'picture'. Example in JS: `formData.append('picture', fileInput.files[0])`. Ensure the file is an image (e.g., .jpg, .png) and keep it under 5MB to avoid timeouts.
func (h *UserHandler) AddProfilePicture(c *gin.Context) {
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		h.logger.Println("Failed to extract user ID:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	file, err := c.FormFile("picture")
	if err != nil {
		h.logger.Println("Invalid file upload:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file upload"})
		return
	}

	fileURL, err := h.file_store.UploadFile(file)
	if err != nil {
		h.logger.Println("Failed to upload file to MinIO:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	f, err := file.Open()
	if err != nil {
		h.logger.Println("Failed to open file:", err)
		if delErr := h.file_store.DeleteFile(fileURL); delErr != nil {
			h.logger.Printf("Failed to clean up file %s from MinIO: %v", fileURL, delErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process file"})
		return
	}
	defer f.Close()

	err = h.repo.AddNewProfilePicture(c.Request.Context(), userID, fileURL)
	if err != nil {
		h.logger.Println("Failed to add profile picture to MongoDB:", err)
		if delErr := h.file_store.DeleteFile(fileURL); delErr != nil {
			h.logger.Printf("Failed to clean up file %s from MinIO: %v", fileURL, delErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": fileURL,
	})
}

// DeleteProfilePicture godoc
// @Summary      Delete a profile picture
// @Description  Removes a profile picture from the user's profile in MongoDB, then deletes it from MinIO. Requires the file URL as a query parameter.
// @Tags         Profile
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token for authentication"
// @Param        file_url       query     string  true  "URL of the profile picture to delete (e.g., '123456789.jpg')"
// @Success      200  {object}  map[string]interface{}  "Confirmation of deletion"
// @Failure      400  {object}  map[string]interface{}  "Missing or invalid file_url"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - missing or invalid token"
// @Failure      500  {object}  map[string]interface{}  "Server error during deletion"
// @Router       /profile/picture [delete]
// @Note        For frontend devs: Pass the file URL (returned from AddProfilePicture) as a query param, e.g., `/profile/picture?file_url=123456789.jpg`. If MinIO deletion fails, the response still succeeds since MongoDB is updated.
func (h *UserHandler) DeleteProfilePicture(c *gin.Context) {
	userID, err := getUserIdFromRequest(c)
	if err != nil {
		h.logger.Println("Failed to extract user ID:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileURL := c.Query("file_url")
	if fileURL == "" {
		h.logger.Println("Missing file_url parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_url is required"})
		return
	}

	if err := h.file_store.DeleteFile(fileURL); err != nil {
		h.logger.Printf("Failed to delete file %s from MinIO, but removed from MongoDB: %v", fileURL, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.repo.DeleteProfilePicture(c.Request.Context(), userID, fileURL)
	if err != nil {
		h.logger.Println("Failed to delete profile picture from MongoDB:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "profile picture deleted"})
}

// GetProfilePictures godoc
// @Summary      Get all profile pictures
// @Description  Retrieves all profile pictures for the authenticated user, sorted by posted date (newest to oldest).
// @Tags         Profile
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token for authentication"
// @Success      200  {object}  map[string]interface{}  "List of profile pictures with URLs and posted dates"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized - missing or invalid token"
// @Failure      500  {object}  map[string]interface{}  "Server error fetching pictures"
// @Router       /profile/pictures [get]
// @Note        For frontend devs: Response includes an array of objects with 'url' (string) and 'posted_at' (ISO 8601 timestamp, e.g., '2025-03-18T12:00:00Z'). Use this to display pics in chronological order.
func (h *UserHandler) GetProfilePictures(c *gin.Context) {
	userIDstr := c.Query("id")
	if len(userIDstr) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no user id provided"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id provided"})
	}

	pics, err := h.repo.GetProfilePictures(c.Request.Context(), userID)
	if err != nil {
		h.logger.Println("Failed to fetch profile pictures:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pics})
}

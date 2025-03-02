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

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/repos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler handles user-related API requests
type UserHandler struct {
	repo   repos.UserRepo
	logger *log.Logger
}

// NewUserHandler initializes a new UserHandler
func NewUserHandler(repo repos.UserRepo, logger *log.Logger) *UserHandler {
	return &UserHandler{repo: repo, logger: logger}
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

	c.JSON(http.StatusOK, gin.H{"token": token})
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

	c.JSON(http.StatusOK, gin.H{"token": token})
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

	c.Status(http.StatusNoContent)
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
	userId, err := getUserIdFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.repo.GetUserByID(c.Request.Context(), userId)
	if err != nil {
		h.logger.Printf("Error fetching user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
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

	c.JSON(http.StatusOK, user)
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

	c.Status(http.StatusOK)
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

	c.Status(http.StatusOK)
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

	c.Status(http.StatusOK)
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
	c.JSON(http.StatusOK, gin.H{"message": "profile visibility updated"})
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

	c.Status(http.StatusOK)
}

func getUserIdFromRequest(c *gin.Context) (primitive.ObjectID, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, fmt.Errorf(" ")
	}

	oid, err := primitive.ObjectIDFromHex(userID.(string))
	return oid, err
}

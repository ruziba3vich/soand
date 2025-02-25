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
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		h.logger.Printf("Error parsing user data: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := h.repo.CreateUser(c.Request.Context(), &user)
	if err != nil {
		h.logger.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// DeleteUser handles user deletion requests
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userId, err := h.getUserIdFromRequest(c)
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
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userId, err := h.getUserIdFromRequest(c)
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
func (h *UserHandler) UpdateFullname(c *gin.Context) {
	var request struct {
		NewFullname string `json:"new_fullname"`
	}

	userId, err := h.getUserIdFromRequest(c)
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
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userId, err := h.getUserIdFromRequest(c)
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
func (h *UserHandler) UpdateUsername(c *gin.Context) {
	var request struct {
		NewUsername string `json:"new_username"`
	}

	userId, err := h.getUserIdFromRequest(c)
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

// ValidateJWT handles JWT validation requests
func (h *UserHandler) ValidateJWT(c *gin.Context) {
	var request struct {
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Printf("Error parsing JWT validation request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID, err := h.repo.ValidateJWT(request.Token)
	if err != nil {
		h.logger.Printf("Invalid JWT token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID})
}

func (h *UserHandler) getUserIdFromRequest(c *gin.Context) (primitive.ObjectID, error) {
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Println("User ID not found in context")
		return primitive.NilObjectID, fmt.Errorf(" ")
	}

	oid, ok := userID.(primitive.ObjectID)
	if !ok {
		h.logger.Println("Invalid user ID type")
		return primitive.NilObjectID, fmt.Errorf(" ")
	}

	return oid, nil
}

package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
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
	user.Status = "basic"
	pp.Println(user)
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

func (h *UserHandler) getUserIdFromRequest(c *gin.Context) (primitive.ObjectID, error) {
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Println("User ID not found in context")
		return primitive.NilObjectID, fmt.Errorf(" ")
	}

	oid, err := primitive.ObjectIDFromHex(userID.(string))

	return oid, err
}

func (h *UserHandler) ChangeProfileVisibility(c *gin.Context) {
	// Extract user ID from context (set by AuthMiddleware)
	userId, err := h.getUserIdFromRequest(c)
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

	if err != nil {
		h.logger.Printf("Invalid user ID format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Call service to update profile visibility
	if err := h.repo.ChangeProfileVisibility(c.Request.Context(), userId, req.Hidden); err != nil {
		h.logger.Printf("Failed to change profile visibility for user %s: %v", userId.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile visibility"})
		return
	}

	h.logger.Printf("Successfully changed profile visibility for user %s", userId.Hex())
	c.JSON(http.StatusOK, gin.H{"message": "profile visibility updated"})
}

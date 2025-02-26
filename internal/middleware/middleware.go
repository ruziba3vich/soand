package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/soand/internal/repos"
)

// AuthHandler holds dependencies for authentication
type AuthHandler struct {
	userRepo repos.UserRepo
	logger   *log.Logger
}

// NewAuthHandler initializes and returns an AuthHandler instance
func NewAuthHandler(userRepo repos.UserRepo, logger *log.Logger) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// AuthMiddleware validates JWT and sets user ID before executing the given handlers
func (a *AuthHandler) AuthMiddleware() func(gin.HandlerFunc) gin.HandlerFunc {
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			tokenString := c.GetHeader("Authorization")
			if tokenString == "" {
				a.logger.Println("Missing authorization token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}

			parts := strings.Split(tokenString, " ")

			userID, err := a.userRepo.ValidateJWT(parts[1])
			if err != nil {
				a.logger.Println("Invalid token:", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			// Set user ID in context
			c.Set("userID", userID)

			// Call the actual handler
			handler(c)
		}
	}
}

func (a *AuthHandler) WebSocketAuthMiddleware() func(gin.HandlerFunc) gin.HandlerFunc {
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			token := c.GetHeader("Authorization") // Extract token from WebSocket header

			if token == "" {
				a.logger.Println("Missing WebSocket authorization token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}

			userID, err := a.userRepo.ValidateJWT(token)
			if err != nil {
				a.logger.Println("Invalid WebSocket token:", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			// Set user ID in context
			c.Set("userID", userID)

			// Call the actual handler
			handler(c)
		}
	}
}

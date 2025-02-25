package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ruziba3vich/soand/internal/repos"
)

// AuthMiddleware validates JWT and sets user ID before executing the given handlers
func AuthMiddleware(userRepo repos.UserRepo, logger *log.Logger) func(gin.HandlerFunc) gin.HandlerFunc {
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			tokenString := c.GetHeader("Authorization")
			if tokenString == "" {
				logger.Println("Missing authorization token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}

			parts := strings.Split(tokenString, " ")

			userID, err := userRepo.ValidateJWT(parts[1])
			if err != nil {
				logger.Println("Invalid token:", err)
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

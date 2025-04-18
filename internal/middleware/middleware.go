package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	limiter "github.com/ruziba3vich/soand/internal/rate_limiter"
	"github.com/ruziba3vich/soand/internal/repos"
)

// AuthHandler holds dependencies for authentication
type AuthHandler struct {
	userRepo repos.UserRepo
	logger   *log.Logger
	limiter  *limiter.TokenBucketLimiter
}

// NewAuthHandler initializes and returns an AuthHandler instance
func NewAuthHandler(userRepo repos.UserRepo, logger *log.Logger, limiter *limiter.TokenBucketLimiter) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		logger:   logger,
		limiter:  limiter,
	}
}

// AuthMiddleware validates JWT and sets user ID before executing the given handlers
func (a *AuthHandler) AuthMiddleware() func(gin.HandlerFunc) gin.HandlerFunc {
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			ip := c.ClientIP() // Get user IP for rate limiting

			allowed, err := a.limiter.AllowRequest(c, ip)
			if err != nil {
				a.logger.Println("Rate limiter error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			if !allowed {
				a.logger.Println("Rate limit exceeded for IP:", ip)
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
				c.Abort()
				return
			}

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
			ctx := context.Background()
			ip := c.ClientIP() // Get user IP for rate limiting

			allowed, err := a.limiter.AllowRequest(ctx, ip)
			if err != nil {
				a.logger.Println("Rate limiter error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			if !allowed {
				a.logger.Println("Rate limit exceeded for IP:", ip)
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
				c.Abort()
				return
			}

			tokenString := c.GetHeader("Authorization")
			if tokenString == "" {
				a.logger.Println("Missing authorization token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}

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

// CommentsMiddleware validates JWT and sets user ID before executing the given handlers
func (a *AuthHandler) CommentsMiddleware() func(gin.HandlerFunc) gin.HandlerFunc {
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			ip := c.ClientIP() // Get user IP for rate limiting

			allowed, err := a.limiter.AllowRequest(c, ip)
			if err != nil {
				a.logger.Println("Rate limiter error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			if !allowed {
				a.logger.Println("Rate limit exceeded for IP:", ip)
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
				c.Abort()
				return
			}

			tokenString := c.GetHeader("Authorization")
			if tokenString == "" {
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

// GetCOmmentsMiddleware validates JWT and sets user ID before executing the given handlers
func (a *AuthHandler) GetCOmmentsMiddleware() func(gin.HandlerFunc) gin.HandlerFunc {
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			ip := c.ClientIP() // Get user IP for rate limiting

			allowed, err := a.limiter.AllowRequest(c, ip)
			if err != nil {
				a.logger.Println("Rate limiter error:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				c.Abort()
				return
			}

			if !allowed {
				a.logger.Println("Rate limit exceeded for IP:", ip)
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
				c.Abort()
				return
			}

			tokenString := c.GetHeader("Authorization")
			if tokenString == "" {
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

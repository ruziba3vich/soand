// Package swagger contains data transfer objects for swagger documentation
package swagger

import "github.com/ruziba3vich/soand/internal/models"

// This is a generic response wrapper for successful API calls
// that contain a single data object.
// It allows us to use the swagger annotation:
// @Success 200 {object} swagger.Response{data=models.Post}
type Response struct {
	Data interface{} `json:"data"`
}

// ErrorResponse is the response sent when an error occurs.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse is used for simple success messages without a complex data object.
type SuccessResponse struct {
	Data string `json:"data"`
}

// PaginatedPostsResponse defines the structure for a paginated list of posts.
type PaginatedPostsResponse struct {
	Data []models.Post `json:"data"`
}

// SearchRequest defines the expected body for the search endpoint.
type SearchRequest struct {
	Query string `json:"query" example:"My awesome post title"`
}

// LikeRequest defines the expected body for the like/unlike endpoint.
type LikeRequest struct {
	Like bool `json:"like" example:"true"`
}

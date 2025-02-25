package service

import (
	"context"
	"log"

	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/internal/storage"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService handles business logic for users
type UserService struct {
	storage *storage.UserStorage
	logger  *log.Logger
}

// NewUserService initializes UserService
func NewUserService(storage *storage.UserStorage, logger *log.Logger) *UserService {
	return &UserService{storage: storage, logger: logger}
}

// CreateUser creates a new user and returns a JWT token
func (s *UserService) CreateUser(ctx context.Context, user *models.User) (string, error) {
	s.logger.Println("Creating new user...")

	token, err := s.storage.CreateUser(ctx, user)
	if err != nil {
		s.logger.Printf("Error creating user: %v\n", err)
		return "", err
	}

	s.logger.Printf("User created successfully, ID: %s\n", user.ID.Hex())
	return token, nil
}

// CreateUser creates a new user and returns a JWT token
func (s *UserService) LoginUser(ctx context.Context, username, password string) (string, error) {

	token, err := s.storage.Login(ctx, username, password)
	if err != nil {
		s.logger.Printf("Error creating user: %v\n", err)
		return "", err
	}

	s.logger.Printf("User logged in successfully, ID: %s\n", username)
	return token, nil
}

// DeleteUser removes a user from the database
func (s *UserService) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	s.logger.Printf("Deleting user with ID: %s\n", userID.Hex())

	err := s.storage.DeleteUser(ctx, userID)
	if err != nil {
		s.logger.Printf("Error deleting user: %v\n", err)
		return err
	}

	s.logger.Println("User deleted successfully")
	return nil
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	s.logger.Printf("Fetching user by ID: %s\n", userID.Hex())

	user, err := s.storage.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Printf("Error fetching user: %v\n", err)
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves a user by their username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	s.logger.Printf("Fetching user by username: %s\n", username)

	user, err := s.storage.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Printf("Error fetching user: %v\n", err)
		return nil, err
	}

	return user, nil
}

// UpdateFullname updates a user's full name
func (s *UserService) UpdateFullname(ctx context.Context, userID primitive.ObjectID, newFullname string) error {
	s.logger.Printf("Updating fullname for user ID: %s\n", userID.Hex())

	err := s.storage.UpdateFullname(ctx, userID, newFullname)
	if err != nil {
		s.logger.Printf("Error updating fullname: %v\n", err)
		return err
	}

	s.logger.Println("User fullname updated successfully")
	return nil
}

// UpdatePassword updates a user's password after checking the old one
func (s *UserService) UpdatePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error {
	s.logger.Printf("Updating password for user ID: %s\n", userID.Hex())

	err := s.storage.UpdatePassword(ctx, userID, oldPassword, newPassword)
	if err != nil {
		s.logger.Printf("Error updating password: %v\n", err)
		return err
	}

	s.logger.Println("User password updated successfully")
	return nil
}

// UpdateUsername updates a user's username
func (s *UserService) UpdateUsername(ctx context.Context, userID primitive.ObjectID, newUsername string) error {
	s.logger.Printf("Updating username for user ID: %s\n", userID.Hex())

	err := s.storage.UpdateUsername(ctx, userID, newUsername)
	if err != nil {
		s.logger.Printf("Error updating username: %v\n", err)
		return err
	}

	s.logger.Println("User username updated successfully")
	return nil
}

// ValidateJWT validates a JWT token and returns the user ID
func (s *UserService) ValidateJWT(tokenString string) (string, error) {
	s.logger.Println("Validating JWT token...")

	userID, err := s.storage.ValidateJWT(tokenString)
	if err != nil {
		s.logger.Printf("Invalid JWT token: %v\n", err)
		return "", err
	}

	s.logger.Printf("JWT token validated successfully, user ID: %s\n", userID)
	return userID, nil
}

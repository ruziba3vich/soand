package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ruziba3vich/soand/internal/models"
	"github.com/ruziba3vich/soand/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage struct {
	db     *mongo.Collection
	secret string
}

// NewUserStorage initializes UserStorage
func NewUserStorage(db *mongo.Collection, config *config.Config) *UserStorage {
	return &UserStorage{db: db, secret: config.JwtSecret}
}

// CreateUser inserts a new user into the database and returns a JWT token
func (s *UserStorage) CreateUser(ctx context.Context, user *models.User) (string, error) {
	if len(user.Password) < 8 {
		return "", fmt.Errorf("user password must be at least 8 characters long")
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return "", err
	}
	user.Password = hashedPassword

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = s.db.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID.Hex(), s.secret)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(userID string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT validates a JWT token and returns the user ID
func (s *UserStorage) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return "", err
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, exists := claims["user_id"].(string); exists {
			return userID, nil
		}
	}

	return "", fmt.Errorf("invalid token")
}

// Login checks user credentials and returns a JWT token
func (s *UserStorage) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	// Check password
	if !CheckPassword(user.Password, password) {
		return "", errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID.Hex(), s.secret)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetUserByID fetches a user by their ID
func (s *UserStorage) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	var user models.User

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.db.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername fetches a user by their username
func (s *UserStorage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.db.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser removes a user from the database
func (s *UserStorage) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.DeleteOne(ctx, bson.M{"_id": userID})
	return err
}

// UpdateUsername updates a user's username
func (s *UserStorage) UpdateUsername(ctx context.Context, userID primitive.ObjectID, newUsername string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"username": newUsername}})
	return err
}

// UpdatePassword updates a user's password after verifying the old password
func (s *UserStorage) UpdatePassword(ctx context.Context, userID primitive.ObjectID, oldPassword, newPassword string) error {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !CheckPassword(user.Password, oldPassword) {
		return errors.New("incorrect old password")
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = s.db.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"password": hashedPassword}})
	return err
}

// UpdateFullname updates a user's full name
func (s *UserStorage) UpdateFullname(ctx context.Context, userID primitive.ObjectID, newFullname string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"fullname": newFullname}})
	return err
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a hashed password
func CheckPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

func (s *UserStorage) ChangeProfileVisibility(ctx context.Context, userID primitive.ObjectID, hidden bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"profile_hidden": hidden}}

	result, err := s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

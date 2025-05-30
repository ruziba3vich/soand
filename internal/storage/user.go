package storage

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	db                 *mongo.Collection
	secret             string
	background_storage *BackgroundStorage
}

// NewUserStorage initializes UserStorage
func NewUserStorage(db *mongo.Collection, config *config.Config, background_storage *BackgroundStorage) *UserStorage {
	return &UserStorage{
		db:                 db,
		secret:             config.JwtSecret,
		background_storage: background_storage,
	}
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
	user.ID = primitive.NewObjectIDFromTimestamp(time.Now())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if user.Username != nil {
		exists, err := s.isUsernameTaken(ctx, *user.Username, user.ID)
		if err != nil {
			return "", fmt.Errorf("failed to check username availability: %s", err.Error())
		}

		if exists {
			return "", fmt.Errorf("this username is already taken")
		}
	}

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
		"exp":     time.Now().Add(time.Hour * 24 * 45).Unix(), // Expires in 45 days
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

// UpdateUsername updates a user's username after checking if it's available
func (s *UserStorage) UpdateUsername(ctx context.Context, userID primitive.ObjectID, newUsername string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check if the username is already taken by another user
	isTaken, err := s.isUsernameTaken(ctx, newUsername, userID)
	if err != nil {
		return fmt.Errorf("username check failed: %v", err)
	}
	if isTaken {
		return fmt.Errorf("username '%s' is already taken by another user", newUsername)
	}

	// Update the username in the database
	_, err = s.db.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{"username": newUsername}})
	if err != nil {
		return fmt.Errorf("failed to update username: %v", err)
	}

	return nil
}

// UpdateUser updates Fullname, Bio, and HiddenProfile fields for a user
func (s *UserStorage) UpdateUser(ctx context.Context, userID primitive.ObjectID, updateFields bson.M) error {

	// Perform the update
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": updateFields}
	result, err := s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
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

func (s *UserStorage) SetBackgroundPic(ctx context.Context, userID primitive.ObjectID, pic_id string) error {
	url, err := s.background_storage.GetBackgroundByID(pic_id)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"background_pic": url}}

	result, err := s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// IsUsernameTaken checks if the given username is already taken by another user (excluding the user with the given userID).
func (s *UserStorage) isUsernameTaken(ctx context.Context, username string, userID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Query the database to find a user with the given username, excluding the current user
	filter := bson.M{
		"username": username,
		"_id":      bson.M{"$ne": userID}, // Exclude the current user
	}

	count, err := s.db.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check username availability: %v", err)
	}

	return count > 0, nil
}

func (s *UserStorage) AddNewProfilePicture(ctx context.Context, userID primitive.ObjectID, fileURL string) error {
	// Fetch the user
	var user models.User
	err := s.db.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to fetch user: %v", err)
	}

	// Check profile picture limit
	if len(user.ProfilePics) > 29 {
		return fmt.Errorf("profile picture limit reached (30)")
	}

	// Create a new ProfilePic entry
	newPic := models.ProfilePic{
		Url:      fileURL,
		PostedAt: time.Now(),
	}

	// Initialize profile_pics if null, then append the new picture
	newProfilePics := []models.ProfilePic{newPic}
	if user.ProfilePics != nil {
		newProfilePics = append(newProfilePics, user.ProfilePics...)
	}

	// Update profile_pics in MongoDB with the new array
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"profile_pics": newProfilePics}}
	_, err = s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user profile pictures: %v", err)
	}

	return nil
}

func (s *UserStorage) DeleteProfilePicture(ctx context.Context, userID primitive.ObjectID, fileURL string) error {
	// Fetch the user
	var user models.User
	err := s.db.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to fetch user: %v", err)
	}

	// Check if the fileURL exists in ProfilePics
	found := false
	for _, pic := range user.ProfilePics {
		if pic.Url == fileURL {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("profile picture %s not found for user %s", fileURL, userID.Hex())
	}

	// Remove from MongoDB only (no MinIO deletion here)
	filter := bson.M{"_id": userID}
	update := bson.M{"$pull": bson.M{"profile_pics": bson.M{"url": fileURL}}}
	_, err = s.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to remove profile picture from MongoDB: %v", err)
	}

	return nil
}

func (s *UserStorage) GetProfilePictures(ctx context.Context, userID primitive.ObjectID) ([]models.ProfilePic, error) {
	// Fetch the user
	var user models.User
	err := s.db.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}

	// Sort profile pictures by PostedAt (newest to oldest)
	sort.Slice(user.ProfilePics, func(i, j int) bool {
		return user.ProfilePics[i].PostedAt.After(user.ProfilePics[j].PostedAt)
	})

	return user.ProfilePics, nil
}

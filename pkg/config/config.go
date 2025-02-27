package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type (
	// Config holds all the configuration settings
	Config struct {
		MongoDB   MongoDBConfig
		MinIO     MinIOConfig
		Redis     RedisConfig
		JwtSecret string
	}

	// MongoDBConfig holds MongoDB settings
	MongoDBConfig struct {
		URI, User, Password, Database string
	}

	// MinIOConfig holds MinIO settings
	MinIOConfig struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		Bucket    string
	}

	// RedisConfig holds Redis settings
	RedisConfig struct {
		Host     string
		Port     string
		Password string
		DB       int
	}
)

// LoadConfig loads configurations from environment variables or .env file
func LoadConfig() *Config {
	// Load .env file (if available)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using system environment variables.")
	}

	return &Config{
		JwtSecret: getEnv("JWT_SECRET", "prodonik"),
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGO_DB", "mydatabase"),
			User:     getEnv("MONGO_USER", "mongodb_pro"),
			Password: getEnv("MONGO_PASSWORD", ""),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "my-access-key"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "my-secret-key"),
			Bucket:    getEnv("MINIO_BUCKET", "my-bucket"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}
}

// getEnv retrieves environment variables with a fallback default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvInt retrieves an integer environment variable
func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intValue int
		_, err := fmt.Sscanf(value, "%d", &intValue)
		if err == nil {
			return intValue
		}
	}
	return fallback
}

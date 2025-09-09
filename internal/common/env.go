// internal/common/env.go
package common

import (
	"os"
	"path/filepath"
	
	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env files
func LoadEnv() error {
	// Find the .env file in the current directory or parent directories
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	
	// Try to load .env file
	_ = godotenv.Load(filepath.Join(dir, ".env"))
	
	return nil
}

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// SetEnv sets an environment variable
func SetEnv(key, value string) error {
	return os.Setenv(key, value)
}


// internal/common/file.go
package common

import (
	"fmt"
	"os"
)

// ReadFile reads the contents of a file and returns it as a string
func ReadFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read file: %w", err)
	}
	return string(data), nil
}

// WriteFile writes content to a file
func WriteFile(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// EnsureDirectoryExists ensures that a directory exists, creating it if necessary
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}


// internal/common/file_test.go
package common

import (
	"os"
	"testing"
)

func TestFileOperations(t *testing.T) {
	// Create a temporary directory
	tmpdir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	
	// Test WriteFile
	content := "Hello, World!"
	filePath := tmpdir + "/test.txt"
	
	err = WriteFile(filePath, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	
	// Test FileExists
	if !FileExists(filePath) {
		t.Error("FileExists returned false for existing file")
	}
	
	if FileExists(tmpdir + "/nonexistent.txt") {
		t.Error("FileExists returned true for non-existent file")
	}
	
	// Test ReadFile
	readContent, err := ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	
	if readContent != content {
		t.Errorf("ReadFile returned %q, expected %q", readContent, content)
	}
	
	// Test EnsureDirectoryExists
	newDir := tmpdir + "/newdir"
	err = EnsureDirectoryExists(newDir)
	if err != nil {
		t.Fatalf("EnsureDirectoryExists failed: %v", err)
	}
	
	if !FileExists(newDir) {
		t.Error("EnsureDirectoryExists did not create directory")
	}
	
	// Test EnsureDirectoryExists on existing directory
	err = EnsureDirectoryExists(newDir)
	if err != nil {
		t.Fatalf("EnsureDirectoryExists failed on existing directory: %v", err)
	}
}


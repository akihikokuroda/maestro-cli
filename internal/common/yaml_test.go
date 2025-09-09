// internal/common/yaml_test.go
package common

import (
	"os"
	"testing"
)

func TestParseYAML(t *testing.T) {
	// Create a temporary YAML file
	content := `
kind: Test
metadata:
  name: test
spec:
  value: 123
---
kind: Test2
metadata:
  name: test2
spec:
  value: 456
`
	
	tmpfile, err := os.CreateTemp("", "test*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	
	// Parse the YAML file
	docs, err := ParseYAML(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}
	
	// Check that we got two documents
	if len(docs) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(docs))
	}
	
	// Check the content of the first document
	if docs[0]["kind"] != "Test" {
		t.Errorf("Expected kind=Test, got %v", docs[0]["kind"])
	}
	
	// Check the content of the second document
	if docs[1]["kind"] != "Test2" {
		t.Errorf("Expected kind=Test2, got %v", docs[1]["kind"])
	}
	
	// Check that source_file was added
	if _, ok := docs[0]["source_file"]; !ok {
		t.Error("source_file not added to document")
	}
}


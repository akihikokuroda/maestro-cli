// internal/common/yaml.go
package common

import (
        "bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"encoding/json"
)

// YAMLDocument represents a parsed YAML document
type YAMLDocument map[string]interface{}

// ParseYAML parses a YAML file and returns a slice of YAML documents
func ParseYAML(filePath string) ([]YAMLDocument, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML file: %w", err)
	}

	// Parse the YAML documents
	var docs []YAMLDocument
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	
	// Read all documents from the YAML file
	for {
		var doc YAMLDocument
		err := decoder.Decode(&doc)
		if err != nil {
			break
		}
		
		// Add source file information
		absPath, _ := filepath.Abs(filePath)
		doc["source_file"] = absPath
		
		docs = append(docs, doc)
	}
	
	if len(docs) == 0 {
		return nil, fmt.Errorf("no valid YAML documents found in file")
	}
	
	return docs, nil
}

func YamlToString(yamls []YAMLDocument) ([]string, error) {
	var yaml_strings []string
	for _, yaml := range yamls {
		yaml_string, err := json.Marshal(yaml)
		if err != nil {
			continue
		}
		yaml_strings = append(yaml_strings, string(yaml_string)) 
	}
	return yaml_strings, nil
}


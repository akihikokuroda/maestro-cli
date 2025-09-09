// internal/commands/validate.go
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	
	"maestro/internal/common"
)

// ValidateCommand implements the validate command
type ValidateCommand struct {
	*BaseCommand
	schemaFile string
	yamlFile   string
}

// NewValidateCommand creates a new validate command
func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [SCHEMA_FILE] YAML_FILE",
		Short: "Validate YAML files against JSON schemas",
		Long:  `Validate YAML files against JSON schemas.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			var schemaFile, yamlFile string
			if len(args) == 1 {
				schemaFile = ""
				yamlFile = args[0]
			} else {
				schemaFile = args[0]
				yamlFile = args[1]
			}
			
			validateCmd := &ValidateCommand{
				BaseCommand: NewBaseCommand(options),
				schemaFile:  schemaFile,
				yamlFile:    yamlFile,
			}
			
			return validateCmd.Run()
		},
	}
	
	return cmd
}

// Run executes the validate command
func (c *ValidateCommand) Run() error {
	if c.schemaFile == "" || c.schemaFile == "" {
		// Try to discover the schema file based on the yaml file
		discoveredSchema, err := c.discoverSchemaFile(c.yamlFile)
		if err != nil {
			c.Console().Error(fmt.Sprintf("Invalid YAML file: %s: %s", c.yamlFile, err))
			return err
		}
		
		if discoveredSchema == "" {
			return nil
		}
		
		return c.validate(discoveredSchema, c.yamlFile)
	}
	
	return c.validate(c.schemaFile, c.yamlFile)
}

// discoverSchemaFile tries to discover the schema file based on the yaml file
func (c *ValidateCommand) discoverSchemaFile(yamlFile string) (string, error) {
	// Read the YAML file
	yamlData, err := common.ParseYAML(yamlFile)
	if err != nil {
		return "", err
	}
	
	if len(yamlData) == 0 {
		return "", fmt.Errorf("no YAML documents found in file")
	}
	
	// Get the kind from the first document
	kind, ok := yamlData[0]["kind"].(string)
	if !ok {
		return "", fmt.Errorf("kind field not found or not a string")
	}
	
	// Determine the schema file based on the kind
	switch kind {
	case "Agent":
		return filepath.Join("schemas", "agent_schema.json"), nil
	case "Tool":
		return filepath.Join("schemas", "tool_schema.json"), nil
	case "MCPTool":
		return filepath.Join("schemas", "tool_toolhive_schema_full.json"), nil
	case "Workflow":
		return filepath.Join("schemas", "workflow_schema.json"), nil
	case "WorkflowRun":
		c.Console().Ok("WorkflowRun is not supported")
		return "", nil
	case "CustomResourceDefinition":
		c.Console().Ok("CustomResourceDefinition is not supported")
		return "", nil
	default:
		return "", fmt.Errorf("unknown kind: %s", kind)
	}
}

// validate validates a YAML file against a JSON schema
func (c *ValidateCommand) validate(schemaFile, yamlFile string) error {
	c.Console().Print(fmt.Sprintf("validating %s with schema %s\n", yamlFile, schemaFile))
	
	// If schema file is not provided, try to infer it from the yaml file name
	if schemaFile == "" {
		if strings.Contains(yamlFile, "agents.yaml") {
			schemaFile = filepath.Join("schemas", "agent_schema.json")
		} else if strings.Contains(yamlFile, "workflow.yaml") {
			schemaFile = filepath.Join("schemas", "workflow_schema.json")
		} else {
			return fmt.Errorf("could not determine schema file from yaml file name")
		}
	}
	
	// Load the schema
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaFile)
	
	// Read the YAML file
	yamlBytes, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("could not read YAML file: %w", err)
	}
	
	// Parse the YAML documents
	decoder := yaml.NewDecoder(strings.NewReader(string(yamlBytes)))
	
	// Validate each document
	for {
		var yamlDoc interface{}
		if err := decoder.Decode(&yamlDoc); err != nil {
			break
		}
		
		// Convert YAML to JSON for validation
		jsonBytes, err := json.Marshal(yamlDoc)
		if err != nil {
			return fmt.Errorf("could not convert YAML to JSON: %w", err)
		}
		
		// Create a JSON document loader
		documentLoader := gojsonschema.NewBytesLoader(jsonBytes)
		
		// Validate
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		
		// Check the result
		if !result.Valid() {
			for _, desc := range result.Errors() {
				c.Console().Error(desc.String())
			}
			return fmt.Errorf("YAML file is NOT valid")
		}
		
		if !c.IsSilent() {
			c.Console().Ok("YAML file is valid.")
		}
	}
	
	return nil
}


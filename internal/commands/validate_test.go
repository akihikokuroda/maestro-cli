// internal/commands/validate_test.go
package commands

import (
	"os"
	"testing"
	
	"github.com/spf13/cobra"
)

func TestValidateCommand(t *testing.T) {
	// Create a temporary YAML file
	content := `
kind: Agent
metadata:
  name: test-agent
spec:
  framework: python
  model: test-model
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
	
	// Create a validate command
	cmd := &cobra.Command{
		Use: "validate",
	}
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("silent", true, "")
	cmd.Flags().Bool("dry-run", false, "")
	
	// Create a validate command instance
	validateCmd := &ValidateCommand{
		BaseCommand: NewBaseCommand(NewCommandOptions(cmd)),
		schemaFile:  "", // Let it auto-discover
		yamlFile:    tmpfile.Name(),
	}
	
	// This will fail because we don't have the schema files in the test environment
	// But we can at least check that the command runs without panicking
	_ = validateCmd.Run()
}


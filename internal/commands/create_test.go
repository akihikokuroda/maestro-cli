// internal/commands/create_test.go
package commands

import (
	"os"
	"testing"
	
	"github.com/spf13/cobra"
)

func TestCreateCommand(t *testing.T) {
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
	
	// Create a create command
	cmd := &cobra.Command{
		Use: "create",
	}
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("silent", true, "")
	cmd.Flags().Bool("dry-run", false, "")
	
	// Create a create command instance
	createCmd := &CreateCommand{
		BaseCommand: NewBaseCommand(NewCommandOptions(cmd)),
		agentsFile:  tmpfile.Name(),
	}
	
	// Run the command
	err = createCmd.Run()
	if err != nil {
		t.Fatalf("CreateCommand.Run failed: %v", err)
	}
}


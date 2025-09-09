// internal/commands/meta_agents.go
package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// MetaAgentsCommand implements the meta-agents command
type MetaAgentsCommand struct {
	*BaseCommand
	textFile string
}

// NewMetaAgentsCommand creates a new meta-agents command
func NewMetaAgentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta-agents TEXT_FILE",
		Short: "Run meta-agents on a text file",
		Long:  `Run meta-agents on a text file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			metaAgentsCmd := &MetaAgentsCommand{
				BaseCommand: NewBaseCommand(options),
				textFile:    args[0],
			}
			
			return metaAgentsCmd.Run()
		},
	}
	
	return cmd
}

// Run executes the meta-agents command
func (c *MetaAgentsCommand) Run() error {
	// Run meta-agents on the text file
	if err := c.runMetaAgents(); err != nil {
		c.Console().Error(fmt.Sprintf("Unable to run meta-agents: %s", err))
		return err
	}
	
	if !c.IsSilent() {
		c.Console().Ok("Running meta-agents\n")
	}
	
	return nil
}

// runMetaAgents runs meta-agents on the text file
func (c *MetaAgentsCommand) runMetaAgents() error {
	// Get the path to the streamlit_meta_agents_deploy.py script
	scriptDir := filepath.Dir(os.Args[0])
	streamlitScript := filepath.Join(scriptDir, "streamlit_meta_agents_deploy.py")
	
	// Build the command
	cmd := exec.Command(
		"uv",
		"run",
		"streamlit",
		"run",
		"--ui.hideTopBar",
		"True",
		"--client.toolbarMode",
		"minimal",
		streamlitScript,
		c.textFile,
	)
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start meta-agents: %w", err)
	}
	
	// Store the process ID for later cleanup
	// TODO: Store the process ID somewhere for the clean command to use
	
	return nil
}


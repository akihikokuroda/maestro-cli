// internal/commands/clean.go
package commands

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

// CleanCommand implements the clean command
type CleanCommand struct {
	*BaseCommand
}

// NewCleanCommand creates a new clean command
func NewCleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up running processes",
		Long:  `Clean up running processes.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			cleanCmd := &CleanCommand{
				BaseCommand: NewBaseCommand(options),
			}
			
			return cleanCmd.Run()
		},
	}
	
	return cmd
}

// Run executes the clean command
func (c *CleanCommand) Run() error {
	// Clean up running processes
	if err := c.cleanProcesses(); err != nil {
		c.Console().Error(fmt.Sprintf("Unable to clean: %s", err))
		return err
	}
	
	return nil
}

// cleanProcesses cleans up running processes
func (c *CleanCommand) cleanProcesses() error {
	// Get all processes
	processes, err := process.Processes()
	if err != nil {
		return fmt.Errorf("failed to get processes: %w", err)
	}
	
	// Find and terminate Streamlit processes
	for _, p := range processes {
		cmdline, err := p.Cmdline()
		if err != nil {
			continue
		}
		
		if strings.Contains(cmdline, "streamlit") {
			// Terminate the process
			if err := p.Terminate(); err != nil {
				c.Console().Warn(fmt.Sprintf("Failed to terminate process %d: %s", p.Pid, err))
			}
		}
	}
	
	return nil
}


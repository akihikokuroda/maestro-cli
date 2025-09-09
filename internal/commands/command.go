// internal/commands/command.go
package commands

import (
	"github.com/spf13/cobra"
	
	"maestro/internal/common"
)

// CommandOptions contains common options for all commands
type CommandOptions struct {
	Verbose bool
	Silent  bool
	DryRun  bool
	Console *common.Console
}

// NewCommandOptions creates a new CommandOptions instance
func NewCommandOptions(cmd *cobra.Command) *CommandOptions {
	verbose, _ := cmd.Flags().GetBool("verbose")
	silent, _ := cmd.Flags().GetBool("silent")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	
	return &CommandOptions{
		Verbose: verbose,
		Silent:  silent,
		DryRun:  dryRun,
		Console: common.NewConsole(verbose, silent),
	}
}

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	Options *CommandOptions
}

// NewBaseCommand creates a new BaseCommand instance
func NewBaseCommand(options *CommandOptions) *BaseCommand {
	return &BaseCommand{
		Options: options,
	}
}

// IsVerbose returns whether verbose mode is enabled
func (c *BaseCommand) IsVerbose() bool {
	return c.Options.Verbose
}

// IsSilent returns whether silent mode is enabled
func (c *BaseCommand) IsSilent() bool {
	return c.Options.Silent
}

// IsDryRun returns whether dry-run mode is enabled
func (c *BaseCommand) IsDryRun() bool {
	return c.Options.DryRun
}

// Console returns the Console instance
func (c *BaseCommand) Console() *common.Console {
	return c.Options.Console
}

// SetDryRunEnv sets the DRY_RUN environment variable if dry-run mode is enabled
func (c *BaseCommand) SetDryRunEnv() {
	if c.IsDryRun() {
		common.SetEnv("DRY_RUN", "True")
	}
}


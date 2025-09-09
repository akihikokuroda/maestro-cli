// internal/commands/create.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	
	"maestro/internal/common"
)

// CreateCommand implements the create command
type CreateCommand struct {
	*BaseCommand
	agentsFile string
}

// NewCreateCommand creates a new create command
func NewCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create AGENTS_FILE",
		Short: "Create agents from a configuration file",
		Long:  `Create agents from the specified configuration file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			createCmd := &CreateCommand{
				BaseCommand: NewBaseCommand(options),
				agentsFile:  args[0],
			}
			
			return createCmd.Run()
		},
	}
	
	return cmd
}

// Run executes the create command
func (c *CreateCommand) Run() error {
	// Parse the agents YAML file
	agentsYaml, err := common.ParseYAML(c.agentsFile)
	if err != nil {
		c.Console().Error(fmt.Sprintf("Unable to parse agents file: %s", err))
		return err
	}
	
	// Create the agents
	if err := c.createAgents(agentsYaml); err != nil {
		c.Console().Error(fmt.Sprintf("Unable to create agents: %s", err))
		return err
	}
	
	return nil
}

// createAgents creates agents from the YAML configuration
func (c *CreateCommand) createAgents(agentsYaml []common.YAMLDocument) error {
	if len(agentsYaml) == 0 {
		return fmt.Errorf("no agents found in YAML file")
	}
	
	// Check the kind of the first document
	kind, ok := agentsYaml[0]["kind"].(string)
	if !ok {
		return fmt.Errorf("kind field not found or not a string")
	}
	
	switch kind {
	case "Agent":
		return c.createAgentsFromYAML(agentsYaml)
	case "MCPTool":
		return c.createMCPToolsFromYAML(agentsYaml)
	default:
		return fmt.Errorf("unsupported kind: %s", kind)
	}
}

// createAgentsFromYAML creates agents from the YAML configuration
func (c *CreateCommand) createAgentsFromYAML(agentsYaml []common.YAMLDocument) error {
	// For now, we'll just print a message
	c.Console().Ok("Creating agents from YAML configuration")
	
	// TODO: Implement the actual agent creation logic
	// This would involve:
	// 1. Parsing the agent definitions
	// 2. Creating the agent instances
	// 3. Registering them with the system

	// Get MCP server URI
	// serverURI, _ := common.GetMCPServerURI(mcpServerURI)
	serverURI, err := common.GetMCPServerURI("")
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Failed to get MCP server URI")
		}
		return err
	}

	if common.Verbose {
		fmt.Printf("Connecting to MCP server at: %s\n", serverURI)
	}

	// Create MCP client
	client, _ := common.NewMCPClient(serverURI)
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Failed to create MCP client")
		}
		return err
	}
	defer client.Close()

	if common.Progress != nil {
		common.Progress.Update("Executing create agents...")
	}
	

	// Call the run_workflow tool
	agent_strings, err := common.YamlToString(agentsYaml)
	if err != nil {
		fmt.Println("agent file error")
	}

	params := map[string]interface{}{
		"agents": agent_strings,
	}
	
	result, err := client.CallMCPServer("create_agents", params)
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Create agent failed")
		}
		return err
	}

	if common.Progress != nil {
		common.Progress.Stop("Create agents completed successfully")
	}

	if !common.Silent {
		fmt.Println("OK")
	}
	fmt.Println(result)
	
	return nil
}

// createMCPToolsFromYAML creates MCP tools from the YAML configuration
func (c *CreateCommand) createMCPToolsFromYAML(agentsYaml []common.YAMLDocument) error {
	// In the Python implementation, this calls create_mcptools from maestro.mcptool
	// We'll need to implement the equivalent functionality in Go
	
	// For now, we'll just print a message
	c.Console().Ok("Creating MCP tools from YAML configuration")
	
	// TODO: Implement the actual MCP tool creation logic
	// This would involve:
	// 1. Parsing the tool definitions
	// 2. Creating the tool instances
	// 3. Registering them with the system
	
	return nil
}


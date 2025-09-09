// internal/commands/serve.go
package commands

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/spf13/cobra"
	
	"maestro/internal/common"
)

// ServeCommand implements the serve command
type ServeCommand struct {
	*BaseCommand
	agentsFile   string
	workflowFile string
	agentName    string
	host         string
	port         int
}

// NewServeCommand creates a new serve command
func NewServeCommand() *cobra.Command {
	serveCmd := &ServeCommand{}
	
	cmd := &cobra.Command{
		Use:   "serve AGENTS_FILE [WORKFLOW_FILE]",
		Short: "Serve agents via HTTP endpoints",
		Long:  `Serve agents via HTTP endpoints.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			serveCmd.BaseCommand = NewBaseCommand(options)
			serveCmd.agentsFile = args[0]
			
			if len(args) > 1 {
				serveCmd.workflowFile = args[1]
			} else {
				serveCmd.workflowFile = ""
			}
			
			// Get flag values
			var err error
			serveCmd.agentName, err = cmd.Flags().GetString("agent-name")
			if err != nil {
				return err
			}
			
			serveCmd.host, err = cmd.Flags().GetString("host")
			if err != nil {
				return err
			}
			
			portStr, err := cmd.Flags().GetString("port")
			if err != nil {
				return err
			}
			
			if portStr == "" {
				serveCmd.port = 8001
			} else {
				serveCmd.port, err = strconv.Atoi(portStr)
				if err != nil {
					return fmt.Errorf("invalid port number: %s", portStr)
				}
			}
			
			return serveCmd.Run()
		},
	}
	
	// Add flags
	cmd.Flags().String("agent-name", "", "Specific agent name to serve (if multiple in file)")
	cmd.Flags().String("host", "127.0.0.1", "Host to bind to")
	cmd.Flags().String("port", "8000", "Port to serve on")
	
	return cmd
}

// Run executes the serve command
func (c *ServeCommand) Run() error {
	if c.workflowFile != "" && c.workflowFile != "None" {
		// Serve workflow
		if err := c.serveWorkflow(); err != nil {
			c.Console().Error(fmt.Sprintf("Unable to serve workflow: %s", err))
			return err
		}
		
		if !c.IsSilent() {
			c.Console().Ok("Workflow server started successfully")
		}
	} else {
		// Serve agent
		if err := c.serveAgent(); err != nil {
			c.Console().Error(fmt.Sprintf("Unable to serve agent: %s", err))
			return err
		}
		
		if !c.IsSilent() {
			c.Console().Ok("Agent server started successfully")
		}
	}
	
	return nil
}

// serveWorkflow serves a workflow via HTTP
func (c *ServeCommand) serveWorkflow() error {
	c.Console().Print(fmt.Sprintf("Serving workflow at %s:%d\n", c.host, c.port))
	
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
		common.Progress.Update("Executing serve workflow...")
	}

	// Read yaml files into string
	data, err := ioutil.ReadFile(c.agentsFile)
	if err != nil {
		return  err
	}
	agent_strings :=  string(data)
	data, err = ioutil.ReadFile(c.workflowFile)
	if err != nil {
		return err
	}
	workflow_strings := string(data)

	params := map[string]interface{}{
		"agents": agent_strings,
		"workflow": workflow_strings,
		"host": c.host,
		"port": c.port,
	}
	
	result, err := client.CallMCPServer("serve_workflow", params)
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Serve workflow failed")
		}
		return err
	}

	if common.Progress != nil {
		common.Progress.Stop("Serve workflow completed successfully")
	}

	if !common.Silent {
		fmt.Println("OK")
	}
	fmt.Println(result)
	
	return nil
}

// serveAgent serves an agent via HTTP
func (c *ServeCommand) serveAgent() error {
	// Get the agent framework
	framework, err := c.getAgentFramework()
	if err != nil {
		return err
	}
	
	if framework == "container" {
		return c.serveContainerAgent()
	} else {
		return c.serveFastAPIAgent()
	}
}

// getAgentFramework gets the framework type of the agent
func (c *ServeCommand) getAgentFramework() (string, error) {
	// Parse the agents YAML file
	agentsYaml, err := common.ParseYAML(c.agentsFile)
	if err != nil {
		return "", fmt.Errorf("unable to parse agents file: %w", err)
	}
	
	// Get the framework from the first agent
	framework := ""
	if spec, ok := agentsYaml[0]["spec"].(common.YAMLDocument); ok {
		if f, ok := spec["framework"].(string); ok {
			framework = f
		}
	}
	
	// If agent name is specified, get the framework for that agent
	if c.agentName != "" {
		for _, agent := range agentsYaml {
			if metadata, ok := agent["metadata"].(common.YAMLDocument); ok {
				if name, ok := metadata["name"].(string); ok && name == c.agentName {
					if spec, ok := agent["spec"].(common.YAMLDocument); ok {
						if f, ok := spec["framework"].(string); ok {
							framework = f
						}
					}
				}
			}
		}
	}
	return framework, nil
}

// serveContainerAgent serves a container agent
func (c *ServeCommand) serveContainerAgent() error {
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
		common.Progress.Update("Executing serve agent...")
	}

	// Parse the agent YAML file
	var agentYaml common.YAMLDocument
	agentsYaml, err := common.ParseYAML(c.agentsFile)
	if err != nil {
		c.Console().Error(fmt.Sprintf("Unable to parse workflow file: %s", err))
		return err
	}

	if c.agentName != "" {
		for _, agent := range agentsYaml {
			if metadata, ok := agent["metadata"].(common.YAMLDocument); ok {
				if name, ok := metadata["name"].(string); ok && name == c.agentName {
					agentYaml = agent
				}
			}
		}
	} else {
		agentYaml = agentsYaml[0]
	}


	image_url := ""
	app_name := ""
	if metadata, ok := agentYaml["metadata"].(common.YAMLDocument); ok {
		if name, ok := metadata["name"].(string); ok {
			app_name = name
		}
	}
	if spec, ok := agentYaml["spec"].(common.YAMLDocument); ok {
		if image, ok := spec["image"].(string); ok {
			image_url = image		
		}
	}
	
	params := map[string]interface{}{
		"image_url": image_url,
		"app_name": app_name,
	}
	
	result, err := client.CallMCPServer("serve_container_agent", params)
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Serve containered agent failed")
		}
		return err
	}

	if common.Progress != nil {
		common.Progress.Stop("Serve containsered agent completed successfully")
	}

	if !common.Silent {
		fmt.Println("OK")
	}
	fmt.Println(result)
	
	return nil
}

// serveFastAPIAgent serves a FastAPI agent
func (c *ServeCommand) serveFastAPIAgent() error {
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
		common.Progress.Update("Executing serve agent...")
	}

	// Read yaml files into string
	data, err := ioutil.ReadFile(c.agentsFile)
	if err != nil {
		return  err
	}
	agent_strings :=  string(data)
	// Call the serve_agent tool
	params := map[string]interface{}{
		"agent": agent_strings,
		"agent_name": c.agentName,
		"host": c.host,
		"port": c.port,
	}
	
	result, err := client.CallMCPServer("serve_agent", params)
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Serve agent failed")
		}
		return err
	}

	if common.Progress != nil {
		common.Progress.Stop("Serve agent completed successfully")
	}

	if !common.Silent {
		fmt.Println("OK")
	}
	fmt.Println(result)
	
	return  nil

}


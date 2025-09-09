// internal/commands/deploy.go
package commands

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	
	"maestro/internal/common"
)

// DeployCommand implements the deploy command
type DeployCommand struct {
	*BaseCommand
	agentsFile   string
	workflowFile string
	url          string
	k8s          bool
	kubernetes   bool
	docker       bool
	streamlit    bool
	autoPrompt   bool
	env          []string
}

// NewDeployCommand creates a new deploy command
func NewDeployCommand() *cobra.Command {
	deployCmd := &DeployCommand{}
	
	cmd := &cobra.Command{
		Use:   "deploy AGENTS_FILE WORKFLOW_FILE [ENV...]",
		Short: "Deploy a workflow to a Kubernetes cluster or local server",
		Long:  `Deploy a workflow to a Kubernetes cluster or local server.`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			deployCmd.BaseCommand = NewBaseCommand(options)
			deployCmd.agentsFile = args[0]
			deployCmd.workflowFile = args[1]
			deployCmd.env = args[2:]
			
			// Get flag values
			var err error
			deployCmd.url, err = cmd.Flags().GetString("url")
			if err != nil {
				return err
			}
			
			deployCmd.k8s, err = cmd.Flags().GetBool("k8s")
			if err != nil {
				return err
			}
			
			deployCmd.kubernetes, err = cmd.Flags().GetBool("kubernetes")
			if err != nil {
				return err
			}
			
			deployCmd.docker, err = cmd.Flags().GetBool("docker")
			if err != nil {
				return err
			}
			
			deployCmd.streamlit, err = cmd.Flags().GetBool("streamlit")
			if err != nil {
				return err
			}
			
			deployCmd.autoPrompt, err = cmd.Flags().GetBool("auto-prompt")
			if err != nil {
				return err
			}
			
			return deployCmd.Run()
		},
	}
	
	// Add flags
	cmd.Flags().String("url", "127.0.0.1:5000", "The deployment URL")
	cmd.Flags().Bool("k8s", false, "Deploy to Kubernetes")
	cmd.Flags().Bool("kubernetes", false, "Deploy to Kubernetes")
	cmd.Flags().Bool("docker", false, "Deploy to Docker")
	cmd.Flags().Bool("streamlit", false, "Deploy as Streamlit application (default)")
	cmd.Flags().Bool("auto-prompt", false, "Run prompt by default if specified")
	
	return cmd
}

// Run executes the deploy command
func (c *DeployCommand) Run() error {
	// Parse the agents and workflow YAML files
	agentsYaml, err := common.ParseYAML(c.agentsFile)
	if err != nil {
		c.Console().Error(fmt.Sprintf("Unable to parse agents file: %s", err))
		return err
	}
	
	workflowYaml, err := common.ParseYAML(c.workflowFile)
	if err != nil {
		c.Console().Error(fmt.Sprintf("Unable to parse workflow file: %s", err))
		return err
	}
	
	// Prepare environment variables
	env := c.prepareEnvironment()
	
	// Deploy the workflow
	if err := c.deployWorkflow(agentsYaml, workflowYaml, env); err != nil {
		c.Console().Error(fmt.Sprintf("Unable to deploy workflow: %s", err))
		return err
	}
	
	return nil
}

// prepareEnvironment prepares the environment variables for deployment
func (c *DeployCommand) prepareEnvironment() string {
	env := strings.Join(c.env, " ")
	
	// Add AUTO_RUN=true if auto-prompt is enabled
	if c.autoPrompt {
		env += " AUTO_RUN=true"
	}
	
	return env
}

// deployWorkflow deploys the workflow to the specified target
func (c *DeployCommand) deployWorkflow(agentsYaml, workflowYaml []common.YAMLDocument, env string) error {
	if c.docker {
		return c.deployToTarget("docker", env)
	} else if c.k8s || c.kubernetes {
		return c.deployToTarget("kubernetes", env)
	} else {
		return c.deployToTarget("streamlit", env)
	}
}

// deploy the workflow to Kubernetes or docker
func (c *DeployCommand) deployToTarget(target string, env string) error {
	c.Console().Ok("Deploying workflow to Kubernetes")
	
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
		common.Progress.Update("Executing deploy...")
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
	workflow_strings :=  string(data)
	
	params := map[string]interface{}{
		"agents": agent_strings,
		"workflow": workflow_strings,
		"target": target,
		"env": env,
	}
	
	result, err := client.CallMCPServer("deploy_workflow", params)
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("deploy failed")
		}
		return err
	}

	if common.Progress != nil {
		common.Progress.Stop("Deploy completed successfully")
	}

	if !common.Silent {
		fmt.Println("OK")
	}
	fmt.Println(result)
	
	return nil
}

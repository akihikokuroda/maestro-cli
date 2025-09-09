// internal/commands/run.go
package commands

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	
	"maestro/internal/common"
)

// RunCommand implements the run command
type RunCommand struct {
	*BaseCommand
	agentsFile   string
	workflowFile string
	prompt       bool
}

// NewRunCommand creates a new run command
func NewRunCommand() *cobra.Command {
	var prompt bool
	
	cmd := &cobra.Command{
		Use:   "run [AGENTS_FILE] WORKFLOW_FILE",
		Short: "Run a workflow with specified agents and workflow files",
		Long:  `Run a workflow with specified agents and workflow files.`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			var agentsFile, workflowFile string
			if len(args) == 1 {
				agentsFile = ""
				workflowFile = args[0]
			} else {
				agentsFile = args[0]
				workflowFile = args[1]
			}
			
			runCmd := &RunCommand{
				BaseCommand:  NewBaseCommand(options),
				agentsFile:   agentsFile,
				workflowFile: workflowFile,
				prompt:       prompt,
			}
			
			return runCmd.Run()
		},
	}
	
	cmd.Flags().BoolVar(&prompt, "prompt", false, "Reads a user prompt and executes workflow with it")
	
	return cmd
}

// Run executes the run command
func (c *RunCommand) Run() error {
	// Set up logging
	logger := c.setupLogger()
	workflowID := c.generateWorkflowID()
	
	// Parse the workflow YAML file
	workflowYaml, err := common.ParseYAML(c.workflowFile)
	if err != nil {
		c.Console().Error(fmt.Sprintf("Unable to parse workflow file: %s", err))
		return err
	}
	
	// Parse the agents YAML file if provided, or try to infer it
	var agentsYaml []common.YAMLDocument
	if c.agentsFile != "" && c.agentsFile != "None" {
		agentsYaml, err = common.ParseYAML(c.agentsFile)
		if err != nil {
			c.Console().Error(fmt.Sprintf("Unable to parse agents file: %s", err))
			return err
		}
	} else {
		// Try to infer the agents file path
		inferredPath := filepath.Join(filepath.Dir(c.workflowFile), "agents.yaml")
		if common.FileExists(inferredPath) {
			agentsYaml, err = common.ParseYAML(inferredPath)
			if err != nil {
				c.Console().Error(fmt.Sprintf("Unable to parse inferred agents file: %s", err))
				return err
			}
			c.Console().Print(fmt.Sprintf("[INFO] Auto-loaded agents.yaml from: %s\n", inferredPath))
		} else {
			c.Console().Warn("⚠️ No agents.yaml path provided or found — skipping custom_agent label handling.")
		}
	}
	
	// Handle prompt if requested
	if c.prompt {
		userPrompt := c.readPrompt()
		if workflow, ok := workflowYaml[0]["spec"].(map[string]interface{}); ok {
			if template, ok := workflow["template"].(map[string]interface{}); ok {
				template["prompt"] = userPrompt
			}
		}
	}
	
	// Run the workflow
	startTime := time.Now().UTC()
	result, err := c.runWorkflow(workflowYaml[0], agentsYaml, workflowID, logger)
	endTime := time.Now().UTC()
	durationMs := int(endTime.Sub(startTime).Milliseconds())
	
	if err != nil {
		c.Console().Error(fmt.Sprintf("Unable to run workflow: %s", err))
		
		// Log the error
		c.logWorkflowRun(
			logger,
			workflowID,
			"UNKNOWN",
			"",
			"",
			[]string{},
			"error",
			startTime,
			endTime,
			durationMs,
		)
		
		return err
	}
	
	// Extract workflow name and prompt
	workflowName := ""
	prompt := ""
	if metadata, ok := workflowYaml[0]["metadata"].(map[string]interface{}); ok {
		if name, ok := metadata["name"].(string); ok {
			workflowName = name
		}
	}
	if spec, ok := workflowYaml[0]["spec"].(map[string]interface{}); ok {
		if template, ok := spec["template"].(map[string]interface{}); ok {
			if p, ok := template["prompt"].(string); ok {
				prompt = p
			}
		}
	}
	
	// Extract models used
	modelsUsed := []string{}
	if agentsYaml != nil {
		for _, agent := range agentsYaml {
			if spec, ok := agent["spec"].(map[string]interface{}); ok {
				if model, ok := spec["model"].(string); ok && model != "" {
					modelsUsed = append(modelsUsed, model)
				} else if metadata, ok := agent["metadata"].(map[string]interface{}); ok {
					if name, ok := metadata["name"].(string); ok {
						modelsUsed = append(modelsUsed, fmt.Sprintf("code:%s", name))
					}
				}
			}
		}
	}
	
	// Extract agent labels
	// excludedCustomAgents := map[string]bool{
	// 	"slack_agent":   true,
	// 	"scoring_agent": true,
	// }
	agentLabels := make(map[string]string)
	if agentsYaml != nil {
		for _, agent := range agentsYaml {
			if metadata, ok := agent["metadata"].(map[string]interface{}); ok {
				if name, ok := metadata["name"].(string); ok {
					if labels, ok := metadata["labels"].(map[string]interface{}); ok {
						if customAgent, ok := labels["custom_agent"].(string); ok {
							agentLabels[strings.ToLower(name)] = customAgent
						}
					}
				}
			}
		}
	}
	
	// Extract output from the result
	output := result["result"].(*common.MCPResponse)
	// TODO: Extract output from the workflow result
	// This would involve:
	// 1. Getting the steps from the workflow
	// 2. Finding the last step that produced a result
	// 3. Using that result as the output
	
	// Log the workflow run
	c.logWorkflowRun(
		logger,
		workflowID,
		workflowName,
		prompt,
		output.Result.(map[string]interface{})["final_prompt"].(string),
		modelsUsed,
		"success",
		startTime,
		endTime,
		durationMs,
	)
	
	return nil
}

// setupLogger sets up the logger for the run command
func (c *RunCommand) setupLogger() *common.Logger {
	// In the Python implementation, this uses FileLogger
	// We'll need to implement the equivalent functionality in Go
	
	// For now, we'll just return a simple logger
	return common.NewLogger()
}

// generateWorkflowID generates a unique workflow ID
func (c *RunCommand) generateWorkflowID() string {
	// In the Python implementation, this uses FileLogger.generate_workflow_id()
	// We'll generate a simple UUID for now
	return fmt.Sprintf("workflow-%d", time.Now().UnixNano())
}

// readPrompt reads a prompt from the user
func (c *RunCommand) readPrompt() string {
	return c.Console().ReadInput("Enter your prompt: ")
}

// runWorkflow runs the workflow with the specified configuration
func (c *RunCommand) runWorkflow(workflow common.YAMLDocument, agents []common.YAMLDocument, workflowID string, logger *common.Logger) (map[string]interface{}, error) {
	
	c.Console().Ok("Running workflow")

	// Get MCP server URI
	// serverURI, _ := common.GetMCPServerURI(mcpServerURI)
	serverURI, err := common.GetMCPServerURI("")
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Failed to get MCP server URI")
		}
		return map[string]interface{}{"result": fmt.Errorf("failed to get MCP server URI: %w", err)}, nil
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
		return map[string]interface{}{"result": fmt.Errorf("failed to create MCP client: %w", err)}, nil
	}
	defer client.Close()

	if common.Progress != nil {
		common.Progress.Update("Executing query...")
	}

	// Call the run_workflow tool
	agent_strings, err := common.YamlToString(agents)
	if err != nil {
		fmt.Println("agent file error")
	}
	workflow_list := []common.YAMLDocument{workflow}
	workflow_strings, err := common.YamlToString(workflow_list)
	if err != nil {
		fmt.Println("workflow file error")
	}
	params := map[string]interface{}{
		"agents": agent_strings,
		"workflow": workflow_strings[0], 
	}
	
	result, err := client.CallMCPServer("run_workflow", params)
	if err != nil {
		if common.Progress != nil {
			common.Progress.StopWithError("Query failed")
		}
		return map[string]interface{}{"result": fmt.Errorf("failed to query vector database: %w", err)}, nil
	}

	if common.Progress != nil {
		common.Progress.Stop("Query completed successfully")
	}

	if !common.Silent {
		fmt.Println("OK")
	}
	fmt.Println(result)
	
	return map[string]interface{}{
		"result": result,
	}, nil
}

// logWorkflowRun logs the workflow run
func (c *RunCommand) logWorkflowRun(logger *common.Logger, workflowID, workflowName, prompt, output string, modelsUsed []string, status string, startTime, endTime time.Time, durationMs int) {
	// In the Python implementation, this calls logger.log_workflow_run()
	// We'll need to implement the equivalent functionality in Go
	
 	// For now, we'll just print a message
	c.Console().Ok(fmt.Sprintf("Workflow %s completed with status: %s", workflowID, status))

	// TODO: Implement the actual logging logic
	// This would involve:
	// 1. Creating a log entry
	// 2. Writing it to the log file
}


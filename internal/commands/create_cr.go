// internal/commands/create_cr.go
package commands

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// CreateCrCommand implements the create-cr command
type CreateCrCommand struct {
	*BaseCommand
	yamlFile string
}

// NewCreateCrCommand creates a new create-cr command
func NewCreateCrCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-cr YAML_FILE",
		Short: "Create Kubernetes custom resources",
		Long:  `Create Kubernetes custom resources from a YAML file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := NewCommandOptions(cmd)
			
			createCrCmd := &CreateCrCommand{
				BaseCommand: NewBaseCommand(options),
				yamlFile:    args[0],
			}
			
			return createCrCmd.Run()
		},
	}
	
	return cmd
}

// Run executes the create-cr command
func (c *CreateCrCommand) Run() error {
	// Create custom resources
	if err := c.createCustomResources(); err != nil {
		c.Console().Error(fmt.Sprintf("Unable to create CR: %s", err))
		return err
	}
	
	return nil
}

// createCustomResources creates Kubernetes custom resources
func (c *CreateCrCommand) createCustomResources() error {
	// Read the YAML file
	yamlBytes, err := os.ReadFile(c.yamlFile)
	if err != nil {
		return fmt.Errorf("could not read YAML file: %w", err)
	}
	
	// Parse the YAML documents
	decoder := yaml.NewDecoder(strings.NewReader(string(yamlBytes)))
	
	// Process each document
	for {
		var yamlDoc map[string]interface{}
		if err := decoder.Decode(&yamlDoc); err != nil {
			break
		}
		
		// Set the API version
		yamlDoc["apiVersion"] = "maestro.ai4quantum.com/v1alpha1"
		
		// Sanitize metadata.name
		if metadata, ok := yamlDoc["metadata"].(map[string]interface{}); ok {
			if name, ok := metadata["name"].(string); ok {
				metadata["name"] = sanitizeName(name)
			}
			
			// Sanitize metadata.labels
			if labels, ok := metadata["labels"].(map[string]interface{}); ok {
				for key, value := range labels {
					if strValue, ok := value.(string); ok {
						labels[key] = sanitizeName(strValue)
					}
				}
			}
		}
		
		// Process workflow-specific fields
		if kind, ok := yamlDoc["kind"].(string); ok && kind == "Workflow" {
			if spec, ok := yamlDoc["spec"].(map[string]interface{}); ok {
				if template, ok := spec["template"].(map[string]interface{}); ok {
					// Remove template.metadata
					delete(template, "metadata")
					
					// Sanitize template.agents
					if agents, ok := template["agents"].([]interface{}); ok {
						sanitizedAgents := make([]interface{}, len(agents))
						for i, agent := range agents {
							if strAgent, ok := agent.(string); ok {
								sanitizedAgents[i] = sanitizeName(strAgent)
							} else {
								sanitizedAgents[i] = agent
							}
						}
						template["agents"] = sanitizedAgents
					}
					
					// Sanitize template.steps
					if steps, ok := template["steps"].([]interface{}); ok {
						for _, step := range steps {
							if stepMap, ok := step.(map[string]interface{}); ok {
								// Sanitize step.agent
								if agent, ok := stepMap["agent"].(string); ok {
									stepMap["agent"] = sanitizeName(agent)
								}
								
								// Sanitize step.parallel
								if parallel, ok := stepMap["parallel"].([]interface{}); ok {
									sanitizedParallel := make([]interface{}, len(parallel))
									for i, agent := range parallel {
										if strAgent, ok := agent.(string); ok {
											sanitizedParallel[i] = sanitizeName(strAgent)
										} else {
											sanitizedParallel[i] = agent
										}
									}
									stepMap["parallel"] = sanitizedParallel
								}
							}
						}
					}
					
					// Sanitize template.exception
					if exception, ok := template["exception"].(map[string]interface{}); ok {
						if agent, ok := exception["agent"].(string); ok {
							exception["agent"] = sanitizeName(agent)
						}
					}
				}
			}
		}
		
		// Write the modified YAML to a temporary file
		tempFile := "temp_yaml"
		file, err := os.Create(tempFile)
		if err != nil {
			return fmt.Errorf("could not create temporary file: %w", err)
		}
		
		encoder := yaml.NewEncoder(file)
		if err := encoder.Encode(yamlDoc); err != nil {
			file.Close()
			return fmt.Errorf("could not encode YAML: %w", err)
		}
		
		file.Close()
		
		// Apply the YAML using kubectl
		cmd := exec.Command("kubectl", "apply", "-f", tempFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("kubectl apply failed: %s: %w", string(output), err)
		}
		
		// Clean up the temporary file
		os.Remove(tempFile)
	}
	
	return nil
}

// sanitizeName sanitizes a name for Kubernetes resources
func sanitizeName(name string) string {
	// Replace non-alphanumeric characters with hyphens
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	newName := re.ReplaceAllString(name, "-")
	
	// Convert to lowercase
	newName = strings.ToLower(newName)
	
	// Replace spaces with hyphens
	newName = strings.ReplaceAll(newName, " ", "-")
	
	// If the name ends with a digit or a dot or a hyphen, add an 'e'
	re = regexp.MustCompile("[.-0-9]$")
	if re.MatchString(newName) {
		newName += "e"
	}
	
	return newName
}


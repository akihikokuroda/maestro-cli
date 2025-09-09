// internal/common/logger.go
package common

import (
	"fmt"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Logger provides logging functionality for workflows
type Logger struct {
	logDir string
}

// NewLogger creates a new Logger instance
func NewLogger() *Logger {
	// Create the log directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return &Logger{logDir: "logs"}
	}
	
	logDir := filepath.Join(homeDir, ".maestro", "logs")
	if err := EnsureDirectoryExists(logDir); err != nil {
		fmt.Println("Error creating log directory:", err)
		return &Logger{logDir: "logs"}
	}
	
	return &Logger{
		logDir: logDir,
	}
}

// GenerateWorkflowID generates a unique workflow ID
func (l *Logger) GenerateWorkflowID() string {
	return fmt.Sprintf("workflow-%d", time.Now().UnixNano())
}

// LogWorkflowRun logs a workflow run
func (l *Logger) LogWorkflowRun(workflowID, workflowName, prompt, output string, modelsUsed []string, status string, startTime, endTime time.Time, durationMs int) error {
	// Create the log file
	logFile := filepath.Join(l.logDir, fmt.Sprintf("%s.log", workflowID))
	
	// Create the log entry
	logEntry := map[string]interface{}{
		"workflow_id":   workflowID,
		"workflow_name": workflowName,
		"prompt":        prompt,
		"output":        output,
		"models_used":   modelsUsed,
		"status":        status,
		"start_time":    startTime.Format(time.RFC3339),
		"end_time":      endTime.Format(time.RFC3339),
		"duration_ms":   durationMs,
	}
	
	// Write the log entry to the file
	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("could not create log file: %w", err)
	}
	defer file.Close()
	
	// Write the log entry as JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(logEntry); err != nil {
		return fmt.Errorf("could not write log entry: %w", err)
	}
	
	return nil
}


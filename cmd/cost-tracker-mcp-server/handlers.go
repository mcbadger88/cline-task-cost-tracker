package main

import (
	"fmt"
	"os"
	"path/filepath"

	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser"
)

// MCPResponse represents an MCP tool response
type MCPResponse struct {
	Content []MCPContent `json:"content"`
}

// MCPContent represents content in an MCP response
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// HandleGenerateCSV processes cost tracking requests and generates CSV files
func HandleGenerateCSV(params map[string]interface{}) (*MCPResponse, error) {
	// Get the file path parameter
	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		// If no file path provided, use the current task
		currentTaskID, err := GetCurrentTaskID()
		if err != nil {
			return nil, fmt.Errorf("failed to get current task ID: %v", err)
		}
		filePath = filepath.Join(GetClineTasksPath(), currentTaskID, UIMessagesFile)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Try to detect the repository root where Cline is working
	repoRoot, err := detectRepositoryRoot(filePath)
	if err != nil {
		// Fallback to current working directory (MCP server directory)
		repoRoot, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error getting current working directory: %v", err)
		}
	}

	// Create the target log directory: {repoRoot}/ui-log-parser
	logBasePath := filepath.Join(repoRoot, "ui-log-parser")

	// Use the new function that creates logs directory in the detected repository
	err = uilogparser.ProcessUILogToCSVAutoAt(filePath, logBasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to process file: %v", err)
	}

	// Extract task ID for response
	taskID := uilogparser.ExtractTaskID(filePath)

	return &MCPResponse{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully processed task %s and generated CSV file at %s/logs/", taskID, logBasePath),
			},
		},
	}, nil
}

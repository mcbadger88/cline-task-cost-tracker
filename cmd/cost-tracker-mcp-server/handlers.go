package main

import (
	"fmt"
	"os"
	"path/filepath"

	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/internal/ui-log-parser"
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

	// Process the file
	err := uilogparser.ProcessUILogToCSVAuto(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to process file: %v", err)
	}

	// Extract task ID for response
	taskID := uilogparser.ExtractTaskID(filePath)

	return &MCPResponse{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Successfully processed task %s and generated CSV file", taskID),
			},
		},
	}, nil
}

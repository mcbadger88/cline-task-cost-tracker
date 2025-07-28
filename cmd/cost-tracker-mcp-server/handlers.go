package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser"
)

// MCPRequest represents an MCP tool request
type MCPRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// MCPResponse represents an MCP tool response
type MCPResponse struct {
	Content []MCPContent `json:"content"`
}

// MCPContent represents content in an MCP response
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// TaskSummary represents a summary of a tracked task
type TaskSummary struct {
	TaskID       string  `json:"task_id"`
	TotalCost    float64 `json:"total_cost"`
	MessageCount int     `json:"message_count"`
	CSVPath      string  `json:"csv_path"`
	LastUpdated  string  `json:"last_updated"`
}

// HandleManualCostTrack processes manual cost tracking requests
func HandleManualCostTrack(params map[string]interface{}) (*MCPResponse, error) {
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

// HandleGetCostSummary returns a summary of all tracked costs
func HandleGetCostSummary(params map[string]interface{}) (*MCPResponse, error) {
	// Ensure logs directory exists
	if err := EnsureLogsDirectory(); err != nil {
		return nil, fmt.Errorf("failed to access logs directory: %v", err)
	}

	// Find all CSV files in logs directory
	pattern := filepath.Join(LogsDirectory, "task_*_costs.csv")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find CSV files: %v", err)
	}

	if len(matches) == 0 {
		return &MCPResponse{
			Content: []MCPContent{
				{
					Type: "text",
					Text: "No tracked tasks found. Run manual_cost_track to generate cost data.",
				},
			},
		}, nil
	}

	var summaries []TaskSummary
	var totalCostAll float64

	for _, csvPath := range matches {
		summary, err := analyzeCsvFile(csvPath)
		if err != nil {
			continue // Skip files that can't be analyzed
		}
		summaries = append(summaries, summary)
		totalCostAll += summary.TotalCost
	}

	// Sort by task ID (most recent first)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].TaskID > summaries[j].TaskID
	})

	// Format response
	var response strings.Builder
	response.WriteString(fmt.Sprintf("Cost Summary - Total Tracked Tasks: %d\n", len(summaries)))
	response.WriteString(fmt.Sprintf("Total Cost Across All Tasks: $%.6f\n\n", totalCostAll))

	for _, summary := range summaries {
		response.WriteString(fmt.Sprintf("Task ID: %s\n", summary.TaskID))
		response.WriteString(fmt.Sprintf("  Total Cost: $%.6f\n", summary.TotalCost))
		response.WriteString(fmt.Sprintf("  Messages: %d\n", summary.MessageCount))
		response.WriteString(fmt.Sprintf("  CSV File: %s\n", summary.CSVPath))
		response.WriteString(fmt.Sprintf("  Last Updated: %s\n\n", summary.LastUpdated))
	}

	return &MCPResponse{
		Content: []MCPContent{
			{
				Type: "text",
				Text: response.String(),
			},
		},
	}, nil
}

// HandleListTrackedTasks returns a list of all tracked tasks
func HandleListTrackedTasks(params map[string]interface{}) (*MCPResponse, error) {
	// Find all CSV files in logs directory
	pattern := filepath.Join(LogsDirectory, "task_*_costs.csv")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find CSV files: %v", err)
	}

	if len(matches) == 0 {
		return &MCPResponse{
			Content: []MCPContent{
				{
					Type: "text",
					Text: "No tracked tasks found.",
				},
			},
		}, nil
	}

	var taskIDs []string
	for _, csvPath := range matches {
		// Extract task ID from filename
		filename := filepath.Base(csvPath)
		parts := strings.Split(filename, "_")
		if len(parts) >= 2 {
			taskIDs = append(taskIDs, parts[1])
		}
	}

	// Sort task IDs (most recent first)
	sort.Slice(taskIDs, func(i, j int) bool {
		return taskIDs[i] > taskIDs[j]
	})

	response := fmt.Sprintf("Tracked Tasks (%d total):\n", len(taskIDs))
	for i, taskID := range taskIDs {
		response += fmt.Sprintf("%d. %s\n", i+1, taskID)
	}

	return &MCPResponse{
		Content: []MCPContent{
			{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// HandleGetCurrentTaskCosts returns costs for the current/most recent task
func HandleGetCurrentTaskCosts(params map[string]interface{}) (*MCPResponse, error) {
	// Get current task ID
	currentTaskID, err := GetCurrentTaskID()
	if err != nil {
		return nil, fmt.Errorf("failed to get current task ID: %v", err)
	}

	// Look for CSV file for this task
	pattern := filepath.Join(LogsDirectory, fmt.Sprintf("task_%s_*_costs.csv", currentTaskID))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find CSV file: %v", err)
	}

	if len(matches) == 0 {
		return &MCPResponse{
			Content: []MCPContent{
				{
					Type: "text",
					Text: fmt.Sprintf("No cost data found for current task %s. Run manual_cost_track to generate cost data.", currentTaskID),
				},
			},
		}, nil
	}

	// Use the most recent CSV file if multiple exist
	csvPath := matches[len(matches)-1]

	summary, err := analyzeCsvFile(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze CSV file: %v", err)
	}

	response := fmt.Sprintf("Current Task Costs (Task ID: %s)\n", summary.TaskID)
	response += fmt.Sprintf("Total Cost: $%.6f\n", summary.TotalCost)
	response += fmt.Sprintf("Total Messages: %d\n", summary.MessageCount)
	response += fmt.Sprintf("CSV File: %s\n", summary.CSVPath)
	response += fmt.Sprintf("Last Updated: %s\n", summary.LastUpdated)

	return &MCPResponse{
		Content: []MCPContent{
			{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// analyzeCsvFile analyzes a CSV file and returns a TaskSummary
func analyzeCsvFile(csvPath string) (TaskSummary, error) {
	// Extract task ID from filename
	filename := filepath.Base(csvPath)
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return TaskSummary{}, fmt.Errorf("invalid CSV filename format")
	}
	taskID := parts[1]

	// Get file info for last updated time
	fileInfo, err := os.Stat(csvPath)
	if err != nil {
		return TaskSummary{}, err
	}

	// Read the CSV file to get cost and message count
	content, err := os.ReadFile(csvPath)
	if err != nil {
		return TaskSummary{}, err
	}

	lines := strings.Split(string(content), "\n")
	messageCount := len(lines) - 2 // Subtract header and empty last line

	// Find the last non-empty line to get total cost
	var totalCost float64
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Parse the line to extract total cost (7th column, index 6)
		fields := strings.Split(line, ",")
		if len(fields) > 6 {
			fmt.Sscanf(fields[6], "%f", &totalCost)
			break
		}
	}

	return TaskSummary{
		TaskID:       taskID,
		TotalCost:    totalCost,
		MessageCount: messageCount,
		CSVPath:      csvPath,
		LastUpdated:  fileInfo.ModTime().Format("2006-01-02 15:04:05"),
	}, nil
}

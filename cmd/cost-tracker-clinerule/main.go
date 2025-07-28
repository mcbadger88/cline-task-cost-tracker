package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/internal/ui-log-parser"
)

// extractWorkingDirectoryFromTask extracts the working directory from ui_messages.json
func extractWorkingDirectoryFromTask(taskFilePath string) string {
	log.Printf("DEBUG: Extracting working directory from: %s", taskFilePath)

	file, err := os.Open(taskFilePath)
	if err != nil {
		log.Printf("DEBUG: Error opening task file: %v", err)
		return ""
	}
	defer file.Close()

	// Read the file content
	content := make([]byte, 10240) // Read first 10KB to find the working directory
	n, err := file.Read(content)
	if err != nil && n == 0 {
		log.Printf("DEBUG: Error reading task file: %v", err)
		return ""
	}

	contentStr := string(content[:n])
	log.Printf("DEBUG: Read %d bytes from task file", n)

	// Look for the pattern "# Current Working Directory (/path/to/directory)"
	workingDirPattern := "# Current Working Directory ("
	startIdx := strings.Index(contentStr, workingDirPattern)
	if startIdx == -1 {
		log.Printf("DEBUG: Working directory pattern not found")
		return ""
	}

	// Find the start of the path
	pathStart := startIdx + len(workingDirPattern)

	// Find the end of the path (closing parenthesis)
	pathEnd := strings.Index(contentStr[pathStart:], ")")
	if pathEnd == -1 {
		log.Printf("DEBUG: Could not find end of working directory path")
		return ""
	}

	workingDir := contentStr[pathStart : pathStart+pathEnd]
	log.Printf("DEBUG: Extracted working directory: %s", workingDir)
	return workingDir
}

// detectRepositoryRoot attempts to find the repository root where Cline is working
func detectRepositoryRoot(taskFilePath string) (string, error) {
	log.Printf("DEBUG: Starting repository detection for task file: %s", taskFilePath)

	// Extract working directory from the ui_messages.json file
	if workingDir := extractWorkingDirectoryFromTask(taskFilePath); workingDir != "" {
		log.Printf("DEBUG: Extracted working directory from task: %s", workingDir)
		log.Printf("DEBUG: Using extracted working directory directly as repository root: %s", workingDir)
		return workingDir, nil
	}

	log.Printf("DEBUG: Could not extract working directory from task context")
	return "", fmt.Errorf("could not extract working directory from task context")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <path_to_ui_messages.json>")
	}

	inputPath := os.Args[1]

	// Try to detect the repository root where Cline is working
	repoRoot, err := detectRepositoryRoot(inputPath)
	if err != nil {
		log.Printf("Warning: Could not detect repository root (%v), falling back to current directory", err)
		// Fallback to current working directory
		repoRoot, err = os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
	}

	// Create the target log directory: {repoRoot}/ui-log-parser
	logBasePath := filepath.Join(repoRoot, "ui-log-parser")

	log.Printf("DEBUG: Detected repository root: %s", repoRoot)
	log.Printf("DEBUG: Using ProcessUILogToCSVAutoAt with basePath: %s", logBasePath)

	// Use the new function that creates logs directory in the detected repository
	if err := uilogparser.ProcessUILogToCSVAutoAt(inputPath, logBasePath); err != nil {
		log.Fatalf("Error processing UI log: %v", err)
	}

	log.Printf("CSV saved to: %s/logs/", logBasePath)
}

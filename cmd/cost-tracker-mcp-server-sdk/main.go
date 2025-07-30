package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/internal/ui-log-parser"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Configuration constants
const (
	UIMessagesFile = "ui_messages.json"
	DebounceDelay  = 1 // seconds
)

// getGitHash returns the current git commit hash
func getGitHash() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// GetClineTasksPath returns the path to Cline tasks directory
func GetClineTasksPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting home directory: %v", err)
		return ""
	}
	return filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks")
}

// GetCurrentTaskID returns the most recent task ID
func GetCurrentTaskID() (string, error) {
	tasksPath := GetClineTasksPath()
	entries, err := os.ReadDir(tasksPath)
	if err != nil {
		return "", err
	}

	var latestTask string
	var latestTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestTask = entry.Name()
			}
		}
	}

	if latestTask == "" {
		return "", fmt.Errorf("no tasks found")
	}

	return latestTask, nil
}

// FileWatcher handles monitoring of Cline task files
type FileWatcher struct {
	watcher       *fsnotify.Watcher
	debounceTimer map[string]*time.Timer
	stopChan      chan bool
}

// NewFileWatcher creates a new file watcher instance
func NewFileWatcher() (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	return &FileWatcher{
		watcher:       watcher,
		debounceTimer: make(map[string]*time.Timer),
		stopChan:      make(chan bool),
	}, nil
}

// Start begins monitoring the Cline tasks directory
func (fw *FileWatcher) Start() error {
	// Add the Cline tasks directory to the watcher
	err := fw.watcher.Add(GetClineTasksPath())
	if err != nil {
		return fmt.Errorf("failed to watch Cline tasks directory: %v", err)
	}

	// Also watch existing task subdirectories
	err = fw.watchExistingTasks()
	if err != nil {
		log.Printf("Warning: failed to watch some existing tasks: %v", err)
	}

	log.Printf("Started watching Cline tasks directory: %s", GetClineTasksPath())

	go fw.watchLoop()
	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() {
	fw.stopChan <- true
	fw.watcher.Close()
}

// watchExistingTasks adds existing task directories to the watcher
func (fw *FileWatcher) watchExistingTasks() error {
	pattern := filepath.Join(GetClineTasksPath(), "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, match := range matches {
		err := fw.watcher.Add(match)
		if err != nil {
			log.Printf("Warning: failed to watch task directory %s: %v", match, err)
		}
	}

	return nil
}

// watchLoop is the main event loop for file watching
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)

		case <-fw.stopChan:
			return
		}
	}
}

// handleEvent processes file system events
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// Check if this is a ui_messages.json file
	if !strings.HasSuffix(event.Name, UIMessagesFile) {
		// Check if it's a new task directory being created
		if event.Op&fsnotify.Create == fsnotify.Create {
			// Add new task directory to watcher
			err := fw.watcher.Add(event.Name)
			if err != nil {
				log.Printf("Failed to watch new task directory %s: %v", event.Name, err)
			} else {
				log.Printf("Now watching new task directory: %s", event.Name)
			}
		}
		return
	}

	// Only process write events for ui_messages.json
	if event.Op&fsnotify.Write == fsnotify.Write {
		fw.debounceProcessing(event.Name)
	}
}

// debounceProcessing implements debouncing to avoid processing rapid file changes
func (fw *FileWatcher) debounceProcessing(filePath string) {
	// Cancel existing timer for this file
	if timer, exists := fw.debounceTimer[filePath]; exists {
		timer.Stop()
	}

	// Create new timer
	fw.debounceTimer[filePath] = time.AfterFunc(DebounceDelay*time.Second, func() {
		fw.processFile(filePath)
		delete(fw.debounceTimer, filePath)
	})
}

// processFile processes a ui_messages.json file and generates CSV
func (fw *FileWatcher) processFile(filePath string) {
	log.Printf("Processing file change: %s", filePath)

	// Try to detect the repository root where Cline is working
	repoRoot, err := detectRepositoryRoot(filePath)
	if err != nil {
		log.Printf("Warning: Could not detect repository root (%v), falling back to current directory", err)
		// Fallback to current working directory
		repoRoot, err = os.Getwd()
		if err != nil {
			log.Printf("Error getting current working directory: %v", err)
			return
		}
	}

	// Create the target log directory: {repoRoot}/ui-log-parser
	logBasePath := filepath.Join(repoRoot, "ui-log-parser")

	// Use the existing function to process and generate CSV
	err = uilogparser.ProcessUILogToCSVAutoAt(filePath, logBasePath)
	if err != nil {
		log.Printf("Error processing file %s: %v", filePath, err)
		return
	}

	log.Printf("Successfully processed and generated CSV for: %s", filePath)
	log.Printf("CSV saved to: %s/logs/", logBasePath)
}

// detectRepositoryRoot attempts to find the repository root where Cline is working
func detectRepositoryRoot(taskFilePath string) (string, error) {
	// Extract working directory from the ui_messages.json file
	if workingDir := extractWorkingDirectoryFromTask(taskFilePath); workingDir != "" {
		return workingDir, nil
	}

	return "", fmt.Errorf("could not extract working directory from task context")
}

// extractWorkingDirectoryFromTask extracts the working directory from ui_messages.json
func extractWorkingDirectoryFromTask(taskFilePath string) string {
	file, err := os.Open(taskFilePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Read the file content
	content := make([]byte, 10240) // Read first 10KB to find the working directory
	n, err := file.Read(content)
	if err != nil && n == 0 {
		return ""
	}

	contentStr := string(content[:n])

	// Look for the pattern "# Current Working Directory (/path/to/directory)"
	workingDirPattern := "# Current Working Directory ("
	startIdx := strings.Index(contentStr, workingDirPattern)
	if startIdx == -1 {
		return ""
	}

	// Find the start of the path
	pathStart := startIdx + len(workingDirPattern)

	// Find the end of the path (closing parenthesis)
	pathEnd := strings.Index(contentStr[pathStart:], ")")
	if pathEnd == -1 {
		return ""
	}

	workingDir := contentStr[pathStart : pathStart+pathEnd]
	return workingDir
}

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create implementation
	impl := &mcp.Implementation{}

	// Create server
	server := mcp.NewServer(impl, &mcp.ServerOptions{})

	// Add generate_csv tool
	tool := &mcp.Tool{
		Name:        "generate_csv",
		Description: "Generate CSV file with cost tracking data from ui_messages.json file",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"file_path": {
					Type:        "string",
					Description: "Optional path to ui_messages.json file. If not provided, uses current task.",
				},
			},
		},
	}

	server.AddTool(tool, handleGenerateCSV)

	// Start file watcher in background
	fileWatcher, err := NewFileWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}

	if err := fileWatcher.Start(); err != nil {
		log.Fatalf("Failed to start file watcher: %v", err)
	}
	defer fileWatcher.Stop()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Create stdio transport and run server
	transport := mcp.NewStdioTransport()
	gitHash := getGitHash()
	log.Printf("Starting Cost Tracker MCP Server with SDK (version: %s)...", gitHash)

	if err := server.Run(ctx, transport); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server shutdown complete")
}

// handleGenerateCSV handles the generate_csv tool call
func handleGenerateCSV(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[any], error) {
	// Extract arguments - params.Arguments is map[string]any for ToolHandler
	arguments := make(map[string]interface{})
	if params.Arguments != nil {
		arguments = params.Arguments
	}

	// Call the existing handler logic
	result, err := HandleGenerateCSV(arguments)
	if err != nil {
		return nil, err
	}

	// Return result
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: result,
			},
		},
	}, nil
}

// HandleGenerateCSV processes cost tracking requests and generates CSV files
func HandleGenerateCSV(params map[string]interface{}) (string, error) {
	// Get the file path parameter
	filePath, ok := params["file_path"].(string)
	if !ok || filePath == "" {
		// If no file path provided, use the current task
		currentTaskID, err := GetCurrentTaskID()
		if err != nil {
			return "", err
		}
		filePath = GetClineTasksPath() + "/" + currentTaskID + "/" + UIMessagesFile
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", err
	}

	// Try to detect the repository root where Cline is working
	repoRoot, err := detectRepositoryRoot(filePath)
	if err != nil {
		// Fallback to current working directory
		repoRoot, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	// Create the target log directory: {repoRoot}/ui-log-parser
	logBasePath := repoRoot + "/ui-log-parser"

	// Use the existing function to process and generate CSV
	err = uilogparser.ProcessUILogToCSVAutoAt(filePath, logBasePath)
	if err != nil {
		return "", err
	}

	// Extract task ID for response
	taskID := uilogparser.ExtractTaskID(filePath)

	return "Successfully processed task " + taskID + " and generated CSV file at " + logBasePath + "/logs/", nil
}

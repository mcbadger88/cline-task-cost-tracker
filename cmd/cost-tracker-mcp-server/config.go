package main

import (
	"os"
	"path/filepath"
)

const (
	// ClineTasksPath is the base path where Cline stores task data
	ClineTasksPath = "/Users/emma/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks"

	// UIMessagesFile is the filename for UI messages in each task directory
	UIMessagesFile = "ui_messages.json"

	// DebounceDelay is the delay in seconds before processing file changes
	DebounceDelay = 1

	// LogsDirectory is where CSV files are saved
	LogsDirectory = "logs"
)

// GetClineTasksPath returns the path to Cline tasks directory
func GetClineTasksPath() string {
	return ClineTasksPath
}

// GetUIMessagesPattern returns the glob pattern for UI messages files
func GetUIMessagesPattern() string {
	return filepath.Join(ClineTasksPath, "*", UIMessagesFile)
}

// EnsureLogsDirectory creates the logs directory if it doesn't exist
func EnsureLogsDirectory() error {
	return os.MkdirAll(LogsDirectory, 0755)
}

// GetCurrentTaskID returns the most recent task ID
func GetCurrentTaskID() (string, error) {
	entries, err := os.ReadDir(ClineTasksPath)
	if err != nil {
		return "", err
	}

	var latestTask string
	var latestTime int64

	for _, entry := range entries {
		if entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Unix() > latestTime {
				latestTime = info.ModTime().Unix()
				latestTask = entry.Name()
			}
		}
	}

	return latestTask, nil
}

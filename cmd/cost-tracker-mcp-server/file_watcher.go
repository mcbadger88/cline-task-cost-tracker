package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser"
)

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
				log.Printf("DEBUG: File watcher events channel closed, exiting watchLoop")
				return
			}
			fw.handleEvent(event)

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				log.Printf("DEBUG: File watcher errors channel closed, exiting watchLoop")
				return
			}
			log.Printf("File watcher error: %v", err)

		case <-fw.stopChan:
			log.Printf("DEBUG: Received stop signal, exiting watchLoop")
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

	// Get the current working directory (project directory)
	projectDir, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v", err)
		return
	}

	log.Printf("DEBUG: Current working directory: %s", projectDir)
	log.Printf("DEBUG: Using ProcessUILogToCSVAutoAt with basePath: %s", projectDir)

	// Use the new function that creates logs directory in the project directory
	err = uilogparser.ProcessUILogToCSVAutoAt(filePath, projectDir)
	if err != nil {
		log.Printf("Error processing file %s: %v", filePath, err)
		return
	}

	log.Printf("Successfully processed and generated CSV for: %s", filePath)
}

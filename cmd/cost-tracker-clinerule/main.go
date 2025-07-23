package main

import (
	"fmt"
	"log"
	"os"

	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <path_to_ui_messages.json>")
	}

	inputPath := os.Args[1]

	// Parse UI messages
	messages, err := uilogparser.ParseUIMessages(inputPath)
	if err != nil {
		log.Fatalf("Error parsing UI messages: %v", err)
	}

	// Process messages into cost records
	records := uilogparser.ProcessMessages(messages)

	// Generate output filename using task start time (first message timestamp)
	taskID := uilogparser.ExtractTaskID(inputPath)
	outputPath := uilogparser.GenerateOutputPath(taskID, messages[0].Timestamp)

	// Create logs directory if it doesn't exist
	if err := uilogparser.EnsureLogsDirectory(); err != nil {
		log.Fatalf("Error creating logs directory: %v", err)
	}

	// Write CSV file
	if err := uilogparser.WriteCSV(outputPath, records); err != nil {
		log.Fatalf("Error writing CSV: %v", err)
	}

	fmt.Printf("Cost tracker CSV generated: %s\n", outputPath)
	fmt.Printf("Total records: %d\n", len(records))
}

package main

import (
	"log"
	"os"

	uilogparser "github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <path_to_ui_messages.json>")
	}

	inputPath := os.Args[1]

	// Process UI log to CSV in a single call
	if err := uilogparser.ProcessUILogToCSVAuto(inputPath); err != nil {
		log.Fatalf("Error processing UI log: %v", err)
	}
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

const VERSION = "v2.2.0-simplified"

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("Cost Tracker MCP Server %s starting...", VERSION)

	// Create MCP server
	server, err := NewMCPServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Wait for either an error or a shutdown signal
	select {
	case err := <-errChan:
		log.Printf("Received errChan %v, shutting down...", err)
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
		server.Stop()
	}

	log.Println("Cost Tracker MCP Server stopped")
}

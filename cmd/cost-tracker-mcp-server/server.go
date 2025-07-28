package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// MCPServer represents the MCP server instance
type MCPServer struct {
	fileWatcher *FileWatcher
	reader      *bufio.Reader
	writer      io.Writer
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() (*MCPServer, error) {
	fileWatcher, err := NewFileWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	return &MCPServer{
		fileWatcher: fileWatcher,
		reader:      bufio.NewReader(os.Stdin),
		writer:      os.Stdout,
	}, nil
}

// Start starts the MCP server
func (s *MCPServer) Start() error {
	log.Println("Starting Cost Tracker MCP Server...")

	// Start file watcher (runs in background)
	if err := s.fileWatcher.Start(); err != nil {
		return fmt.Errorf("failed to start file watcher: %v", err)
	}

	// Send initialization response
	if err := s.sendInitialization(); err != nil {
		return fmt.Errorf("failed to send initialization: %v", err)
	}

	log.Println("MCP Server started successfully - File watcher active and MCP protocol ready")

	// Start message loop (this will block, but file watcher continues in background)
	return s.messageLoop()
}

// Stop stops the MCP server
func (s *MCPServer) Stop() {
	log.Println("Stopping Cost Tracker MCP Server...")
	s.fileWatcher.Stop()
}

// sendInitialization sends the MCP initialization response
func (s *MCPServer) sendInitialization() error {
	initResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "cost-tracker-mcp-server",
				"version": "1.0.0",
			},
		},
	}

	return s.sendResponse(initResponse)
}

// messageLoop handles incoming MCP messages
func (s *MCPServer) messageLoop() error {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("MCP client disconnected (EOF), but server continues running...")
				// Don't exit - keep the server running for file watching
				// The server should only stop on explicit signals
				select {} // Block forever until signal
			}
			return fmt.Errorf("failed to read message: %v", err)
		}

		if err := s.handleMessage(line); err != nil {
			log.Printf("Error handling message: %v", err)
		}
	}

	return nil
}

// handleMessage processes an incoming MCP message
func (s *MCPServer) handleMessage(message string) error {
	var request map[string]interface{}
	if err := json.Unmarshal([]byte(message), &request); err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}

	method, ok := request["method"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid method")
	}

	id := request["id"]

	switch method {
	case "tools/list":
		return s.handleToolsList(id)
	case "tools/call":
		return s.handleToolsCall(request)
	case "initialize":
		return s.sendInitialization()
	default:
		return s.sendError(id, fmt.Sprintf("unknown method: %s", method))
	}
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsList(id interface{}) error {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "generate_csv",
					"description": "Generate CSV file with cost tracking data from ui_messages.json file",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"file_path": map[string]interface{}{
								"type":        "string",
								"description": "Path to ui_messages.json file (optional, defaults to current task)",
							},
						},
					},
				},
			},
		},
	}

	return s.sendResponse(response)
}

// handleToolsCall handles tool execution requests
func (s *MCPServer) handleToolsCall(request map[string]interface{}) error {
	id := request["id"]

	params, ok := request["params"].(map[string]interface{})
	if !ok {
		return s.sendError(id, "missing or invalid params")
	}

	name, ok := params["name"].(string)
	if !ok {
		return s.sendError(id, "missing or invalid tool name")
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	var response *MCPResponse
	var err error

	switch name {
	case "generate_csv":
		response, err = HandleGenerateCSV(arguments)
	default:
		return s.sendError(id, fmt.Sprintf("unknown tool: %s", name))
	}

	if err != nil {
		return s.sendError(id, err.Error())
	}

	result := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  response,
	}

	return s.sendResponse(result)
}

// sendResponse sends a response to the client
func (s *MCPServer) sendResponse(response interface{}) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %v", err)
	}

	_, err = fmt.Fprintf(s.writer, "%s\n", data)
	return err
}

// sendError sends an error response to the client
func (s *MCPServer) sendError(id interface{}, message string) error {
	errorResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    -1,
			"message": message,
		},
	}

	return s.sendResponse(errorResponse)
}

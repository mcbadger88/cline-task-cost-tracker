# Cost Tracker MCP Server

An MCP (Model Context Protocol) server that automatically monitors Cline tasks and generates CSV cost tracking files without manual intervention.

## Features

- **Automatic File Monitoring**: Watches Cline task directories for `ui_messages.json` changes
- **Debounced Processing**: 1-second delay to avoid processing rapid file changes
- **CSV Generation**: Uses the proven `ui-log-parser` library to generate 15-column CSV files
- **MCP Tools**: Provides 4 tools for cost tracking and analysis
- **Graceful Error Handling**: Robust error handling for file access issues

## Installation

### Prerequisites

- Go 1.24.5 or later
- Access to Cline task files at: `/Users/emma/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/`

### Build the Server

```bash
cd cmd/cost-tracker-mcp-server
go mod tidy
go build -o cost-tracker-mcp-server
```

### Run the Server

```bash
./cost-tracker-mcp-server
```

## MCP Tools

The server provides 4 MCP tools:

### 1. `manual_cost_track`
Manually trigger cost tracking for a specific ui_messages.json file.

**Parameters:**
- `file_path` (optional): Path to ui_messages.json file. Defaults to current task if not provided.

**Example:**
```json
{
  "name": "manual_cost_track",
  "arguments": {
    "file_path": "/Users/emma/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/1753250474282/ui_messages.json"
  }
}
```

### 2. `get_cost_summary`
Get a summary of costs across all tracked tasks.

**Parameters:** None

**Returns:** Summary with total costs, task counts, and individual task details.

### 3. `list_tracked_tasks`
List all tasks that have been tracked for costs.

**Parameters:** None

**Returns:** List of task IDs that have generated CSV files.

### 4. `get_current_task_costs`
Get cost information for the current/most recent task.

**Parameters:** None

**Returns:** Detailed cost information for the most recent task.

## Cline MCP Configuration

Add this configuration to your Cline MCP settings:

```json
{
  "mcpServers": {
    "cost-tracker": {
      "command": "/path/to/cost-tracker-mcp-server",
      "args": [],
      "env": {}
    }
  }
}
```

Replace `/path/to/cost-tracker-mcp-server` with the actual path to your built executable.

## CSV Output Format

The server generates CSV files with 15 columns:

1. **Request Summary** - Categorized request types (Task Request, User Input, API Request)
2. **Ask/Say** - Message type and category  
3. **Cost** - Individual API request cost
4. **Text** - Full message content
5. **Timestamp** - Formatted timestamp
6. **Context tokens used** - Token usage information
7. **Total cost** - Cumulative cost
8. **Cline_Action** - Extracted Cline actions
9. **Tool_Used** - Tools invoked during the task
10. **Has_Images** - Whether images were involved
11. **Phase** - Task phase (Initial, Processing, Completion)
12. **Context_Percentage** - Context window usage percentage
13. **Search_Term_In_Transcript** - Unique search identifiers
14. **Cost_Notes** - Additional cost-related notes
15. **Time_Approx** - Approximate time (HH:MM format)

## File Locations

- **CSV Output**: `logs/task_{task_id}_{timestamp}_costs.csv`
- **Monitored Path**: `/Users/emma/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/*/ui_messages.json`

## Example Output

```
logs/task_1753250474282_2025-07-23_16-01-14_costs.csv
```

## Validation

Compare the MCP server's CSV output with the existing CLI tool output to ensure identical results:

```bash
# CLI tool
cd ../cost-tracker-clinerule
go run main.go "/path/to/ui_messages.json"

# MCP server (automatic via file watching or manual via tool)
# Files should be identical
```

## Troubleshooting

### File Permission Issues
Ensure the server has read access to the Cline tasks directory and write access to the logs directory.

### File Watching Not Working
Check that the Cline tasks directory exists and contains task subdirectories.

### CSV Generation Errors
Verify that ui_messages.json files are valid JSON and contain the expected message structure.

## Dependencies

- `github.com/fsnotify/fsnotify` - File system watching
- `github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser` - CSV generation logic

## Architecture

```
cmd/cost-tracker-mcp-server/
├── main.go           # Entry point
├── server.go         # MCP protocol implementation
├── handlers.go       # MCP tools implementation
├── file_watcher.go   # File monitoring with debouncing
├── config.go         # Configuration constants
└── README.md         # This file
```

The server uses a modular architecture with separate concerns for MCP protocol handling, file watching, and cost processing.
